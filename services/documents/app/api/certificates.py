from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import Response
from pydantic import BaseModel
from typing import Dict, Any
import uuid

from app.services.pdf_generator import DocumentGenerator
from app.core.auth import get_current_user, require_company
from app.core.db import Session, get_company

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
    data = req.dict()
    async with Session() as session:
        company = await get_company(session, user["tenant_id"], require_company(user))
    if company is None:
        raise HTTPException(status_code=404, detail="Empresa no encontrada")
    # Identidad de la empresa siempre desde Core, nunca desde el body.
    data["company_name"] = company["name"]
    data["legal_rep"] = company["legal_rep"] or data["legal_rep"]
    pdf_bytes = gen.generate("certificado_capacitacion", data, doc_id)
    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={"Content-Disposition": f"attachment; filename=certificado-{req.folio[:8]}.pdf"},
    )

# La verificación pública por folio la hace Core (/v1/public/certificates/:folio/verify),
# contra su base de datos. Esta ruta antes respondía "válido" para cualquier folio.
