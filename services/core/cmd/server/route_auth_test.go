package main

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

// This is a structural regression guard, not a runtime RBAC matrix: it reads
// main.go's own source and checks that every route group other than the
// explicitly-public ones (health, login/refresh, Google sign-in, the public
// JWT-free endpoints under /v1/auth) includes jwt as a middleware. A full
// runtime matrix (role × tenant × company × verb against every one of
// Core's ~127 routes) is a much larger undertaking than this test attempts;
// this only catches the common, high-consequence mistake of registering a
// new route group and forgetting the auth middleware entirely — exactly the
// shape of gap found in evidence/PHVA/medical exams before they were wired
// with company scoping.
var explicitlyPublicGroups = map[string]bool{
	`app.Group("/v1/auth")`: true,
	// Public certificate verification — what a QR code printed on a
	// certificate resolves to. No session exists at scan time, so it cannot
	// require jwt; it is rate-limited and returns a minimized DTO instead
	// (see TrainingService.VerifyCertificate / CertificateQRPNG).
	`app.Group("/v1/public/certificates", publicCertLimiter)`: true,
}

func TestEveryNonPublicRouteGroupRequiresJWT(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("could not read main.go: %v", err)
	}
	text := string(src)

	groupPattern := regexp.MustCompile(`(?m)^\s*\w+\s*:?=\s*app\.Group\(([^)]*)\)`)
	matches := groupPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		t.Fatal("expected to find at least one app.Group(...) call — did main.go's structure change?")
	}

	for _, m := range matches {
		full := m[0]
		args := m[1]
		if explicitlyPublicGroups[strings.TrimSpace("app.Group("+args+")")] {
			continue
		}
		if strings.Contains(args, `"/v1/auth"`) {
			continue // the auth group itself mixes public and JWT-protected routes per-handler
		}
		if !strings.Contains(args, "jwt") {
			t.Errorf("route group appears to skip JWT auth entirely: %s", full)
		}
	}
}

// Standalone app.Get/Post/... calls outside any group are rarer and easier to
// miss in review. Only /v1/health may skip jwt.
func TestEveryStandaloneRouteRequiresJWTExceptHealth(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("could not read main.go: %v", err)
	}
	text := string(src)

	routePattern := regexp.MustCompile(`(?m)^\s*app\.(Get|Post|Put|Patch|Delete)\(([^)]*(?:\([^)]*\)[^)]*)*)\)`)
	matches := routePattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		t.Fatal("expected to find standalone app.* route registrations — did main.go's structure change?")
	}

	for _, m := range matches {
		full := m[0]
		args := m[2]
		if strings.Contains(args, `"/v1/health"`) {
			continue
		}
		if !strings.Contains(args, "jwt") {
			t.Errorf("standalone route appears to skip JWT auth: %s", full)
		}
	}
}
