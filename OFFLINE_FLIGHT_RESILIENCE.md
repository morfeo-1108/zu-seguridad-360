# Operación de Drones Sin Red / Señal Intermitente (v16)

## El problema que se resolvió
El diseño previo asumía que el "cerebro" del piloto recibía telemetría por red
(MQTT). Si la señal se perdía, el dron quedaba sin decisiones. Para volar sobre
personas en zonas sin cobertura, eso era una brecha real. Ya está resuelta.

## Principio de ingeniería (cómo vuela un piloto profesional sin cobertura)
El dron NO depende del enlace de red para hacer su trabajo ni para volver:
1. **La misión se carga COMPLETA a bordo antes de despegar** (ruta, geocerca,
   política de pérdida de enlace, punto de despliegue). Si no se confirma la
   carga, el vuelo no inicia.
2. **Navega por GNSS** (GPS/GLONASS/Galileo/BeiDou) de forma autónoma: cada
   waypoint es una coordenada absoluta que el aparato alcanza con su propio fix.
3. **El retorno al punto de despliegue es por coordenadas GNSS**, no por red: el
   dron siempre sabe dónde está (satélite) y dónde está su home (coordenada fija).
4. La red solo transporta telemetría y comandos de conveniencia — nunca es punto
   único de falla.

## Componentes (services/drone/internal/flight/)
- **link_resilience.go**: `LinkMonitor` (detecta OK / intermitente / perdido por
  timeout de heartbeat) y la escalera de failsafes `LinkLossPolicy` con 4 acciones:
  continuar misión onboard, RTH inmediato, hover-luego-RTH, o aterrizar en sitio.
- **onboard_navigation.go**: `OnboardMission` (paquete autónomo cargado al dron),
  `ValidateForOfflineOps` (no se vuela sin home, geocerca, ruta y política a bordo),
  y `GNSSNavigator` (rumbo y distancia a home por coordenadas, sin red).
- **flight_controller.go**: `EvaluateOnboardLinkLoss` modela la lógica de failsafe
  que corre A BORDO del dron (en el Onboard SDK/firmware), independiente del servidor.

## Escenarios cubiertos (todos con test que pasa)
| Situación | Comportamiento |
|-----------|----------------|
| Enlace intermitente | El dron sigue su misión onboard sin interrupción |
| Enlace perdido + GNSS OK + política "continuar" | Termina la misión y RTH por GNSS |
| Enlace perdido + política "RTH" | Retorno inmediato al punto de despliegue por coordenadas |
| Enlace perdido + GNSS degradado | Aterrizaje controlado en sitio (RTH no seguro) |
| Sin home point / sin carga onboard | El pre-vuelo BLOQUEA el despegue |

## Honestidad de ingeniería
- En el despliegue real, `EvaluateOnboardLinkLoss` se ejecuta en el **firmware /
  Onboard SDK del dron**, no en el servidor. Aquí se modela y prueba para que el
  comportamiento sea correcto y verificable; la integración con el SDK del
  fabricante (DJI Onboard SDK / MSDK) es trabajo de campo con el hardware.
- La evasión de obstáculos en tiempo real sigue siendo del hardware (APAS DJI).
- Todo esto opera bajo la autoridad del PIC humano (Aerocivil RAC 91/100).
