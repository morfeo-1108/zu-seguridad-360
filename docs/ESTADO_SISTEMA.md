# Estado real del sistema — auditoría ejecutada 2026-10-19

Todo lo descrito aquí fue **ejecutado y verificado en esta sesión**, en un entorno local
con Docker Engine 28.5 / Compose v2.40, Go 1.26, Node 24, Python 3.14, contra el snapshot
`zu-seguridad-360-gate2-2026-10-19.zip`. Es la primera vez que este código corre contra
PostgreSQL real: las auditorías previas documentadas en el propio repo se hicieron en un
sandbox sin Docker.

## 1.1 Microservicios

| Servicio | Compila (`go build`) | Tests Go | Levantado en Docker hoy |
|---|---|---|---|
| core | Sí | 7 paquetes con tests, todos pasan | Sí — probado end-to-end |
| mail | Sí | Sin tests | No probado hoy |
| notifications | Sí | Sin tests | No probado hoy |
| autopilot | Sí | Sin tests | No probado hoy |
| attention | Sí | Sin tests | No probado hoy |
| onboarding | Sí | Sin tests | No probado hoy |
| billing | Sí | Sin tests | No probado hoy (bug encontrado y corregido por inspección, ver abajo) |
| sla | Sí | Sin tests | No probado hoy |
| drone | Sí | 1 paquete con test (`internal/flight`) | No probado hoy |
| media | Sí | Sin tests | No probado hoy |
| ai (Python) | `py_compile` OK | 42 tests, 6 fallan en este entorno (ver 1.1.1) | No (bloqueado por diseño: `AI_SERVICE_API_ENABLED=false`, etc.) |
| documents (Python) | `py_compile` OK | 10/10 pasan de forma aislada; **nunca se ejecutan dentro de `make test`** porque el script aborta antes por el fallo de `ai` (`set -e`) | No probado hoy |

**De los 10 módulos Go, solo 2 (`core`, `drone`) tienen algún test unitario real.** Los
otros 8 compilan pero no tienen una sola prueba. Eso no es automáticamente un defecto —
varios están deliberadamente apagados por feature flags— pero "compila" y "está probado"
son cosas distintas y el inventario anterior los mezclaba.

### 1.1.1 Hallazgo: 6 tests de IA fallan en este entorno por incompatibilidad con Python 3.14

`tests/test_biometric.py` y un caso de `test_inference.py` usan `asyncio.get_event_loop()`
sin loop activo, patrón que Python 3.14 ya no soporta (antes emitía un *warning*, ahora
lanza `RuntimeError`). Es un bug de la *test suite*, no del motor de biometría/visión en sí
(que de todos modos está bloqueado por flags). Se recomienda cambiar ese helper a
`asyncio.run(...)`.

## 1.2 El bug más importante encontrado hoy: JSONB roto en Core y Billing

**Antes de esta sesión, nadie había podido verificar login ni ninguna lectura de Core
contra PostgreSQL real**, porque el sandbox de todas las auditorías previas no tenía
Docker. Al correrlo por primera vez:

- `services/core/internal/model/company.go` define `type JSONB map[string]interface{}`
  **sin implementar `sql.Scanner` ni `driver.Valuer`**. Resultado: **toda lectura de un
  usuario con esa columna no-nula fallaba**, incluido el login — y el error se enmascaraba
  como "credenciales inválidas" (401) porque `Login()` trata cualquier error de
  `FindByEmail` como `ErrInvalidCredentials`. No era un problema de contraseñas ni de
  hashing (se verificó con tres implementaciones independientes de bcrypt: Postgres
  `crypt()`, Python `bcrypt`, y Go `golang.org/x/crypto/bcrypt` aislado — las tres
  confirmaron que el hash y la contraseña coincidían).
- El mismo patrón rompía la lectura de `employees` (`active_restrictions`, `epp_required`),
  y habría roto `incidents` (`evidence_urls`), `documents` (`signers`, `norm_refs`),
  `risk` (`controls`) y `medical_exam` (`restrictions`) en cuanto esas filas tuvieran datos.
- El mismo tipo `JSONB`, definido por separado en `services/billing/internal/model/idempotency.go`
  para las tablas de **idempotencia de cobros y outbox de eventos**, tiene exactamente el
  mismo defecto. No se verificó en vivo (billing no se levantó hoy), pero es el mismo código,
  el mismo bug.

**Se corrigió en esta sesión** (commits `4743927` y `a27072d`): se implementó `Scan`/`Value`
para `JSONB`, y se creó un segundo tipo `JSONBArray` para las columnas cuyo valor por
defecto es un arreglo (`evidence_urls`, `signers`, `norm_refs`, `controls`, `restrictions`,
`active_restrictions`, `epp_required`) — un `map` de Go no puede deserializar un arreglo
JSON. Tras el fix, login y lectura de `employees` funcionan contra Postgres real.

**Por qué importa:** esto no es un detalle. Significa que el sistema, tal como estaba
empaquetado, **nunca podría haber iniciado sesión un usuario real contra una base de datos
real**, sin que ningún test unitario lo detectara (los tests unitarios no tocan Postgres).
Es la prueba más concreta de por qué "los tests pasan" y "el sistema funciona" son
afirmaciones distintas.

## 1.3 Gate 1 (aislamiento multi-tenant): verificado end-to-end por primera vez

Con el bug de arriba corregido, se ejecutó el runbook completo (Fases A/B/C):

- `make setup-core-test` adaptado a esta máquina (ver nota de puertos abajo) levantó
  Postgres + Redis + RabbitMQ + Core, migró el esquema, y pobló datos sintéticos.
- `scripts/smoke-company-isolation.py`: **25/25 verificaciones pasan**, incluyendo:
  login real, listado de memberships exacto por usuario, acceso a empresa activa,
  denegación 403 estricta en todo intento cruzado A→B y B→A (empresa y trabajadores),
  rechazo de headers/query forjados (`X-Company-ID`, `X-Tenant-ID`, `?company_id=`),
  `switch-company` emitiendo tokens correctamente escopados, y bloqueo de cambio de tenant.
- La multi-tenencia es **monolítica con columna `tenant_id`**, no esquema-por-tenant
  (confirmado leyendo `internal/repository/postgres.go`): el aislamiento depende enteramente
  de que cada query incluya el filtro correcto, no de una barrera estructural de PostgreSQL.
  El resultado de hoy es evidencia de que, en los endpoints probados, ese filtro está bien
  aplicado — pero es un control de aplicación, no de base de datos.

**Esto no es Gate 1 aprobado para producción.** Cubre los endpoints que el smoke ejercita
(companies, employees, switch-company), no una matriz completa de rol×tenant×objeto×verbo
para las ~127 rutas que registra Core. Tampoco se probó con RabbitMQ consumiendo colas con
reintentos, ni con carga concurrente, ni egress de red, ni pentest.

### Nota de entorno (no es un bug del proyecto)

Esta máquina ya tiene otro proyecto Docker corriendo con Postgres en 5432 y Redis en 6379.
Se remapearon esos puertos del host a 55432/56379 solo para esta sesión de pruebas (no se
tocó ningún contenedor ajeno). El cambio no se dejó commiteado porque es específico de esta
máquina, no del proyecto.

Separado de eso, se encontró un bug real de empaquetado: `scripts/setup-core-test.sh` solo
crea `.env.dev` para `core`, pero `docker compose config` valida el archivo completo
(incluye `env_file` de los 11 servicios restantes) y fallaba por archivos inexistentes.
Se generaron localmente desde `.env.example` para poder continuar; el script debería
hacerlo automáticamente.

También se encontró que el perfil "solo Core" que describe el runbook **no es
independiente**: Core no arranca su *health check* sin Redis y RabbitMQ también corriendo
(reintenta indefinidamente sin fallar ruidosamente). El runbook debería decirlo.

## 1.4 Infraestructura de datos

- **PostgreSQL**: multi-tenant monolítico por columna, no por esquema (ver 1.3). Las
  migraciones (`AUTO_MIGRATE=true` vía GORM AutoMigrate) corrieron limpias hoy sobre una
  base nueva.
- **Weaviate (RAG normativo)**: el esquema (`ZUNormativaSST`, `ZUDocumentTemplates`) se crea
  vacío en el arranque (`_ensure_schema` en `services/ai/app/core/weaviate_client.py`). Hay
  un endpoint `index_normative_document` para cargar normas una por una. **No existe en el
  repositorio ningún corpus pre-cargado de las "50+ normas"**; no se buscó ni se encontró
  ningún dataset de normativa colombiana en ningún formato. Confirmado: está vacío, no es
  un problema de configuración.
- **Neo4j**: usado de verdad por `services/ai` (`neo4j_client.py`, `cognition.py`,
  `roster.py`) — no es un contenedor huérfano. Pero el servicio `ai` completo está detrás
  de flags apagados (`AI_SERVICE_API_ENABLED=false`), así que hoy no hay datos ni grafo
  poblado verificable.
- **EMQX/MQTT**: usado de verdad por `services/drone` (telemetría de vuelo). No huérfano,
  pero drone/autopilot están fuera de alcance del piloto.
- **RabbitMQ**: usado de verdad por core, mail, notifications, onboarding, billing. No
  huérfano.
- **Redis**: usado de verdad por core (refresh tokens, health check). No huérfano.

**Conclusión de esta sección:** ningún contenedor del `docker-compose.dev.yml` es
decorativo — todos tienen código real que los consume. El problema no es infraestructura
huérfana; es que los módulos de mayor riesgo (IA, biometría, drones) están, correctamente,
apagados por flag, y por eso su infraestructura de soporte está vacía.

## 1.5 Secretos (escaneo real con `detect-secrets`, no revisión manual)

61 coincidencias en todo el árbol (excluyendo `node_modules`). Triaje:

- **20 falsos positivos**: `services/ai/app/skills/domain_packs/zu360-skills-v3/manifest.json`
  son checksums SHA-256 de integridad de archivos, no secretos.
- **~30 falsos positivos de bajo riesgo**: contraseñas de ejemplo en `.env.example` de cada
  servicio y en `docker-compose.dev.yml` (`zupassword`, `zuredis123`, `zuemqx123`, etc.).
  Son intencionalmente de desarrollo. **Dato relevante y positivo**: `services/core/config/config.go`
  tiene un `ValidateSyntheticPilot`/validación de arranque que **rechaza explícitamente**
  arrancar en `ENV=production` si `DB_PASSWORD`, `REDIS_PASSWORD` o `JWT_SECRET` siguen
  siendo esos valores de desarrollo. Es un control real, no solo documentación.
- **1 hallazgo de higiene menor, no bloqueante**: `services/drone/internal/repository/postgres.go`
  tiene un DSN de respaldo hardcodeado (`user=zu password=zu`) si `DRONE_DB_DSN` no está
  seteada, en vez de fallar cerrado. Bajo riesgo porque drone está fuera de alcance del
  piloto, pero debería corregirse antes de habilitar ese módulo.
- **No se encontró ninguna credencial real** de Wompi, DIAN, SendGrid/SES, WhatsApp
  Business, AWS, Sophos, ni de ninguna ARL, en ningún formato, en ningún archivo.
- **Limitación de este escaneo**: el ZIP de origen excluye `.git` explícitamente (según su
  propio manifiesto de entrega), así que no hay historial de git que inspeccionar. Si en
  algún momento se commiteó y luego se removió un secreto real en el repositorio original,
  esta auditoría no puede verlo desde este snapshot.
- **HashiCorp Vault**: está definido en `docker-compose.dev.yml` (`infra/vault`) pero no se
  encontró ningún código en los 12 servicios que lo consulte en tiempo de ejecución — hoy
  todos los servicios leen configuración de variables de entorno/`.env`, no de Vault. Vault
  está desplegable, pero no integrado.

## 1.6 Frontend

- `next build` real: **24 páginas generadas**, sin errores, con Next.js 15.5.28.
- Rutas con datos reales conectados a Core: `/`, `/auth/login`, `/dashboard`,
  `/dashboard/companies` (+ `[id]`), `/dashboard/employees` (+ `[id]`), `/dashboard/incidents`,
  `/dashboard/training`.
- Rutas que son pantallas de bloqueo idénticas (1.13 kB cada una, sin lógica propia):
  `agents`, `attention`, `autopilot`, `billing`, `documents`, `drones`, `drones/flight`,
  `management`, `medical`, `risks`, `settings`, `siso`, `zones`. No son mocks con datos
  falsos — son bloqueos explícitos, consistente con lo que documentan el TODO y el manifiesto
  de entrega — pero tampoco son funcionalidad.

## 1.7 Modelo de negocio: plan por profesional vs. plan por empresa

Hallazgo fuera del checklist original de la Fase 1, surgido de una pregunta del usuario:
**no existe en el código ningún concepto de "plan que da derecho a N empresas por
profesional"** (el negocio describe básico=10, medio=20, premium=50 empresas por
profesional). `services/billing` modela `Subscription` con `CompanyID` como
`uniqueIndex` — es decir, **una suscripción por empresa individual**, no una suscripción
por profesional con cupo de empresas. Si el modelo de negocio real es el segundo, el
esquema de `billing` necesita rediseño, no solo configuración. El propio
`docs/production/AUDITORIA_TECNICA_BASELINE.md` ya señalaba como pendiente "implementar
catálogo versionado de planes, entitlements, cuotas, seat/storage limits".
