# Estado de preparación Gate 1 y Gate 2 — Piloto sintético

**Fecha de corte:** 19 de octubre de 2026  
**Decisión técnica:** se puede preparar un piloto funcional con datos exclusivamente sintéticos en un stack local; el sistema sigue en **NO-GO para producción y para cualquier dato real de trabajadores o empresas**.

## Decisión por gate

### Gate 1 — aislamiento multi-tenant: PARCIAL, NO-GO para datos reales

Core valida JWT firmado, extrae `tenant_id`, `user_id` y `company_id` de sus claims, y exige que la empresa del recurso coincida con la compañía activa del token. Un middleware posterior verifica una membership persistida y vigente. El seed representa dos profesionales en tenants separados, con dos memberships y dos empresas sintéticas por profesional. El selector cambia el contexto mediante `switch-company`, que sólo emite un token nuevo si el usuario tiene la membership.

Hay pruebas unitarias de denegación de scope A→B, incluido el intento de modificar scope mediante headers, query o body. El smoke HTTP fue actualizado para recorrer ambas empresas, el cambio de scope y las rutas sintéticas de trabajadores. **No se ejecutó contra PostgreSQL/Core reales** porque el sandbox no tiene Docker (`docker: command not found`). Por ello, todavía no hay evidencia E2E de autorización por objeto ni de persistencia de las denegaciones.

### Gate 2 — habilitación funcional: ABIERTO CON MITIGACIÓN TÉCNICA, NO-GO para producción

El Core y el dashboard permiten probar localmente los flujos implementados: altas/edición/archivo lógico de compañías de prueba, perfiles sintéticos, incidentes sin lesionados y capacitación con constancias inválidas. Los registros y formularios requieren marcadores reservados; el modo sintético sólo se habilita con configuración de servidor y stack de desarrollo/pruebas local. El servidor bloquea operaciones con compañías no marcadas como sintéticas en los flujos de empresa, y el score de riesgo ilustrativo no se ofrece en este modo.

AI externa, OCR médico, visión/biometría, drones/autopilot, expedientes médicos, documentos/PHVA, verificación pública, pagos/cobros y exportación divulgativa no quedan funcionales. Las pantallas de bloqueo informan que no ejecutan la función ni simulan éxito. No se emitió revisión jurídica ni dictamen de cumplimiento para Resolución 0312 de 2019 o Decreto 1072 de 2015.

## Evidencia de validación local

- `make test`: pasó en diez módulos Go, 42 pruebas de AI, 10 pruebas de privacidad de Documents y el guard de Docker local simulado.
- `make build`: pasó para los diez módulos Go y la compilación sintáctica de Python.
- `npm run typecheck`: pasó.
- `npm run build`: pasó con Next.js 15.5.28 y generó las 24 páginas del App Router.
- `python3 -m py_compile scripts/smoke-company-isolation.py` y `bash -n scripts/seed.sh scripts/setup-core-test.sh`: pasaron.

Son pruebas unitarias, de compilación y sintaxis. No demuestran que el stack de servicios esté integrado, que el navegador complete los flujos ni que el comportamiento de autorización funcione frente a PostgreSQL, Redis y RabbitMQ levantados.

## Riesgos que siguen abiertos

1. **Integración no observada:** no se ejecutó `make setup-core-test` ni `make test-company-isolation`; el smoke de A/B está preparado, pero pendiente de una estación con Docker Engine/Compose v2.
2. **Egress de contenedores:** publicar puertos locales y validar `DOCKER_HOST` no es una regla de bloqueo de salida. El bridge estándar no demuestra que Core no pueda alcanzar Internet. Antes de considerar datos reales, se debe definir y probar una política de egress independiente.
3. **Cobertura de autorización:** las pruebas unitarias actuales no sustituyen una matriz E2E de rol, tenant, empresa, identificador de objeto y verbo HTTP para cada ruta sensible.
4. **Validación externa:** falta revisión independiente de privacidad, tratamiento de datos, retención/borrado, contratos con encargados, pentest/DAST, recuperación de backup y aprobación profesional/legal de reglas SST.
5. **Alcance incompleto:** CSV, exámenes médicos, PHVA, documentos, eventos de agente, notificaciones regulatorias y certificados consultables no forman parte del conjunto funcional habilitado. Deben mantenerse bloqueados hasta implementar sus controles y pruebas.

## Paso controlado al entorno Docker

En una máquina de pruebas local, dedicada y no compartida con producción, revisar primero el [Runbook de pruebas controladas](../testing/RUNBOOK_PRUEBAS_CONTROLADAS.md). El procedimiento comienza con `make setup-core-test`, que configura `ENV=development`, `PILOT_SYNTHETIC_ENABLED=true`, `DB_HOST=postgres` y la base `zu_core_synthetic`. El seed requiere confirmar la base local y usa exclusivamente cuentas `@zu360.test` con contraseñas de prueba ingresadas en la terminal.

Cuando Core y PostgreSQL estén sanos, ejecutar `make test-company-isolation`. Para aprobar esta etapa, el test debe demostrar que A sólo ve A1/A2, B sólo ve B1/B2, el switch emite tokens de la empresa autorizada y todos los accesos A↔B devuelven exactamente 403. Después deben probarse persistencia, operaciones CRUD sintéticas y denegaciones sin token. Registrar salidas redactadas; nunca guardar contraseñas ni tokens.

La ejecución en Docker no equivale a autorización para cargar datos reales. Gate 1 y Gate 2 seguirán en NO-GO hasta completar pruebas independientes, cierre de egress y aprobaciones externas.

## Archivos operativos actualizados

- [Runbook de pruebas controladas](../testing/RUNBOOK_PRUEBAS_CONTROLADAS.md)
- [TODO con estados y criterios](../../TODO.md)
- `scripts/smoke-company-isolation.py` — smoke de membership, `switch-company` y denegación A/B
- `scripts/seed.sh` — dos profesionales, cuatro empresas, cuatro memberships y cuatro perfiles de trabajador sintéticos


## Anexo 19 de octubre de 2026 — acceso Google aislado del piloto

Se incorporó una ruta separada, inactiva por defecto, para autoregistro y acceso con Google Identity Services. Core valida el ID token firmado con la biblioteca oficial de Google (`aud`, `iss`, `exp`, `iat`, `email_verified`, dominio autoritativo y `nonce` de un solo uso); no vincula cuentas existentes sólo por email. Un alta nueva crea transaccionalmente usuario SST, tenant UUID propio, empresa inicial y membership. Términos/aviso se aceptan por separado y se guardan con versiones/timestamps. No se asignan plan ni clasificación de riesgo; NIT/CIIU/ARL son datos declarados, no verificados. Los endpoints exigen Origin permitido, límites de cuerpo y cuota Redis compartida; la verificación de los certificados Google requiere egress explícito.

La pantalla de login sólo presenta Google cuando el API indica configuración habilitada. El flag sigue en `false` en la plantilla y `PILOT_SYNTHETIC_ENABLED=true` rechaza el registro. Se corrigió la telemetría de Core para no asumir claims de tenant en rutas públicas.

**Validación observada tras el cambio:** `go test ./...`, `go build ./...` y `go mod verify` del Core pasaron; `make test` pasó tras anteponer el virtualenv de pruebas Python (`42` pruebas AI, `10` de Documents y los módulos Go); `make build`, `npm run typecheck` y `npm run build` pasaron. El build produjo las 24 páginas del App Router. `docker`/`docker compose` no existen en este sandbox, así que no se aplicó la migración ni se ejecutó Google OAuth real, Redis/Postgres, navegador E2E o el smoke A/B.

**Conclusión:** esto amplía el código y la interfaz, pero **no abre Gate 1 ni Gate 2**. Falta Client ID/orígenes verificados en Google Cloud, términos y aviso reales aprobados/publicados, preflight de duplicados y aplicación de migración SQL, ambiente staging, configuración de TLS/egress/proxy confiable, pruebas E2E/replay/rollback y revisión independiente de privacidad/seguridad. Las cuentas Google reales contienen datos personales del profesional; no se deben habilitar hasta autorizar expresamente el tratamiento. Se mantienen bloqueados datos de trabajadores, funciones médicas, IA, biometría y drones.
