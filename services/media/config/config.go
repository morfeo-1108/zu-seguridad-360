package config

import "os"

type Config struct {
	Port, AWSRegion, S3Bucket, SophosKey, JWTSecret string
	S3Endpoint, S3PublicEndpoint, AWSAccessKeyID, AWSSecretAccessKey string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Port: env("PORT", "4004"), AWSRegion: env("AWS_REGION", "us-east-1"),
		S3Bucket: env("S3_BUCKET", "zu-documents"), SophosKey: env("SOPHOS_API_KEY", ""),
		JWTSecret: env("JWT_SECRET", "zu-super-secret-key-change-in-prod"),

		S3Endpoint:         env("S3_ENDPOINT", ""),
		S3PublicEndpoint:   env("S3_PUBLIC_ENDPOINT", ""),
		AWSAccessKeyID:     env("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: env("AWS_SECRET_ACCESS_KEY", ""),

		VaultEnabled: env("VAULT_ENABLED", "false") == "true",
		VaultAddr:    env("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   env("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    env("VAULT_SECRET_PATH", "secret/data/zu360/media"),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
