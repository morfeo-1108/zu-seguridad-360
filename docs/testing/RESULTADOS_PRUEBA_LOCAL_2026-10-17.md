# Resultados de pruebas locales — ZU Seguridad 360

**Fecha:** 19 de octubre de 2026  
**Entorno:** sandbox aislado, sin Docker ni datos de usuario/empresa.  
**Conclusión:** **PASS para preparación de pruebas locales sintéticas; NO-GO para producción, licencias pagas o datos personales/sensibles.**

**Actualización:** 19 de octubre de 2026 — además del guard local Docker, se añadieron cierres fail-closed a rutas y servicios Core de trabajadores, incidentes y capacitaciones; se cerraron AI/Documents y se neutralizaron citas normativas no verificadas. El primer intento usando Python global no encontró `pytest`; la suite final, con `PATH=/tmp/zu360-test-venv/bin:$PATH`, completó todas las pruebas.

## Verificaciones ejecutadas

| Verificación | Resultado | Evidencia y alcance |
|---|---|---|
| `PATH=/tmp/zu360-test-venv/bin:$PATH make test` | **PASS** | Los 10 módulos Go (`core`, `mail`, `notifications`, `autopilot`, `attention`, `onboarding`, `billing`, `sla`, `drone`, `media`) completaron `go test ./... -count=1`. Los paquetes sin tests sólo se compilaron; no equivalen a cobertura funcional. |
| Tests Python AI | **PASS: 42** | `pytest tests/ -v --asyncio-mode=auto` en el virtualenv temporal; incluye gates de privacidad y errores genéricos, sin PyTorch ni modelos de inferencia. |
| Tests Python Documents | **PASS: 10** | Sólo prueba el gate local de rutas; verifica que únicamente `/health` está permitido y que generación, reportes, certificados/OpenAPI quedan cerrados. No sustituye prueba HTTP del servicio en contenedor. |
| `go vet ./...` | **PASS** | Ejecutado en los 10 módulos Go. |
| `make build` | **PASS** | Compila los 10 módulos Go y ejecuta `compileall` para AI y Documents. Media usa automáticamente la toolchain Go 1.24.0 declarada en su `go.mod`. |
| `scripts/sync-go-deps.sh --clean-cache` | **PASS** | Se limpió la caché global Go; los 10 módulos ejecutaron `go mod tidy`, `go mod verify` y `go build ./...`; todos generaron/verificaron `go.sum` y compilaron. |
| Aislamiento company A→B unitario | **PASS** | Middleware Fiber devuelve 403 cuando el token A solicita la ruta de empresa B; casos positivos prueban que header/query/body no sobrescriben el scope firmado. Esto no es una prueba con DB ni cubre todas las rutas. |
| Sintaxis shell y smoke de aislamiento | **PASS estático** | `bash -n` sobre scripts Bash; el nuevo smoke HTTP Python compila, pero no se pudo ejecutar sin servicios Docker. |
| Guard Docker local/remote | **PASS (18 oct.)** | `test-local-docker-guard.sh` aceptó sockets locales simulados y rechazó endpoint TCP y `DOCKER_HOST`; `make test` ahora ejecuta esta prueba. No se contactó ningún daemon real. |
| Suite unitaria completa (19 oct.) | **PASS** | Los diez módulos Go, 42 tests AI, 10 tests Documents y el guard Docker pasaron usando el virtualenv aislado, incluyendo negativas de servicios Core antes de DB/colas. |
| Sintaxis Python | **PASS** | `python3 -m compileall -q services/ai/app services/documents/app`. |
| Sintaxis YAML | **PASS** | Parseo de `docker-compose.dev.yml`, los workflows de GitHub Actions, Kong y Prometheus. Las 12 referencias de `env_file` tienen plantilla `.env.example`. |
| Referencias estáticas de Compose | **PASS estático** | Las 24 definiciones de servicio, 12 referencias build/Dockerfile y dependencias directas resuelven; todos los puertos publicados aparecen enlazados a loopback. No valida el esquema oficial de Docker Compose. |
| `docker compose config`, build de imágenes y stack | **NO EJECUTADOS** | El sandbox no tiene Docker Engine ni plugin Compose. Se instaló Podman rootless para intentar probar el runtime: `hello-world` funcionó, pero las imágenes del stack no pueden extraerse porque el espacio de usuario del sandbox sólo mapea un UID/GID y no admite los IDs que requieren las capas; por tanto no se construyó ni inició ningún servicio del sistema. El parseo YAML/podman-compose no sustituye Docker Compose. |

## Hardening que limita el piloto

- Billing arranca en **health-only**: se retiraron migraciones automáticas, cron y rutas financieras sin autenticación. No probar suscripciones ni pagos con esta copia.
- JWT ahora incluye company_id firmado desde usuario y comprueba pertenencia company→tenant al login/refresh. CompanyScope comprueba scope antes del cierre de privacidad: acceso A→B sigue en 403; las APIs laborales propias responden 404 hasta cerrar Gate 2.
- Evidencia, PHVA por objeto y eventos de agente están sin registrar porque aún no se ha probado ownership por objeto. El seed ahora crea sólo dos cuentas/empresas sintéticas en tenants distintos.
- AI requiere aprobación de tratamiento además de sus flags; OCR, biometría y todas las rutas de inspección por zona quedan bloqueadas aun con flags activas. Errores AI devuelven respuesta genérica y el fallback OCR no incluye detalle del proveedor. Drone y Autopilot son health-only.
- EmployeeService, IncidentService, TrainingService, exámenes y exportación fallan cerrados antes de DB/colas; también se bloquean certificados/FURAT. Captura/evaluación de consentimiento y revocación no se aceptan hasta integrar evidencia/política y actor con auditoría. Supresión sigue bloqueada hasta revisar schema y retención.
- Documents permite únicamente `/health`; generación, certificados, reportes y OpenAPI responden 404. El bootstrap local añade trigger contra UPDATE/DELETE/TRUNCATE en `audit.event_log`; no se verificó contra PostgreSQL ni equivale a migración productiva.
- Las herramientas ejecutables AI empiezan vacías; el test usa sólo un handler sintético local. Skills y protocolos normativos de alto riesgo siguen deshabilitados; las referencias conocidas no verificadas se marcan pendientes, sin reemplazo supuesto. Exportación general y verificación pública de certificados permanecen deshabilitadas.
- El setup local crea datos sintéticos e inicia servicios de desarrollo. No cambies `.env.dev` para apuntar a staging/producción ni cargues información real.

## Próximo paso para el equipo

En una máquina controlada con Docker Engine y Docker Compose v2, crear el venv indicado en el README, correr primero la suite, revisar el diff y luego ejecutar `make setup-core-test` para levantar sólo Postgres + Core con dos profesionales/empresas sintéticos. Después, correr `make test-company-isolation`: A→A y B→B deben dar 200 para empresas; A→B debe dar 403; acceso propio a `/employees` debe dar 404 y A→B a `/employees` debe dar 403. En este sandbox no se pudo ejecutar esa prueba HTTP porque Docker no está instalado. Cuando esa etapa esté validada, el stack backend opcional es `make setup`; los perfiles high-risk siguen desactivados. Confirmar `/v1/health` con `./scripts/smoke-test.sh local`; dicho smoke test no valida los otros servicios ni los flujos funcionales.

Antes de introducir cualquier dato no sintético siguen siendo obligatorios: cobertura negativa de todos los objetos/endpoints (incluyendo dos empresas dentro de un tenant), membership multiempresa, privacidad/consentimiento y retención aprobados, pruebas de integración y restore, CI limpia, seguridad externa y revisión normativa por profesionales competentes. No se declara habilitado ningún uso real.
