package config

import "os"

type Config struct {
	Port, AWSRegion, S3Bucket, SophosKey string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Port: env("PORT", "4004"), AWSRegion: env("AWS_REGION", "us-east-1"),
		S3Bucket: env("S3_BUCKET", "zu-documents"), SophosKey: env("SOPHOS_API_KEY", ""),

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
