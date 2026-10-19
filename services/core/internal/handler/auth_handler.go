package handler

import (
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	accessToken, refreshToken, err := h.authSvc.Login(c.Context(), req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		if err == service.ErrAccountInactive {
			return fiber.NewError(fiber.StatusForbidden, "account is inactive")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "login failed")
	}

	return c.JSON(loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    28800, // 8 hours
	})
}

func (h *AuthHandler) GoogleConfig(c *fiber.Ctx) error {
	return c.JSON(h.authSvc.GoogleSignupPublicConfig())
}

func (h *AuthHandler) GoogleChallenge(c *fiber.Ctx) error {
	if err := h.checkGoogleRateLimit(c, "challenge", 30, time.Minute); err != nil {
		return err
	}
	nonce, err := h.authSvc.NewGoogleSignupChallenge(c.Context())
	if err != nil {
		if errors.Is(err, service.ErrGoogleSignupDisabled) {
			return fiber.NewError(fiber.StatusNotFound, "Google sign-in is unavailable")
		}
		return fiber.NewError(fiber.StatusServiceUnavailable, "Google sign-in is temporarily unavailable")
	}
	return c.JSON(fiber.Map{"nonce": nonce})
}

func (h *AuthHandler) GoogleStart(c *fiber.Ctx) error {
	if err := h.checkGoogleRateLimit(c, "start", 10, time.Minute); err != nil {
		return err
	}
	var req struct {
		Credential string `json:"credential"`
		Nonce      string `json:"nonce"`
	}
	if err := c.BodyParser(&req); err != nil || req.Credential == "" || req.Nonce == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Google credential and challenge are required")
	}
	result, err := h.authSvc.BeginGoogleAuthentication(c.Context(), req.Credential, req.Nonce)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGoogleSignupDisabled):
			return fiber.NewError(fiber.StatusNotFound, "Google sign-in is unavailable")
		case errors.Is(err, service.ErrGoogleTokenInvalid), errors.Is(err, service.ErrGoogleNonceInvalid), errors.Is(err, service.ErrGoogleEmailNotVerified):
			return fiber.NewError(fiber.StatusUnauthorized, "Google sign-in could not be verified")
		case errors.Is(err, service.ErrGoogleAccountConflict):
			return fiber.NewError(fiber.StatusConflict, "This email cannot be registered with Google. Sign in using the existing account or contact support.")
		case errors.Is(err, service.ErrAccountInactive):
			return fiber.NewError(fiber.StatusForbidden, "This account is unavailable")
		default:
			return fiber.NewError(fiber.StatusServiceUnavailable, "Google sign-in is temporarily unavailable")
		}
	}
	if result.RegistrationRequired {
		return c.JSON(fiber.Map{"registration_required": true, "registration_token": result.RegistrationToken, "profile": result.Profile})
	}
	return c.JSON(loginResponse{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken, TokenType: "Bearer", ExpiresIn: 28800})
}

func (h *AuthHandler) GoogleComplete(c *fiber.Ctx) error {
	if err := h.checkGoogleRateLimit(c, "complete", 5, time.Hour); err != nil {
		return err
	}
	var req struct {
		RegistrationToken         string                       `json:"registration_token"`
		Company                   service.GoogleInitialCompany `json:"company"`
		TermsAccepted             bool                         `json:"terms_accepted"`
		PrivacyNoticeAcknowledged bool                         `json:"privacy_notice_acknowledged"`
	}
	if err := c.BodyParser(&req); err != nil || req.RegistrationToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "invalid registration request")
	}
	accessToken, refreshToken, err := h.authSvc.CompleteGoogleSignup(c.Context(), service.GoogleSignupCompletion{
		RegistrationToken: req.RegistrationToken, Company: req.Company,
		TermsAccepted: req.TermsAccepted, PrivacyNoticeAcknowledged: req.PrivacyNoticeAcknowledged,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrGoogleSignupDisabled):
			return fiber.NewError(fiber.StatusNotFound, "Google sign-in is unavailable")
		case errors.Is(err, service.ErrGoogleTermsNotAccepted), errors.Is(err, service.ErrGoogleRegistrationInvalid):
			return fiber.NewError(fiber.StatusBadRequest, "Review the company details and required legal acknowledgements")
		case errors.Is(err, service.ErrGoogleRegistrationExpired):
			return fiber.NewError(fiber.StatusUnauthorized, "Google registration expired. Start again.")
		case errors.Is(err, service.ErrGoogleAccountConflict):
			return fiber.NewError(fiber.StatusConflict, "This account or company identifier cannot be registered")
		default:
			return fiber.NewError(fiber.StatusServiceUnavailable, "Registration is temporarily unavailable")
		}
	}
	return c.JSON(loginResponse{AccessToken: accessToken, RefreshToken: refreshToken, TokenType: "Bearer", ExpiresIn: 28800})
}

func (h *AuthHandler) checkGoogleRateLimit(c *fiber.Ctx, operation string, limit int64, window time.Duration) error {
	allowed, err := h.authSvc.CheckGoogleSignupRateLimit(c.Context(), operation, c.IP(), limit, window)
	if errors.Is(err, service.ErrGoogleSignupDisabled) {
		return fiber.NewError(fiber.StatusNotFound, "Google sign-in is unavailable")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Google sign-in is temporarily unavailable")
	}
	if !allowed {
		return fiber.NewError(fiber.StatusTooManyRequests, "Too many Google sign-in attempts. Try again later.")
	}
	return nil
}

func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	accessToken, err := h.authSvc.RefreshToken(c.Context(), body.RefreshToken)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	return c.JSON(fiber.Map{"access_token": accessToken, "token_type": "Bearer"})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, err := h.authSvc.GetCurrentUser(c.Context(), c.Locals("user_id").(string))
	if err != nil || !user.IsActive || user.TenantID.String() != c.Locals("tenant_id").(string) {
		return fiber.NewError(fiber.StatusUnauthorized, "session unavailable")
	}
	companyID, _ := c.Locals("company_id").(string)
	return c.JSON(fiber.Map{
		"id": user.ID, "email": user.Email, "role": user.Role,
		"first_name": user.FirstName, "last_name": user.LastName,
		"tenant_id": user.TenantID, "company_id": companyID,
		"is_synthetic": user.IsSynthetic,
	})
}

func (h *AuthHandler) SwitchCompany(c *fiber.Ctx) error {
	var body struct {
		CompanyID string `json:"company_id"`
	}
	if err := c.BodyParser(&body); err != nil || body.CompanyID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "company_id required")
	}
	accessToken, refreshToken, err := h.authSvc.SwitchCompany(
		c.Context(), c.Locals("user_id").(string), c.Locals("tenant_id").(string), body.CompanyID,
	)
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "company membership required")
	}
	return c.JSON(fiber.Map{
		"access_token": accessToken, "refresh_token": refreshToken,
		"token_type": "Bearer", "active_company_id": body.CompanyID,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.authSvc.Logout(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "logout failed")
	}
	return c.JSON(fiber.Map{"message": "logged out successfully"})
}

// GlobalErrorHandler maneja todos los errores de forma centralizada
func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	msg := "internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		msg = e.Message
	}
	if code >= fiber.StatusInternalServerError {
		// Do not return or log raw error messages: database/provider errors may
		// contain personal data, query fragments, or infrastructure details.
		log.Printf("request failure request_id=%v status=%d error_type=%T", c.Locals("requestid"), code, err)
		msg = "internal server error"
	}

	return c.Status(code).JSON(fiber.Map{
		"error":      msg,
		"status":     code,
		"request_id": c.Locals("requestid"),
	})
}
