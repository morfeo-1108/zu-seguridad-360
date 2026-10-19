# ZU 360 como App Instalable (PWA)

## Qué es
La aplicación web ZU 360 ahora es una **PWA (Progressive Web App)**: se puede
"instalar" en el escritorio (Windows/Mac/Linux) o en el móvil, abriéndose en su
propia ventana con ícono propio, sin la barra del navegador. Se siente como una
app de escritorio, pero es la misma web app — sin duplicar código ni instaladores.

## Qué se implementó (frontend/web/)
- **public/manifest.json** — identidad de la app instalable: nombre, íconos de
  marca ZU 360, color de tema (verde lima), display standalone y 3 accesos
  directos (Inbox, PHVA, Drones).
- **public/sw.js** — Service Worker con estrategia segura:
  - Navegación: network-first, cae al shell/offline si no hay red.
  - **Las llamadas a la API NUNCA se cachean** (datos SST/salud siempre frescos).
  - Assets estáticos: cache-first (rápido).
- **public/offline.html** — pantalla de marca cuando no hay conexión.
- **public/icons/** — íconos 192/512 + maskable + apple-touch, con la marca.
- **components/pwa/register-sw.tsx** — registra el SW (solo en producción).
- **components/pwa/install-button.tsx** — botón "Instalar app" que aparece cuando
  el navegador la considera instalable.
- **app/layout.tsx** — metadata PWA (manifest, theme-color, apple-web-app, viewport).

## Cómo se instala (usuario final)
- **Escritorio (Chrome/Edge):** aparece un ícono de instalación en la barra de
  direcciones, o el botón "Instalar app" dentro de la plataforma. Queda con ícono
  propio en el menú de aplicaciones.
- **Móvil (Android/iOS):** "Añadir a pantalla de inicio" desde el navegador.

## Nota honesta
- La PWA requiere **HTTPS** para activarse (el SW no corre en http://, salvo
  localhost). En producción sobre CloudFront/EKS esto ya se cumple.
- El caché offline cubre el **shell** de la app (que abra sin red); los datos SST
  siguen requiriendo conexión por seguridad. El trabajo de campo offline pesado
  (fotos, GPS sin señal) es el rol de la **app móvil nativa** (Expo), que es
  complementaria a la PWA de escritorio.
- Verificación real de instalación/offline: en el Sprint 0, con `npm run build`
  y servida sobre HTTPS.
