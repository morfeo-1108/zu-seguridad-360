"""Persistence for generated documents (schema `documents`, owned by this
service's role) and read-only lookups of the caller's company in Core
(`public.companies`, SELECT granted on a few columns)."""
from sqlalchemy import Column, DateTime, Integer, LargeBinary, String, UniqueConstraint, func, text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import declarative_base

from app.core.config import settings

Base = declarative_base()


class GeneratedDocument(Base):
    __tablename__ = "generated_documents"
    __table_args__ = (UniqueConstraint("company_id", "template_key", "version", name="uq_doc_version"),
                      {"schema": "documents"})

    id = Column(UUID(as_uuid=False), primary_key=True, server_default=text("gen_random_uuid()"))
    tenant_id = Column(UUID(as_uuid=False), nullable=False, index=True)
    company_id = Column(UUID(as_uuid=False), nullable=False, index=True)
    template_key = Column(String(64), nullable=False)
    title = Column(String(255), nullable=False)
    version = Column(Integer, nullable=False)
    sha256 = Column(String(64), nullable=False)
    size_bytes = Column(Integer, nullable=False)
    content = Column(LargeBinary, nullable=False)
    created_by = Column(UUID(as_uuid=False), nullable=False)
    created_at = Column(DateTime(timezone=True), nullable=False, server_default=func.now())


def _asyncpg_url(url: str) -> str:
    # Vault stores one libpq-style URL for every service; asyncpg takes ssl= instead of sslmode=.
    url = url.replace("postgres://", "postgresql+asyncpg://", 1).replace("postgresql://", "postgresql+asyncpg://", 1)
    return url.replace("sslmode=require", "ssl=require")


engine = create_async_engine(_asyncpg_url(settings.DATABASE_URL), pool_pre_ping=True)
Session = async_sessionmaker(engine, expire_on_commit=False, class_=AsyncSession)


async def init_db() -> None:
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)


async def get_company(session: AsyncSession, tenant_id: str, company_id: str) -> dict | None:
    row = (await session.execute(text(
        "SELECT id::text, name, coalesce(nit,'') AS nit, coalesce(legal_rep,'') AS legal_rep, coalesce(city,'') AS city "
        "FROM public.companies WHERE id = :c AND tenant_id = :t AND deleted_at IS NULL"),
        {"c": company_id, "t": tenant_id})).mappings().first()
    return dict(row) if row else None
