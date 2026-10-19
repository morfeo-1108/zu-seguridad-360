package config

import "os"

type Config struct{ Port, AWSRegion, S3Bucket, SophosKey string }

func Load() *Config {
	return &Config{
		Port: env("PORT", "4004"), AWSRegion: env("AWS_REGION", "us-east-1"),
		S3Bucket: env("S3_BUCKET", "zu-documents"), SophosKey: env("SOPHOS_API_KEY", ""),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
