from typing import List

from pydantic_settings import BaseSettings

from app.core.vault import load_vault_secrets

load_vault_secrets()  # must run before Settings() reads os.environ below


class Settings(BaseSettings):
    PORT: int = 4003
    ENV: str = "development"
    ALLOWED_ORIGINS: List[str] = ["http://localhost:3000"]
    DATABASE_URL: str = "postgresql+asyncpg://zuadmin:zupassword@postgres:5432/zu_core"
    RABBITMQ_URL: str = "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"
    AWS_REGION: str = "us-east-1"
    S3_BUCKET: str = "zu-documents"
    CORE_URL: str = "http://core:4001"
    JWT_SECRET: str = "zu-super-secret-key-change-in-prod"

    class Config:
        env_file = ".env.dev"
        extra = "ignore"  # .env.example carries template vars (REDIS_*, JWT_SECRET,
        # AI_URL) this service's Settings model doesn't declare as fields.


settings = Settings()
