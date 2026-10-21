package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zu-seguridad/core/config"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
)

type ZUClaims struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	CompanyID string `json:"company_id,omitempty"`
	Role      string `json:"role"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type AuthService struct {
	userRepo        *repository.UserRepository
	redis           *redis.Client
	cfg             *config.Config
	googleValidator googleIDTokenValidator
}

func NewAuthService(userRepo *repository.UserRepository, redis *redis.Client, cfg *config.Config) (*AuthService, error) {
	var googleValidator googleIDTokenValidator
	if cfg != nil && cfg.GoogleSignupEnabled {
		validator, err := newGoogleIDTokenValidator(context.Background())
		if err != nil {
			return nil, fmt.Errorf("Google ID token validator initialization failed")
		}
		googleValidator = validator
	}
	return &AuthService{userRepo: userRepo, redis: redis, cfg: cfg, googleValidator: googleValidator}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", ErrInvalidCredentials
	}

	if !user.CheckPassword(password) {
		return "", "", ErrInvalidCredentials
	}

	if !user.IsActive {
		return "", "", ErrAccountInactive
	}
	companyID := companyIDForUser(user)
	if companyID == "" {
		companies, listErr := s.userRepo.ListMemberCompanies(ctx, user.ID, user.TenantID)
		if listErr != nil || len(companies) == 0 {
			return "", "", ErrInvalidCredentials
		}
		companyID = companies[0].ID.String()
	}
	companyUUID, err := uuid.Parse(companyID)
	if err != nil || companyUUID == uuid.Nil {
		return "", "", ErrInvalidCredentials
	}
	member, err := s.userRepo.HasActiveCompanyMembership(ctx, user.ID, user.TenantID, companyUUID)
	if err != nil || !member {
		return "", "", ErrInvalidCredentials
	}
	scopedUser := copyUserWithCompany(user, companyUUID)

	accessToken, err = s.generateAccessToken(scopedUser)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = s.generateRefreshToken(scopedUser)
	if err != nil {
		return "", "", err
	}

	// Store refresh token in Redis
	key := fmt.Sprintf("refresh:%s", user.ID.String())
	if err := s.redis.Set(ctx, key, refreshToken, time.Duration(s.cfg.JWTRefreshDays)*24*time.Hour).Err(); err != nil {
		return "", "", fmt.Errorf("refresh token storage unavailable")
	}

	// Update last login
	s.userRepo.UpdateLastLogin(ctx, user.ID)

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.validateToken(refreshToken)
	if err != nil {
		return "", ErrTokenInvalid
	}

	key := fmt.Sprintf("refresh:%s", claims.UserID)
	stored, err := s.redis.Get(ctx, key).Result()
	if err != nil || stored != refreshToken {
		return "", ErrTokenInvalid
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", ErrTokenInvalid
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", ErrTokenInvalid
	}
	if !user.IsActive || user.TenantID.String() != claims.TenantID {
		return "", ErrTokenInvalid
	}
	if claims.CompanyID == "" {
		// Tenant-only session (account without companies yet): never widen it
		// to a company scope here; that requires an explicit switch-company.
		tenantOnly := *user
		tenantOnly.CompanyID = nil
		return s.generateAccessToken(&tenantOnly)
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil || companyID == uuid.Nil {
		return "", ErrTokenInvalid
	}
	member, err := s.userRepo.HasActiveCompanyMembership(ctx, user.ID, user.TenantID, companyID)
	if err != nil || !member {
		return "", ErrTokenInvalid
	}

	return s.generateAccessToken(copyUserWithCompany(user, companyID))
}

func (s *AuthService) SwitchCompany(ctx context.Context, userIDText, tenantIDText, companyIDText string) (string, string, error) {
	userID, err := uuid.Parse(userIDText)
	if err != nil || userID == uuid.Nil {
		return "", "", ErrTokenInvalid
	}
	tenantID, err := uuid.Parse(tenantIDText)
	if err != nil || tenantID == uuid.Nil {
		return "", "", ErrTokenInvalid
	}
	companyID, err := uuid.Parse(companyIDText)
	if err != nil || companyID == uuid.Nil {
		return "", "", ErrTokenInvalid
	}
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || !user.IsActive || user.TenantID != tenantID {
		return "", "", ErrTokenInvalid
	}
	member, err := s.userRepo.HasActiveCompanyMembership(ctx, userID, tenantID, companyID)
	if err != nil || !member {
		return "", "", ErrTokenInvalid
	}
	scopedUser := copyUserWithCompany(user, companyID)
	accessToken, err := s.generateAccessToken(scopedUser)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.generateRefreshToken(scopedUser)
	if err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("refresh:%s", user.ID.String())
	if err := s.redis.Set(ctx, key, refreshToken, time.Duration(s.cfg.JWTRefreshDays)*24*time.Hour).Err(); err != nil {
		return "", "", fmt.Errorf("refresh token storage unavailable")
	}
	return accessToken, refreshToken, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userIDText string) (*model.User, error) {
	userID, err := uuid.Parse(userIDText)
	if err != nil || userID == uuid.Nil {
		return nil, ErrTokenInvalid
	}
	return s.userRepo.FindByID(ctx, userID)
}

func (s *AuthService) Logout(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh:%s", userID)
	return s.redis.Del(ctx, key).Err()
}

func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	claims := ZUClaims{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID.String(),
		CompanyID: companyIDForUser(user),
		Role:      string(user.Role),
		Email:     user.Email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTExpireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) generateRefreshToken(user *model.User) (string, error) {
	claims := ZUClaims{
		UserID:    user.ID.String(),
		TenantID:  user.TenantID.String(),
		CompanyID: companyIDForUser(user),
		Role:      string(user.Role),
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.cfg.JWTRefreshDays) * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *AuthService) validateToken(tokenStr string) (*ZUClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &ZUClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*ZUClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	if claims.TokenType != "refresh" || claims.UserID == "" || claims.TenantID == "" || claims.CompanyID == "" || claims.Subject != claims.UserID {
		return nil, ErrTokenInvalid
	}
	if _, err := uuid.Parse(claims.UserID); err != nil {
		return nil, ErrTokenInvalid
	}
	if _, err := uuid.Parse(claims.TenantID); err != nil {
		return nil, ErrTokenInvalid
	}
	if claims.CompanyID != "" {
		companyID, err := uuid.Parse(claims.CompanyID)
		if err != nil || companyID == uuid.Nil {
			return nil, ErrTokenInvalid
		}
	}
	return claims, nil
}

func companyIDForUser(user *model.User) string {
	if user == nil || user.CompanyID == nil || *user.CompanyID == uuid.Nil {
		return ""
	}
	return user.CompanyID.String()
}

func copyUserWithCompany(user *model.User, companyID uuid.UUID) *model.User {
	copy := *user
	copy.CompanyID = &companyID
	return &copy
}
