# ZU Seguridad Integral 360 — entorno de pruebas

> **Estado:** copia de hardening para pruebas controladas. **No está certificada para producción** ni debe recibir historias médicas, datos personales reales, credenciales de clientes o pagos reales. La amplitud del repositorio y sus claims originales no equivalen a funciones integradas/verificadas.

## Propósito y alcance

El producto busca ayudar a profesionales de Seguridad y Salud en el Trabajo (SST) a operar procesos y varias empresas. El código contiene servicios Go/Python y clientes web/móvil; cada módulo se debe tratar como experimental hasta superar pruebas funcionales y de seguridad. No se promete cumplimiento normativo automático, operación autónoma, reducción porcentual, capacidad de escalar a un número determinado de empresas, firma legal, facturación electrónica ni compatibilidad de publicación en tiendas.

En esta copia, Billing queda en health-only; AI API, agentes, visión, biometría/OCR, Drone y Autopilot quedan apagados o health-only. El scope JWT de empresa sigue comprobándose antes de los cierres de privacidad. Mientras Gate 2 permanezca abierto, las rutas de empleados, incidentes, capacitaciones y OCR-RUT responden 404; un cruce A→B sigue respondiendo 403 antes de ese cierre. El servicio Documents también devuelve 404 para sus APIs de generación, certificados y reportes sin autenticación por objeto. Los servicios de exámenes y exportación fallan cerrados; evidencia, PHVA por objeto y eventos de agente siguen sin rutas. También permanecen deshabilitados protocolos SST de alto riesgo, supresión y verificación pública de certificados. La evaluación PHVA no afirma cumplimiento automático. Ver [estado de hardening](docs/production/ESTADO_HARDENING_2026-10-17.md), [gates para datos reales](docs/production/GATES_DATOS_REALES_TRABAJADORES.md), [resultados de pruebas locales](docs/testing/RESULTADOS_PRUEBA_LOCAL_2026-10-17.md) y [opinión técnica](RESPUESTA_AUDITORIA_ELITE.md).

## Arquitectura declarada

El árbol contiene servicios para core, IA, documentos, media, mail, notificaciones, autopilot, atención, onboarding, billing, SLA y drones, además de frontends web, portal y móvil. Esta lista describe directorios y propósito pretendido; **no significa que cada servicio compile, se autentique entre sí o esté listo para operar**. Docker Compose sólo define servicios backend/infraestructura; no levanta los frontends.

| Área | Tecnologías declaradas | Nota de prueba |
|---|---|---|
| Core | Go/Fiber | Scope JWT company_id en empresas/empleados/incidentes/capacitaciones; resto de superficies sigue en revisión |
| IA y documentos | Python/FastAPI | API AI cerrada por defecto; agentes/visión/OCR con flags separados, sin claves de producción |
| Billing y notificaciones | Go | No usar pagos, mensajes ni webhooks reales durante pruebas |
| Drones/visión | Go/MQTT | Drone y Autopilot sólo health; no hay rutas de vuelo/triggers en este build |
| Web/portal/móvil | Next.js/React Native | No incluidos en Compose; validar aparte antes de usuarios reales |

## Requisitos para pruebas locales

- Go 1.24+, Python 3.12, `pip`, Docker Engine y Docker Compose v2. El sandbox de auditoría no tiene Docker instalado, por lo que aquí no se pudo arrancar Compose.
- Para tests Go, las sumas de módulos `go.sum` están versionadas. Para tests Python, crear un entorno virtual e instalar `services/ai/requirements-test.txt` (no descarga PyTorch/modelos) y ejecutar desde el repositorio.
- Sólo datos sintéticos. Los `.env.example` y contraseñas de Compose son exclusivamente locales y jamás se deben reutilizar en staging/producción.

## Preparar entorno local

Todos los scripts que levantan, migran o siembran el stack y los targets Make que invocan Docker directamente ejecutan `scripts/assert-local-docker.sh`. Sólo se permiten endpoints locales `unix://` o `npipe://`; se rechaza `DOCKER_HOST` explícito y cualquier endpoint TCP/SSH. El guard no acredita que el daemon local sea seguro ni protege scripts externos al repositorio.

Primero corre unit tests sin levantar servicios. `setup.sh` no es sólo un copiador de plantillas: valida Compose, inicia infraestructura, aplica SQL, levanta Core, pide la contraseña de seed sintético y luego construye/levanta el stack local.

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r services/ai/requirements-test.txt

# Suite Go + Python; falla ante cualquier test fallido.
# También prueba el guard Docker con un CLI simulado; no contacta ningún daemon.
make test

# Opcional: borra la caché global Go antes de regenerar go.sum/verificar/compilar.
make sync-go-deps-clean

# Primera prueba funcional recomendada: sólo Postgres + Core, migra y crea seed sintético.
make setup-core-test

# Prueba HTTP real A→B con las dos cuentas sintéticas; solicita contraseñas sin eco.
make test-company-isolation

# Cuando Core esté validado, iniciar el stack backend/infraestructura completo.
make setup
```

Consulta el [runbook de pruebas controladas](docs/testing/RUNBOOK_PRUEBAS_CONTROLADAS.md) para escenarios, límites y gates de aceptación.

El seed crea sólo dos profesionales sintéticos en tenants/empresas aislados: `sst-a@zu360.test` (empresa `22222222-2222-4222-8222-222222222222`) y `sst-b@zu360.test` (empresa `bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb`). Introduce dos contraseñas distintas y únicas al ejecutar el seed; no se guardan en el repositorio ni se imprimen. No crea empleados, historias médicas, accidentes ni información de personas reales. El seed exige la marca de DB local y confirmación explícita; nunca lo ejecutes contra una base remota.

Para sólo validar que los archivos Compose son consistentes, ejecuta `docker compose -f docker-compose.dev.yml config --quiet`. Los puertos publicados quedan enlazados a `127.0.0.1`. `make dev-down` detiene servicios sin borrar volúmenes; borrar datos persistidos exige el target explícito `make reset-dev-volumes CONFIRM_DEV_VOLUME_RESET=YES`.

## Qué no hacer durante el piloto

- No usar nombres, NIT, documentos, diagnósticos, exámenes, fotos, ubicaciones ni accidentes reales.
- No cargar credenciales de OpenAI/Anthropic, correo, Wompi/DIAN, AWS ni producción.
- No conectar vuelos/drones, enviar avisos a trabajadores ni ejecutar webhooks de cobro.
- No publicar las imágenes ni desplegar los workflows a staging/producción sin revisión y aprobación.
- No afirmar que el software certifica cumplimiento SG-SST ni sustituye al profesional competente.

## Documentación operativa

- [Checklist de producción](docs/production/CHECKLIST_LISTO_PARA_PRODUCCION.md)
- [Plan maestro de hardening y lanzamiento](docs/production/PLAN_MAESTRO_HARDENING_Y_LANZAMIENTO.md)
- [Comparativa de software SG-SST Colombia](docs/market/COMPARATIVA_COMPETIDORES_SG-SST_COLOMBIA.md)
- [Manejo de skills v3](docs/skills/INTEGRACION_SEGURA_SKILLS_V3.md)
- [Verificación normativa inicial](docs/skills/VERIFICACION_NORMATIVA_INICIAL_COLOMBIA.md)

## Criterio

Las pruebas locales comprueban comportamiento del código, no certifican cumplimiento legal, privacidad, seguridad de infraestructura ni disponibilidad. La salida a piloto con datos reales requiere gates separados de aislamiento multiempresa, revisión normativa/privacidad, restore de backup, pentest, respuesta a incidentes y aprobación contractual.
