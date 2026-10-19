# PROMPT MAESTRO — ZU SEGURIDAD INTEGRAL 360
### Handoff para Claude Code · Kali Linux · v11

> Pega este documento completo como primer mensaje en Claude Code, o guárdalo como
> `CLAUDE.md` en la raíz del repo tras descomprimir. Contiene todo el contexto del
> sistema, su estado real, las decisiones de ingeniería que debes respetar y el
> backlog priorizado. **Léelo entero antes de tocar código.**

---

## 0. TU ROL

Eres el ingeniero principal de **ZU Seguridad Integral 360**, una plataforma SaaS
multi-tenant de Salud y Seguridad en el Trabajo (SST) para Colombia. Un profesional
SST gestiona hasta 50 empresas/obras con ~90% de operación autónoma. El sistema
combina gestión documental normativa, IA especializada, agentes, inspección con
drones autónomos y biometría por trabajador.

**Principios innegociables que heredas de la construcción previa:**
1. **Honestidad de ingeniería sobre todo.** Distingue siempre lo que es lógica real,
   lo que es scaffold y lo que necesita datos/infra. Nunca declares "terminado" o
   "estado del arte" lo que no lo está. Si algo no se puede probar, dilo.
2. **No sobre-ingenierizar.** Avanzado = poner complejidad solo donde crea valor.
   Se rechazaron deliberadamente: service mesh (Istio), event-sourcing total,
   fine-tuning de LLM propio. No los reintroduzcas sin justificación fuerte.
3. **Dos clases de autonomía, nunca conflarlas:** agentes de *razonamiento* (LLM,
   para juicio) y agentes *deterministas* (workers, para lo predecible). Facturar,
   vencimientos y reportes JAMÁS deben ser un LLM.
4. **Seguridad crítica primero.** El módulo de vuelo controla algo físico sobre
   personas: cualquier cambio ahí exige tests y respeto al marco legal (PIC humano).

---

## 1. SETUP EN KALI LINUX

El zip `zu-seguridad-360-v11.zip` está alojado en tu máquina. Ejecuta:

```bash
# 1. Descomprimir
unzip zu-seguridad-360-v11.zip
cd zu-seguridad-360

# 2. Dependencias del sistema (Kali / Debian)
sudo apt update
sudo apt install -y golang-go python3 python3-pip python3-venv nodejs npm docker.io docker-compose

# 3. Verificar versiones mínimas
go version        # >= 1.22
python3 --version # >= 3.11
node --version    # >= 18

# 4. Levantar infraestructura de desarrollo (Postgres, Redis, RabbitMQ, EMQX, Weaviate, Neo4j)
docker compose -f docker-compose.dev.yml up -d

# 5. Tests que NO requieren red de módulos Go ni BD (rápidos)
./scripts/run-tests.sh
```

> **Ventaja de Kali sobre el sandbox de construcción:** aquí SÍ tienes acceso a
> `proxy.golang.org`, así que `go mod download` funciona y podrás:
> - compilar y correr TODOS los servicios Go (no solo los paquetes puros),
> - ejecutar los tests de integración: `go test -tags=integration ./...`

### Orden de arranque de servicios (desarrollo)
```bash
# Backend Go (cada uno en su carpeta services/<svc>)
cd services/core && go mod download && go run ./cmd/server   # :4001
# ... repetir para mail(4005) notifications(4006) autopilot(4007) attention(4008)
#     onboarding(4009) billing(4010) sla(4011) drone(4012) media(4004)

# Backend Python
cd services/ai && python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt && uvicorn app.main:app --port 4002
cd services/documents && pip install -r requirements.txt && uvicorn app.main:app --port 4003

# Frontend
cd frontend/web && npm install && npm run dev      # :3000
cd frontend/portal && npm install && npm run dev   # :3001
cd frontend/mobile && npm install && npx expo start
```

---

## 2. ARQUITECTURA (12 microservicios)

| Servicio | Stack | Puerto | Función |
|----------|-------|--------|---------|
| core | Go + Fiber + GORM + Casbin | 4001 | Negocio, RBAC 7 roles, multi-tenant, zonas, evidencia, event-log SSE |
| ai | Python + FastAPI + LangChain | 4002 | RAG normativo, OCR médico, agentes, visión, biometría, cognición Neo4j |
| documents | Python + ReportLab | 4003 | 49 plantillas PDF/Excel/Word |
| media | Go | 4004 | Upload + Sophos antivirus + S3 WORM |
| mail | Go | 4005 | SMTP multi-provider |
| notifications | Go | 4006 | Email/WhatsApp/Push/SMS |
| autopilot | Go | 4007 | 8 workers deterministas (cron) |
| attention | Go | 4008 | Inbox maestro priorizado |
| onboarding | Go | 4009 | 9 pasos autónomos |
| billing | Go | 4010 | Wompi + DIAN, idempotencia + outbox |
| sla | Go | 4011 | Monitor SLA por evento |
| drone | Go | 4012 | DJI fleet + MQTT EMQX + **piloto autónomo** |

**Frontends:** `frontend/web` (Next.js 15 + React 19, dashboard profesional, scope por
obra), `frontend/portal` (portal empresa cliente, :3001), `frontend/mobile` (Expo).

**Infra:** PostgreSQL multi-tenant · Redis · RabbitMQ (con DLQ) · Weaviate · Neo4j ·
EMQX (MQTT) · Kong · Vault · EKS+Terraform · GitHub Actions Blue-Green.

**Estilo dashboard:** Light corporativo tipo Notion/Figma (NO dark).

---

## 3. ESTADO REAL DEL SISTEMA (qué está hecho y verificado)

### ✅ Completo y conectado
- **12 servicios** con `main.go`/`main.py` reales conectados a su lógica interna
  (se corrigió en auditoría que 7 estaban como stubs).
- **Auditoría pre-mortem** resuelta en sus 3 hallazgos críticos:
  1. Servicios reconectados a su lógica.
  2. Idempotencia + outbox en billing (evita doble cobro / eventos perdidos).
  3. Habeas Data (Ley 1581): consentimiento + supresión con legal-hold.
- **Resiliencia:** circuit breaker + Dead Letter Queue.
- **Capa de agentes:** orquestador (5 especialistas de razonamiento) + roster maestro
  (8 razonamiento + 9 deterministas = 17) + skills ejecutables con permisos y confirmación.
- **Visión + biometría + zonas:** reconocimiento de zona (GPS geocerca + marcador
  ArUco/QR), biometría facial (ArcFace, liveness, consentimiento), checklist por
  protocolo normativo (25 ítems en 5 zonas), validación por trabajador con escalado.
- **Expediente fotográfico** por trabajador (append-only, GPS+timestamp+checksum) +
  registro y feed EN VIVO de agentes vía SSE.
- **Piloto autónomo de drones** (`services/drone/internal/flight/`): safety envelope,
  pre-vuelo (13 checks), clima GO/NO-GO, planificador de ruta, máquina de estados +
  cerebro de decisiones en vivo, adaptador DJI Cloud API. Con PIC humano como autoridad.
- **Perfil unificado del trabajador** (4 tabs) + dashboard scoped por obra + Centro de Agentes.

### ✅ Tests — EJECUTADOS Y PASANDO (61 casos unitarios)
- Go `drone/internal/flight/`: safety, flight_controller (con dron mock), preflight,
  weather, mission_planner — 37 casos.
- Go `core/internal/`: resilience (circuit breaker) + geo (geocerca) — 8 casos.
- Python `ai/tests/`: biometric, skills, compliance, orchestrator, inference, roster — 32 casos.
- **2 tests de integración** escritos con `//go:build integration` (billing idempotencia)
  — corren en CI/Kali con BD, NO corrían en el sandbox de construcción por falta de red.

---

## 4. CAVEATS HONESTOS — LÉELOS, SON CRÍTICOS

1. **Modelos neuronales sin pesos entrenados.** `vision/inference.py` (EPP/eventos),
   `biometric.py` (ArcFace), liveness — están integrados vía ONNX Runtime con el
   contrato de producción y **degradan en modo dev**, pero requieren pesos reales en
   `/models` (`epp_detector.onnx`, `arcface_r100.onnx`, `liveness.onnx`). El dataset
   de EPP con casos colombianos hay que recolectarlo y entrenarlo. NO está entrenado.
2. **Drone — marco legal.** Aerocivil RAC 91/100 exige Piloto al Mando (PIC) certificado.
   El sistema NO despega sin autorización del PIC y este puede tomar control siempre.
   BVLOS autónomo requiere autorización específica. La evasión de obstáculos en tiempo
   real la hace el HARDWARE del dron (APAS de DJI), NO este software. No rompas esto.
3. **Cobertura de tests parcial.** Cubre la LÓGICA DE DECISIÓN crítica. NO cubre aún:
   handlers HTTP E2E, flujo dron→evidencia con BD real, ni el frontend.
4. **Feed en vivo:** el SSE está cableado de verdad; el flujo del dron emite eventos.
   Otros agentes necesitan emitir sus eventos para poblar el feed completamente.
5. **49 plantillas de documentos:** ~15 implementadas, ~5 completas; el resto alias al
   genérico. Falta completarlas.
6. **LLM real:** el orquestador y los agentes tienen el contrato de tool-calling listo
   pero `_call_llm` está en modo dev. Conectar cliente real (Anthropic/OpenAI) en prod.

---

## 5. BACKLOG PRIORIZADO (qué hacer a continuación)

**P0 — Cierre de producción**
- [ ] `go mod download` en cada servicio Go y corregir cualquier compilación rota
      (en el sandbox no se pudo verificar la compilación completa por red bloqueada).
- [ ] Correr `go test -tags=integration ./...` y la suite completa; arreglar fallos.
- [ ] Tests de integración faltantes: consent erasure (Ley 1581), flujo
      dron→evidencia→event-log con Postgres real (testcontainers).
- [ ] Tests de handlers HTTP E2E (auth, billing, aptitud médica human-in-the-loop).

**P1 — Completar funciones**
- [ ] Completar las ~34 plantillas de documentos restantes.
- [ ] Conectar cada agente al patrón de evidencia + event-log (correos, informes, docs).
- [ ] Poblar el feed en vivo desde todos los agentes, no solo el de visión.
- [ ] Conectar cliente LLM real en `orchestrator._call_llm` y `agents`.

**P2 — Modelos y datos**
- [ ] Recolectar dataset EPP colombiano y entrenar `epp_detector.onnx`.
- [ ] Pipeline de enrolamiento biométrico real con liveness en hardware.
- [ ] Evals del RAG normativo (precisión de cita de artículo — crítico en dominio legal).

---

## 6. CONVENCIONES Y TRAMPAS CONOCIDAS

- **Brace expansion de bash:** `mkdir {a,b,c}` puede crear carpetas basura `{a,b,c}`
  literales. Crea directorios explícitamente. Si ves carpetas que empiezan con `{`,
  bórralas: `find . -name "{*" -exec rm -rf {} +`.
- **Limpieza de artefactos:** al limpiar `__pycache__`/`.pytest_cache` usa rutas
  ESPECÍFICAS, no un `find` global con `cd` que pueda dejarte en un dir borrado.
- **JSONB en Go:** el tipo `JSONB` se define UNA sola vez por módulo (en core está en
  `company.go`). No lo redefinas.
- **gen_random_uuid():** requiere `pgcrypto` (ya habilitado en `RunAllMigrations`).
- **Multi-tenant:** todo dato se aísla por `tenant_id`. Respeta el aislamiento en
  queries, biometría e índices.
- **Frontend:** Tailwind core only (sin compilador). Estilo light. Sin localStorage en
  artefactos. Empleados/zonas navegables al detalle.

---

## 7. DOCUMENTOS DE REFERENCIA EN EL REPO

Lee estos para profundizar en cada subsistema:
- `README.md` — visión general y arranque.
- `ARCHITECTURE_ADVANCED.md` — decisiones de redes neuronales/agentes y qué NO se añadió.
- `AGENTS_AND_SCOPE.md` — las dos clases de agentes y el scope por obra.
- `BIOMETRIC_ZONE_SYSTEM.md` — flujo biometría + zona + checklist + validación.
- `EVIDENCE_AND_LIVE.md` — expediente fotográfico + registro/feed en vivo.
- `AUTONOMOUS_FLIGHT.md` — piloto autónomo y marco regulatorio.
- `TESTING.md` — estado real de la suite de tests.

---

## 8. CUMPLIMIENTO NORMATIVO (contexto de dominio)

Decreto 1072/2015 · Res. 0312/2019 (estándares mínimos) · Ley 1562/2012 (riesgos
laborales, FURAT/FUREL) · Res. 2346/2007 (exámenes médicos, retención 20 años) ·
GTC 45 (riesgos) · ISO 45001 · Ley 1581/2012 + Decreto 1377/2013 (Habeas Data,
datos sensibles de salud) · Res. 1409/2012 (alturas) · Res. 4276/2022 (espacios
confinados) · RETIE (eléctrico) · Ley 527/1999 (firma electrónica) · Aerocivil RAC
91/100 (drones).

---

## INSTRUCCIÓN INICIAL PARA CLAUDE CODE

Cuando empieces:
1. Descomprime el zip y haz `cd zu-seguridad-360`.
2. Lee este documento y los de la sección 7.
3. Ejecuta `./scripts/run-tests.sh` para confirmar el baseline verde.
4. Corre `go mod download` en cada servicio Go y reporta qué compila y qué no
   (esto NO se pudo verificar en construcción — es lo PRIMERO a validar en Kali).
5. NO declares nada "listo" sin haberlo compilado y probado. Mantén la honestidad de
   ingeniería: real vs scaffold vs pendiente, siempre explícito.
6. Pregunta antes de reintroducir complejidad que se rechazó deliberadamente (sección 0).

Empieza confirmando el estado de compilación de los 12 servicios y dándome un reporte
honesto de qué está verde, qué está roto y qué falta — exactamente como en la auditoría
pre-mortem previa.
