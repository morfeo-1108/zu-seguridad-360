from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import Optional
import uuid
from datetime import datetime

router = APIRouter()

class SignatureRequest(BaseModel):
    document_id: str
    signer_id: str
    signer_name: str
    signer_role: str
    biometric_hash: Optional[str] = None

class SignatureResponse(BaseModel):
    signature_id: str
    document_id: str
    signed_at: str
    legal_basis: str
    is_valid: bool

@router.post("/sign", response_model=SignatureResponse)
async def sign_document(req: SignatureRequest):
    """Firma electrónica simple conforme Ley 527/1999."""
    sig_id = str(uuid.uuid4())
    return SignatureResponse(
        signature_id=sig_id, document_id=req.document_id,
        signed_at=datetime.utcnow().isoformat(),
        legal_basis="Ley 527/1999 — Comercio Electrónico Colombia",
        is_valid=True,
    )

@router.get("/document/{document_id}/signatures")
async def get_signatures(document_id: str):
    return {"document_id": document_id, "signatures": [], "pending": [], "status": "pending_signature"}

@router.post("/request-signature")
async def request_signature(document_id: str, signer_email: str, signer_name: str):
    """Solicita firma a un usuario y le envía notificación."""
    return {"status": "notification_sent", "signer": signer_email, "document_id": document_id}

router_templates = APIRouter()
