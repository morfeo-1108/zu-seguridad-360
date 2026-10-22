"""Vault secret loading for the AI service.

Mirrors services/core/config/vault.go: overrides secret env vars from Vault
KV v2 when VAULT_ENABLED=true, and is a no-op in local/dev setups that never
set that flag (the default). When explicitly enabled, a failed read is
fatal — reaching this path means an operator intended to run on
Vault-managed secrets, and silently falling back to whatever default/env
value is already configured would defeat the point.

Must run (load_vault_secrets()) before Settings() is instantiated in
config.py, since pydantic-settings reads os.environ/.env at construction
time, not lazily.
"""
import os

import hvac

REQUIRED_KEYS = {
    "jwt_secret": "JWT_SECRET",
    "weaviate_api_key": "WEAVIATE_API_KEY",
}

# Only needed by the high-risk modules (off in production today).
OPTIONAL_KEYS = {
    "database_url": "DATABASE_URL",
    "rabbitmq_url": "RABBITMQ_URL",
    "redis_url": "REDIS_URL",
    "neo4j_password": "NEO4J_PASSWORD",
    "openai_api_key": "OPENAI_API_KEY",
    "anthropic_api_key": "ANTHROPIC_API_KEY",
    "aws_access_key_id": "AWS_ACCESS_KEY_ID",
    "aws_secret_access_key": "AWS_SECRET_ACCESS_KEY",
}


def load_vault_secrets() -> None:
    if os.getenv("VAULT_ENABLED", "false").lower() != "true":
        return

    addr = os.getenv("VAULT_ADDR", "http://localhost:8200")
    token = os.getenv("VAULT_TOKEN", "zu-vault-dev-token")
    token_file = os.getenv("VAULT_TOKEN_FILE")
    if token_file:
        with open(token_file) as fh:
            token = fh.read().strip()
    path = os.getenv("VAULT_SECRET_PATH", "secret/data/zu360/ai")

    client = hvac.Client(url=addr, token=token)
    mount_point, _, secret_path = path.partition("/data/")
    if secret_path:
        response = client.secrets.kv.v2.read_secret_version(path=secret_path, mount_point=mount_point)
        data = (response or {}).get("data", {}).get("data")
    else:
        response = client.secrets.kv.v1.read_secret(path=path)
        data = (response or {}).get("data")

    if not data:
        raise RuntimeError(f"vault secret at {path} not found")

    missing = [key for key in REQUIRED_KEYS if not data.get(key)]
    if missing:
        raise RuntimeError(f"vault secret at {path} is missing required keys: {missing}")

    for vault_key, env_key in REQUIRED_KEYS.items():
        os.environ[env_key] = data[vault_key]
    for vault_key, env_key in OPTIONAL_KEYS.items():
        value = data.get(vault_key)
        if value:
            os.environ[env_key] = value
