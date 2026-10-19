package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zu-seguridad/core/config"
	"google.golang.org/api/idtoken"
)

type fixedGoogleValidator struct {
	payload *idtoken.Payload
	err     error
}

func (v fixedGoogleValidator) Validate(context.Context, string, string) (*idtoken.Payload, error) {
	return v.payload, v.err
}

func TestEnsureGoogleSignupEnabledRejectsSyntheticAndUnconfigured(t *testing.T) {
	for name, svc := range map[string]*AuthService{
		"nil service": nil,
		"nil config":  {cfg: nil},
		"synthetic":   {cfg: &config.Config{GoogleSignupEnabled: true, SyntheticPilotEnabled: true, GoogleClientID: "web.apps.googleusercontent.com"}},
		"disabled":    {cfg: &config.Config{GoogleClientID: "web.apps.googleusercontent.com"}},
		"no client":   {cfg: &config.Config{GoogleSignupEnabled: true}},
	} {
		t.Run(name, func(t *testing.T) {
			if err := svc.ensureGoogleSignupEnabled(); !errors.Is(err, ErrGoogleSignupDisabled) {
				t.Fatalf("expected disabled error, got %v", err)
			}
		})
	}
}

func TestGoogleAuthorityRequiresGmailOrHostedDomain(t *testing.T) {
	cases := []struct {
		email string
		hd    string
		want  bool
	}{
		{"professional@gmail.com", "", true},
		{"person@company.example", "company.example", true},
		{"person@third-party.example", "", false},
	}
	for _, tc := range cases {
		if got := isGoogleAuthoritativeEmail(tc.email, tc.hd); got != tc.want {
			t.Errorf("isGoogleAuthoritativeEmail(%q, %q) = %t, want %t", tc.email, tc.hd, got, tc.want)
		}
	}
}

func TestBeginGoogleAuthenticationRejectsInvalidTokenClaimsBeforeStorage(t *testing.T) {
	future := time.Now().Add(time.Hour).Unix()
	validBase := idtoken.Payload{Issuer: "https://accounts.google.com", Audience: "web.apps.googleusercontent.com", Expires: future, IssuedAt: time.Now().Unix()}
	cases := []struct {
		name string
		edit func(*idtoken.Payload)
	}{
		{"wrong audience", func(p *idtoken.Payload) { p.Audience = "other.apps.googleusercontent.com" }},
		{"wrong issuer", func(p *idtoken.Payload) { p.Issuer = "https://evil.example" }},
		{"expired", func(p *idtoken.Payload) { p.Expires = time.Now().Add(-time.Minute).Unix() }},
		{"future issued-at", func(p *idtoken.Payload) { p.IssuedAt = time.Now().Add(10 * time.Minute).Unix() }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := validBase
			tc.edit(&payload)
			svc := &AuthService{
				cfg:             &config.Config{GoogleSignupEnabled: true, GoogleClientID: "web.apps.googleusercontent.com"},
				googleValidator: fixedGoogleValidator{payload: &payload},
			}
			_, err := svc.BeginGoogleAuthentication(context.Background(), "signed-looking-token", "random-nonce")
			if !errors.Is(err, ErrGoogleTokenInvalid) {
				t.Fatalf("expected invalid token error, got %v", err)
			}
		})
	}
}

func TestValidateGoogleInitialCompanyRequiresDeclaredFieldsAndValidShapes(t *testing.T) {
	valid := GoogleInitialCompany{
		Name: "Empresa declarada SAS", NIT: "900123-4", CIIU: "6201", ARL: "ARL declarada",
		LegalRep: "Representante declarado", LegalRepEmail: "rep@example.co",
	}
	company, err := validateGoogleInitialCompany(valid)
	if err != nil {
		t.Fatalf("expected well-shaped declared data, got %v", err)
	}
	if company.IsSynthetic || company.RiskClass != "" || company.RiskScore != 0 || company.PlanID != nil {
		t.Fatalf("new company must be real-profile, unclassified, unscored and without plan: %+v", company)
	}

	invalidCases := []GoogleInitialCompany{
		{},
		{Name: "Empresa", NIT: "TEST-NIT-1", CIIU: "6201", ARL: "ARL", LegalRep: "Persona", LegalRepEmail: "x@example.co"},
		{Name: "Empresa", NIT: "900123", CIIU: "CIUU INVALIDO", ARL: "ARL", LegalRep: "Persona", LegalRepEmail: "x@example.co"},
		{Name: "Empresa", NIT: "900123", CIIU: "6201", ARL: "ARL", LegalRep: "Persona", LegalRepEmail: "no-es-correo"},
	}
	for i, input := range invalidCases {
		if _, err := validateGoogleInitialCompany(input); !errors.Is(err, ErrGoogleRegistrationInvalid) {
			t.Errorf("case %d: expected invalid registration data, got %v", i, err)
		}
	}
}
