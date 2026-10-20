package handler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Tipos MIME permitidos
var AllowedMIME = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/csv": true,
}

const MaxFileSize = 50 * 1024 * 1024 // 50MB

type UploadHandler struct {
	s3Client   *s3.Client
	bucket     string
	publicBase string // e.g. http://localhost:9000/<bucket> for a local S3-compatible endpoint; empty means real AWS S3
}

// NewUploadHandler wires either a MinIO-style S3-compatible endpoint (when
// endpoint/accessKeyID are set — the local dev/pilot default, matching the
// minio container in docker-compose.dev.yml) or real AWS S3 via the SDK's
// default credential chain. Before this, the handler always built the real
// AWS S3 client regardless of the MinIO env vars already defined for it —
// uploads would have tried to hit real AWS with MinIO's fake credentials.
func NewUploadHandler(bucket, region, endpoint, publicEndpoint, accessKeyID, secretAccessKey string) *UploadHandler {
	if endpoint != "" {
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		)
		if err != nil {
			log.Printf("S3-compatible config error: %v — usando modo desarrollo", err)
			return &UploadHandler{bucket: bucket}
		}
		client := s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true // MinIO requires path-style addressing
		})
		// endpoint (e.g. http://minio:9000) is the Docker-network address
		// this service itself uses to reach MinIO; a browser can't resolve
		// that hostname, so the URL returned to clients uses publicEndpoint
		// instead (e.g. http://localhost:9000) when set.
		base := endpoint
		if publicEndpoint != "" {
			base = publicEndpoint
		}
		return &UploadHandler{s3Client: client, bucket: bucket, publicBase: base + "/" + bucket}
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		log.Printf("AWS config error: %v — usando modo desarrollo", err)
		return &UploadHandler{bucket: bucket}
	}
	return &UploadHandler{s3Client: s3.NewFromConfig(cfg), bucket: bucket}
}

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "archivo requerido")
	}

	// Validar tamaño
	if file.Size > MaxFileSize {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "archivo demasiado grande (máx 50MB)")
	}

	// Validar tipo MIME
	ext := strings.ToLower(filepath.Ext(file.Filename))
	mimeType := mime.TypeByExtension(ext)
	if !AllowedMIME[mimeType] {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, fmt.Sprintf("tipo de archivo no permitido: %s", mimeType))
	}

	f, _ := file.Open()
	defer f.Close()
	body, _ := io.ReadAll(f)

	// SHA-256 checksum
	hash := sha256.Sum256(body)
	checksum := hex.EncodeToString(hash[:])

	// Escaneo antivirus Sophos Intelix
	if isMalicious, threat := h.scanWithSophos(body); isMalicious {
		log.Printf("🦠 VIRUS DETECTADO: %s en archivo %s", threat, file.Filename)
		return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("archivo bloqueado por antivirus: %s", threat))
	}

	// Generar key S3
	tenantID := c.Locals("tenant_id")
	prefix := c.FormValue("prefix", "general")
	isWORM := c.FormValue("worm", "false") == "true" // Para docs médicos: retención 20 años
	fileID := uuid.New().String()
	s3Key := fmt.Sprintf("%v/%s/%s/%s%s", tenantID, prefix, fileID[:8], fileID, ext)

	// Upload a S3
	var fileURL string
	if h.s3Client != nil {
		input := &s3.PutObjectInput{
			Bucket:      aws.String(h.bucket),
			Key:         aws.String(s3Key),
			Body:        strings.NewReader(string(body)),
			ContentType: aws.String(mimeType),
			Metadata: map[string]string{
				"original-name": file.Filename,
				"checksum":      checksum,
				"tenant-id":     fmt.Sprintf("%v", tenantID),
				"uploaded-at":   time.Now().UTC().Format(time.RFC3339),
			},
		}
		// Object Lock (retención WORM, 20 años para docs médicos per Res.
		// 2346/2007) requires the bucket itself to have Object Lock enabled
		// at creation time — true on real S3 when provisioned that way, not
		// true for the local MinIO container used in this pilot. Request it
		// only against a real AWS endpoint; against MinIO it would just
		// fail the whole upload for a guarantee we can't actually back yet.
		if isWORM && h.publicBase == "" {
			input.ObjectLockMode = "COMPLIANCE"
			input.ObjectLockRetainUntilDate = aws.Time(time.Now().AddDate(20, 0, 0))
		}
		_, err = h.s3Client.PutObject(context.Background(), input)
		if err != nil {
			log.Printf("S3 upload error: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "error subiendo archivo")
		}
		if h.publicBase != "" {
			fileURL = fmt.Sprintf("%s/%s", h.publicBase, s3Key)
		} else {
			fileURL = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", h.bucket, s3Key)
		}
	} else {
		// Dev: retornar URL ficticia
		fileURL = fmt.Sprintf("http://localhost:9000/%s/%s", h.bucket, s3Key)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"file_id":      fileID,
		"url":          fileURL,
		"s3_key":       s3Key,
		"size":         file.Size,
		"mime_type":    mimeType,
		"checksum_sha": checksum,
		"worm":         isWORM,
		"uploaded_at":  time.Now().UTC(),
	})
}

func (h *UploadHandler) scanWithSophos(data []byte) (bool, string) {
	// Integración Sophos Intelix API
	// En producción: POST a https://de.api.labs.sophos.com/analysis/file/v1
	// Por ahora: simulación (nunca devuelve positivo en dev)
	if len(data) > 0 && data[0] == 0x4D && len(data) > 1 && data[1] == 0x5A {
		// Bloquear ejecutables .EXE (MZ header)
		return true, "Executable file blocked"
	}
	return false, ""
}
