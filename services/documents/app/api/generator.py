from fastapi import APIRouter, HTTPException, Depends, BackgroundTasks
from fastapi.responses import Response
from pydantic import BaseModel
from typing import Dict, Any, Optional
import uuid

from app.services.pdf_generator import DocumentGenerator
from app.core.config import settings

router = APIRouter()
generator = DocumentGenerator()


class GenerateRequest(BaseModel):
    template_key: str
    data: Dict[str, Any]
    format: str = "pdf"
    async_mode: bool = False


class GenerateResponse(BaseModel):
    doc_id: str
    template_key: str
    format: str
    status: str
    download_url: Optional[str] = None


@router.post("/generate", response_model=GenerateResponse)
async def generate_document(req: GenerateRequest, background_tasks: BackgroundTasks):
    """Genera un documento SST en PDF/Word/Excel según template."""
    doc_id = str(uuid.uuid4())

    if req.format != "pdf":
        raise HTTPException(status_code=422, detail=f"Formato '{req.format}' — use PDF por ahora")

    try:
        pdf_bytes = generator.generate(req.template_key, req.data, doc_id)

        # En producción: subir a S3 y retornar URL firmada
        # Por ahora retornar inline para desarrollo
        return GenerateResponse(
            doc_id=doc_id,
            template_key=req.template_key,
            format="pdf",
            status="generated",
            download_url=f"/v1/documents/{doc_id}/download",
        )

    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error generando documento: {str(e)}")


@router.post("/generate/inline")
async def generate_document_inline(req: GenerateRequest):
    """Retorna el PDF directamente en la respuesta (para preview)."""
    doc_id = str(uuid.uuid4())
    try:
        pdf_bytes = generator.generate(req.template_key, req.data, doc_id)
        return Response(
            content=pdf_bytes,
            media_type="application/pdf",
            headers={"Content-Disposition": f"inline; filename=zu-doc-{doc_id[:8]}.pdf"},
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
