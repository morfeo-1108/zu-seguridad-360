package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"google.golang.org/api/idtoken"
	"gorm.io/gorm"
)

var (
	ErrGoogleSignupDisabled      = errors.New("Google signup is disabled")
	ErrGoogleTokenInvalid        = errors.New("Google credential is invalid")
	ErrGoogleNonceInvalid        = errors.New("Google challenge is invalid or expired")
	ErrGoogleEmailNotVerified    = errors.New("Google email is not verified or authoritative")
	ErrGoogleAccountConflict     = errors.New("account or company identifier already exists")
	ErrGoogleRegistrationExpired = errors.New("Google registration has expired")
	ErrGoogleTermsNotAccepted    = errors.New("terms and privacy notice must be accepted")
	ErrGoogleRegistrationInvalid = errors.New("Google registration data is invalid")
)

const (
	googleChallengeTTL    = 5 * time.Minute
	googleRegistrationTTL = 10 * time.Minute
)

var (
	companyNITPattern  = regexp.MustCompile(`^[0-9][0-9.\-]{4,19}$`)
	companyCIUUPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,10}$`)
)

type googleIDTokenValidator interface {
	Validate(ctx context.Context, idToken string, audience string) (*idtoken.Payload, error)
}

type GoogleSignupProfile struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GoogleAuthResult struct {
	RegistrationRequired bool
	RegistrationToken    string
	Profile              GoogleSignupProfile
	AccessToken          string
	RefreshToken         string
}

type GoogleSignupPublicConfig struct {
	Enabled        bool   `json:"enabled"`
	SyntheticPilot bool   `json:"synthetic_pilot"`
	ClientID       string `json:"client_id,omitempty"`
	TermsURL       string `json:"terms_url,omitempty"`
	PrivacyURL     string `json:"privacy_url,omitempty"`
	TermsVersion   string `json:"terms_version,omitempty"`
	PrivacyVersion string `json:"privacy_version,omitempty"`
}

type GoogleInitialCompany struct {
	Name          string `json:"name"`
	NIT           string `json:"nit"`
	CIIU          string `json:"ciiu"`
	ARL           string `json:"arl"`
	LegalRep      string `json:"legal_rep"`
	LegalRepEmail string `json:"legal_rep_email"`
}

type GoogleSignupCompletion struct {
	RegistrationToken string
	// Company is optional: the professional can create the account first and
	// register companies later from the dashboard (POST /v1/companies).
	Company                   *GoogleInitialCompany
	TermsAccepted             bool
	PrivacyNoticeAcknowledged bool
}

type googleRegistrationIdentity struct {
	Subject   string `json:"sub"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func newGoogleIDTokenValidator(ctx context.Context) (*idtoken.Validator, error) {
	return idtoken.NewValidator(ctx)
}

func (s *AuthService) ensureGoogleSignupEnabled() error {
	if s == nil || s.cfg == nil || !s.cfg.GoogleSignupEnabled || s.cfg.SyntheticPilotEnabled {
		return ErrGoogleSignupDisabled
	}
	if strings.TrimSpace(s.cfg.GoogleClientID) == "" {
		return ErrGoogleSignupDisabled
	}
	return nil
}

// NewGoogleSignupChallenge creates an unpredictable, short-lived nonce. Google
// signs this value into the ID token; Core consumes it exactly once on callback.
func (s *AuthService) NewGoogleSignupChallenge(ctx context.Context) (string, error) {
	if err := s.ensureGoogleSignupEnabled(); err != nil {
		return "", err
	}
	nonce, err := randomOpaqueValue(32)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, googleRedisKey("challenge", nonce), "issued", googleChallengeTTL).Err(); err != nil {
		return "", fmt.Errorf("Google challenge storage unavailable")
	}
	return nonce, nil
}

// CheckGoogleSignupRateLimit uses Redis so the limit survives restarts and is
// shared by Core replicas. The raw client IP is never stored in Redis.
func (s *AuthService) CheckGoogleSignupRateLimit(ctx context.Context, operation, clientIP string, limit int64, window time.Duration) (bool, error) {
	if err := s.ensureGoogleSignupEnabled(); err != nil {
		return false, err
	}
	if operation == "" || clientIP == "" || limit < 1 || window < time.Second {
		return false, fmt.Errorf("invalid rate limit parameters")
	}
	identityHash := sha256.Sum256([]byte(clientIP))
	key := fmt.Sprintf("google:signup:rate:%s:%s", operation, hex.EncodeToString(identityHash[:]))
	script := `local n = redis.call('INCR', KEYS[1]); if n == 1 then redis.call('EXPIRE', KEYS[1], ARGV[1]); end; return n`
	count, err := s.redis.Eval(ctx, script, []string{key}, int64(window.Seconds())).Int64()
	if err != nil {
		return false, fmt.Errorf("signup rate limiter unavailable")
	}
	return count <= limit, nil
}

// BeginGoogleAuthentication logs in an already-linked identity, or returns a
// short-lived server-side registration ticket for a verified, first-time user.
func (s *AuthService) BeginGoogleAuthentication(ctx context.Context, rawIDToken, nonce string) (*GoogleAuthResult, error) {
	if err := s.ensureGoogleSignupEnabled(); err != nil {
		return nil, err
	}
	if s.googleValidator == nil || strings.TrimSpace(rawIDToken) == "" || len(rawIDToken) > 16384 || strings.TrimSpace(nonce) == "" || len(nonce) > 128 {
		return nil, ErrGoogleTokenInvalid
	}
	payload, err := s.googleValidator.Validate(ctx, rawIDToken, s.cfg.GoogleClientID)
	if err != nil || payload == nil {
		return nil, ErrGoogleTokenInvalid
	}
	if payload.Audience != s.cfg.GoogleClientID || (payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com") || payload.Expires <= time.Now().Unix() || payload.IssuedAt > time.Now().Add(time.Minute).Unix() {
		return nil, ErrGoogleTokenInvalid
	}
	claims := payload.Claims
	claimNonce, _ := claims["nonce"].(string)
	if claimNonce == "" || claimNonce != nonce {
		return nil, ErrGoogleNonceInvalid
	}
	challenge, err := s.redis.GetDel(ctx, googleRedisKey("challenge", nonce)).Result()
	if err != nil || challenge != "issued" {
		return nil, ErrGoogleNonceInvalid
	}

	subject := strings.TrimSpace(payload.Subject)
	email, _ := claims["email"].(string)
	email = strings.ToLower(strings.TrimSpace(email))
	parsedGoogleEmail, emailParseErr := mail.ParseAddress(email)
	verified, _ := claims["email_verified"].(bool)
	hostedDomain, _ := claims["hd"].(string)
	if subject == "" || len(subject) > 255 || email == "" || emailParseErr != nil || parsedGoogleEmail.Address != email || !verified || !isGoogleAuthoritativeEmail(email, hostedDomain) {
		return nil, ErrGoogleEmailNotVerified
	}
	firstName, _ := claims["given_name"].(string)
	lastName, _ := claims["family_name"].(string)
	if strings.TrimSpace(firstName) == "" {
		fullName, _ := claims["name"].(string)
		parts := strings.Fields(fullName)
		if len(parts) > 0 {
			firstName = parts[0]
		}
		if len(parts) > 1 && strings.TrimSpace(lastName) == "" {
			lastName = strings.Join(parts[1:], " ")
		}
	}
	firstName, lastName = strings.TrimSpace(firstName), strings.TrimSpace(lastName)
	if firstName == "" || len(firstName) > 100 || len(lastName) > 100 {
		return nil, ErrGoogleRegistrationInvalid
	}
	identity := googleRegistrationIdentity{Subject: subject, Email: email, FirstName: firstName, LastName: lastName}

	user, err := s.userRepo.FindByGoogleSubject(ctx, subject)
	if err == nil {
		if !user.IsActive {
			return nil, ErrAccountInactive
		}
		companyID := companyIDForUser(user)
		if companyID == "" {
			companies, listErr := s.userRepo.ListMemberCompanies(ctx, user.ID, user.TenantID)
			if listErr != nil {
				return nil, fmt.Errorf("company membership lookup unavailable")
			}
			if len(companies) == 0 {
				// Account created without a company yet: issue a tenant-only
				// session so the professional can register the first one.
				access, refresh, sessionErr := s.issueGoogleSession(ctx, user)
				if sessionErr != nil {
					return nil, sessionErr
				}
				return &GoogleAuthResult{AccessToken: access, RefreshToken: refresh}, nil
			}
			companyID = companies[0].ID.String()
		}
		companyUUID, parseErr := uuid.Parse(companyID)
		if parseErr != nil {
			return nil, ErrAccountInactive
		}
		member, membershipErr := s.userRepo.HasActiveCompanyMembership(ctx, user.ID, user.TenantID, companyUUID)
		if membershipErr != nil || !member {
			return nil, ErrAccountInactive
		}
		access, refresh, sessionErr := s.issueGoogleSession(ctx, copyUserWithCompany(user, companyUUID))
		if sessionErr != nil {
			return nil, sessionErr
		}
		return &GoogleAuthResult{AccessToken: access, RefreshToken: refresh}, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("Google account lookup unavailable")
	}

	// Never auto-link to a password account merely because the email matches.
	existing, err := s.userRepo.FindByEmailAnyStatus(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrGoogleAccountConflict
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("Google account lookup unavailable")
	}
	registrationToken, err := randomOpaqueValue(32)
	if err != nil {
		return nil, err
	}
	serialized, err := json.Marshal(identity)
	if err != nil {
		return nil, err
	}
	if err := s.redis.Set(ctx, googleRedisKey("registration", registrationToken), serialized, googleRegistrationTTL).Err(); err != nil {
		return nil, fmt.Errorf("Google registration storage unavailable")
	}
	return &GoogleAuthResult{RegistrationRequired: true, RegistrationToken: registrationToken, Profile: GoogleSignupProfile{
		Email: email, FirstName: firstName, LastName: lastName,
	}}, nil
}

// CompleteGoogleSignup consumes the one-time registration ticket and provisions
// the user and isolated tenant — plus, when one is given, the initial company
// and membership — in one DB transaction.
func (s *AuthService) CompleteGoogleSignup(ctx context.Context, input GoogleSignupCompletion) (string, string, error) {
	if err := s.ensureGoogleSignupEnabled(); err != nil {
		return "", "", err
	}
	if !input.TermsAccepted || !input.PrivacyNoticeAcknowledged {
		return "", "", ErrGoogleTermsNotAccepted
	}
	if strings.TrimSpace(s.cfg.GoogleSignupTermsVersion) == "" || strings.TrimSpace(s.cfg.GoogleSignupPrivacyVersion) == "" {
		return "", "", ErrGoogleSignupDisabled
	}
	var company *model.Company
	if input.Company != nil {
		validated, err := validateGoogleInitialCompany(*input.Company)
		if err != nil {
			return "", "", err
		}
		company = validated
	}
	if strings.TrimSpace(input.RegistrationToken) == "" || len(input.RegistrationToken) > 128 {
		return "", "", ErrGoogleRegistrationExpired
	}
	registrationKey := googleRedisKey("registration", input.RegistrationToken)
	encodedIdentity, err := s.redis.Get(ctx, registrationKey).Result()
	if err != nil {
		return "", "", ErrGoogleRegistrationExpired
	}
	var identity googleRegistrationIdentity
	if err := json.Unmarshal([]byte(encodedIdentity), &identity); err != nil || identity.Subject == "" || identity.Email == "" {
		return "", "", ErrGoogleRegistrationExpired
	}
	if existing, findErr := s.userRepo.FindByEmailAnyStatus(ctx, identity.Email); findErr == nil && existing != nil {
		return "", "", ErrGoogleAccountConflict
	} else if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
		return "", "", fmt.Errorf("Google account lookup unavailable")
	}
	if company != nil {
		companyNITExists, err := s.userRepo.CompanyNITExists(ctx, company.NIT)
		if err != nil {
			return "", "", fmt.Errorf("company identifier lookup unavailable")
		}
		if companyNITExists {
			return "", "", ErrGoogleAccountConflict
		}
	}
	consumedIdentity, err := s.redis.GetDel(ctx, registrationKey).Result()
	if err != nil || consumedIdentity != encodedIdentity {
		return "", "", ErrGoogleRegistrationExpired
	}

	now := time.Now().UTC()
	subject := identity.Subject
	user := &model.User{
		ID: uuid.New(), TenantID: uuid.New(), Email: identity.Email,
		GoogleSub: &subject, AuthProvider: "google", Role: model.RoleProfSST,
		FirstName: identity.FirstName, LastName: identity.LastName,
		IsActive: true, IsSynthetic: false, TermsAcceptedAt: &now, PrivacyAcknowledgedAt: &now,
		TermsVersion:         s.cfg.GoogleSignupTermsVersion,
		PrivacyNoticeVersion: s.cfg.GoogleSignupPrivacyVersion,
		// PasswordHash remains an empty, non-authenticating value for Google-only accounts.
	}
	if company != nil {
		company.TenantID = user.TenantID
	}
	if err := s.userRepo.CreateGoogleRegistration(ctx, user, company); err != nil {
		if errors.Is(err, repository.ErrIdentityConflict) {
			return "", "", ErrGoogleAccountConflict
		}
		return "", "", fmt.Errorf("Google registration could not be committed")
	}
	return s.issueGoogleSession(ctx, user)
}

func (s *AuthService) issueGoogleSession(ctx context.Context, user *model.User) (string, string, error) {
	access, err := s.generateAccessToken(user)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("refresh:%s", user.ID.String())
	if err := s.redis.Set(ctx, key, refresh, time.Duration(s.cfg.JWTRefreshDays)*24*time.Hour).Err(); err != nil {
		return "", "", fmt.Errorf("refresh token storage unavailable")
	}
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)
	return access, refresh, nil
}

func (s *AuthService) GoogleSignupIsEnabled() bool {
	return s != nil && s.cfg != nil && s.cfg.GoogleSignupEnabled && !s.cfg.SyntheticPilotEnabled
}

func (s *AuthService) GoogleSignupPublicConfig() GoogleSignupPublicConfig {
	if s == nil || s.cfg == nil {
		return GoogleSignupPublicConfig{}
	}
	return GoogleSignupPublicConfig{
		Enabled: s.GoogleSignupIsEnabled(), SyntheticPilot: s.cfg.SyntheticPilotEnabled,
		ClientID: s.cfg.GoogleClientID, TermsURL: s.cfg.GoogleSignupTermsURL,
		PrivacyURL: s.cfg.GoogleSignupPrivacyURL, TermsVersion: s.cfg.GoogleSignupTermsVersion,
		PrivacyVersion: s.cfg.GoogleSignupPrivacyVersion,
	}
}

func validateGoogleInitialCompany(input GoogleInitialCompany) (*model.Company, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.NIT = strings.TrimSpace(input.NIT)
	input.CIIU = strings.TrimSpace(input.CIIU)
	input.ARL = strings.TrimSpace(input.ARL)
	input.LegalRep = strings.TrimSpace(input.LegalRep)
	input.LegalRepEmail = strings.ToLower(strings.TrimSpace(input.LegalRepEmail))
	parsedEmail, emailErr := mail.ParseAddress(input.LegalRepEmail)
	if input.Name == "" || len(input.Name) > 255 || !companyNITPattern.MatchString(input.NIT) || !companyCIUUPattern.MatchString(input.CIIU) || input.ARL == "" || len(input.ARL) > 100 || input.LegalRep == "" || len(input.LegalRep) > 255 || emailErr != nil || parsedEmail.Address != input.LegalRepEmail {
		return nil, ErrGoogleRegistrationInvalid
	}
	return &model.Company{
		ID: uuid.New(), Name: input.Name, NIT: input.NIT, CIIU: input.CIIU,
		ARL: input.ARL, LegalRep: input.LegalRep, LegalRepEmail: input.LegalRepEmail,
		IsSynthetic: false, RiskClass: "", RiskScore: 0,
	}, nil
}

func isGoogleAuthoritativeEmail(email, hostedDomain string) bool {
	return strings.HasSuffix(strings.ToLower(email), "@gmail.com") || strings.TrimSpace(hostedDomain) != ""
}

func randomOpaqueValue(bytes int) (string, error) {
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func googleRedisKey(purpose, value string) string {
	sum := sha256.Sum256([]byte(value))
	return "google:signup:" + purpose + ":" + hex.EncodeToString(sum[:])
}
