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

## 2. Correcciones y desbloqueos ejecutados en la misma sesión, por decisión explícita del usuario

Todo lo siguiente se hizo **después** de la Fase 2 (gap analysis), con aprobación explícita
del usuario para: (a) quitar los flags de bloqueo de AI/biometría/OCR/medical/PHVA/
Documents, (b) dejar scaffolding de integraciones externas sin credenciales reales,
(c) rediseñar billing a plan-por-profesional, (d) ejecutar el resto de ingeniería pura
sin pedir permiso paso a paso. Cada punto se verificó con `go build`/`go test`/`pytest`
reales, y los más críticos además contra Docker real.

### 2.1 Pipeline de tests arreglado de punta a punta
`make test` ya corre limpio en una sola pasada (antes, el primer fallo de AI cortaba la
ejecución con `set -e` y los 10 tests de Documents nunca llegaban a correr). Se arregló
`asyncio.get_event_loop()` roto en Python 3.14, se corrigió que `setup-core-test.sh` solo
generaba `.env.dev` para `core` (rompía la validación de `docker compose config` con los
otros 11 servicios), y se descubrió que el perfil "solo Core" nunca podía pasar su propio
healthcheck sin Redis+RabbitMQ también corriendo (reintenta indefinidamente sin fallar).

### 2.2 Backup y restore reales, no solo el script
`make backup-db` y `make test-restore` existen y se ejecutaron de verdad: backup de 74 KB,
restore en una base descartable con conteo de filas idéntico en las 4 tablas comparadas.
Antes de hoy no existía ningún mecanismo de backup.

### 2.3 Medical exams, PHVA y Evidence: desbloqueados, pero no como se pidió literalmente
Al intentar simplemente "quitar el flag", se encontró que:
- `MedicalExamService` era un stub completo (todo método devolvía el mismo error sin
  tocar la base). Se implementó de verdad, respetando el mismo gate de `PILOT_SYNTHETIC_ENABLED`
  que ya usan employees/incidents/trainings.
- Evidence y PHVA solo validaban `tenant_id`, no `company_id`, y tomaban el `company_id`
  del body/query enviado por el cliente — un profesional con token activo en la empresa A1
  podía leer/escribir evidencia y registros PHVA de la empresa A2 de su mismo tenant,
  saltándose el flujo de `switch-company`. Se corrigió forzando `company_id` desde el
  claim del token en cada punto, nunca desde el cliente.
- Verificado contra Postgres real: programar y revisar un examen médico bajo A1 funciona
  (200/201); el mismo examen es inalcanzable al cambiar a A2 (400 "record not found");
  PHVA init y registro de evidencia funcionan bajo A1.
- **Drones/autopilot se dejaron bloqueados a propósito.** No son un flag: el binario real
  (`main.go`) rechaza arrancar con `log.Fatal` si `DRONE_ENABLED=true`, con el mensaje
  "disabled until verified identity, tenant scope, operator authorization, and hardware
  safety checks are implemented". No hay código de autorización ni de seguridad física
  para control de vuelo real detrás de ese flag — forzarlo no es "desbloquear", es exponer
  control de hardware sin ningún control de seguridad. Se dejó intacto.

### 2.4 AI/visión/biometría/OCR: desbloqueados, con dos huecos de autenticación cerrados
Al revisar `app/core/feature_gates.py` (que bloqueaba OCR y zone-inspection incluso con
todos los flags en `true`, "pending verified Core consent/company authorization
integration"), se encontró:
- `/v1/vision/infer` no tenía autenticación alguna; `company_id` era un campo de
  formulario controlado por el cliente. Se agregó `Depends(get_current_user)` y ahora
  `company_id` viene del token.
- `/v1/zone-inspection/enroll-biometric` solo validaba `tenant_id`; un profesional podía
  enrolar datos biométricos de un trabajador de otra empresa de su mismo tenant. Se agregó
  verificación real de membership contra Core antes de enrolar.
- `agents.py` y `zone_inspection.py` importaban `get_current_user`, que **no existía** en
  `app/core/auth.py` — el servicio de IA no podía arrancar en absoluto, nunca se había
  probado porque Docker nunca estuvo disponible antes de esta sesión.
- `ai/.env.example` y `documents/.env.example` tenían `DATABASE_URL` con el esquema de Go
  (`postgres://`) en vez de SQLAlchemy (`postgresql+asyncpg://`) — ambos servicios habrían
  caído al arrancar.
- `consent_verified` en el enrolamiento biométrico sigue siendo un booleano que reporta el
  propio cliente, sin verificación real. Esto no se arregló ni se puede arreglar sin un
  sistema real de captura de consentimiento — queda documentado como hueco, no oculto.

### 2.5 Documents: se le agregó la autenticación que nunca tuvo
Antes de abrir el allowlist, se encontró que `generate_document`, `generate_certificate` y
ambos reportes **no tenían ningún `Depends`, ningún JWT, nada** — era un generador de
documentos/certificados completamente anónimo. Se agregó `app/core/auth.py` (mismo
esquema JWT de Core) y se exige en cada ruta de generación. `/v1/certificates/verify/*`
sigue público a propósito (verificación por QR). Límite que sigue abierto: este servicio
no tiene acceso directo a la base de Core, así que no puede verificar que los campos del
body (nombre, empresa, NIT) correspondan a la empresa activa del que llama — solo que
está autenticado.

### 2.6 Billing: rediseñado a plan-por-profesional, con cupo aplicado de verdad
`Subscription` pasó de ser por empresa (`CompanyID uniqueIndex`) a ser por
tenant/profesional (`TenantID uniqueIndex`, `MaxCompanies`: básico=10, estándar=20,
premium=50). Billing expone `GET /v1/tenants/:id/can-add-company` (JWT real, un tenant
solo puede consultar su propio cupo) y Core la llama antes de crear una empresa no
sintética, **fallando cerrado** si billing no responde (no se asume permiso). Cubierto con
tests que mockean la llamada HTTP (permitir/denegar/no disponible/omitido en modo
sintético) — no se probó aún contra un par Core+Billing reales en Docker porque la
descarga de la imagen base se volvió muy lenta a mitad de sesión (condición de red de esta
máquina, no un bug de código).

### 2.7 Notificaciones: WhatsApp y push dejaron de simular éxito
`sendWhatsApp` y `sendPush` registraban un log y devolvían éxito **sin intentar nada**,
sin credenciales, sin llamada real. Para un canal que escala hallazgos críticos de
seguridad (falta de arnés, etc.), reportar éxito falso es peor que no tener el canal. Ambos
ahora hacen una llamada HTTP real (WhatsApp Business Cloud API / FCM legacy) y devuelven
`Success: false` con el motivo exacto si no hay credenciales o la llamada falla.

### 2.8 Wompi: cliente real, no simulado
`internal/wompi` implementa `CreateTransaction` (API de pagos) y `VerifyEventSignature`
(checksum de webhooks), ambos honestos: sin llave privada configurada, devuelven error
explícito, nunca una transacción fabricada. Están modelados sobre el contrato
públicamente documentado de Wompi, pero **nunca se ejecutaron contra el sandbox o
producción reales de Wompi** porque no hay credenciales en este entorno — antes de un
cobro real hay que verificar los nombres/formas exactas de los campos contra la
documentación vigente de Wompi. `BillingService.ChargeInvoice` y
`POST /v1/webhooks/wompi` ya están conectados y probados con un servidor Wompi simulado.

### 2.9 DIAN, ARL (FURAT), Textract, Sophos: estado real, sin inventar nada
- **DIAN**: no se construyó un cliente real. Facturación electrónica real requiere una
  resolución de facturación vigente y un certificado digital para firmar XML UBL 2.1 —
  son credenciales legales que solo el usuario puede tramitar, y sin ellas cualquier
  cliente que yo construya sería, en el mejor caso, decorativo. Se deja explícitamente
  pendiente, no simulado.
- **FURAT/ARLs**: `GenerateFURAT` ya era honesto antes de esta sesión (`"simulated_preview"`,
  no se transmite a ARL) y se dejó así. Construir la transmisión real requiere el formato
  específico de cada ARL (Sura/Positiva/Colmena no comparten API), que no está documentado
  en este repositorio ni se puede inventar de forma responsable.
- **AWS Textract**: ya estaba implementado con el SDK real (`boto3`) en
  `services/ai/app/api/ocr.py`; solo falta que el usuario cargue credenciales AWS reales.
- **Sophos Intelix**: `scanWithSophos` ya era honesto (comentario explícito "simulación";
  solo bloquea ejecutables por cabecera MZ) — no mentía sobre haber escaneado con Sophos,
  así que no se tocó.

### 2.10 Vault: integración real en Core; patrón replicable en los otros 11
Antes, `VaultAddr`/`VaultToken` solo se usaban para rechazar el token de desarrollo en
producción — ninguna credencial se leía realmente de Vault. Se agregó
`Config.LoadVaultSecrets()` (KV v2), gateado por `VAULT_ENABLED=false` por defecto (no
cambia ningún comportamiento ya verificado). Si se activa, una falla de lectura es fatal
(no se sigue con el valor local). Probado contra un servidor Vault simulado. **No
replicado a los otros 11 servicios** — es el mismo patrón mecánico, pero no alcanzó el
tiempo de esta sesión.

### 2.11 Guardia estructural de autenticación (no la matriz RBAC completa)
Se agregó un test que lee el propio `main.go` y falla si algún grupo de rutas o ruta
suelta no incluye el middleware JWT — se verificó que detecta la regresión real
(se quitó `jwt` de un grupo a propósito, el test falló, se restauró). Esto **no es** la
matriz completa de rol×tenant×empresa×verbo sobre las ~127 rutas de Core que pediría una
auditoría RBAC exhaustiva — esa requiere una base real con fixtures de los 7 roles y es un
trabajo considerablemente más grande. Lo que existe hoy es una red de seguridad barata
contra el error más costoso (registrar una ruta nueva sin autenticación), no un sustituto
de la matriz completa.
