package config

import (
	"fmt"
	"os"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// LoadVaultSecrets overrides secret fields with values read from Vault KV v2
// when VAULT_ENABLED=true. It is a no-op in local/dev setups that never set
// that flag (the default), so it changes nothing about how the synthetic
// pilot or any already-verified Docker flow behaves. When explicitly
// enabled, a failed read is fatal — reaching this path means an operator
// intended to run on Vault-managed secrets, and silently falling back to
// whatever default/env value is already in Config would defeat the point.
func (c *Config) LoadVaultSecrets() error {
	if !c.VaultEnabled {
		return nil
	}
	// The token normally arrives as a Docker secret file, so it never sits in
	// an env file or in `docker inspect` output.
	if f := os.Getenv("VAULT_TOKEN_FILE"); f != "" {
		raw, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("vault token file: %w", err)
		}
		c.VaultToken = strings.TrimSpace(string(raw))
	}
	client, err := vault.NewClient(&vault.Config{Address: c.VaultAddr})
	if err != nil {
		return fmt.Errorf("vault client init failed: %w", err)
	}
	client.SetToken(c.VaultToken)

	secret, err := client.Logical().Read(c.VaultPath)
	if err != nil {
		return fmt.Errorf("vault read of %s failed: %w", c.VaultPath, err)
	}
	if secret == nil || secret.Data == nil {
		return fmt.Errorf("vault secret at %s not found", c.VaultPath)
	}
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		// KV v1 mounts return the fields directly under secret.Data.
		data = secret.Data
	}

	required := map[string]*string{
		"db_password":    &c.DBPassword,
		"redis_password": &c.RedisPassword,
		"jwt_secret":     &c.JWTSecret,
		"rabbitmq_url":   &c.RabbitMQURL,
	}
	var missing []string
	for key, target := range required {
		value, ok := data[key].(string)
		if !ok || value == "" {
			missing = append(missing, key)
			continue
		}
		*target = value
	}
	if len(missing) > 0 {
		return fmt.Errorf("vault secret at %s is missing required keys: %v", c.VaultPath, missing)
	}

	optional := map[string]*string{
		"aws_access_key_id":     &c.AWSAccessKeyID,
		"aws_secret_access_key": &c.AWSSecretAccessKey,
		"google_client_id":      &c.GoogleClientID,
	}
	for key, target := range optional {
		if value, ok := data[key].(string); ok && value != "" {
			*target = value
		}
	}
	return nil
}
