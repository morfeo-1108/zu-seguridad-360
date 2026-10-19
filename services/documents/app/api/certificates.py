from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import Response
from pydantic import BaseModel
from typing import Dict, Any
import uuid

from app.services.pdf_generator import DocumentGenerator
from app.core.auth import get_current_user

router = APIRouter()
gen = DocumentGenerator()


class CertificateRequest(BaseModel):
    employee_name: str
    document_number: str
    company_name: str
    training_title: str
    norm_ref: str
    duration_hours: str
    issued_at: str
    expires_at: str
    folio: str
    legal_rep: str
    sst_name: str


@router.post("/generate")
async def generate_certificate(req: CertificateRequest, user=Depends(get_current_user)):
    """Genera certificado PDF con QR de verificación."""
    doc_id = req.folio
    pdf_bytes = gen.generate("certificado_capacitacion", req.dict(), doc_id)
    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={"Content-Disposition": f"attachment; filename=certificado-{req.folio[:8]}.pdf"},
    )


@router.get("/verify/{folio}")
async def verify_certificate_public(folio: str):
    """Endpoint público para verificación QR de certificados."""
    # En producción: consultar BD para validar folio
    return {
        "folio": folio,
        "status": "valid",
        "message": "Certificado válido — verificación QR ZU Seguridad Integral 360",
    }
