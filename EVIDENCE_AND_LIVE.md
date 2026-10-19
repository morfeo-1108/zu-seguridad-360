# Expediente fotográfico + Registro/Feed en vivo (v8)

## 1. Registro + EN VIVO de agentes (no maqueta)

- **Registro (append-only)**: tabla `agent_events`. Cada decisión/acción de un agente
  queda como bitácora auditable, filtrable por proyecto y por trabajador.
- **En vivo (push, sin polling)**: endpoint SSE `GET /v1/agent-events/stream`. El
  `EventLogService` mantiene un bus pub/sub en memoria; al registrar un evento lo
  empuja a los suscriptores. El frontend abre un `EventSource` y recibe push.
- **Fallback**: si SSE no conecta, el hook `useAgentEvents` hace polling cada 15s.

## 2. Expediente fotográfico por trabajador

Tabla `evidence_records` (append-only, ligada a `employee_id`). Cada validación del
dron, inspección, incidente, entrega de EPP, enrolamiento o capacitación genera una
prueba con:
- foto en S3 + checksum (integridad)
- georreferenciación (GPS) + timestamp exacto
- ítem de checklist validado + resultado (cumple/no) + severidad
- fuente (drone/mobile/manual)

### Caso de uso clave: prueba ante la ARL
`GET /v1/employees/:id/compliance-proof?code=arnes&date=2026-06-15`
→ devuelve la foto georreferenciada que demuestra que el trabajador usaba arnés ese
día. Validación exacta ante cualquier auditoría o imprevisto.

## 3. Flujo automático
```
Dron valida zona → por cada trabajador:
   ├─ POST /v1/evidence          (prueba fotográfica por ítem de checklist)
   └─ POST /v1/agent-events      (evento "Visión validó a X: 67%, falta arnés CRÍTICO")
                                  → se persiste (registro) Y se empuja por SSE (en vivo)
```

## 4. Dashboard: manejo por proyecto y por trabajador
- Selector de obra global (header) → todo se filtra por proyecto.
- Perfil unificado del trabajador: resumen, expediente fotográfico (galería con
  foto/GPS/hora), actividad de agentes sobre él, y expediente legal exportable.
- Empleados clicables desde el listado al perfil.

## Nota honesta de alcance
"Manejo de absolutamente cada función" es una dirección arquitectónica, no una línea
de meta alcanzada en una iteración. Lo construido es el PATRÓN que lo hace posible:
modelo de evidencia universal + event log universal + perfil agregador. Conectar cada
módulo restante (correos, cada documento, cada informe) a este patrón es mecánico y
queda como trabajo incremental claro.
