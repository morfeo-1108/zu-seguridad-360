# Suite de Tests — ZU 360 (v11)

## Filosofía
Priorizamos los flujos donde un bug tiene consecuencias reales: seguridad de vuelo,
cumplimiento legal y resiliencia. Tests honestos: los que corren aquí están marcados;
los de integración (requieren BD) están escritos y corren en CI.

## ✅ Tests unitarios — EJECUTADOS Y PASANDO (61 casos)

### Go — drone/internal/flight/ (lo más crítico: vuela sobre personas) — 37 casos
- **safety_test.go** — geocerca, Haversine, reserva de batería, veredicto de seguridad.
- **flight_controller_test.go** — máquina de estados completa con dron simulado:
  requiere autorización PIC, aborta en pre-vuelo fallido, completa misión autorizada,
  aterrizaje de emergencia por batería crítica, RTH autónomo, **PIC override desactiva
  la autonomía**, registro de caja negra.
- **mission_planner_test.go** — ruta válida, límite de 120m, priorización.
- **preflight_test.go** — autoriza con estado óptimo; bloquea fallas críticas y clima NO-GO.
- **weather_test.go** — decisiones GO/NO-GO.

### Go — core/internal/ — 8 casos
- **resilience/circuit_breaker_test.go** — abre tras N fallas, recupera, resetea.
- **geo/geofence_test.go** — point-in-polygon para reconocimiento de zona por GPS.

### Python — ai/tests/ — 32 casos
- **test_biometric.py** — coseno, umbral, bloqueo sin consentimiento (Ley 1581), multi-tenant.
- **test_skills.py** — permisos por rol, confirmación de acciones destructivas.
- **test_compliance.py** — cumplimiento por trabajador, faltas críticas, % parcial.
- **test_orchestrator.py** — ruteo a agente correcto por intención.
- **test_inference.py** — severidad, eventos críticos, violaciones EPP, modo degradado.
- **test_roster.py** — las dos clases de autonomía; billing JAMÁS es LLM.

## 📝 Tests de integración — ESCRITOS, corren en CI (requieren BD)
Marcados con `//go:build integration`. No corren en sandbox sin acceso a módulos Go,
pero sí en CI con Postgres/SQLite disponible.
- **billing_idempotency_test.go** — evita doble cobro (hallazgo de auditoría #2).
- *(pendiente)* consent erasure con legal-hold (Ley 1581).

## Correr
```bash
./scripts/run-tests.sh                          # unitarios (rápido)
go test -tags=integration ./services/.../...    # integración (CI, con BD)
```

## Honestidad sobre cobertura
Esto cubre la LÓGICA DE DECISIÓN crítica. NO cubre aún: handlers HTTP de extremo a
extremo, el flujo completo dron→evidencia con BD real, ni el frontend. Es la base
correcta sobre la cual crecer, no cobertura total.
