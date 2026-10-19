package config

import "testing"

func TestValidateProductionRejectsDevelopmentDefaults(t *testing.T) {
	cfg := &Config{Env: "production"}
	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected development defaults to be rejected")
	}
}

func TestValidateProductionAcceptsExplicitSecureSettings(t *testing.T) {
	cfg := &Config{
		Env:            "production",
		JWTSecret:      "a-long-unique-production-secret-with-more-than-32-bytes",
		DBPassword:     "unique-db-password",
		RedisPassword:  "unique-redis-password",
		RabbitMQURL:    "amqp://service-user:unique-secret@rabbitmq.internal:5672/zu_vhost",
		VaultToken:     "unique-vault-token",
		DBSSLMode:      "verify-full",
		AllowedOrigins: "https://app.example.co,https://portal.example.co",
	}
	if err := cfg.ValidateProduction(); err != nil {
		t.Fatalf("expected explicitly secure settings to pass, got: %v", err)
	}
}

func TestValidateProductionAlsoRecognizesProdAlias(t *testing.T) {
	cfg := &Config{Env: "prod"}
	if err := cfg.ValidateProduction(); err == nil {
		t.Fatal("expected prod alias to enforce production validation")
	}
}

func TestValidateProductionRejectsAutoMigrate(t *testing.T) {
	cfg := &Config{Env: "production", AutoMigrate: true}
	if err := cfg.ValidateProduction(); err == nil {
		t.Fatalf("expected explicit AutoMigrate refusal, got: %v", err)
	}
}

func validSyntheticConfig() *Config {
	return &Config{
		Env:                   "development",
		SyntheticPilotEnabled: true,
		DBHost:                "postgres",
		DBName:                "zu_core_synthetic",
		RedisAddr:             "redis:6379",
		RabbitMQURL:           "amqp://user:pass@rabbitmq:5672/zu_vhost",
	}
}

func TestValidateSyntheticPilotAllowsDedicatedLocalStack(t *testing.T) {
	if err := validSyntheticConfig().ValidateSyntheticPilot(); err != nil {
		t.Fatalf("expected local synthetic configuration to pass: %v", err)
	}
}

func TestValidateSyntheticPilotRejectsProductionAndRemoteStores(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
	}{
		{"production", func(c *Config) { c.Env = "production" }},
		{"remote database", func(c *Config) { c.DBHost = "db.example.com" }},
		{"shared database name", func(c *Config) { c.DBName = "zu_core" }},
		{"remote redis", func(c *Config) { c.RedisAddr = "redis.example.com:6379" }},
		{"remote rabbitmq", func(c *Config) { c.RabbitMQURL = "amqp://user:pass@mq.example.com:5672/zu" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validSyntheticConfig()
			tt.edit(cfg)
			if err := cfg.ValidateSyntheticPilot(); err == nil {
				t.Fatal("unsafe synthetic configuration was accepted")
			}
		})
	}
}

func TestValidateSyntheticPilotIsOptionalWhenDisabled(t *testing.T) {
	cfg := validSyntheticConfig()
	cfg.SyntheticPilotEnabled = false
	cfg.Env = "production"
	cfg.DBHost = "db.example.com"
	if err := cfg.ValidateSyntheticPilot(); err != nil {
		t.Fatalf("disabled synthetic mode should not add validation requirements: %v", err)
	}
}

func validGoogleSignupConfig() *Config {
	return &Config{
		Env: "staging", GoogleSignupEnabled: true,
		GoogleClientID:             "web-client.apps.googleusercontent.com",
		GoogleSignupTermsURL:       "https://example.co/terms",
		GoogleSignupPrivacyURL:     "https://example.co/privacy",
		GoogleSignupTermsVersion:   "2026-10-01",
		GoogleSignupPrivacyVersion: "2026-10-01",
	}
}

func TestValidateGoogleSignupAcceptsConfiguredStaging(t *testing.T) {
	if err := validGoogleSignupConfig().ValidateGoogleSignup(); err != nil {
		t.Fatalf("expected valid staging config, got: %v", err)
	}
}

func TestValidateGoogleSignupFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		edit func(*Config)
	}{
		{"local environment", func(c *Config) { c.Env = "development" }},
		{"synthetic mode", func(c *Config) { c.SyntheticPilotEnabled = true }},
		{"invalid client ID", func(c *Config) { c.GoogleClientID = "" }},
		{"non-HTTPS legal URL", func(c *Config) { c.GoogleSignupTermsURL = "http://example.co/terms" }},
		{"missing policy version", func(c *Config) { c.GoogleSignupPrivacyVersion = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validGoogleSignupConfig()
			tt.edit(cfg)
			if err := cfg.ValidateGoogleSignup(); err == nil {
				t.Fatal("unsafe Google signup configuration was accepted")
			}
		})
	}
}

func TestValidateGoogleSignupDisabledIsNoOp(t *testing.T) {
	if err := (&Config{Env: "development"}).ValidateGoogleSignup(); err != nil {
		t.Fatalf("disabled Google signup should remain optional: %v", err)
	}
}
