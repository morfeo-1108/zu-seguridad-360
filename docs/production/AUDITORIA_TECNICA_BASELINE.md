# Auditoría técnica ZU Seguridad 360

## Informe para fundador / comprador técnico

**Objeto:** evaluación estática de preparación técnica para una plataforma SST multicliente, licenciable por planes, con módulos de IA, biometría y drones.

**Base de este informe:** síntesis de los hallazgos estructurados entregados y verificaciones reproducibles complementarias. No se re-investigó el ZIP ni se ejecutó funcionalmente la aplicación; no se agregan evidencias externas.

---

## 1. Veredicto ejecutivo

### ¿Se puede vender o desplegar hoy?

**No.** El producto **no está listo para comercialización general, cobro ni despliegue productivo como SaaS SST multicliente**. Tampoco debe operar hoy con datos médicos, evidencias personales, biometría, decisiones automatizadas de IA o vuelos reales.

Los bloqueadores no son únicamente de acabado: afectan la **frontera de seguridad del producto**:

- El core contiene referencias estáticas a handlers no declarados y el repositorio carece de `go.sum` en los módulos Go; por ello no existe un artefacto de build reproducible verificable.
- La autorización por rol, objeto, empresa y tenant sólo cubre una parte del core. Varias rutas sensibles aceptan cualquier JWT válido y varias consultas filtran por UUID sin validar pertenencia.
- Exportación, evidencias, PHVA, datos médicos, zonas y algunos servicios auxiliares presentan patrones de IDOR/BOLA o reciben el tenant desde la petición.
- Billing tiene modelos y estados declarados, pero no hay enforcement efectivo de planes, autenticación interna consistente, webhook de pago verificable, outbox operativo ni máquina de estados conectada al acceso del core.
- Web, portal y móvil contienen flujos visuales, datos fijos o simulaciones que no demuestran login, persistencia ni operación SST real.
- IA/biometría contiene imports rotos, autorización incompleta, consentimiento biométrico controlado por el cliente, inferencia degradada y salidas que pueden parecer conclusiones médicas o legales sin aprobación obligatoria.
- Drones contiene lógica de planificación y seguridad con pruebas unitarias, pero no demuestra autenticación del broker, preflight autoritativo, ACK físico, telemetría conectada al controlador, persistencia ni integración DJI de campo.
- Infraestructura, migraciones, backups, TLS, smoke tests y cadena de suministro no son reproducibles ni verificables con lo entregado.

### Decisión recomendada

1. **No vender licencias ni cobrar clientes.**
2. **No desplegar producción con datos reales.**
3. **No anunciar como disponibles** funciones de firma, IA, biometría, RAG, inspección por visión, SLA, notificaciones, exportación segura o vuelos autónomos que el código no demuestra.
4. El uso admisible en el estado observado es de **laboratorio o desarrollo**, y sólo podría plantearse un **piloto cerrado con datos sintéticos** después de completar los controles P0 aplicables, con IA/biometría/drones desactivados por defecto.
5. El core SaaS SST y las capacidades de mayor riesgo deben tener **gates de lanzamiento separados**. Aprobar el core no autoriza automáticamente IA, biometría ni vuelos.

Esta conclusión es técnica y operativa. **No es opinión jurídica, certificación regulatoria, certificación de seguridad ni validación de cumplimiento SST.**

---

## 2. Límites, método e incertidumbres de la auditoría

### Alcance realizado

La evaluación es una revisión **estática** de rutas, símbolos, modelos, servicios, consultas, configuraciones, Docker/Compose, Terraform, Kubernetes, frontends, documentación y pruebas fuente citadas en los hallazgos entregados.

Se distinguieron tres categorías:

- **Hecho observado:** derivado directamente de una ruta, símbolo, modelo, consulta o configuración citada.
- **Impacto inferido:** consecuencia técnica razonable del hecho, sin afirmar que una explotación haya sido reproducida.
- **Aspiración/documentación:** capacidad anunciada o comentada que no se toma como evidencia de operación.

### No realizado

- No se ejecutaron servicios, pruebas dinámicas, pruebas de API, pruebas E2E ni pruebas de penetración.
- Se intentó ejecutar `scripts/run-tests.sh` y `go test` directamente, pero el sandbox no tenía el binario `go`; también se intentó `pytest`, que no estaba instalado. Por ello no se ejecutaron `go build`, `go test` ni la suite Python.
- Sí se compiló sintácticamente con Python estándar el contenido de 42 archivos `.py`: no se encontraron errores de sintaxis. Esta comprobación no equivale a pruebas funcionales.
- No se validó configuración de producción, proveedores externos, esquema de base de datos fuera del ZIP, secretos reales, red, firewall, despliegue o hardware.
- No se probó un vuelo, un dispositivo DJI, EMQX, ACK físico, pérdida de enlace, restauración de backup ni webhook real de pagos.
- No se verificó contractual ni legalmente la retención, transferencia internacional, consentimiento, firma, licencia profesional o regulación aplicable.

### Confianza y cautelas

La confianza es **alta** para los defectos que se desprenden de símbolos, rutas, predicados, modelos y configuraciones citados. Es **media** para efectos operativos que podrían cambiar si existe infraestructura o un esquema externo no incluido en el ZIP. La ausencia de una prueba dinámica significa que este informe **no afirma que una vulnerabilidad haya sido explotada**, sino que el código observado no demuestra el control necesario.

---

## 3. Madurez cualitativa

Se usa una escala orientativa, no una puntuación de certificación:

- **Demo:** interfaz o flujo ilustrativo, sin contrato operativo confiable.
- **Scaffold:** estructura, modelos o intención presentes, con wiring o controles incompletos.
- **Prototipo controlado:** lógica parcial útil en laboratorio, no apta para datos o acciones reales.
- **Operable restringido:** funciona con controles, pruebas y límites documentados para un alcance pequeño.
- **Comercial producible:** build reproducible, seguridad multicliente, operación, cobro y soporte demostrados.

| Dominio | Madurez observada | Justificación | Decisión actual |
|---|---|---|---|
| **Core SaaS SST y tenancy** | **Scaffold avanzado / inicial** | Hay JWT HMAC, bcrypt, Redis para refresh, algunos repositorios con `tenant_id`, Casbin, auditoría prevista y pasos humanos en medicina/PHVA. Sin embargo, hay wiring roto, autorización parcial, consultas por objeto sin tenant/company y defensa de base de datos insuficiente. | **No comercializable** |
| **Licencias, planes y billing** | **Scaffold incompleto** | Existen códigos/precios, estados de suscripción/factura y una intención de idempotencia/outbox. Faltan entitlements, cuotas, autorización interna, ciclo de vida, webhook, facturación verificable y coherencia de tipos/puertos. | **No cobrar** |
| **Web, portal y móvil** | **Demo visual** | Hay pantallas y navegación, pero login web/móvil, portal, reportes de campo, inspecciones y módulos SST contienen mocks, fixtures, `setTimeout` o acciones sin persistencia. | **No entregable como producto** |
| **IA, RAG, OCR y documentos** | **Prototipo de alto riesgo** | Hay módulos y algunas dependencias de autenticación, fuentes RAG y banderas de revisión humana. También hay imports rotos, rutas sin auth, egress no controlado, confianza ornamental y resultados médicos/legales sin workflow bloqueante. | **Sólo laboratorio sintético y opt-in** |
| **Biometría y visión** | **Prototipo / modo demo riesgoso** | Se observan umbrales, consentimiento booleano y estructuras de índice, pero faltan ciclo de vida de consentimiento, borrado, modelos verificables y el modo degradado fabrica señales o detecciones. | **Desactivado por defecto** |
| **Drones y seguridad de vuelo** | **Lógica de laboratorio** | Hay estados de vuelo, preflight, geocerca, clima, pérdida de enlace y tests unitarios. No se demuestra authz, broker seguro, anti-replay, ACK físico, telemetría al controlador, misión onboard ni persistencia. | **No vuelos reales** |
| **Delivery, infraestructura y operación** | **Declarativo incompleto** | Hay CI, manifiestos parciales y controles positivos en algunos pods, pero faltan `go.sum`, módulos Terraform, overlays Kubernetes, TLS, backups/restore, migraciones confiables y gates de vulnerabilidades. | **No reproducible** |

**Conclusión de madurez:** el conjunto se comporta como una **colección de scaffolds y prototipos con piezas rescatables**, no como una plataforma SaaS SST comercialmente operable.

---

## 4. Matriz de dominios: construido versus scaffold/documentado

| Dominio | Construido u observado | Scaffold, placeholder o sólo documentado | Riesgo para el comprador |
|---|---|---|---|
| **Core SST** | Filtros `tenant_id` en varios CRUD de compañías, empleados e incidentes; JWT que rechaza tokens ausentes/inválidos/expirados; bcrypt; refresh en Redis; modelo Casbin; intención de auditoría WORM; revisión humana médica y PHVA. | `exportH` y `metaH` no están cableados; `Authorize` no cubre la mayoría de recursos; consultas de evidencia, management y exportación omiten tenant/company; no hay RLS visible ni constraints de pertenencia suficientes; metadata afirma schema aislado que no corresponde con tablas compartidas. | Fuga o mutación cross-tenant y ausencia de artefacto ejecutable. |
| **Tenancy y autorización** | Existe `TenantID` en numerosas entidades y un patrón reutilizable de repositorios scoped. | El `User.CompanyID` no gobierna consultas; Casbin no modela tenant/company/objeto; `X-Tenant-ID` puede ser una fuente alternativa; JWT confía en claims sin revalidar membership. | La frontera de seguridad no está demostrada. |
| **Planes y billing** | Catálogo de códigos/precios, estados de suscripción/factura, clave de idempotencia y modelos de outbox. | Modelos y servicios tienen incompatibilidades estáticas; faltan entitlements/cuotas; billing no instala JWT; no hay webhook autenticado ni Wompi operativo; facturas no son idempotentes por período; estados no gobiernan el core. | No se puede vender una licencia ni confiar en cobros/ingresos. |
| **Onboarding, SLA y atención** | Hay endpoints, estados, targets SLA, inbox y colas declarativas. | Onboarding puede completar pasos sin ejecutar trabajo real; SLA usa objetivos globales y no notifica; attention acepta tenant del llamador; notificaciones/mailer marcan éxito o hacen ACK antes de procesar. | Se crean compromisos comerciales sin entrega durable. |
| **Clientes web/portal/móvil** | Existe UI, rutas, stores, páginas SST y navegación. | Login web/móvil navega sin autenticarse; tokens usan claves distintas; portal muestra empresa/KPIs fijos; móvil usa `setTimeout`, botones sin persistencia y almacenamiento no seguro; múltiples pantallas tienen fixtures. | El comprador vería una demo, no un cliente operativo. |
| **Media y documentos** | Hay endpoint de upload, checksum/límites previstos y generación de PDFs. | Media no instala JWT/tenant middleware; puertos se publican; antivirus es simulado; documentos/generación no exigen auth; URLs/almacenamiento y expiración no están demostrados. | Exposición de datos, almacenamiento sin segregación y documentos no confiables. |
| **IA/RAG/OCR** | Existe RAG con fuentes, OCR con `requires_human_review`, predictor separado de GTC 45 y roster/skills con cierta jerarquía. | Imports de auth inexistentes bloquean AI; RAG no filtra tenant; WebSocket no está autenticado; OCR puede convertir keywords en aptitud; proveedor externo recibe datos sin egress visible; confianza constante; acciones AI no llaman al core. | Riesgo de automatización falsa y transferencia de información sensible. |
| **Biometría/visión** | Hay umbrales de liveness/calidad/match y estructura de índice. | `consent_verified` es un booleano del cliente; embeddings sólo en memoria; falta borrado/TTL; ONNX ausente activa hash/ceros y detecciones sintéticas; asociación persona-detección es incorrecta. | No apto para identificación, disciplina, EPP o decisiones SST reales. |
| **Drones** | Hay máquina de estados, safety checks, planner, geocerca y tests unitarios. | HTTP sin auth; broker sin TLS/mTLS/ACL; preflight recibido del cliente; sin anti-replay/ACK; telemetría se descarta o se reparte incorrectamente; errores ignorados; onboard no conectado; reportes en memoria. | Riesgo físico y ausencia de evidencia operacional. |
| **Delivery/infra** | Hay workflows CI, HPA, algunos hardening flags en Kubernetes, EKS privado/KMS declarado y pruebas unitarias. | Faltan `go.sum`, módulos Terraform, overlays K8s, TLS/Ingress, backups/restore, migrator versionado, smoke coherente, SBOM/firma y gates efectivos de Trivy; scripts apuntan a archivos inexistentes. | No hay instalación ni promoción reproducible. |

---

## 5. Riesgos principales ordenados por severidad

### R1 — Bloqueadores de build y de release (**Crítico**)

**Evidencia:** `services/core/cmd/server/main.go:180-184` usa `exportH` y `metaH` sin declaración/instanciación; los constructores existen pero no están cableados. Billing tiene inconsistencias entre `OutboxEvent`, `outbox.go`, `billing.go` y repositorios. Los Dockerfiles Go hacen `COPY go.mod go.sum`, pero no se entregan `go.sum` en los módulos `services/*`.

**Impacto:** no existe evidencia de un artefacto compilable y reproducible de core/billing. Las rutas de exportación, metadata y licenciamiento no son publicables sólo por existir en el fuente. No se afirma un resultado de build ejecutado: el entorno de revisión no tenía Go; el bloqueo se deriva de incompatibilidades estáticas y del contrato de build observado.

**Recomendación:** reconciliar modelos/repositorios/migraciones, cablear o retirar handlers, generar y revisar `go.sum`, fijar toolchain y exigir en CI `gofmt`, `go vet`, `go build`, `go test` e import-smoke de Python para todos los servicios.

### R2 — Frontera de tenancy y RBAC incompleta en el core (**Crítico**)

**Evidencia:** sólo `companies` y `employees` reciben `middleware.Authorize` en el wiring descrito; medical, incidents, trainings, zones, evidence, management, regulatory y agent-events usan principalmente JWT. Casbin evalúa role/resource/action, pero no tenant, company ni objeto.

**Impacto:** un JWT válido puede alcanzar operaciones de lectura/escritura sensibles sin demostrar pertenencia a empresa ni permiso de operación. El rol declarado en metadata no equivale a enforcement real.

**Recomendación:** aplicar deny-by-default a cada operación; resolver principal, tenant, company, objeto y acción en un PDP/repository obligatorio; separar leer, escribir, revisar, validar y exportar; cubrir la matriz de roles con pruebas positivas y negativas.

### R3 — IDOR/BOLA en exportación, evidencia, PHVA, zonas y datos médicos (**Crítico**)

**Evidencia:** exportación recibe `tenantID` pero `data_export_service.go:46-79` carga por `company_id` sin `tenant_id`; evidence consulta por `employee_id/project_id`; management consulta/actualiza por `systemID/standardID`; zonas resuelve por `project_id/zone_id`; medical crea con IDs del path/body sin comprobar pertenencia. Las relaciones de empleados, proyectos, zonas y empresas no se validan siempre como una cadena de ownership.

**Impacto:** con un UUID conocido o filtrado, podría leerse o mutarse información de otra empresa/tenant: expedientes médicos, fotos, GPS, documentos, incidentes, puntuaciones PHVA, checklists o un export completo.

**Recomendación:** recibir tenant/principal en todo método; usar `id AND tenant_id` y `company_id` cuando corresponda; resolver la relación padre antes de crear o mutar; ignorar IDs de relación no confiables enviados por el cliente; añadir FKs/constraints/RLS o una defensa equivalente; probar dos tenants con UUIDs adversariales.

### R4 — Licencia, planes y cobro no gobiernan el producto (**Crítico**)

**Evidencia:** `Company` sólo contiene un `PlanID` opcional; no se observan cuotas/entitlements. La UI promete límites de trabajadores, horas, visitas, drones, estándares y SLA, pero el backend no los aplica. Billing no instala JWT, acepta `CompanyID`/`TenantID` del body y el dashboard toma tenant de la URL. Los estados `past_due`, `cancelled` y `suspended` no gobiernan core; Wompi/webhooks no están conectados.

**Impacto:** un cliente puede usar módulos fuera de su plan o consumir recursos sin límite; una cuenta vencida podría mantener acceso; una factura puede marcarse sin reconciliación externa. El precio comercial no se convierte en una licencia operable.

**Recomendación:** catálogo versionado de planes y entitlements por tenant/company; contadores atómicos; middleware de features; autorización y relación company-tenant dentro de billing; máquina de estados; webhook firmado/idempotente; outbox real; pruebas de alta, upgrade, downgrade, gracia, suspensión, cancelación y reactivación.

### R5 — Servicios auxiliares y flujos operativos aceptan identidad del llamador (**Crítico**)

**Evidencia:** onboarding, SLA, attention, media, mail y parte de notifications no instalan JWT o aceptan `tenantId`, `companyId`, destinatarios o eventos desde body/path/header. Media publica puerto y usa `c.Locals('tenant_id')` sin middleware visible. Attention filtra por tenant proporcionado por URL/header. Notifications/mailer confirman o hacen ACK antes de garantizar procesamiento.

**Impacto:** lectura/escritura cross-tenant, relay de correo/spam, pérdida silenciosa de notificaciones, estados de onboarding falsos y compromisos SLA que no llegan a atención.

**Recomendación:** autenticación en cada servicio, identidad de servicio para llamadas internas, derivar tenant/company desde claims o relación server-side, cerrar puertos directos, allowlists, ack después de persistencia/entrega, reintentos/DLQ/deduplicación y pruebas cross-tenant.

### R6 — Control de vuelo no es una frontera segura para acciones físicas (**Crítico**)

**Evidencia:** endpoints de drone no instalan auth local; MQTT no muestra TLS/mTLS, credenciales, certificados ni ACL; `PlanRequest` y `PreflightResult` reciben del cliente estados que deberían ser autoritativos; no hay firma, nonce, TTL, anti-replay ni ACK físico; la telemetría se descarta o se reparte mal; varios errores de comandos se ignoran; onboard mission no está conectado al flujo real.

**Impacto:** una identidad no verificada puede planear/autorizar/interrumpir vuelos; un caller puede presentar un preflight aprobado; el sistema puede indicar RTH/aterrizaje sin que el dron haya ejecutado la acción; telemetría y failsafe pueden no activarse.

**Recomendación:** no operar vuelos reales hasta implementar authz interna y por scopes, broker privado con mTLS/ACL por dispositivo, preflight recalculado con telemetría confiable, comandos firmados con `command_id`/nonce/TTL, ACK físico correlacionado, fan-out correcto de telemetría, estados fail-closed, persistencia y campañas en hardware.

### R7 — IA, biometría y datos sensibles carecen de ciclo de vida y egress controlado (**Alto**)

**Evidencia:** rutas de visión/documentos no exigen autenticación consistente; RAG no filtra tenant; el consentimiento biométrico es `consent_verified` controlado por el cliente; embeddings quedan en memoria sin borrado/TTL; OCR envía bytes completos a AWS Textract; RAG/documentos/incidentes envían datos a proveedores LLM sin redacción, allowlist, región, retención o logging de egress visible.

**Impacto:** exposición o mezcla cross-tenant, consentimiento no verificable, imposibilidad de revocar/borrar biometría y transferencia de información médica/identificable sin controles técnicos demostrados.

**Recomendación:** consentimiento server-side firmado y versionado por empleado/tenant/finalidad; revocación y borrado cascada; clasificación y minimización; proveedores y regiones allowlist; egress bloqueado por defecto; URLs firmadas de corta duración; retención/TTL; auditoría de acceso y pruebas negativas.

### R8 — JWT, secretos y almacenamiento de tokens permiten suplantación si se reutilizan defaults (**Alto**)

**Evidencia:** existen fallbacks como `zu-super-secret-key-change-in-prod`, credenciales DB/Redis/Rabbit/Vault de desarrollo y `.env.dev` con secretos. Middleware copia role/tenant desde claims y permite `X-Tenant-ID` como fuente alternativa según los hallazgos. Web usa `localStorage`; móvil usa `AsyncStorage`; refresh no rota ni detecta reuse por dispositivo.

**Impacto:** con defaults expuestos podría forjarse un token con rol/tenant arbitrario. Con token robado, la sesión puede persistir; XSS/extracción de almacenamiento eleva el impacto por tratarse de datos SST sensibles.

**Recomendación:** fallar al arrancar fuera de development si faltan secretos o hay valores conocidos; secret manager/KMS, rotación y TLS; fijar algoritmo/issuer/audience y revalidar usuario/membership; ignorar o comparar `X-Tenant-ID`; cookies HttpOnly o access token en memoria; SecureStore/Keychain/Keystore; refresh rotativo con detección de reuse.

### R9 — Interfaces y acciones visibles no prueban producto operativo (**Alto**)

**Evidencia:** login web sólo hace `router.push`; mobile sólo navega a `Main`; portal fija empresa, persona y KPIs; varias páginas web fabrican empleados, zonas, KPIs y estados; incidentes/inspecciones usan `setTimeout` o botones sin handlers. SISO/AI devuelve mensajes locales y el ejecutor de acciones retorna estados `pending` sin llamar al core.

**Impacto:** un cliente puede ver datos demo o una confirmación falsa de registro, accidente, inspección, firma o acción AI. La interfaz no demuestra autenticación, persistencia, permisos ni trazabilidad.

**Recomendación:** eliminar fixtures de rutas productivas; conectar login, API, mutations y persistencia reales; asociar empresa/proyecto/zona desde membresías del servidor; implementar loading/error/empty, idempotencia offline y E2E de los flujos SST principales; marcar explícitamente cualquier demo.

### R10 — Delivery, migraciones, backups y supply chain no son operables (**Alto**)

**Evidencia:** Terraform referencia módulos ausentes; Kubernetes sólo contiene core mientras CI espera overlays y 12 deployments; smoke test usa rutas/header no coincidentes; migrator fija localhost y silencia errores con `|| true`; no se observan backups/restore/RPO/RTO; Trivy no tiene gate de salida y hay tags flotantes; `run-tests.sh` puede ocultar fallos por el pipeline de `grep`.

**Impacto:** instalación limpia, staging, promoción, rollback, restauración y detección de vulnerabilidades no son demostrables. Se puede publicar una imagen o ejecutar una migración sin evidencia suficiente.

**Recomendación:** completar módulos/manifiestos, versionar migraciones con locks y códigos estrictos, añadir probes y gates, backups cifrados/versionados con restore drills, fijar imágenes/acciones por digest/commit, SBOM/firma/attestation y hacer obligatorios todos los jobs de integración.

---

## 6. Plan de remediación priorizado

### P0 — Obligatorio antes de cualquier piloto

El piloto sólo puede ser **cerrado, con datos sintéticos o expresamente no sensibles, sin cobro y sin vuelos reales** después de que estos gates pasen:

1. **Build reproducible y arranque limpio**
   - Corregir símbolos no declarados de core y tipos/columnas de billing.
   - Generar `go.sum` para cada módulo; fijar toolchain y dependencias.
   - Ejecutar build/test/import-smoke en entorno limpio para core, billing, AI y servicios Python.
   - Retirar rutas no implementadas o hacer que compilen y tengan contrato verificable.

2. **Tenant/RBAC/objeto como frontera de seguridad**
   - Derivar tenant del token/membership, no de body, path o `X-Tenant-ID` libre.
   - Aplicar autorización a todas las rutas y operaciones sensibles, con deny-by-default.
   - Corregir exportación, evidence, management, medical, zones, projects, regulatory y employee/company ownership.
   - Validar cada relación `tenant → company → employee/project/zone/resource` en lecturas, altas y mutaciones.
   - Crear pruebas negativas con dos tenants y UUIDs conocidos; cero accesos cruzados.

3. **Auth de servicios y superficie de red**
   - Instalar autenticación/autorización dentro de billing, onboarding, SLA, attention, media, mail, notifications, drone, AI y documents.
   - Cerrar puertos internos directos; permitir sólo gateway/red privada; configurar TLS/mTLS según servicio.
   - Sustituir defaults y secretos de desarrollo; fail-closed en producción.

4. **Datos sensibles y media**
   - Media con authz, tenant/company scope, MIME/magic-byte, límites, AV fail-closed y URLs firmadas cortas.
   - Separar roles de empleador, profesional SST, IPS, auditor, trabajador y contratista.
   - No persistir ni notificar aptitud médica/OCR hasta revisión humana identificada.
   - Deshabilitar biometría, visión y acciones AI en el piloto salvo datos sintéticos, consentimiento de prueba y kill switch.

5. **Migración y datos**
   - Migrador versionado, idempotente, con historial y lock; sin `AutoMigrate` silencioso en producción.
   - FKs, unicidad y constraints de tenant/company donde correspondan; evaluar RLS o defensa equivalente.
   - Verificar que un entorno vacío crea todas las tablas antes de servir tráfico.

6. **Evidencia y observabilidad mínima**
   - Request ID, actor, tenant, recurso, acción, estado anterior/posterior en operaciones sensibles.
   - Errores públicos genéricos; detalles sólo en logs protegidos.
   - Readiness real para dependencias y alertas de fallos de cola, DB, storage y auth.

7. **Criterio de piloto**
   - Contrato explícito: sin datos médicos reales, biometría, decisiones automáticas, cobro, compromisos SLA ni vuelos.
   - Lista de features desactivadas y banner de prototipo; consentimiento informado de usuarios del piloto.

### P1 — Obligatorio antes de cobrar o vender una licencia

1. Implementar catálogo versionado de planes, entitlements, cuotas, seat/storage limits y feature flags por tenant/company.
2. Implementar billing autenticado con ownership, máquina de estados, webhook firmado, idempotencia por evento y período, reconciliación y outbox real con reintentos/DLQ.
3. Conectar estados `active`, `past_due`, `suspended`, `cancelled` y gracia al servicio de entitlements consumido por core/gateway.
4. Crear entidad verificable de profesional/licencia: número, emisor, vigencia, alcance, relación con tenant/company, auditoría y gate de firma/aval.
5. Hacer onboarding durable: pasos reales, orden, idempotencia, locks, errores persistidos y confirmación de consumidores antes de `done`.
6. Hacer SLA dependiente de plan/contrato, calendario y reloj hábil; persistir breach/resolución y entregar escalaciones con outbox.
7. Hacer portal/web/móvil reales: login, refresh/logout, permisos, datos dinámicos, mutations, persistencia offline e idempotencia; eliminar fixtures.
8. Completar mail/notifications/attention con proveedores reales configurados, ACK posterior a persistencia/entrega, retry/DLQ y deduplicación.
9. Definir retención, borrado y exportación por tipo de dato; registrar accesos y exportaciones; no exponer PII innecesaria en verificación pública de certificados.
10. Obtener evidencia de E2E del core: login, creación de empresa/empleado, incidente, evidencia, revisión médica protegida, PHVA, exportación permitida/denegada y suspensión por plan.

### P2 — Obligatorio antes de escalar, automatizar o vender módulos de alto riesgo

1. **Defensa en profundidad multicliente:** RLS o mecanismo equivalente, FKs/constraints de pertenencia, repositorios tenant-scoped obligatorios y pruebas continuas de aislamiento.
2. **Operación resiliente:** backups cifrados e inmutables, restore drills con RPO/RTO, runbooks, PDB, probes y rollback probado.
3. **Supply chain:** imágenes y acciones fijadas por digest/commit, SBOM, firma, provenance, escaneo de dependencias/imágenes con gate efectivo y revisión de vulnerabilidades.
4. **IA gobernada:** modelos versionados, evaluación fuera de muestra, calibración, abstención, drift, métricas por subgrupo, provenance por afirmación y workflow `pending_review/approved/rejected`.
5. **Egress y privacidad AI:** minimización/redacción, proveedores/regiones autorizados, no-training/retención configurados cuando aplique, auditoría y bloqueo de datos sensibles sin consentimiento/política.
6. **Biometría/visión:** modelos presentes y firmados, no modo degradado en producción, FAR/FRR calibrados, asociación persona-detección validada, consentimiento revocable, TTL y borrado verificable.
7. **Drones:** identidad criptográfica por dispositivo, broker mTLS/ACL, comandos firmados/anti-replay, preflight autoritativo, ACK físico, telemetría fan-out, failsafe onboard, persistencia y pruebas hardware de pérdida de enlace/GNSS/geocerca/batería.
8. **Escalabilidad comercial:** contabilidad de uso, límites atómicos, colas observables, pruebas de carga, aislamiento de ruido entre tenants, SLO/SLA respaldados por métricas históricas y escalación real.

---

## 7. Condiciones mínimas de go-live

### Condiciones generales, no negociables

El go-live debe bloquearse si queda abierto cualquier riesgo **crítico** y, para el alcance afectado, cualquier riesgo **alto** sin aceptación formal del riesgo por el responsable técnico y de negocio. Como mínimo debe existir evidencia reproducible de:

- Build limpio y firmado de cada servicio, con dependencias fijadas.
- Migración desde base vacía, verificación de esquema, constraints y rollback/forward plan.
- Tests de autorización por rol, tenant, company y objeto; matriz negativa cross-tenant.
- Secret manager, rotación, TLS, red privada y ausencia de defaults de desarrollo.
- Backups, restore probado, RPO/RTO y runbook de incidentes.
- Logs estructurados, auditoría de acciones sensibles, métricas, alertas, request IDs y error handling seguro.
- Health/readiness por servicio y smoke tests que realmente atraviesen gateway y upstreams.
- Política de datos, retención/borrado, acceso interno, exportación y respuesta a incidentes.
- Accesibilidad mínima y estados de error/loading/busy en los flujos de cliente críticos.

### Go-live del núcleo SaaS SST

Además de lo anterior:

- Login, refresh, logout y recuperación de cuenta probados en web, portal y móvil.
- Tenant y company derivados de identidad/membership; ningún body/path/header del cliente puede ampliar el alcance.
- CRUD de empresas, empleados, incidentes, proyectos, zonas, evidencias, documentos, exámenes médicos y PHVA cubierto por authz de objeto.
- Medical/IPS/profesional SST con roles separados, revisión humana y auditoría de acceso.
- Exportación sólo con permiso explícito, consulta `company + tenant`, job controlado, URL temporal y registro de descarga.
- Planes, límites, entitlements y suspensión conectados al acceso antes de facturar.
- Portal y pantallas SST sin datos fijos, sin `setTimeout` como confirmación y sin acciones anunciadas que no persistan.

### Go-live de IA, biometría y visión

No debe incluirse en el go-live general por defecto. Requiere un gate independiente:

- Authz por tenant/objeto en cada ruta, WebSocket autenticado y acciones server-side idempotentes.
- Consentimiento verificable por persona, finalidad, versión, fecha, actor, revocación y borrado.
- Modelo presente, firmado, versionado y validado; ningún modo hash/ceros/detección de ejemplo.
- Resultados rotulados como asistencia/borrador; no aptitud clínica, documento legal ni reporte oficial sin profesional identificado.
- Workflow bloqueante de revisión humana y prohibición técnica de firmar/notificar/usar downstream antes de aprobación.
- Egress minimizado y autorizado; retención y borrado comprobables en DB, vector store y object store.
- Evidencia de calibración, abstención, falsos positivos/negativos, drift y límites de uso.

### Go-live de drones

**No hay condición suficiente con el código observado para autorizar vuelos reales.** Como mínimo, antes de considerar ese módulo:

- Integración DJI/SDK o dispositivo objetivo documentada y probada en hardware.
- Authz local del servicio y scopes por acción de vuelo.
- Broker privado con TLS/mTLS, credencial por dispositivo, ACL por topic y revocación.
- Preflight recalculado server-side con telemetría/clima confiables y confirmación onboard.
- Comandos firmados, anti-replay, TTL, correlación, ACK físico y estados de fallo.
- Telemetría correctamente distribuida al controlador de seguridad, con sincronización y pérdida de enlace probadas.
- Geocerca de puntos y segmentos, batería/reserva/home/clima con datos frescos y validación estricta.
- Persistencia durable de misión, comando, decisión, ACK, telemetría relevante, PIC y reporte.
- Pruebas de GNSS degradado, batería baja, broker caído, reboot, RTH, aterrizaje de emergencia y pérdida de red.

---

## 8. Fortalezas aprovechables

La decisión de no lanzar hoy no implica que todo deba desecharse. Se observan activos técnicos rescatables:

- Algunos repositorios aplican filtros reales de `tenant_id` para compañías, empleados e incidentes.
- JWT valida presencia, expiración y firma HMAC; las contraseñas usan bcrypt y refresh tokens se almacenan en Redis.
- Existe un modelo Casbin, limiter, CORS configurable, request ID, recover y timeouts en core.
- El dominio SST conserva intención de revisión humana en exámenes médicos y validación humana de estándares PHVA.
- Hay catálogo inicial de planes/precios/estados, intención de idempotencia y estructura outbox, aunque aún no gobiernan el flujo.
- El módulo de vuelo contiene lógica legible de seguridad, geocerca, preflight, estados y pérdida de enlace; las pruebas unitarias de esa lógica son una base, no una prueba de integración.
- Kubernetes core incluye algunos hardening flags (`runAsNonRoot`, filesystem de sólo lectura, sin privilegios); EKS declara endpoint privado, KMS y logs.
- La documentación reconoce en algunos lugares que la integración SDK/hardware y E2E quedan pendientes. Esa transparencia debe conservarse y extenderse a todo el producto.

---

## 9. Conclusión para el fundador / comprador técnico

ZU Seguridad 360 tiene una **dirección de producto clara y piezas técnicas útiles**, pero el estado entregado no constituye todavía una plataforma SST licenciable de forma segura. La brecha principal no es estética: es la falta de una frontera confiable entre **tenant, empresa, usuario, objeto, plan y acción**, reforzada por build reproducible, infraestructura operable y pruebas negativas.

La recomendación es **detener la comercialización general**, separar tres líneas de trabajo y no mezclarlas en el mismo compromiso contractual:

1. **Core SaaS SST:** reconstruir primero la frontera multicliente, autorización por objeto, build, migraciones, portal real y entitlements.
2. **IA/biometría/documentos:** tratarlo como una superficie de datos sensibles independiente, con consentimiento, egress, revisión humana, retención y kill switch.
3. **Drones:** tratarlo como un sistema ciberfísico independiente; la lógica Go y los tests unitarios no autorizan vuelos ni sustituyen hardware, broker seguro, ACK y failsafe onboard.

El producto puede avanzar, pero el siguiente hito correcto no es “vender el primer plan”: es demostrar un **vertical slice seguro y reproducible** de un tenant, una empresa, roles reales, datos sintéticos, migración limpia, autorización negativa, auditoría, límites de plan y recuperación. Sólo después de esa evidencia tiene sentido abrir P1 de cobro y, posteriormente, gates separados para IA, biometría y drones.
