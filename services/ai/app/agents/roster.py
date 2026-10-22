"""
Roster maestro de agentes de ZU 360 — la fuente de verdad de "qué da vida al sistema".

PRINCIPIO DE DISEÑO CLAVE: hay DOS clases de autonomía y NO deben conflarse.

  1. AGENTES DE RAZONAMIENTO (LLM): para tareas que requieren juicio, lenguaje,
     clasificación o interpretación. Pueden equivocarse, así que NUNCA ejecutan
     acciones irreversibles sin confirmación humana.

  2. AGENTES DETERMINISTAS (workers programados, sin LLM): para tareas que DEBEN
     ser predecibles y auditables. Jamás dejarías a un LLM decidir si cobrar a un
     cliente o si un certificado venció. Eso es código determinista.

El dashboard ultra-avanzado muestra ambos vivos, pero etiquetados por clase, para
que la profesional SST entienda qué está razonando y qué está ejecutando reglas.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum


class AgentClass(str, Enum):
    REASONING = "reasoning"      # LLM
    DETERMINISTIC = "deterministic"  # worker programado


@dataclass
class AgentSpec:
    key: str
    name: str
    cls: AgentClass
    role: str                    # qué hace, en una línea
    triggers: list[str]          # qué lo activa
    skills: list[str] = field(default_factory=list)   # qué puede ejecutar
    autonomy: str = "asistido"   # "asistido" (pide confirmación) | "autonomo"
    norm_refs: list[str] = field(default_factory=list)


# ── AGENTES DE RAZONAMIENTO (LLM) ────────────────────────────
REASONING_AGENTS = [
    AgentSpec(
        key="orchestrator", name="Orquestador",
        cls=AgentClass.REASONING,
        role="Entiende la intención y coordina a los especialistas; compone respuestas multi-dominio.",
        triggers=["mensaje de la profesional", "tarea que cruza dominios"],
        skills=["routing", "composición"], autonomy="asistido",
    ),
    AgentSpec(
        key="compliance", name="Agente de Cumplimiento",
        cls=AgentClass.REASONING,
        role="Detecta brechas frente a Res. 0312, vigila plazos legales y propone acciones.",
        triggers=["vencimiento próximo", "auditoría", "consulta de cumplimiento"],
        skills=["query_normativa", "generate_document"], autonomy="asistido",
        norm_refs=["Res. 0312/2019", "Decreto 1072/2015"],
    ),
    AgentSpec(
        key="incident", name="Agente de Incidentes",
        cls=AgentClass.REASONING,
        role="Clasifica AT/Incidente/EL, árbol de causas (5 Porqués), Bow-Tie y medidas correctivas.",
        triggers=["reporte de incidente", "evento desde app móvil o dron"],
        skills=["create_incident", "generate_document"], autonomy="asistido",
        norm_refs=["Ley 1562/2012"],
    ),
    AgentSpec(
        key="vision", name="Agente de Visión",
        cls=AgentClass.REASONING,
        role="Interpreta detecciones del dron por zona y trabajador; decide severidad y escalado.",
        triggers=["frame del dron", "alerta de visión IA"],
        skills=["schedule_drone_round", "send_reminder"], autonomy="autonomo",
        norm_refs=["Pendiente de validación en fuente oficial vigente"],
    ),
    AgentSpec(
        key="normative", name="Agente Normativo (RAG)",
        cls=AgentClass.REASONING,
        role="Responde sobre 50+ normas SST con cita del artículo exacto; jamás inventa contenido legal.",
        triggers=["consulta normativa", "duda legal"],
        skills=["query_normativa"], autonomy="asistido",
        norm_refs=["50+ normas SST Colombia"],
    ),
    AgentSpec(
        key="document", name="Agente Documental",
        cls=AgentClass.REASONING,
        role="Genera los 49 documentos parametrizados por sector CIIU, tamaño e historial.",
        triggers=["solicitud de documento", "onboarding de empresa"],
        skills=["generate_document"], autonomy="asistido",
        norm_refs=["Ley 527/1999"],
    ),
    AgentSpec(
        key="briefing", name="Agente de Briefing",
        cls=AgentClass.REASONING,
        role="Compone el briefing matutino: prioriza el portafolio y estima la carga del día.",
        triggers=["6:00 AM diario"],
        skills=["query_normativa"], autonomy="autonomo",
    ),
    AgentSpec(
        key="cognition", name="Agente de Cognición",
        cls=AgentClass.REASONING,
        role="Aprende patrones por empresa (Neo4j), genera insights predictivos y ajusta umbrales.",
        triggers=["cada interacción", "cierre de incidente"],
        skills=[], autonomy="autonomo",
    ),
]

# ── AGENTES DETERMINISTAS (workers, sin LLM) ─────────────────
DETERMINISTIC_AGENTS = [
    AgentSpec("monthly_report", "Worker Reporte Mensual", AgentClass.DETERMINISTIC,
              "Genera y envía el informe de indicadores el día 5 de cada mes.", ["cron: día 5, 7AM"], autonomy="autonomo"),
    AgentSpec("expiration", "Worker Vencimientos", AgentClass.DETERMINISTIC,
              "Detecta y notifica vencimientos 30/15/7/0 días antes.", ["cron: diario 6AM"], autonomy="autonomo"),
    AgentSpec("copasst", "Worker Actas COPASST", AgentClass.DETERMINISTIC,
              "Genera el acta mensual del COPASST con orden del día.", ["cron: día 1"], autonomy="autonomo"),
    AgentSpec("risk_matrix", "Worker Matriz de Riesgos", AgentClass.DETERMINISTIC,
              "Actualiza el NR de peligros tras nuevos incidentes.", ["cron: diario 2AM"], autonomy="autonomo"),
    AgentSpec("res0312", "Worker Informe Res. 0312", AgentClass.DETERMINISTIC,
              "Genera el informe de estándares mínimos mensual.", ["cron: día 28"], autonomy="autonomo",
              norm_refs=["Res. 0312/2019"]),
    AgentSpec("drone_schedule", "Worker Rondas de Dron", AgentClass.DETERMINISTIC,
              "Programa y ejecuta rondas de inspección por calendario.", ["cron: cada 6h"], autonomy="autonomo"),
    AgentSpec("cert_renewal", "Worker Renovación Certificados", AgentClass.DETERMINISTIC,
              "Inicia el flujo de renovación de certificados próximos a vencer.", ["cron: diario 9AM"], autonomy="autonomo"),
    AgentSpec("billing", "Worker Facturación", AgentClass.DETERMINISTIC,
              "Cobra, emite factura DIAN y suspende por mora (idempotente).", ["cron: día 1, 6AM"], autonomy="autonomo"),
    AgentSpec("sla_monitor", "Worker SLA", AgentClass.DETERMINISTIC,
              "Vigila incumplimientos de SLA por tipo de evento.", ["cron: cada minuto"], autonomy="autonomo"),
]

ALL_AGENTS = REASONING_AGENTS + DETERMINISTIC_AGENTS


def roster_summary() -> dict:
    return {
        "total": len(ALL_AGENTS),
        "reasoning": len(REASONING_AGENTS),
        "deterministic": len(DETERMINISTIC_AGENTS),
        "agents": [
            {
                "key": a.key, "name": a.name, "class": a.cls.value, "role": a.role,
                "triggers": a.triggers, "skills": a.skills, "autonomy": a.autonomy,
                "norm_refs": a.norm_refs,
            }
            for a in ALL_AGENTS
        ],
    }
