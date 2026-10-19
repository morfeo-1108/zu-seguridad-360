from fastapi import APIRouter, HTTPException, Depends
from fastapi.responses import Response
from pydantic import BaseModel
from typing import Dict, Any, List
import uuid
from datetime import datetime

from app.services.pdf_generator import DocumentGenerator

router = APIRouter()
gen = DocumentGenerator()


class MonthlyReportRequest(BaseModel):
    company_name: str
    nit: str
    period: str  # "Abril 2026"
    employee_count: int
    if_value: float
    is_value: float
    ili_value: float
    incident_count: int
    compliance_pct: float
    trainings_completed: int
    exams_completed: int
    legal_rep: str
    sst_name: str


@router.post("/monthly/generate")
async def generate_monthly_report(req: MonthlyReportRequest):
    """Genera informe mensual SST con indicadores IF, IS, ILI."""
    doc_id = str(uuid.uuid4())
    data = req.dict()
    data["title"] = f"Informe SST Mensual — {req.period}"
    data["generated_at"] = datetime.now().strftime("%d/%m/%Y")

    pdf_bytes = gen.generate("informe_indicadores", data, doc_id)
    return Response(
        content=pdf_bytes,
        media_type="application/pdf",
        headers={"Content-Disposition": f"attachment; filename=informe-sst-{req.period.replace(' ','-')}.pdf"},
    )


@router.post("/res0312/generate")
async def generate_res0312_report(data: Dict[str, Any]):
    """Genera checklist de cumplimiento Res. 0312/2019 con estado por estándar."""
    doc_id = str(uuid.uuid4())
    pdf_bytes = gen.generate("politica_sst", {"title": "Evaluación Res. 0312/2019", **data}, doc_id)
    return Response(content=pdf_bytes, media_type="application/pdf")
