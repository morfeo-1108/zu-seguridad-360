# Runbook de pruebas controladas — ZU Seguridad 360

**Objetivo:** ejecutar un piloto funcional exclusivamente sintético en un stack local aislado. Esto no autoriza datos personales reales, atención a clientes, producción, licenciamiento ni cobro.

## Estado y límites del piloto

- **Gate 1: PARCIAL / NO-GO para datos reales.** Hay middleware JWT + membership persistida que exige que la compañía del recurso coincida con el `company_id` firmado y verifica membership activa. Las pruebas unitarias negativas de header/query/body existen. Falta ejecutar el smoke HTTP contra PostgreSQL real y ampliar la cobertura independiente a todas las rutas y objetos.
- **Gate 2: ABIERTO CON MITIGACIÓN TÉCNICA / NO-GO para producción.** En modo sintético local, Core y el dashboard permiten compañías, memberships, perfiles de trabajador de prueba, incidentes sintéticos y capacitaciones/constancias inválidas. Los datos médicos reales, AI/OCR/visión/biometría, drones/autopilot, verificación pública, pagos/cobros y exportación divulgativa permanecen bloqueados. No se afirma cumplimiento normativo ni preparación de producción.
- El Compose local publica puertos sólo en `127.0.0.1`; `assert-local-docker.sh` rechaza `DOCKER_HOST` explícito y endpoints distintos de `unix://`/`npipe://`. **Esto no bloquea por sí mismo el egress de contenedores**; no se ha demostrado aislamiento de red externo. No configurar claves ni consumir proveedores externos.
- No introducir claves de IA, correo, Wompi/DIAN, AWS, ni conectar hardware o drones. No guardar contraseñas, tokens, datos personales ni payloads en logs o evidencias.
- El Compose del repositorio no levanta frontends móviles. Los módulos médicos, PHVA, documentos, exportación y rutas no incluidas en este alcance siguen cerrados o sin validación E2E; no deben interpretarse como funcionales por el hecho de que aparezcan en el menú.

## Fase A — build y tests sin Docker

Requisitos: Go 1.24+, Python 3.12+, Node.js/npm del proyecto y dependencias instaladas. Para tests AI, instalar `services/ai/requirements-test.txt` en un entorno virtual; esto no instala modelos de producción ni PyTorch.

```bash
make test
make build
cd frontend/web
npm run typecheck
npm run build
```

`make test` ejecuta el guard Docker simulado, los tests Go disponibles por módulo y los tests Python. `make build` compila los módulos Go y verifica sintaxis Python. No requieren Docker. El build Next.js/typecheck comprueba el frontend, no sus llamadas E2E a Core.

En la corrida final del 19 de octubre de 2026, `make test` pasó en los diez módulos Go, los 42 tests AI, los 10 tests de privacidad de Documents y el test del guard Docker local. `make build`, `npm run typecheck` y `npm run build` pasaron; Next generó las 24 páginas del App Router. `py_compile` del smoke y `bash -n` del seed/setup también pasaron. La prueba unitaria de CompanyScope cubre A→B con 403 y rechaza suplantación por header/query/body. El smoke HTTP contra Postgres y las pruebas de interacción del navegador siguen pendientes porque Docker no está instalado en este sandbox; no se interpretan los builds como E2E.

## Fase B — Core y PostgreSQL locales con datos reservados de prueba

Usar una máquina con Docker Engine y Compose v2. Ejecutar únicamente el perfil local de Core:

```bash
make setup-core-test
```

El script valida el daemon Docker local, crea `services/core/.env.dev` sólo si no existe, fuerza `ENV=development`, `AUTO_MIGRATE=true`, `PILOT_SYNTHETIC_ENABLED=true`, `DB_HOST=postgres` y `DB_NAME=zu_core_synthetic`, y levanta Postgres + dependencias locales de Core. Aplica migraciones con marca `zu360-local-test-only`, construye Core y comprueba `/v1/health` antes de poblar datos. Si una fase falla, detenerse y revisar `docker compose -f docker-compose.dev.yml logs <servicio>`; no sustituir endpoints ni credenciales por valores remotos.

El seed solicita confirmación explícita y dos contraseñas distintas de al menos 16 caracteres, leídas sin eco o tomadas de variables de entorno locales. Sólo acepta correos en `@zu360.test`. Crea dos profesionales SST sintéticos en tenants separados, cuatro empresas sintéticas, cuatro memberships (dos compañías por profesional) y cuatro perfiles de trabajador ficticios. El seed puede rotar las contraseñas de esas cuentas al repetirse. No crea personas ni empresas reales.

IDs reservados:

- Profesional A (`sst-a@zu360.test`): tenant `11111111-1111-4111-8111-111111111111`; empresas A1 `22222222-2222-4222-8222-222222222222` y A2 `66666666-6666-4666-8666-666666666666`.
- Profesional B (`sst-b@zu360.test`): tenant `aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa`; empresas B1 `bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb` y B2 `cccccccc-cccc-4ccc-8ccc-cccccccccccc`.

No usar contraseñas reales ni enviarlas por chat, ticket o evidencia. Mantener `services/core/.env.dev` fuera del ZIP/commit. El perfil reducido no construye AI, Drone ni Autopilot. **El network bridge de Docker no se ha demostrado como barrera de egress**, por lo que la ausencia de claves/configuración externa no equivale a una política de red probada.

Para detener sin borrar datos: `make dev-down`. No ejecutar `make reset-dev-volumes` salvo que realmente se quiera borrar el volumen local y se haya confirmado esa acción en la terminal.

## Fase C — smoke HTTP multiempresa A/B

Con el stack local sano, ejecutar:

```bash
make test-company-isolation
```

El script solicita las dos contraseñas de prueba sin eco y no las imprime ni guarda tokens. Debe comprobar como mínimo:

1. A obtiene sólo sus memberships A1 y A2; B obtiene sólo B1 y B2. Los IDs y tenants ajenos no aparecen en el contexto del otro usuario.
2. A consulta A1 con normalidad. Headers `X-Company-ID`/`X-Tenant-ID` y `?company_id=` forjados no alteran el scope firmado.
3. A→B1 (empresa y trabajadores) devuelve estrictamente 403. Antes del switch, A→A2 también devuelve 403 porque el token está activo para A1.
4. El endpoint `switch-company` emite tokens de A2 y B2 cuando las memberships existen. Cada token sólo consulta su empresa activa y sus trabajadores; intentar cambiar de tenant (A→B1 o B→A1) devuelve 403.
5. B consulta B1 y sus trabajadores; B→A2 devuelve estrictamente 403.

Si un caso falla, no continuar a pruebas manuales ni usar datos reales. El smoke comprueba un conjunto acotado de endpoints Core y no sustituye una matriz automatizada de autorización por objeto, tenant, rol y verbo HTTP.

Para health básico, ejecutar `./scripts/smoke-test.sh local`; valida únicamente `http://127.0.0.1:4001/v1/health`.

## Fase D — flujos web sintéticos manuales

Sólo después de que Fase A y Fase C pasen en un checkout limpio:

1. Iniciar el dashboard local siguiendo `frontend/web/README` y usar sólo cuentas `@zu360.test`.
2. Confirmar login contra Core, logout, refresh y mensajes de error genéricos. Cambiar de empresa desde el selector y comprobar que el nuevo token emitido por Core delimita el `company_id` activo. En Empresas, crear una compañía con nombre `ZU360 PILOTO SINTÉTICO`, ID `TEST-NIT-*`, CIIU `TEST`, representante ficticio y correo `@zu360.test`; editar sólo esos campos habilitados y archivar sólo una compañía sintética no activa, tras confirmación.
3. Desde A1/A2/B1, listar/crear/editar/eliminar perfiles de prueba y confirmar que toda solicitud usa la compañía activa. No subir CSV ni documentos reales.
4. Crear un incidente sin lesionados con ubicación `SITIO SINTÉTICO` y descripción `SIMULACIÓN:`; revisar, cerrar y previsualizar FURAT. La previsualización debe decir “NO VÁLIDO — SINTÉTICO” y no se transmite a ARL ni autoridad.
5. Crear un curso de demostración y registrar asistencia sólo con perfiles sintéticos de la compañía activa. Las constancias deben ser inválidas, sin QR ni URL pública, y no acreditan formación ni vigencia normativa.
6. Confirmar bloqueos de rutas no habilitadas y conservar errores visibles; nunca presentar mocks o datos de demostración como ejecución real.
7. Inspeccionar logs por request ID y comprobar que no impriman contraseñas, tokens, PII, datos médicos ni payloads innecesarios.

Registrar versión/commit, fecha, rol sintético, tenant/empresa ficticios, precondición, request/captura redactada, resultado esperado/real, severidad y responsable. No incluir contraseñas, tokens ni información personal.

## Gate de datos reales y producción — NO APROBADO

Antes de cualquier prueba con personas o empresas reales se requieren, como mínimo:

- Builds y pruebas integrales verdes en CI desde checkout limpio; dependencias, acciones e imágenes escaneadas.
- Tests negativos automatizados de todas las rutas/objetos relevantes, con dos tenants y escenarios de múltiples empresas por profesional, y smoke HTTP independiente contra PostgreSQL local.
- Revisión independiente de privacidad, membership, cifrado, retención, borrado, portabilidad y contratos con encargados.
- Política efectiva de egress para aislar el piloto sintético de proveedores externos; configuración de secretos segura y rotación probada.
- Revisión profesional/legal de cada regla SST; protocolos peligrosos y resultados IA permanecen como borrador con aprobación humana.
- Pentest/DAST y remediación; restore de backup medido, plan de incidentes, monitoreo y responsables.
- Entitlements y ciclo de suscripción comprobados sin cargos reales; webhooks de pago en sandbox.
- Consentimiento, aviso, minimización de datos y acuerdo de tratamiento aprobados antes de importar cualquier dato de una empresa piloto.

Un build o test unitario verde no sustituye estos gates ni constituye certificación de cumplimiento. No declarar producción GO ni habilitar datos reales hasta cerrar las aprobaciones externas y pruebas independientes.

## Anexo — autorregistro Google para un entorno real separado

El flujo de registro Google se implementó en Core y en la página de acceso, pero **no forma parte del piloto sintético y no queda habilitado por defecto**. `GOOGLE_SIGNUP_ENABLED=false`; el arranque falla si se intenta activarlo en modo sintético o en `development`. La configuración exige `ENV=staging` o `production`, un Client ID Web de Google, orígenes HTTPS, URLs HTTPS publicadas para Términos/Aviso de privacidad y versiones vigentes. No se debe inventar ni utilizar páginas legales provisionales como si estuvieran aprobadas.

El flujo comprueba localmente firma, audiencia, emisor, expiración, `iat`, correo verificado/autoritativo y un `nonce` aleatorio de un solo uso. Google `sub` es la clave federada; no se vincula automáticamente una cuenta contraseña existente por coincidencia de email. El primer registro exige datos declarados de empresa y dos confirmaciones legales versionadas; crea atómicamente usuario profesional SST, tenant UUID propio, empresa inicial y membership. No asigna un plan ni clasifica el riesgo ARL automáticamente. El NIT, CIIU, ARL, identidad de representante y existencia legal no se verifican contra fuentes oficiales.

Antes de hacer una prueba de staging, en un entorno limpio y dedicado:

1. Obtener aprobación profesional/legal y publicar los Términos y el Aviso de privacidad en dominios HTTPS controlados, con titular, fines, bases, retención, derechos, canales y versiones. Configurar versiones inmutables; no reutilizar el piloto sintético.
2. En Google Cloud Console, crear OAuth Client de tipo Web, registrar exactamente los orígenes JavaScript HTTPS y completar branding, dominio autorizado, soporte, privacidad y publicación/usuarios de prueba. Usar los scopes OIDC mínimos `openid email profile`; el Client ID es público y **no** equivale a un secreto.
3. Aplicar primero el esquema base revisado. Antes de la migración nueva, detectar duplicados case-insensitive de correo y NIT numéricamente normalizado, resolverlos manualmente y hacer backup probado. La migración crea los campos de identidad/aceptación e índices únicos:

   ```sql
   SELECT LOWER(email), COUNT(*) FROM users GROUP BY LOWER(email) HAVING COUNT(*) > 1;
   SELECT REGEXP_REPLACE(nit, '[^0-9]', '', 'g'), COUNT(*)
     FROM companies GROUP BY REGEXP_REPLACE(nit, '[^0-9]', '', 'g') HAVING COUNT(*) > 1;
   ```

   Si ambas consultas no devuelven filas y el backup es restaurable, aplicar con el pipeline de migraciones aprobado (no con `AUTO_MIGRATE` en producción):

   ```bash
   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f services/core/migrations/20261019_google_self_signup.sql
   ```

4. Proveer por secret manager/configuración de despliegue: `GOOGLE_SIGNUP_ENABLED=true`, `GOOGLE_CLIENT_ID`, `GOOGLE_SIGNUP_TERMS_URL`, `GOOGLE_SIGNUP_PRIVACY_URL`, `GOOGLE_SIGNUP_TERMS_VERSION` y `GOOGLE_SIGNUP_PRIVACY_VERSION`. Mantener `PILOT_SYNTHETIC_ENABLED=false`. Configurar lista `ALLOWED_ORIGINS` con orígenes exactos HTTPS y egress del Core sólo hacia Google para recuperar certificados OIDC; no desactivar la verificación TLS.
5. Asegurar Redis 6.2 o posterior para `GETDEL`, rate limits compartidos y tickets/nonce con TTL. Si existe un reverse proxy, la IP usada para rate-limit debe provenir de un proxy confiable; no aceptar `X-Forwarded-For` arbitrario. Aplicar también rate limit en el edge y monitoreo sin registrar tokens, email, NIT ni payloads.
6. Ejecutar pruebas E2E independientes en staging: primera alta/tenant aislado, login recurrente por Google, nonce expirado/replay, aud/iss inválidos, email no verificado, conflicto con login existente sin auto-link, email con diferencias de mayúsculas, duplicado NIT con puntuación distinta, rollback de la transacción ante fallos, aceptación de términos/privacidad y refresh/logout. Verificar que el usuario A no ve recursos B. No usar datos de trabajadores en estas pruebas.

**Estado del anexo:** implementación y compilación local verificadas; no se probó OAuth con credenciales reales, PostgreSQL/Redis, el nuevo SQL ni navegador contra un despliegue. En este sandbox no existe `docker`; tampoco se configuraron Client ID, URLs legales, dominios/orígenes ni secretos de staging. Por tanto la función permanece deshabilitada y esto **no cambia Gate 1/Gate 2 ni autoriza datos reales o producción**. La cuenta Google real contiene datos personales del profesional; la aprobación de finalidad, aviso y tratamiento debe completarse antes de habilitar el endpoint.

Fuentes oficiales consultadas para el flujo: [verificación server-side de Google ID tokens](https://developers.google.com/identity/gsi/web/guides/verify-google-id-token), [OpenID Connect de Google](https://developers.google.com/identity/openid-connect/openid-connect) y [configuración del Client ID](https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid). Aplicar estas guías junto con la revisión de seguridad de la implementación; el cumplimiento de OIDC no reemplaza las obligaciones de privacidad colombianas.
