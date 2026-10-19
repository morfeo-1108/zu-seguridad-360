package config

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadVaultSecretsIsNoopWhenDisabled(t *testing.T) {
	c := &Config{VaultEnabled: false, DBPassword: "unchanged"}
	if err := c.LoadVaultSecrets(); err != nil {
		t.Fatalf("expected no-op when VaultEnabled is false, got %v", err)
	}
	if c.DBPassword != "unchanged" {
		t.Fatalf("expected DBPassword untouched, got %q", c.DBPassword)
	}
}

func TestLoadVaultSecretsOverridesFromKV2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vault-Token") != "real-token" {
			t.Fatalf("expected the configured vault token to be sent")
		}
		fmt.Fprint(w, `{"data":{"data":{
			"db_password":"vault-db-pw",
			"redis_password":"vault-redis-pw",
			"jwt_secret":"vault-jwt-secret-at-least-32-bytes-long",
			"rabbitmq_url":"amqp://vault:vault@rabbitmq:5672/zu_vhost"
		}}}`)
	}))
	defer server.Close()

	c := &Config{
		VaultEnabled: true,
		VaultAddr:    server.URL,
		VaultToken:   "real-token",
		VaultPath:    "secret/data/zu360/core",
		DBPassword:   "zupassword",
	}
	if err := c.LoadVaultSecrets(); err != nil {
		t.Fatalf("expected successful vault load, got %v", err)
	}
	if c.DBPassword != "vault-db-pw" {
		t.Fatalf("expected DBPassword from vault, got %q", c.DBPassword)
	}
	if c.JWTSecret != "vault-jwt-secret-at-least-32-bytes-long" {
		t.Fatalf("expected JWTSecret from vault, got %q", c.JWTSecret)
	}
}

func TestLoadVaultSecretsFailsClosedWhenMissingKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":{"data":{"db_password":"only-one-key"}}}`)
	}))
	defer server.Close()

	c := &Config{VaultEnabled: true, VaultAddr: server.URL, VaultToken: "x", VaultPath: "secret/data/zu360/core"}
	if err := c.LoadVaultSecrets(); err == nil {
		t.Fatal("expected an error when required keys are missing from the vault secret")
	}
}

func TestLoadVaultSecretsFailsClosedWhenUnreachable(t *testing.T) {
	c := &Config{VaultEnabled: true, VaultAddr: "http://127.0.0.1:1", VaultToken: "x", VaultPath: "secret/data/zu360/core"}
	if err := c.LoadVaultSecrets(); err == nil {
		t.Fatal("expected an error when vault is unreachable, got nil (fail-open is not acceptable)")
	}
}
