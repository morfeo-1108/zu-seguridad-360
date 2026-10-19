# Piloto Autónomo de Drones (v9)

## El sistema vuela como un piloto profesional

El módulo replica la secuencia mental completa de un piloto profesional, encodeada
en 6 componentes que trabajan juntos:

| Componente | Qué hace (como un piloto) |
|-----------|----------------------------|
| `safety.go` | Define los límites que jamás se cruzan: 120m AGL legal, geocerca, reserva de batería para volver, viento máximo. Evalúa seguridad en cada ciclo. |
| `preflight.go` | Checklist de 13 verificaciones antes de despegar (batería, GPS, brújula, IMU, firmware, home point, clima). Si algo crítico falla, NO se vuela. |
| `weather.go` | Decisión GO/NO-GO meteorológica (viento, ráfagas, lluvia, visibilidad, temperatura para baterías LiPo). |
| `mission_planner.go` | Ruta óptima: prioridad + nearest-neighbor para minimizar distancia, perfil de altitud de crucero, ángulos de gimbal por inspección, estimación de batería/tiempo. |
| `flight_controller.go` | El cerebro: máquina de estados (PREVUELO→AUTORIZACIÓN→DESPEGUE→EN RUTA→INSPECCIÓN→RTH→ATERRIZAJE) + motor de decisiones que corre en cada frame de telemetría y dispara RTH/emergencia/hover. |
| `dji_commander.go` | Traduce decisiones a comandos DJI Cloud API vía MQTT. |

## Decisiones en vivo (el "cerebro" durante el vuelo)
En cada frame de telemetría, `OnTelemetry` evalúa el envelope y decide como un
piloto vigilante:
- Batería en reserva mínima → **RTH autónomo**
- Batería crítica (≤10%) → **aterrizaje de emergencia**
- Fuera de geocerca / excede 120m / viento alto → **RTH**
- Anomalía no fatal → **hover de precaución** y vigila

Todas las decisiones quedan en una **caja negra** (PilotDecision) para auditoría.

## Honestidad regulatoria y de seguridad (IMPORTANTE)
1. **PIC obligatorio**: Aerocivil (RAC 91/100) exige un Piloto al Mando certificado.
   El sistema NO despega sin autorización del PIC (con nombre + licencia), y el PIC
   puede **tomar control manual** en cualquier momento (`TakeControl`).
2. **La evasión de obstáculos la hace el HARDWARE** del dron (sensores APAS de DJI),
   no este software. La seguridad crítica no depende solo de la red: el dron tiene
   sus propios sistemas redundantes.
3. **BVLOS autónomo** (más allá de línea visual) requiere autorización específica de
   Aerocivil. El sistema soporta la operación; el permiso es responsabilidad del operador.

Esto NO es un piloto que reemplaza la responsabilidad legal humana. Es un sistema que
**decide y ejecuta con criterio profesional bajo la autoridad de un PIC**.

## Endpoints
- `POST /v1/flight/plan` — genera plan + pre-vuelo + dictamen de clima (no vuela)
- `POST /v1/flight/:missionId/authorize` — el PIC autoriza (requiere licencia) y arranca
- `POST /v1/flight/:missionId/take-control` — el PIC toma control manual
- `GET  /v1/flight/:missionId/status` — fase actual + caja negra de decisiones
