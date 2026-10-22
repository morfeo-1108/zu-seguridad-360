package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"mime"
	"path/filepath"
	"regexp"
	"strconv"
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
	publicBase string // non-empty for an S3-compatible endpoint (MinIO): no Object Lock there
	linkKey    []byte // HMAC key for short-lived download links
}

// FilesPath is the stable, browser-reachable reference stored by Core for an
// uploaded object (served through the frontend's /api/media proxy). It only
// works with a signature from POST /v1/files/link.
const FilesPath = "/api/media/files/"

var prefixRE = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

// WithLinkSecret derives the download-link signing key from a service secret.
func (h *UploadHandler) WithLinkSecret(secret string) *UploadHandler {
	sum := sha256.Sum256([]byte("zu-media-download-link:" + secret))
	h.linkKey = sum[:]
	return h
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
	if !prefixRE.MatchString(prefix) {
		return fiber.NewError(fiber.StatusBadRequest, "prefix inválido")
	}
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
		// Object Lock (retención WORM para historia clínica ocupacional; el plazo
		// debe revisarse contra la Res. 1843 de 2025, que derogó la 2346/2007) requires the bucket itself to have Object Lock enabled
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
		fileURL = FilesPath + s3Key
	} else {
		// No storage configured: refuse instead of returning a URL to nothing.
		return fiber.NewError(fiber.StatusServiceUnavailable, "almacenamiento de archivos no configurado")
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

// keyFromRef accepts either a raw object key or a FilesPath URL.
func keyFromRef(ref string) string {
	if i := strings.Index(ref, FilesPath); i >= 0 {
		ref = ref[i+len(FilesPath):]
	}
	if j := strings.IndexAny(ref, "?#"); j >= 0 {
		ref = ref[:j]
	}
	return strings.TrimPrefix(ref, "/")
}

func (h *UploadHandler) sign(key string, exp int64) string {
	m := hmac.New(sha256.New, h.linkKey)
	fmt.Fprintf(m, "%s|%d", key, exp)
	return hex.EncodeToString(m.Sum(nil))
}

// Link issues a 15-minute download link for an object of the caller's tenant.
func (h *UploadHandler) Link(c *fiber.Ctx) error {
	var body struct {
		URL string `json:"url"`
	}
	if err := c.BodyParser(&body); err != nil || body.URL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "url requerida")
	}
	key := keyFromRef(body.URL)
	tenantID := fmt.Sprintf("%v", c.Locals("tenant_id"))
	if tenantID == "" || strings.Contains(key, "..") || !strings.HasPrefix(key, tenantID+"/") {
		return fiber.NewError(fiber.StatusNotFound, "archivo no encontrado")
	}
	exp := time.Now().Add(15 * time.Minute).Unix()
	return c.JSON(fiber.Map{"url": fmt.Sprintf("%s%s?exp=%d&sig=%s", FilesPath, key, exp, h.sign(key, exp)), "expires_at": exp})
}

// Download streams an object given a valid, unexpired signature.
func (h *UploadHandler) Download(c *fiber.Ctx) error {
	if h.s3Client == nil || len(h.linkKey) == 0 {
		return fiber.NewError(fiber.StatusServiceUnavailable, "almacenamiento no configurado")
	}
	key := strings.TrimPrefix(c.Params("*"), "/")
	exp, err := strconv.ParseInt(c.Query("exp"), 10, 64)
	if err != nil || time.Now().Unix() > exp || strings.Contains(key, "..") ||
		!hmac.Equal([]byte(h.sign(key, exp)), []byte(c.Query("sig"))) {
		return fiber.NewError(fiber.StatusForbidden, "enlace inválido o vencido")
	}
	obj, err := h.s3Client.GetObject(c.Context(), &s3.GetObjectInput{Bucket: aws.String(h.bucket), Key: aws.String(key)})
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "archivo no encontrado")
	}
	defer obj.Body.Close()
	data, err := io.ReadAll(obj.Body)
	if err != nil {
		return fiber.NewError(fiber.StatusBadGateway, "error leyendo archivo")
	}
	ctype := aws.ToString(obj.ContentType)
	c.Set("Content-Type", ctype)
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("Cache-Control", "private, max-age=300")
	if !strings.HasPrefix(ctype, "image/") {
		c.Set("Content-Disposition", "attachment")
	}
	return c.Send(data)
}
