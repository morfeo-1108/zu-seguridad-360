package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "test-secret-test-secret-test-secret"

func sign(t *testing.T, claims jwt.MapClaims, method jwt.SigningMethod, key interface{}) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func base() jwt.MapClaims {
	return jwt.MapClaims{
		"user_id": "11111111-1111-1111-1111-111111111111", "sub": "11111111-1111-1111-1111-111111111111",
		"tenant_id": "22222222-2222-2222-2222-222222222222", "role": "sst_professional",
		"token_type": "access", "exp": time.Now().Add(time.Minute).Unix(),
	}
}

func TestParseAcceptsCoreAccessToken(t *testing.T) {
	c := base()
	c["company_id"] = "33333333-3333-3333-3333-333333333333"
	id, err := Parse(sign(t, c, jwt.SigningMethodHS256, []byte(secret)), secret)
	if err != nil || id.CompanyID != "33333333-3333-3333-3333-333333333333" || id.TenantID == "" {
		t.Fatalf("expected valid identity, got %+v %v", id, err)
	}
}

func TestParseRejects(t *testing.T) {
	cases := map[string]func() string{
		"refresh token": func() string {
			c := base()
			c["token_type"] = "refresh"
			return sign(t, c, jwt.SigningMethodHS256, []byte(secret))
		},
		"expired": func() string {
			c := base()
			c["exp"] = time.Now().Add(-time.Minute).Unix()
			return sign(t, c, jwt.SigningMethodHS256, []byte(secret))
		},
		"no exp": func() string {
			c := base()
			delete(c, "exp")
			return sign(t, c, jwt.SigningMethodHS256, []byte(secret))
		},
		"sub mismatch": func() string { c := base(); c["sub"] = "x"; return sign(t, c, jwt.SigningMethodHS256, []byte(secret)) },
		"bad tenant": func() string {
			c := base()
			c["tenant_id"] = "nope"
			return sign(t, c, jwt.SigningMethodHS256, []byte(secret))
		},
		"nil company": func() string {
			c := base()
			c["company_id"] = "00000000-0000-0000-0000-000000000000"
			return sign(t, c, jwt.SigningMethodHS256, []byte(secret))
		},
		"wrong secret": func() string { return sign(t, base(), jwt.SigningMethodHS256, []byte("other-secret")) },
		"HS512":        func() string { return sign(t, base(), jwt.SigningMethodHS512, []byte(secret)) },
		"alg none":     func() string { return sign(t, base(), jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType) },
	}
	for name, mk := range cases {
		if _, err := Parse(mk(), secret); err == nil {
			t.Errorf("%s: expected rejection", name)
		}
	}
}
