"""API del sistema multi-agente: punto de entrada para el profesional SST."""
from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel
from typing import Any

from app.agents.orchestrator import orchestrator, AgentContext
from app.skills.registry import registry
from app.core.auth import get_current_user

router = APIRouter()


class AgentRequest(BaseModel):
    message: str
    company_id: str | None = None
    history: list[dict] = []


class ConfirmRequest(BaseModel):
    skill: str
    args: dict[str, Any]


@router.post("/invoke")
async def invoke_agent(req: AgentRequest, user=Depends(get_current_user)):
    """Punto de entrada principal: el orquestador rutea y ejecuta."""
    ctx = AgentContext(
        tenant_id=user["tenant_id"], user_id=user["user_id"], role=user["role"],
        token=user.get("token", ""), company_id=req.company_id,
        conversation_history=req.history,
    )
    result = await orchestrator.handle(req.message, ctx)
    return result


@router.post("/confirm")
async def confirm_action(req: ConfirmRequest, user=Depends(get_current_user)):
    """Confirma y ejecuta una acción destructiva que quedó pendiente."""
    ctx = {"tenant_id": user["tenant_id"], "user_id": user["user_id"],
           "role": user["role"], "token": user.get("token", "")}
    result = await registry.execute(req.skill, req.args, ctx, confirmed=True)
    if not result.ok:
        raise HTTPException(status_code=400, detail=result.error)
    return {"executed": True, "data": result.data}


@router.get("/skills")
async def list_skills(user=Depends(get_current_user)):
    """Lista las skills disponibles para el rol del usuario."""
    return {"skills": registry.tool_specs(user["role"])}


from app.agents.roster import roster_summary

@router.get("/roster")
async def get_roster(user=Depends(get_current_user)):
    """Devuelve el roster completo de agentes (razonamiento + deterministas)."""
    return roster_summary()


@router.get("/activity")
async def agent_activity(project_id: str = None, user=Depends(get_current_user)):
    """
    Actividad reciente de los agentes (feed en vivo del dashboard).
    En producción: lee de un stream/log de ejecuciones por tenant y proyecto.
    """
    # Estructura del contrato; el dashboard la consume en tiempo real
    return {
        "project_id": project_id,
        "events": [
            # Poblado por el event log real en producción
        ],
    }
