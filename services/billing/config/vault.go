package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	vault "github.com/hashicorp/vault/api"
)

// LoadVaultSecrets overrides secret fields with values read from Vault KV v2
// when VAULT_ENABLED=true. It is a no-op in local/dev setups that never set
// that flag (the default). When explicitly enabled, a failed read is fatal —
// reaching this path means an operator intended to run on Vault-managed
// secrets, and silently falling back to whatever default/env value is
// already in Config would defeat the point. Mirrors services/core/config/vault.go.
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
		data = secret.Data // KV v1 mounts return the fields directly under secret.Data.
	}

	required := map[string]*string{
		"database_url": &c.DatabaseURL,
		"rabbitmq_url": &c.RabbitMQURL,
		"jwt_secret":   &c.JWTSecret,
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
		"wompi_public_key":       &c.WompiPublicKey,
		"wompi_private_key":      &c.WompiPrivateKey,
		"wompi_events_secret":    &c.WompiEventsSecret,
		"wompi_integrity_secret": &c.WompiIntegritySecret,
		"dian_issuer_nit":        &c.DianIssuerNIT,
		"dian_issuer_name":       &c.DianIssuerName,
		"dian_issuer_address":    &c.DianIssuerAddress,
		"dian_resolution":        &c.DianResolution,
		"dian_prefix":            &c.DianPrefix,
		"dian_technical_key":     &c.DianTechnicalKey,
		"dian_valid_from":        &c.DianValidFrom,
		"dian_valid_to":          &c.DianValidTo,
		"dian_software_id":       &c.DianSoftwareID,
		"dian_software_pin":      &c.DianSoftwarePIN,
		"dian_env":               &c.DianEnv,
	}
	for key, target := range optional {
		if value, ok := data[key].(string); ok && value != "" {
			*target = value
		}
	}
	for key, target := range map[string]*int64{"dian_from": &c.DianFrom, "dian_to": &c.DianTo} {
		if raw, ok := data[key].(string); ok && raw != "" {
			n, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				return fmt.Errorf("vault key %s is not an integer", key)
			}
			*target = n
		}
	}
	return nil
}
