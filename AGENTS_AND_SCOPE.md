# Agentes que dan vida al sistema + Scope por obra (v7)

## Principio: DOS clases de autonomía (no conflar)

| | Agentes de Razonamiento (8) | Agentes Deterministas (9) |
|--|------------------------------|----------------------------|
| Motor | LLM | Worker programado (sin LLM) |
| Para | juicio, lenguaje, clasificación | tareas predecibles y auditables |
| Riesgo | pueden errar → piden confirmación | reglas fijas → ejecutan solos |
| Ejemplo | "clasifica este incidente" | "cobra el día 1" |

Regla de oro: **nunca dejas a un LLM decidir si cobrar a un cliente**; nunca
pones reglas rígidas donde se necesita interpretar lenguaje o una imagen.

## Los 8 agentes de razonamiento (LLM)
1. **Orquestador** — rutea y compone respuestas multi-dominio.
2. **Cumplimiento** — plazos legales, brechas Res. 0312.
3. **Incidentes** — AT/EL, 5 Porqués, Bow-Tie, FURAT.
4. **Visión** — interpreta el dron por zona y trabajador, decide escalado. (autónomo)
5. **Normativo (RAG)** — 50+ normas con cita del artículo.
6. **Documental** — los 49 documentos parametrizados.
7. **Briefing** — briefing matutino. (autónomo)
8. **Cognición** — aprende patrones por empresa en Neo4j. (autónomo)

## Los 9 agentes deterministas (workers)
Facturación · Vencimientos · Reporte Mensual · Actas COPASST · Matriz de Riesgos ·
Informe Res. 0312 · Rondas de Dron · Renovación de Certificados · Monitor SLA

## Scope por obra/proyecto
Todo el dashboard se filtra por la OBRA activa (selector global en el header):
- Briefing, KPIs, zonas, trabajadores, cumplimiento EPP → de esa obra.
- Cada obra tiene sus propias zonas, cada zona su checklist normativo.
- El feed "Agentes en vivo" muestra qué está haciendo cada agente en esa obra.

## Endpoints de agentes
- `GET /v1/agents/roster` — roster completo (17 agentes, 2 clases)
- `GET /v1/agents/activity?project_id=` — feed de actividad por obra
- `POST /v1/agents/invoke` — el orquestador rutea y ejecuta
- `POST /v1/agents/confirm` — confirma acción destructiva pendiente
