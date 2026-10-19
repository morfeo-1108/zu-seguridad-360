# ZU 360 — Arquitectura Avanzada (v5)

## Decisiones de ingeniería: qué se añadió y POR QUÉ

### ✅ Redes neuronales — SÍ, donde crean valor
- **Visión del dron** (`ai/app/vision/inference.py`): detección de objetos (YOLO/ONNX)
  para EPP, caídas, fuego, humo, pose. Las reglas no resuelven visión — aquí la red es
  indispensable. Escalado automático por contexto (sin arnés en altura = crítico).
- **RAG normativo**: embeddings (multilingual-e5) + reranker (BGE) — ya eran neurales.
- **OCR médico**: LayoutLM para formatos variables de IPS.
- **Riesgo por CIIU**: scaffold ML listo; requiere N>50 empresas para entrenar (honesto:
  no se finge un modelo entrenado sin datos).

### ✅ Skills ejecutables — SÍ
`ai/app/skills/registry.py`: capacidades atómicas que los agentes invocan para producir
efectos reales (crear incidente, generar documento, programar dron, consultar norma,
enviar recordatorio). Tool-calling estándar con validación de permisos por rol y
confirmación obligatoria en acciones destructivas.

### ✅ Multi-agente — SÍ, porque el dominio lo justifica
`ai/app/agents/orchestrator.py`: NO es multi-agente por moda. Hay sub-problemas con
objetivos y herramientas distintas (cumplimiento, incidentes, visión, normativa) que un
prompt monolítico degradaría. Patrón Supervisor/Router + 5 especialistas + memoria.

### ✅ Core óptimo — resiliencia añadida
- Circuit breaker (`core/internal/resilience`): evita fallos en cascada.
- Idempotencia + Outbox (billing): cero doble-cobro, cero eventos perdidos.
- DLQ RabbitMQ: mensajes fallidos no se evaporan.
- Habeas Data (Ley 1581): consentimiento + supresión + anonimización con legal-hold.

## Lo que deliberadamente NO se añadió (sería over-engineering)
- Service mesh (Istio): innecesario para 12 servicios; Kong + breakers bastan.
- Event sourcing total: el outbox cubre la consistencia sin la complejidad operativa.
- Fine-tuning de LLM propio: el RAG + prompting especializado es superior en costo/calidad
  para este volumen.
