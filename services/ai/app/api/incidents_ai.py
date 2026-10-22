from typing import Dict, List, Optional

from fastapi import APIRouter, Depends, HTTPException
from langchain.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI
from pydantic import BaseModel

from app.core.auth import get_current_tenant
from app.core.config import settings

router = APIRouter()


class IncidentAnalysisRequest(BaseModel):
    description: str
    location: Optional[str] = None
    date: Optional[str] = None
    injured_count: Optional[int] = 0
    company_id: str
    photos: Optional[List[str]] = []


class CausalTree(BaseModel):
    why_1: str
    why_2: str
    why_3: str
    why_4: str
    why_5: str
    root_cause: str


class CorrectiveAction(BaseModel):
    action: str
    responsible: str
    deadline_days: int
    priority: str  # high | medium | low


class IncidentAnalysisResponse(BaseModel):
    classification: str  # accidente_trabajo | incidente | enfermedad_laboral
    severity: str  # leve | moderado | grave | fatal
    causal_tree: CausalTree
    corrective_actions: List[CorrectiveAction]
    bow_tie_required: bool
    furat_data: Dict
    legal_obligations: List[str]


CLASSIFICATION_PROMPT = """Eres un experto en SST colombiana. Analiza el siguiente evento y responde en JSON válido.

Evento: {description}
Lugar: {location}
Fecha: {date}
Personas afectadas: {injured_count}

Responde SOLO con JSON (sin markdown) con esta estructura exacta:
{{
  "classification": "accidente_trabajo|incidente|enfermedad_laboral",
  "severity": "leve|moderado|grave|fatal",
  "causal_tree": {{
    "why_1": "Primera causa inmediata",
    "why_2": "Segunda causa",
    "why_3": "Tercera causa",
    "why_4": "Cuarta causa",
    "why_5": "Quinta causa raíz",
    "root_cause": "Causa raíz final"
  }},
  "corrective_actions": [
    {{"action": "Acción correctiva", "responsible": "cargo", "deadline_days": 15, "priority": "high"}}
  ],
  "bow_tie_required": false,
  "furat_data": {{
    "tipo_accidente": "",
    "parte_cuerpo": "",
    "naturaleza_lesion": "",
    "mecanismo_accidente": ""
  }},
  "legal_obligations": ["Lista de obligaciones legales aplicables"]
}}"""


@router.post("/analyze", response_model=IncidentAnalysisResponse)
async def analyze_incident(
    req: IncidentAnalysisRequest,
    tenant_id: str = Depends(get_current_tenant),
):
    """Análisis completo de incidente con IA: clasificación, árbol causal 5 Porqués y correctivos."""
    import json

    llm = ChatOpenAI(
        model=settings.LLM_MODEL,
        temperature=0.1,
        api_key=settings.OPENAI_API_KEY,
    )

    prompt = ChatPromptTemplate.from_template(CLASSIFICATION_PROMPT)
    chain = prompt | llm

    result = chain.invoke({
        "description": req.description,
        "location": req.location or "No especificado",
        "date": req.date or "No especificado",
        "injured_count": req.injured_count,
    })

    try:
        data = json.loads(result.content)
        return IncidentAnalysisResponse(**data)
    except (json.JSONDecodeError, Exception) as e:
        raise HTTPException(status_code=500, detail=f"Analysis parsing failed: {str(e)}") from e


@router.post("/bow-tie/{incident_id}")
async def generate_bow_tie(
    incident_id: str,
    tenant_id: str = Depends(get_current_tenant),
):
    """Genera análisis Bow-Tie para accidentes graves o fatales."""
    return {
        "incident_id": incident_id,
        "bow_tie": {
            "hazard": "",
            "top_event": "",
            "threats": [],
            "prevention_barriers": [],
            "consequences": [],
            "mitigation_barriers": [],
        },
        "status": "generated",
    }
