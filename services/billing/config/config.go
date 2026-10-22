package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Env, Port, DatabaseURL, RabbitMQURL, JWTSecret string
	AutoMigrate                                    bool
	PublicAppURL                                   string

	// Wompi: the key prefix selects the environment (pub_test_ → sandbox,
	// pub_prod_ → production). Going live = replacing these four in Vault.
	WompiPublicKey, WompiPrivateKey, WompiEventsSecret, WompiIntegritySecret string

	// DIAN (factura electrónica). Defaults are DIAN's habilitación test
	// resolution (prefijo SETP), ambiente 2 = pruebas: documents produced
	// with them have no fiscal validity. The issuer NIT/name must be ZU360's.
	DianEnv, DianResolution, DianPrefix, DianTechnicalKey, DianValidFrom, DianValidTo string
	DianFrom, DianTo                                                                  int64
	DianIssuerNIT, DianIssuerName, DianIssuerAddress, DianIssuerEmail, DianTaxLevel   string
	DianSoftwareID, DianSoftwarePIN                                                   string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Env:          env("ENV", "development"),
		Port:         env("PORT", "4010"),
		DatabaseURL:  env("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL:  env("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		JWTSecret:    env("JWT_SECRET", "zu-super-secret-key-change-in-prod"),
		AutoMigrate:  env("AUTO_MIGRATE", "false") == "true",
		PublicAppURL: env("PUBLIC_APP_URL", "https://www.zuseguridad360.com"),

		WompiPublicKey:       env("WOMPI_PUBLIC_KEY", ""),
		WompiPrivateKey:      env("WOMPI_PRIVATE_KEY", ""),
		WompiEventsSecret:    env("WOMPI_EVENTS_SECRET", ""),
		WompiIntegritySecret: env("WOMPI_INTEGRITY_SECRET", ""),

		DianEnv:           env("DIAN_ENV", "2"),
		DianResolution:    env("DIAN_RESOLUTION", "18760000001"),
		DianPrefix:        env("DIAN_PREFIX", "SETP"),
		DianFrom:          envInt("DIAN_FROM", 990000000),
		DianTo:            envInt("DIAN_TO", 995000000),
		DianValidFrom:     env("DIAN_VALID_FROM", "2019-01-19"),
		DianValidTo:       env("DIAN_VALID_TO", "2030-01-19"),
		DianTechnicalKey:  env("DIAN_TECHNICAL_KEY", "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c"),
		DianIssuerNIT:     env("DIAN_ISSUER_NIT", ""),
		DianIssuerName:    env("DIAN_ISSUER_NAME", ""),
		DianIssuerAddress: env("DIAN_ISSUER_ADDRESS", ""),
		DianIssuerEmail:   env("DIAN_ISSUER_EMAIL", "facturacion@zuseguridad360.com"),
		DianTaxLevel:      env("DIAN_TAX_LEVEL", "R-99-PN"),
		DianSoftwareID:    env("DIAN_SOFTWARE_ID", ""),
		DianSoftwarePIN:   env("DIAN_SOFTWARE_PIN", ""),

		VaultEnabled: env("VAULT_ENABLED", "false") == "true",
		VaultAddr:    env("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   env("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    env("VAULT_SECRET_PATH", "secret/data/zu360/billing"),
	}
}

// ValidateProduction refuses unsafe combinations outside development.
func (c *Config) ValidateProduction() error {
	if c.Env == "development" {
		return nil
	}
	if !c.VaultEnabled {
		return fmt.Errorf("ENV=%s requires VAULT_ENABLED=true", c.Env)
	}
	if len(c.JWTSecret) < 32 || c.JWTSecret == "zu-super-secret-key-change-in-prod" {
		return fmt.Errorf("JWT_SECRET must be a real secret")
	}
	if c.DianEnv == "1" && c.DianPrefix == "SETP" {
		return fmt.Errorf("DIAN_ENV=1 (producción) with the habilitación test resolution SETP is not allowed")
	}
	return nil
}

func envInt(k string, d int64) int64 {
	if n, err := strconv.ParseInt(os.Getenv(k), 10, 64); err == nil {
		return n
	}
	return d
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
