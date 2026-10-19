# Integración segura de `zu360-skills-v3`

## Estado actual

El ZIP contiene 20 `SKILL.md` de dominio más un `CLAUDE.md`. Ya se conservaron copias fuente inalteradas en `docs/skills/source-v3/`. Por ahora son **contenido de referencia no aprobado para ejecución**. No deben insertarse completos en un prompt de producción ni sus instrucciones deben ejecutarse automáticamente. El paquete sí está integrado al workspace y a su proceso de revisión; todavía no está conectado al registry como herramienta ejecutable.

El backend existente `services/ai/app/skills/registry.py` no es un cargador de markdown: registra cinco herramientas HTTP (`create_incident`, `generate_document`, `schedule_drone_round`, `query_normativa`, `send_reminder`). Algunas llamadas no propagan el token y la política local de roles sólo modela administrador/profesional SST/empleador/trabajador. La integración directa de los textos como tools o instrucciones activas no es segura ni corresponde al contrato actual.

## Contenido recibido

Dominios: alturas, auditoría, capacitaciones, contratistas, COPASST, correo corporativo, correos, documental, emergencias, empleados, empresas, EPP, equipos, exámenes, incidentes, indicadores, inspecciones, organizador, riesgos y master. `CLAUDE.md` es una guía de otra plataforma de agente y describe ubicaciones, herramientas y acciones que no existen en este backend.

## Hallazgos de revisión que exigen tratamiento antes de RAG

- **Claims normativos con fecha/obligatoriedad fija:** el skill de exámenes se basa en la Resolución 2346 de 2007 y afirma categorías/frecuencias, plazos y retención en forma categórica. El Ministerio indexa la Resolución 1843 de 2025 sobre evaluaciones médicas ocupacionales. Se necesita revisión profesional y versionamiento antes de ofrecer respuestas normativas. La mención oficial no basta para validar cada regla del documento.
- **Alturas y seguridad física:** el skill de alturas contiene umbral fijo de 1,5 m, referencias a normas internacionales, cargas numéricas, criterios de aprobación/rechazo y un “concepto final” que un modelo no debe emitir desde fotografías. Requiere revisión normativa vigente y validación técnica independiente; el agente sólo puede producir observaciones/borradores con estado `pending_review`.
- **Datos personales y médicos:** los prompts proponen datos de identificación, aptitud, restricciones, trabajadores y fotos. Deben minimizarse, autorizarse por rol/tenant, cifrarse y no enviarse a un modelo externo sin una política técnica de egress aprobada.
- **Acciones con efectos:** `zu360-correo-corp` describe configurar credenciales y enviar emails; `zu360-organizador` describe escanear/mover archivos locales. Ninguno de esos poderes existe en el backend del producto. Mantenerlos como instrucciones de diseño/referencia, no ejecución autónoma.
- **Firmas y certificaciones:** los documentos hablan de “firmar”, “certificar” y generar conceptos. El sistema no debe afirmar firma válida, licencia verificada ni concepto aprobado sólo por producir un PDF.
- **Puntuaciones no gobernadas:** el skill de empresas incluye pesos y umbrales de scoring no respaldados en el paquete por validación histórica/calibración; mostrarlos como fórmula configurable de borrador, nunca como evaluación científica establecida.
- **Instrucciones de interacción:** el `CLAUDE.md` ordena activar skills, generar archivos y operar sobre carpetas locales, pero no es una política de autorización del SaaS y no puede prevalecer sobre controles de servidor, permisos o confirmaciones.

## Arquitectura objetivo

1. Mantener cada skill como recurso versionado, con `skill_id`, versión, propietario, estado, jurisdicción, fuentes, fecha de revisión, clasificación de sensibilidad, capacidades, schemas y estado de aprobación.
2. Separar **conocimiento consultable** de **tools ejecutables**. Un skill de dominio puede aportar contexto, pero no concede acceso a datos ni permisos.
3. Retrieval con tenant/empresa, clasificación del documento y allowlist de fuentes; las normas públicas se guardan en colecciones globales versionadas y el contenido propio del cliente se aísla por tenant. El prompt y el filtro se forman en servidor, nunca a partir de filtros que el cliente pueda ampliar.
4. Para cada acción, el servidor deriva principal/tenant desde token y membership, verifica el entitlement del plan, valida objeto/relaciones y aplica RBAC. No confiar en `company_id`, `tenant_id`, URLs o destinatarios sugeridos por el modelo.
5. La salida de IA es un borrador. Operaciones de enviar, firmar, reportar, exportar, revisar concepto médico, cambiar estado o programar operación física requieren una tool estrecha, validación de negocio y confirmación humana según riesgo. La autorización se vuelve a comprobar al ejecutar, no se “recuerda” desde el prompt.
6. Registrar `skill_id`, versión, modelo, fuentes, actor, tenant, request ID, decisión, cambios sugeridos y aprobación; excluir o redactar secretos, historia clínica y prompts innecesarios de los logs.
7. Para aprendizaje por retroalimentación, guardar la corrección por separado y bajo tenant. No entrenar ni compartir feedback entre clientes por defecto; sólo incorporarlo a un conjunto global con base legítima documentada, anonimización revisada, consentimiento/política aplicable y revisión humana. En v1, el feedback mejora ejemplos/evaluaciones aislados; no altera automáticamente modelos o reglas productivas.

## Gates de publicación por skill

- [ ] Revisar texto y exactitud técnica/normativa con persona competente; incorporar fuentes oficiales y jurisdicción.
- [ ] Corregir afirmaciones absolutas, plazos, fórmulas y referencias derogadas/superadas.
- [ ] Definir datos de entrada/salida, sensibilidad, minimización, retención y permisos.
- [ ] Especificar si el skill es sólo conocimiento o si requiere tools y cuáles.
- [ ] Testear prompt injection, datos cruzados y denegación por rol/tenant.
- [ ] Para salida física, médica, legal, firma o notificación, configurar `pending_review` y gate humano obligatorio.
- [ ] Medir calidad en un set curado, registrando versión y criterio de aceptación.
- [ ] Publicar una versión inmutable, habilitada mediante feature flag y observable; conservar rollback.

## Prioridad

Se pueden preparar primero las skills informativas de empresas, gestión documental, incidentes e indicadores, tras verificar contenidos y fuentes. Exámenes médicos, alturas, EPP, equipos, emergencias, correo corporativo y organizador requieren gates reforzados o integraciones externas específicas y deben permanecer inactivos hasta completar revisión. Ninguna entrada de este ZIP se declara legalmente vigente, clínicamente validada ni técnicamente aprobada por el hecho de haberse incluido en el repositorio.


## Cambio aplicado en el hardening (17-oct-2026)

Aunque la implementación de esas cinco herramientas se conserva como código de referencia, se quitó su registro automático en el singleton: el registry inicia vacío. La descripción del estado inicial arriba es evidencia del riesgo encontrado, no el estado runtime final de esta copia. El registro sólo debe reactivarse después de agregar y probar autorización por empresa/objeto, auth servicio-a-servicio para cada destino, controles de confirmación, redacción de datos y auditoría.
