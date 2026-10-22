from typing import List

from pydantic_settings import BaseSettings

from app.core.vault import load_vault_secrets

load_vault_secrets()  # must run before Settings() reads os.environ below


class Settings(BaseSettings):
    # Server
    PORT: int = 4002
    ENV: str = "development"
    ALLOWED_ORIGINS: List[str] = ["http://localhost:3000"]

    # Database
    DATABASE_URL: str = "postgresql+asyncpg://zuadmin:zupassword@localhost:5432/zu_core"

    # Redis
    REDIS_URL: str = "redis://:zuredis123@localhost:6379/0"

    # RabbitMQ
    RABBITMQ_URL: str = "amqp://zurabbit:zurabbit123@localhost:5672/zu_vhost"

    # Weaviate
    WEAVIATE_URL: str = "http://localhost:8080"
    WEAVIATE_API_KEY: str = ""

    # Neo4j
    NEO4J_URI: str = "bolt://localhost:7687"
    NEO4J_USER: str = "neo4j"
    NEO4J_PASSWORD: str = "zuneo4j123"

    # AI Models
    OPENAI_API_KEY: str = ""
    ANTHROPIC_API_KEY: str = ""
    EMBEDDING_MODEL: str = "multilingual-e5-large"
    LLM_MODEL: str = "gpt-4o"

    # Pilot safety gates: APIs and high-risk features are closed by default.
    AI_SERVICE_API_ENABLED: bool = False
    AI_AUTONOMOUS_ENABLED: bool = False
    VISION_ENABLED: bool = False
    BIOMETRIC_ENABLED: bool = False
    SENSITIVE_AI_ENABLED: bool = False
    # Does not replace legal/privacy review; must remain false until the data
    # processing purpose, controller instructions, and vendor controls are approved.
    PRIVACY_PROCESSING_APPROVED: bool = False

    # AWS
    AWS_REGION: str = "us-east-1"
    AWS_ACCESS_KEY_ID: str = ""
    AWS_SECRET_ACCESS_KEY: str = ""
    S3_BUCKET: str = "zu-documents"

    # Service URLs (comunicación inter-servicios)
    CORE_URL: str = "http://core:4001"
    DOCUMENTS_URL: str = "http://documents:4003"
    NOTIFICATIONS_URL: str = "http://notifications:4006"
    DRONE_URL: str = "http://drone:4012"

    # JWT
    JWT_SECRET: str = "zu-super-secret-key-change-in-prod"

    class Config:
        env_file = ".env.dev"
        extra = "ignore"  # .env.example carries template vars (e.g. AI_URL) other
        # services use to reach this one; they aren't settings this service reads.


settings = Settings()
