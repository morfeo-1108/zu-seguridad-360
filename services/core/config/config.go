package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port                       string
	Env                        string
	AutoMigrate                bool
	SyntheticPilotEnabled      bool
	AllowedOrigins             string
	GoogleSignupEnabled        bool
	GoogleClientID             string
	GoogleSignupTermsURL       string
	GoogleSignupPrivacyURL     string
	GoogleSignupTermsVersion   string
	GoogleSignupPrivacyVersion string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	RabbitMQURL string

	JWTSecret      string
	JWTExpireHours int
	JWTRefreshDays int

	VaultAddr  string
	VaultToken string

	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	S3Bucket           string

	OTELEndpoint string

	BillingURL string
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using env vars: %v", err)
	}

	return &Config{
		Port:                       getEnv("PORT", "4001"),
		Env:                        getEnv("ENV", "development"),
		AutoMigrate:                strings.EqualFold(getEnv("AUTO_MIGRATE", "false"), "true"),
		SyntheticPilotEnabled:      strings.EqualFold(getEnv("PILOT_SYNTHETIC_ENABLED", "false"), "true"),
		AllowedOrigins:             getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		GoogleSignupEnabled:        strings.EqualFold(getEnv("GOOGLE_SIGNUP_ENABLED", "false"), "true"),
		GoogleClientID:             getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSignupTermsURL:       getEnv("GOOGLE_SIGNUP_TERMS_URL", ""),
		GoogleSignupPrivacyURL:     getEnv("GOOGLE_SIGNUP_PRIVACY_URL", ""),
		GoogleSignupTermsVersion:   getEnv("GOOGLE_SIGNUP_TERMS_VERSION", ""),
		GoogleSignupPrivacyVersion: getEnv("GOOGLE_SIGNUP_PRIVACY_VERSION", ""),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "zuadmin"),
		DBPassword: getEnv("DB_PASSWORD", "zupassword"),
		DBName:     getEnv("DB_NAME", "zu_core"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", "zuredis123"),

		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@localhost:5672/zu_vhost"),

		JWTSecret:      getEnv("JWT_SECRET", "zu-super-secret-key-change-in-prod"),
		JWTExpireHours: 8,
		JWTRefreshDays: 7,

		VaultAddr:  getEnv("VAULT_ADDR", "http://localhost:8200"),
		VaultToken: getEnv("VAULT_TOKEN", "zu-vault-dev-token"),

		AWSRegion: getEnv("AWS_REGION", "us-east-1"),
		S3Bucket:  getEnv("S3_BUCKET", "zu-documents"),

		OTELEndpoint: getEnv("OTEL_ENDPOINT", "http://localhost:4317"),
		BillingURL:   getEnv("BILLING_URL", "http://billing:4010"),
	}
}

// ValidateProduction blocks known development defaults and insecure settings.
// It does not replace secret management, infrastructure review or key rotation.
func (c *Config) ValidateProduction() error {
	if c == nil {
		return nil
	}
	env := strings.ToLower(strings.TrimSpace(c.Env))
	if env != "production" && env != "prod" && env != "staging" {
		return nil
	}
	if c.AutoMigrate {
		return fmt.Errorf("AUTO_MIGRATE must be false in production; use a reviewed migration pipeline")
	}
	if len([]byte(c.JWTSecret)) < 32 || c.JWTSecret == "zu-super-secret-key-change-in-prod" {
		return fmt.Errorf("JWT_SECRET must be a unique secret of at least 32 bytes")
	}
	if c.DBPassword == "" || c.DBPassword == "zupassword" {
		return fmt.Errorf("DB_PASSWORD must not be empty or a development default")
	}
	if c.RedisPassword == "" || c.RedisPassword == "zuredis123" {
		return fmt.Errorf("REDIS_PASSWORD must not be empty or a development default")
	}
	if strings.Contains(c.RabbitMQURL, "zurabbit:zurabbit123@") || c.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL must not use development credentials")
	}
	if c.VaultToken == "" || c.VaultToken == "zu-vault-dev-token" {
		return fmt.Errorf("VAULT_TOKEN must not be empty or a development default")
	}
	switch strings.ToLower(strings.TrimSpace(c.DBSSLMode)) {
	case "require", "verify-ca", "verify-full":
	default:
		return fmt.Errorf("DB_SSLMODE must be require, verify-ca, or verify-full in production")
	}
	origins := strings.Split(c.AllowedOrigins, ",")
	if len(origins) == 0 || strings.TrimSpace(c.AllowedOrigins) == "" {
		return fmt.Errorf("ALLOWED_ORIGINS must be explicitly configured")
	}
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "*" || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") || !strings.HasPrefix(strings.ToLower(origin), "https://") {
			return fmt.Errorf("ALLOWED_ORIGINS contains a wildcard, development origin, or non-HTTPS origin")
		}
	}
	return nil
}

// ValidateSyntheticPilot prevents the synthetic-data routes from connecting to
// a production or externally hosted data plane. Enabling the flag alone is not
// sufficient: the application environment, database name, and dependent stores
// must all identify a local test stack.
func (c *Config) ValidateSyntheticPilot() error {
	if c == nil {
		return nil
	}
	if c.SyntheticPilotEnabled && c.GoogleSignupEnabled {
		return fmt.Errorf("Google signup cannot be enabled in the synthetic pilot")
	}
	if !c.SyntheticPilotEnabled {
		return nil
	}
	env := strings.ToLower(strings.TrimSpace(c.Env))
	if env != "development" && env != "test" {
		return fmt.Errorf("PILOT_SYNTHETIC_ENABLED is permitted only in development or test environments")
	}
	if !isAllowedLocalHost(c.DBHost) {
		return fmt.Errorf("synthetic pilot requires a local PostgreSQL host")
	}
	if !strings.HasSuffix(strings.ToLower(strings.TrimSpace(c.DBName)), "_synthetic") {
		return fmt.Errorf("synthetic pilot requires a dedicated database name ending in _synthetic")
	}
	if !isAllowedLocalHost(endpointHost(c.RedisAddr)) {
		return fmt.Errorf("synthetic pilot requires a local Redis host")
	}
	rabbitURL, err := url.Parse(c.RabbitMQURL)
	if err != nil || !isAllowedLocalHost(rabbitURL.Hostname()) {
		return fmt.Errorf("synthetic pilot requires a local RabbitMQ host")
	}
	return nil
}

// ValidateGoogleSignup keeps real Google identities outside the synthetic
// pilot and fails closed until provider branding/legal configuration is present.
func (c *Config) ValidateGoogleSignup() error {
	if c == nil || !c.GoogleSignupEnabled {
		return nil
	}
	if c.SyntheticPilotEnabled {
		return fmt.Errorf("Google signup cannot be enabled in the synthetic pilot")
	}
	env := strings.ToLower(strings.TrimSpace(c.Env))
	if env != "staging" && env != "production" && env != "prod" {
		return fmt.Errorf("Google signup is permitted only in staging or production")
	}
	if !strings.HasSuffix(strings.TrimSpace(c.GoogleClientID), ".apps.googleusercontent.com") {
		return fmt.Errorf("GOOGLE_CLIENT_ID must be a configured Google Web OAuth client ID")
	}
	for key, value := range map[string]string{
		"GOOGLE_SIGNUP_TERMS_URL":   c.GoogleSignupTermsURL,
		"GOOGLE_SIGNUP_PRIVACY_URL": c.GoogleSignupPrivacyURL,
	} {
		parsed, err := url.Parse(strings.TrimSpace(value))
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must be an absolute HTTPS URL", key)
		}
	}
	if strings.TrimSpace(c.GoogleSignupTermsVersion) == "" || strings.TrimSpace(c.GoogleSignupPrivacyVersion) == "" {
		return fmt.Errorf("Google signup requires versioned terms and privacy notice configuration")
	}
	return nil
}

func endpointHost(address string) string {
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return ""
		}
		return u.Hostname()
	}
	if strings.HasPrefix(address, "[") {
		if end := strings.Index(address, "]"); end > 0 {
			return strings.Trim(address[1:end], "[]")
		}
	}
	if i := strings.LastIndex(address, ":"); i >= 0 {
		return address[:i]
	}
	return address
}

func isAllowedLocalHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1", "postgres", "zu-postgres", "redis", "zu-redis", "rabbitmq", "zu-rabbitmq":
		return true
	default:
		return false
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if viper.IsSet(key) {
		if value := viper.GetString(key); value != "" {
			return value
		}
	}
	return fallback
}
