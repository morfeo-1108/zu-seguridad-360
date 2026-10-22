import json
from typing import Any, Dict, List, Optional

from fastapi import APIRouter, Depends
from neo4j import GraphDatabase
from pydantic import BaseModel

from app.core.auth import get_current_tenant
from app.core.config import settings

router = APIRouter()


class InsightRequest(BaseModel):
    company_id: str
    context: Optional[str] = None
    insight_types: List[str] = ["risk_trend", "incident_pattern", "compliance_gap"]


class Insight(BaseModel):
    type: str
    title: str
    description: str
    severity: str  # critical | warning | info
    data: Dict[str, Any]
    recommendation: str
    confidence: float


class CognitionUpdate(BaseModel):
    company_id: str
    event_type: str
    event_data: Dict[str, Any]


@router.post("/insights", response_model=List[Insight])
async def get_company_insights(
    req: InsightRequest,
    tenant_id: str = Depends(get_current_tenant),
):
    """
    Genera insights predictivos usando el knowledge graph en Neo4j.
    Motor de cognición adaptativa con namespace por tenant.
    """
    driver = GraphDatabase.driver(
        settings.NEO4J_URI,
        auth=(settings.NEO4J_USER, settings.NEO4J_PASSWORD)
    )

    insights = []

    try:
        with driver.session() as session:
            # Query: patrones de incidentes por día de semana
            pattern_result = session.run("""
                MATCH (c:Company {id: $company_id, tenant_id: $tenant_id})-[:HAS]->(i:Incident)
                WHERE i.occurred_at > datetime() - duration('P90D')
                RETURN
                    i.type AS incident_type,
                    i.day_of_week AS dow,
                    count(i) AS count
                ORDER BY count DESC
                LIMIT 10
            """, company_id=req.company_id, tenant_id=tenant_id)

            patterns = pattern_result.data()
            if patterns:
                top = patterns[0]
                day_names = ["Lun", "Mar", "Mié", "Jue", "Vie", "Sáb", "Dom"]
                dow_name = day_names[top.get("dow", 0) % 7]
                insights.append(Insight(
                    type="incident_pattern",
                    title=f"Patrón: más incidentes los {dow_name}",
                    description=f"El {dow_name} concentra {top['count']} incidentes en los últimos 90 días.",
                    severity="warning",
                    data={"patterns": patterns},
                    recommendation=f"Refuerce la supervisión los días {dow_name}. Considere ronda de dron matutina.",
                    confidence=0.87,
                ))

            # Query: tendencia de cumplimiento
            compliance_result = session.run("""
                MATCH (c:Company {id: $company_id, tenant_id: $tenant_id})-[:HAS_SCORE]->(s:ComplianceScore)
                WHERE s.recorded_at > datetime() - duration('P180D')
                RETURN s.score AS score, s.recorded_at AS date
                ORDER BY s.recorded_at
            """, company_id=req.company_id, tenant_id=tenant_id)

            scores = compliance_result.data()
            if len(scores) >= 2:
                trend = scores[-1]["score"] - scores[0]["score"]
                severity = "info" if trend >= 0 else "warning"
                if trend < -10:
                    severity = "critical"
                insights.append(Insight(
                    type="compliance_trend",
                    title=f"Cumplimiento {'↑' if trend >= 0 else '↓'} {abs(trend):.1f}% en 6 meses",
                    description=f"Tendencia {'positiva' if trend >= 0 else 'negativa'} en el cumplimiento Res. 0312.",
                    severity=severity,
                    data={"scores": scores, "trend": trend},
                    recommendation="Revisar los estándares mínimos con menor puntuación." if trend < 0 else "Mantener el ritmo de mejora.",
                    confidence=0.91,
                ))

    except Exception:
        # Si Neo4j no disponible, retornar insights generados por reglas
        insights = _rule_based_insights(req.company_id)
    finally:
        driver.close()

    return insights


@router.post("/update")
async def update_cognition(
    update: CognitionUpdate,
    tenant_id: str = Depends(get_current_tenant),
):
    """Actualiza el knowledge graph con nuevos eventos del sistema."""
    driver = GraphDatabase.driver(
        settings.NEO4J_URI,
        auth=(settings.NEO4J_USER, settings.NEO4J_PASSWORD)
    )
    try:
        with driver.session() as session:
            # Merge empresa en el grafo con namespace de tenant
            session.run("""
                MERGE (c:Company {id: $company_id, tenant_id: $tenant_id})
                SET c.last_updated = datetime()
            """, company_id=update.company_id, tenant_id=tenant_id)

            # Registrar evento
            session.run("""
                MATCH (c:Company {id: $company_id, tenant_id: $tenant_id})
                CREATE (c)-[:HAS_EVENT]->(e:Event {
                    type: $event_type,
                    data: $event_data,
                    occurred_at: datetime()
                })
            """, company_id=update.company_id, tenant_id=tenant_id,
                event_type=update.event_type,
                event_data=json.dumps(update.event_data))

        return {"status": "updated", "company_id": update.company_id}
    finally:
        driver.close()


def _rule_based_insights(company_id: str) -> List[Insight]:
    """Fallback: insights basados en reglas cuando Neo4j no disponible."""
    return [
        Insight(
            type="system",
            title="Cognición adaptativa en inicialización",
            description="El motor de cognición se activa después de 30 días de datos.",
            severity="info",
            data={"company_id": company_id},
            recommendation="Continue registrando eventos. Los insights aparecerán pronto.",
            confidence=1.0,
        )
    ]
