import hashlib
import uuid
from typing import Any, Dict, Optional

from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import Response
from pydantic import BaseModel
from sqlalchemy import func, select

from app.core.auth import get_current_user, require_company
from app.core.db import GeneratedDocument, Session, get_company
from app.services.pdf_generator import TEMPLATES, DocumentGenerator

router = APIRouter()
generator = DocumentGenerator()

# Fields that identify the company on the document. They always come from
# Core for the session's active company, never from the request body.
COMPANY_FIELDS = {"company_name": "name", "nit": "nit", "legal_rep": "legal_rep", "city": "city"}


class GenerateRequest(BaseModel):
    template_key: str
    data: Dict[str, Any] = {}
    format: str = "pdf"


class DocumentMeta(BaseModel):
    id: str
    template_key: str
    title: str
    version: int
    sha256: str
    size_bytes: int
    created_by: str
    created_at: Optional[str] = None
    download_url: str


def _meta(d: GeneratedDocument) -> DocumentMeta:
    return DocumentMeta(id=d.id, template_key=d.template_key, title=d.title, version=d.version, sha256=d.sha256,
                        size_bytes=d.size_bytes, created_by=d.created_by,
                        created_at=d.created_at.isoformat() if d.created_at else None,
                        download_url=f"/v1/documents/{d.id}/download")


async def _render(req: GenerateRequest, user: dict, session, doc_id: str) -> tuple[bytes, dict]:
    if req.format != "pdf":
        raise HTTPException(status_code=422, detail="Solo se genera PDF por ahora")
    if req.template_key not in TEMPLATES:
        raise HTTPException(status_code=422, detail=f"Plantilla desconocida: {req.template_key}")
    company_id = require_company(user)
    company = await get_company(session, user["tenant_id"], company_id)
    if company is None:
        raise HTTPException(status_code=404, detail="Empresa no encontrada")
    data = dict(req.data)
    for field, col in COMPANY_FIELDS.items():
        data[field] = company[col]
    try:
        return generator.generate(req.template_key, data, doc_id), data
    except (KeyError, ValueError, TypeError) as e:
        raise HTTPException(status_code=422, detail=f"Datos incompletos para la plantilla: {e}")


@router.get("/templates")
async def list_templates(user=Depends(get_current_user)):
    """Plantillas disponibles. `kind` dice la verdad sobre cada una: "estructurada"
    tiene contenido y formato propios; "generica" es título + el texto que se envíe."""
    return {"templates": [{"key": k, **v} for k, v in TEMPLATES.items()]}


@router.post("/generate", response_model=DocumentMeta, status_code=201)
async def generate_document(req: GenerateRequest, user=Depends(get_current_user)):
    """Genera el PDF y lo guarda con versión por (empresa, plantilla), autor y SHA-256."""
    doc_id = str(uuid.uuid4())
    async with Session() as session:
        async with session.begin():
            pdf, _ = await _render(req, user, session, doc_id)
            company_id = user["company_id"]
            # Serializes concurrent generations of the same template for the company.
            await session.execute(select(func.pg_advisory_xact_lock(func.hashtext(f"{company_id}:{req.template_key}"))))
            current = await session.scalar(select(func.max(GeneratedDocument.version)).where(
                GeneratedDocument.company_id == company_id, GeneratedDocument.template_key == req.template_key))
            doc = GeneratedDocument(
                id=doc_id, tenant_id=user["tenant_id"], company_id=company_id, template_key=req.template_key,
                title=TEMPLATES[req.template_key]["title"], version=(current or 0) + 1,
                sha256=hashlib.sha256(pdf).hexdigest(), size_bytes=len(pdf), content=pdf, created_by=user["user_id"])
            session.add(doc)
        await session.refresh(doc)
        return _meta(doc)


@router.post("/generate/inline")
async def generate_document_inline(req: GenerateRequest, user=Depends(get_current_user)):
    """Vista previa: devuelve el PDF sin guardarlo."""
    async with Session() as session:
        pdf, _ = await _render(req, user, session, str(uuid.uuid4()))
    return Response(content=pdf, media_type="application/pdf",
                    headers={"Content-Disposition": "inline; filename=vista-previa.pdf"})


@router.get("")
async def list_documents(user=Depends(get_current_user)):
    company_id = require_company(user)
    async with Session() as session:
        rows = (await session.scalars(select(GeneratedDocument).where(
            GeneratedDocument.tenant_id == user["tenant_id"], GeneratedDocument.company_id == company_id,
        ).order_by(GeneratedDocument.created_at.desc()).limit(200))).all()
    return {"data": [_meta(d) for d in rows], "total": len(rows)}


@router.get("/{doc_id}/download")
async def download_document(doc_id: str, user=Depends(get_current_user)):
    company_id = require_company(user)
    try:
        uuid.UUID(doc_id)
    except ValueError:
        raise HTTPException(status_code=404, detail="Documento no encontrado")
    async with Session() as session:
        doc = await session.scalar(select(GeneratedDocument).where(
            GeneratedDocument.id == doc_id, GeneratedDocument.tenant_id == user["tenant_id"],
            GeneratedDocument.company_id == company_id))
    if doc is None:
        raise HTTPException(status_code=404, detail="Documento no encontrado")
    return Response(content=doc.content, media_type="application/pdf", headers={
        "Content-Disposition": f'attachment; filename="{doc.template_key}-v{doc.version}.pdf"',
        "X-Content-SHA256": doc.sha256,
    })
