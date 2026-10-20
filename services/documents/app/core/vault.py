"""Vault secret loading for the Documents service.

Mirrors services/core/config/vault.go and services/ai/app/core/vault.py:
overrides secret env vars from Vault KV v2 when VAULT_ENABLED=true, no-op
otherwise. When explicitly enabled, a failed read is fatal.

Must run (load_vault_secrets()) before Settings() is instantiated in
config.py, since pydantic-settings reads os.environ/.env at construction
time, not lazily.
"""
import os

import hvac

REQUIRED_KEYS = {
    "database_url": "DATABASE_URL",
    "rabbitmq_url": "RABBITMQ_URL",
    "jwt_secret": "JWT_SECRET",
}

OPTIONAL_KEYS = {
    "aws_access_key_id": "AWS_ACCESS_KEY_ID",
    "aws_secret_access_key": "AWS_SECRET_ACCESS_KEY",
}


def load_vault_secrets() -> None:
    if os.getenv("VAULT_ENABLED", "false").lower() != "true":
        return

    addr = os.getenv("VAULT_ADDR", "http://localhost:8200")
    token = os.getenv("VAULT_TOKEN", "zu-vault-dev-token")
    path = os.getenv("VAULT_SECRET_PATH", "secret/data/zu360/documents")

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
