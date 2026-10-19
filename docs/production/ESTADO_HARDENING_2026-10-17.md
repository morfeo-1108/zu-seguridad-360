# Estado de hardening y decisión de salida — ZU Seguridad 360

**Corte:** 19 de octubre de 2026  
**Decisión:** **NO-GO para producción, cobro de licencias o datos reales de salud.** El código entregado en esta copia es una primera remediación defensiva, no una certificación ni una puesta en producción.

## Resumen ejecutivo

El producto puede llegar a diferenciarse como una plataforma de operación para profesionales SST que administran un portafolio de empresas. En el estado encontrado no es responsable venderlo como “cumple automáticamente” ni activar todas las funciones: faltan aislamiento verificable por empresa, build reproducible, controles comerciales de planes, validación legal completa, protección de datos sensibles y operación productiva. La IA/visión/drones podrían ser una ventaja, pero por ahora son una hipótesis técnica, no una ventaja probada ni habilitada.

En esta copia se implementaron barreras defensivas: JWT exige claims verificadas; company_id proviene del usuario en base de datos y se valida contra el tenant al login/refresh; CompanyScope rechaza cruce A→B. Mientras Gate 2 esté abierto, APIs laborales, incidentes/capacitaciones y OCR-RUT se cierran con 404 tras comprobar scope/rol, y un cruce A→B conserva 403. EmployeeService, IncidentService y TrainingService también fallan cerrados antes de DB/colas. Documents, OCR médico, biometría, exámenes, exportación y captura/validación de consentimiento permanecen deshabilitados; se eliminaron errores que exponían mensajes internos. Aun así, no hay cobertura E2E de objetos ni aprobación legal/privacidad/retención. No habilitar datos personales reales.

## Cambios aplicados en esta copia

- **Auth y configuración:** separación explícita `access`/`refresh`, rechazo de algoritmo distinto de HS256, `sub == user_id`, claims obligatorios y parseables, rechazo de refresh como access, reconsulta de usuario/estado/company al renovar, company_id firmado derivado del usuario y validado como compañía del tenant tanto en login como refresh. No se confía en `X-Tenant-ID`/`X-Company-ID`; guardias productivas fallan ante defaults/secretos débiles, TLS de DB desactivado o CORS de desarrollo.
- **Autorización:** Casbin propaga fallos de inicialización/persistencia y el proceso aborta al fallar. Políticas heredadas de empleador/trabajador/IPS/auditor/contratista se retiran al arranque. professional_sst ahora requiere company_id firmado en CompanyScope; queda habilitado en las rutas con scope implementado, pero todo el API/membership multiempresa aún requiere auditoría.
- **Company scope:** la ruta de empresas mantiene el acceso del piloto sintético. Los grupos de empleados/incidentes/capacitaciones verifican JWT, company scope y rol antes del cierre 404; sus pruebas confirman 403 A→B y que header/query/body no sobrescriben el claim. Los filtros de servicio no reemplazan la cobertura E2E; las funciones sensibles no se habilitan para datos reales.
- **PHVA y evidencias:** se conservan servicios tenant-scoped y evaluación PHVA conservadora, pero las rutas PHVA/evidencias/eventos de agente están sin registrar hasta que exista prueba de ownership por objeto y cobertura de integración.
- **Rutas y operaciones retiradas temporalmente:** exámenes médicos, exportación, OCR-RUT, APIs Documents, empleados/incidentes/capacitaciones, biometría e inspección por zona están cerrados hasta aprobar autorización, privacidad y retención. AI requiere `PRIVACY_PROCESSING_APPROVED` además de sus flags; perfil `high-risk` sigue desactivado. Billing, Drone y Autopilot están health-only.
- **Privacidad/consentimiento:** capturar consentimiento ya no crea un `granted=true` implícito; la validación responde denegado con error si no existe integración aprobada, y la revocación queda cerrada hasta incluir actor y auditoría. Exámenes/exportación fallan también a nivel de servicio. Supresión/anonimización sigue deshabilitada, no se inventa retención por categoría y el bootstrap local añade triggers de inmutabilidad al audit log; producción requiere migración y privilegios separados.
- **Skills:** el paquete v3 quedó como contenido versionado y con todas sus entradas deshabilitadas en el manifest. No se trata como código ejecutable ni como fuente normativa aprobada.
- **Trazabilidad:** se agregaron notas de metadata con estados parciales/deshabilitados para no presentar capacidades no comprobadas como garantías.

## Validaciones ejecutadas

| Verificación | Resultado | Alcance |
|---|---|---|
| `PATH=/tmp/zu360-test-venv/bin:$PATH make test` | **PASS** | Los 10 módulos Go ejecutaron `go test ./... -count=1`; pasaron 42 tests AI, 10 tests Documents y el guard Docker local/remoto. Los paquetes Go sin tests sólo se compilaron. |
| Guard de Docker local | **PASS** | Se aplicó a `setup-core-test.sh`, `migrate.sh`, `seed.sh` y targets Make con invocación directa; el test simulado acepta `unix://`/`npipe://` y rechaza TCP/`DOCKER_HOST`. |
| `go vet ./...` | **PASS** | Ejecutado en los 10 módulos Go. |
| `make build` | **PASS** | Compila los diez módulos Go y valida sintaxis Python AI/Documents; Media usa toolchain Go 1.24.0. |
| `scripts/sync-go-deps.sh --clean-cache` | **PASS** | Caché Go limpiada; los diez módulos ejecutaron tidy, verify y build. |
| Sintaxis shell/Python/YAML | **PASS** | `bash -n`, `compileall` y parseo YAML de Compose, tres workflows, Kong y Prometheus. Las 12 rutas `env_file` tienen `.env.example`. |
| Parseo JSON de `manifest.json` | **PASS** | Formato; no habilita ninguna skill. |
| `docker compose config`, build de imágenes y stack | **NO EJECUTADOS** | El sandbox carece de Docker/Compose; parsear YAML no valida esquema Compose ni inicia servicios. |
| Suite de integración con Postgres/Redis/RabbitMQ/S3 | **PENDIENTE** | No se levantó el stack local; no se han probado migraciones, flujos HTTP ni mensajería E2E. |
| CompanyScope unit A→B | **PASS** | Casos Fiber demuestran 403 para A→B, pero no acceden a PostgreSQL ni verifican todas las rutas/objetos. |
| Smoke HTTP A/B con Postgres | **PENDIENTE** | `scripts/smoke-company-isolation.py` creado; no ejecutado, porque Docker/Compose no está disponible en este sandbox. |
| Pruebas multi-tenant cross-company y de roles completas | **PENDIENTES** | No existe una batería de integración que demuestre aislamiento en todo el API/objetos/archivos/jobs. |
| Despliegue, restore de backup, carga, DAST/pentest, observabilidad en prod | **NO EJECUTADOS** | No se usaron credenciales ni infraestructura de producción |

## Bloqueos P0 antes de producción o datos reales

1. **CI/artefactos de despliegue:** unit tests, `go vet`, `make build` y `go.sum` ya pasan en el sandbox. Aún falta ejecutar CI limpia en GitHub, lint, build de imágenes, validar Compose con Docker, compilar web/móvil y probar migraciones/stack desde un checkout nuevo.
2. **Autorización empresa–usuario:** el scope firmado y el 403 A→B están cubiertos por unit tests; APIs laborales están temporalmente cerradas. Falta membership explícita usuario↔varias empresas, cobertura de todos los objetos y APIs (incluyendo rutas legacy, medios, archivos, jobs, evidencia y certificados), tests con dos empresas por tenant y prueba E2E de DB.
3. **RLS/DB y migraciones:** adoptar RLS o un control equivalente de defensa en profundidad; revisar claves foráneas y restricciones compuestas `(tenant_id, company_id)`; validar borrado lógico, cascadas, migración hacia atrás y consistencia entre servicios Go/Python.
4. **JWT/sesiones:** implementar rotación de refresh tokens, invalidación inmediata de access al cerrar sesión o reducir ventana, rotación de llaves, issuer/audience, política de expiración y pruebas; el logout actual no revoca un access ya emitido.
5. **Datos sensibles:** clasificación/retención/borrado/portabilidad aprobados por privacidad y asesoría legal; cifrado y gestión de llaves; control de acceso a exámenes, restricciones, diagnósticos, biometría y fotos; consentimiento, bitácora, acuerdos con encargados/proveedores y respuesta a incidentes. La exportación permanece retirada a propósito.
6. **Normativa:** revisión por profesional SST competente y asesoría jurídica de normas, vigencias y alcance. Se confirma que 4272/2021 derogó 1409/2012; 1843/2025 derogó 2346/2007; y 491/0491/2020 regula espacios confinados. Hay una cita `4276/2022` sin soporte oficial identificado. Las plantillas de alto riesgo siguen deshabilitadas.
7. **Planes y cobro:** no hay evidencia de entitlements ejecutables por plan ni de un circuito de suscripción/pago/reconciliación probado. Definir matriz de planes y límites; backend como autoridad, idempotencia, suspensión/renovación/cancelación, facturación, impuestos y políticas de sobreuso antes de vender licencia.
8. **Operación:** infraestructura TLS, secretos fuera de `.env`, backup cifrado y prueba real de restore, RPO/RTO/SLA, monitoreo/alertas, logs con redacción de datos, rate limit por usuario/IP, WAF, continuidad, runbooks, responsable de guardia y plan de respuesta a incidentes.
9. **Seguridad externa:** SBOM, escaneo de dependencias/secretos, SAST, DAST, pentest independiente, threat model por tenant y reporte de remediación antes de admitir información de empresas.

## Mercado colombiano y tesis de diferenciación

La comparación anexa contrasta páginas oficiales/comerciales de ISOTools, Pensemos, Novasoft, Heinsohn, Verifty y Kondory, además de alternativas de ARL como ALISSTA/Positiva y la herramienta SG-SST de ARL SURA. Es una comparación de presencia y capacidades publicadas; no sustituye demo con pruebas, cotizaciones ni evaluación de contratos, onboarding, servicio y costos. No se encontraron precios públicos comparables suficientes.

La tesis más defendible es **“operación SST de portafolio para profesionales independientes/consultores, con trazabilidad y asistencia”**, no “software que garantiza cumplimiento”. Para marcar diferencia de verdad hay que medir: minutos para incorporar empresa, tiempo para cerrar acciones, evidencia recuperable en auditoría, exactitud/seguridad de asistencias IA, y retención/disposición a pagar. IA, visión y drones sólo deben convertirse en propuesta comercial después de validación independiente de precisión, sesgos, seguridad y aprobaciones humanas.

## Corte actualizado de Gates 1 y 2 — 19 de octubre de 2026

- **Gate 1 (build e aislamiento por empresa): PARCIAL / NO-GO.** La suite unitaria se repitió y pasó; el guard Docker evita seleccionar endpoints remotos en los scripts cubiertos. El aislamiento A→B demostrado sigue limitado al middleware/rutas ya protegidas: no hay Docker disponible aquí para ejecutar migraciones ni el smoke HTTP/Postgres, y siguen pendientes membresía multiempresa, cobertura por objeto y defensa de base de datos.
- **Gate 2 (privacidad y revisión normativa): ABIERTO / NO-GO, con mitigaciones técnicas aplicadas.** Se cerraron APIs y servicios internos de trabajadores/incidentes/capacitaciones, Documents, exámenes, exportación, OCR médico, biometría e inspección por zona; las referencias de alto riesgo se marcan pendientes en los artefactos actualizados. El inventario de código es preliminar. Siguen pendientes aprobación legal/privacidad, contratos, finalidad/base/autorización, retención por categoría, ciclo de vida de revocación/borrado, migración productiva de auditoría, revisión normativa competente y pruebas de integración/DB.
- **Alcance autorizado:** sólo preparación y pruebas locales con datos sintéticos. No autoriza piloto con trabajadores/empresas reales, operación comercial, licencias pagas ni afirmaciones de cumplimiento.

## Orden de ejecución recomendado

**Gate 1:** build/CI y autorización por empresa.  
**Gate 2:** privacidad y revisión normativa, con módulos médicos/alto riesgo/exportación públicos apagados.  
**Gate 3:** planes/entitlements/pagos + observabilidad y recuperación probada.  
**Gate 4:** piloto cerrado con datos sintéticos; luego piloto con consentimiento y DPA; pentest/remediación; sólo entonces GA.

## Artefactos asociados

- [Auditoría técnica baseline](AUDITORIA_TECNICA_BASELINE.md)
- [Checklist amplio listo para ejecución](CHECKLIST_LISTO_PARA_PRODUCCION.md)
- [Gates críticos antes de datos reales de trabajadores](GATES_DATOS_REALES_TRABAJADORES.md)
- [Inventario técnico preliminar de privacidad Gate 2](INVENTARIO_TECNICO_PRIVACIDAD_GATE2_2026-10-18.md)
- [Resultados de pruebas locales](../testing/RESULTADOS_PRUEBA_LOCAL_2026-10-17.md)
- [Plan maestro de hardening y lanzamiento](PLAN_MAESTRO_HARDENING_Y_LANZAMIENTO.md)
- [Comparativa SG-SST Colombia](../market/COMPARATIVA_COMPETIDORES_SG-SST_COLOMBIA.md)
- [Integración segura del paquete de skills v3](../skills/INTEGRACION_SEGURA_SKILLS_V3.md)
- [Verificación normativa inicial Colombia](../skills/VERIFICACION_NORMATIVA_INICIAL_COLOMBIA.md)


## Aclaración de última revisión — herramientas ejecutables de IA

El paquete `domain_packs/zu360-skills-v3` no se carga automáticamente, pero `services/ai/app/skills/registry.py` registraba herramientas propias en el singleton al iniciar, incluidas creación de incidentes, generación de documentos, recordatorios y misiones de dron. Ese registro automático se retiró: el registry arranca vacío hasta que cada herramienta tenga autenticación servicio-a-servicio, autorización por empresa/objeto, confirmación según acción, registro auditable y pruebas. El manifest del paquete importado continúa con todas las entradas `approved_for_runtime=false`.


### Actualización final — PHVA

La evaluación automática PHVA quedó aún más conservadora: ya no consulta conteos de tablas para inferir señal; devuelve el estándar como `pendiente` con instrucción de validar aplicabilidad, fuente vigente y suficiencia de evidencia. La afirmación previa sobre “no marcar cumple por contar filas” se reemplaza por este comportamiento final.
