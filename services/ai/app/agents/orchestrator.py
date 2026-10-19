"""
Orquestador multi-agente de ZU 360.

DECISIÓN DE DISEÑO (importante): NO es multi-agente por moda. Es multi-agente
porque el dominio tiene sub-problemas con objetivos, contexto y herramientas
DISTINTOS, y un solo prompt monolítico degradaría en todos. Cada agente es
especialista; el orquestador rutea por intención y coordina cuando una tarea
cruza dominios.

Patrón: Supervisor/Router + agentes especialistas + memoria compartida.
- ComplianceAgent: vigila plazos legales, genera documentos normativos.
- IncidentAgent: clasifica AT/EL, árbol de causas, FURAT, medidas correctivas.
- VisionAgent: interpreta detecciones del dron, decide alertas y escalado.
- NormativeAgent: RAG sobre 50+ normas, responde con cita de artículo.
- OrchestratorAgent: enruta y compone respuestas multi-dominio.
"""
from __future__ import annotations
from dataclasses import dataclass, field
from enum import Enum
from typing import Any
import json

from app.skills.registry import registry, SkillResult
from app.core.config import settings


class AgentRole(str, Enum):
    ORCHESTRATOR = "orchestrator"
    COMPLIANCE = "compliance"
    INCIDENT = "incident"
    VISION = "vision"
    NORMATIVE = "normative"


@dataclass
class AgentContext:
    tenant_id: str
    user_id: str
    role: str
    token: str = ""
    company_id: str | None = None
    conversation_history: list[dict] = field(default_factory=list)
    memory: dict[str, Any] = field(default_factory=dict)

    def to_skill_ctx(self) -> dict:
        return {"tenant_id": self.tenant_id, "user_id": self.user_id,
                "role": self.role, "token": self.token, "company_id": self.company_id}


# ── Definición de cada agente especialista ───────────────────
AGENT_SYSTEM_PROMPTS = {
    AgentRole.ORCHESTRATOR: (
        "Eres el orquestador de ZU Seguridad 360, plataforma SST colombiana. "
        "Tu trabajo es entender la intención del profesional SST y enrutar a el/los "
        "agentes especialistas correctos, luego componer una respuesta coherente. "
        "Si la tarea cruza dominios (ej: 'analiza el incidente y genera el FURAT y "
        "notifica a la ARL'), coordinas la secuencia. Nunca inventas datos legales."
    ),
    AgentRole.COMPLIANCE: (
        "Eres el agente de apoyo normativo. No determines cumplimiento ni plazos sin "
        "fuente oficial vigente comprobada para la fecha y actividad. Cita la fuente "
        "y su alcance; si no puedes verificarla, marca el punto como pendiente y "
        "solicita revisión de un profesional SST competente. No inventes referencias."
    ),
    AgentRole.INCIDENT: (
        "Eres el agente de investigación de incidentes. Clasificas AT/Incidente/EL, "
        "construyes árbol de causas (5 Porqués), aplicas Bow-Tie en eventos graves y "
        "propones medidas correctivas jerarquizadas. Generas FURAT con datos verificados."
    ),
    AgentRole.VISION: (
        "Eres el agente de visión. Interpretas detecciones del dron (EPP, caídas, fuego, "
        "posturas) y decides severidad y escalado. Una violación de EPP en zona de alto "
        "riesgo se escala de inmediato; una postura subóptima genera recomendación."
    ),
    AgentRole.NORMATIVE: (
        "Eres el agente normativo. Respondes consultas sobre las 50+ normas SST "
        "colombianas usando RAG, SIEMPRE con cita del artículo exacto. Si no hay fuente, "
        "lo dices explícitamente; jamás inventas contenido legal."
    ),
}


# Reglas de ruteo por palabras clave de intención (capa rápida antes del LLM router)
ROUTING_HINTS = {
    AgentRole.INCIDENT: ["incidente", "accidente", "furat", "lesión", "causas", "investiga"],
    AgentRole.COMPLIANCE: ["vence", "vencimiento", "plazo", "cumplimiento", "0312", "auditoría", "documento", "acta", "certificado"],
    AgentRole.VISION: ["dron", "epp", "casco", "cámara", "detección", "ronda"],
    AgentRole.NORMATIVE: ["qué dice", "norma", "resolución", "decreto", "ley", "artículo", "gtc"],
}


class Orchestrator:
    """Punto de entrada. Enruta, ejecuta y compone."""

    def __init__(self, llm_client=None):
        self.llm = llm_client  # cliente LLM inyectado (OpenAI/Anthropic compatible)

    def route(self, message: str) -> AgentRole:
        """Ruteo heurístico rápido; el LLM puede refinar en casos ambiguos."""
        msg = message.lower()
        scores = {role: sum(1 for kw in hints if kw in msg) for role, hints in ROUTING_HINTS.items()}
        best = max(scores, key=scores.get)
        return best if scores[best] > 0 else AgentRole.NORMATIVE

    async def handle(self, message: str, ctx: AgentContext) -> dict:
        """
        Flujo completo:
        1. Rutea a agente especialista.
        2. El agente decide si necesita skills (tool-calling).
        3. Ejecuta skills (con confirmación si son destructivas).
        4. Compone respuesta final con trazabilidad.
        """
        target = self.route(message)
        system_prompt = AGENT_SYSTEM_PROMPTS[target]
        tools = registry.tool_specs(ctx.role)

        # Llamada al LLM con tools disponibles
        llm_response = await self._call_llm(system_prompt, message, ctx, tools)

        executed_skills = []
        pending_confirmations = []

        # Procesar tool calls que el agente decidió usar
        for tool_call in llm_response.get("tool_calls", []):
            result = await registry.execute(
                tool_call["name"], tool_call["args"], ctx.to_skill_ctx(),
                confirmed=tool_call.get("confirmed", False),
            )
            if result.requires_confirmation:
                pending_confirmations.append({"skill": tool_call["name"], "args": tool_call["args"], "message": result.data.get("message")})
            else:
                executed_skills.append({"skill": tool_call["name"], "ok": result.ok, "data": result.data, "error": result.error})

        return {
            "agent": target.value,
            "response": llm_response.get("text", ""),
            "executed_skills": executed_skills,
            "pending_confirmations": pending_confirmations,
            "sources": llm_response.get("sources", []),
        }

    async def _call_llm(self, system: str, message: str, ctx: AgentContext, tools: list[dict]) -> dict:
        """
        Invoca el LLM con function calling. En producción usa el cliente real.
        Aquí dejamos la integración lista para conectar (OpenAI/Anthropic/Claude API).
        """
        if self.llm is None:
            # Modo desarrollo: respuesta estructurada de ejemplo con el contrato correcto
            return {
                "text": f"[{system[:40]}...] Procesando: '{message}'. "
                        f"Tools disponibles: {len(tools)}. Conectar LLM real en producción.",
                "tool_calls": [],
                "sources": [],
            }
        # Producción:
        return await self.llm.complete(system=system, user=message, tools=tools,
                                       history=ctx.conversation_history)


orchestrator = Orchestrator()
