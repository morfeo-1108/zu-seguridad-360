from typing import Any, Dict, List, Optional

import boto3
from fastapi import APIRouter, Depends, File, HTTPException, UploadFile
from pydantic import BaseModel

from app.core.auth import get_current_tenant
from app.core.config import settings
from app.core.privacy_errors import ocr_failure_payload

router = APIRouter()


class OCRResult(BaseModel):
    employee_id: str
    exam_type: str
    ips_name: str
    doctor_name: Optional[str]
    aptitude_result: str  # apto | apto_con_restriccion | no_apto
    restrictions: List[str]
    performed_at: Optional[str]
    expires_at: Optional[str]
    confidence: float
    raw_fields: Dict[str, Any]
    requires_human_review: bool


APTITUDE_KEYWORDS = {
    "apto": ["apto", "apta", "aptitud favorable", "puede desempeñar"],
    "apto_con_restriccion": ["apto con restricción", "restricción", "limitación", "condicionado"],
    "no_apto": ["no apto", "no apta", "inapto", "no puede", "incapacitado"],
}

RESTRICTION_KEYWORDS = [
    "no levantar", "no cargar", "no alturas", "no conducir",
    "no turnos nocturnos", "no trabajo en frío", "no exposición a ruido",
    "no bipedestación prolongada", "no movimientos repetitivos",
]


@router.post("/medical-exam", response_model=OCRResult)
async def process_medical_exam(
    file: UploadFile = File(...),
    employee_id: str = "",
    exam_type: str = "periodico",
    tenant_id: str = Depends(get_current_tenant),
):
    """
    Procesa imagen/PDF de examen médico con AWS Textract + LayoutLM.
    Pipeline human-in-the-loop: toda aptitud requiere validación del profesional.
    La ruta permanece cerrada en el piloto; este código no afirma cumplimiento normativo.
    """
    content = await file.read()
    if len(content) > 10 * 1024 * 1024:  # 10MB max
        raise HTTPException(status_code=413, detail="Archivo demasiado grande (máx 10MB)")

    # AWS Textract para extracción estructurada
    extracted = await _textract_extract(content, file.content_type)

    # Análisis del texto extraído
    text = extracted.get("text", "").lower()
    aptitude = _detect_aptitude(text)
    restrictions = _detect_restrictions(text)
    confidence = extracted.get("confidence", 0.0)

    # Revisión humana obligatoria; no determina aptitud de forma autónoma.
    requires_review = True

    result = OCRResult(
        employee_id=employee_id,
        exam_type=exam_type,
        ips_name=extracted.get("ips_name", ""),
        doctor_name=extracted.get("doctor_name"),
        aptitude_result=aptitude,
        restrictions=restrictions,
        performed_at=extracted.get("date"),
        expires_at=_calculate_expiry(exam_type, extracted.get("date")),
        confidence=confidence,
        raw_fields=extracted.get("fields", {}),
        requires_human_review=requires_review,
    )

    return result


async def _textract_extract(content: bytes, content_type: str) -> Dict:
    """Extrae texto con AWS Textract + análisis de layout."""
    try:
        client = boto3.client("textract", region_name=settings.AWS_REGION)
        response = client.detect_document_text(Document={"Bytes": content})

        blocks = response.get("Blocks", [])
        text_blocks = [b["Text"] for b in blocks if b["BlockType"] == "LINE" and "Text" in b]
        full_text = "\n".join(text_blocks)

        avg_confidence = sum(b.get("Confidence", 0) for b in blocks) / max(len(blocks), 1)

        return {
            "text": full_text,
            "confidence": avg_confidence / 100,
            "fields": _extract_fields(full_text),
            "blocks": len(blocks),
        }
    except Exception:
        # Fallback sin filtrar mensajes de proveedor o contenido del documento.
        return ocr_failure_payload()


def _detect_aptitude(text: str) -> str:
    """Detecta resultado de aptitud en texto extraído."""
    text = text.lower()
    for aptitude, keywords in APTITUDE_KEYWORDS.items():
        for kw in keywords:
            if kw in text:
                return aptitude
    return "pendiente_revision"


def _detect_restrictions(text: str) -> List[str]:
    """Detecta restricciones médicas en el texto."""
    found = []
    text_lower = text.lower()
    for restriction in RESTRICTION_KEYWORDS:
        if restriction in text_lower:
            found.append(restriction)
    return found


def _extract_fields(text: str) -> Dict:
    """Extrae campos estructurados del texto del examen."""
    import re
    fields = {}
    patterns = {
        "nit_ips": r"NIT[:\s]+(\d{6,12})",
        "license_doctor": r"licencia[:\s]+(\d{4,8})",
        "date": r"fecha[:\s]+(\d{1,2}[/\-]\d{1,2}[/\-]\d{2,4})",
    }
    for key, pattern in patterns.items():
        match = re.search(pattern, text, re.IGNORECASE)
        if match:
            fields[key] = match.group(1)
    return fields


def _calculate_expiry(exam_type: str, date_str: Optional[str]) -> Optional[str]:
    """No infer an exam validity/retention period without an approved policy."""
    return None
