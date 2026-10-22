// Service Worker — ZU Seguridad Integral 360 (PWA)
// Estrategia: app-shell cacheado + network-first para datos frescos.
// Objetivo: que la app abra aunque la red esté intermitente, sin servir
// datos sensibles obsoletos (las respuestas de API NO se cachean).
//
// Nada de lo cacheado queda congelado: las páginas se piden primero a la red
// y los archivos del shell se sirven desde caché pero se revalidan en segundo
// plano (stale-while-revalidate), así que un archivo reemplazado en public/
// llega en la visita siguiente. Lo que no está en el shell no se guarda aquí:
// va a la red y lo gobierna la caché HTTP normal.
//
// Subir la versión solo hace falta si cambia la lista SHELL o la estrategia;
// al activarse, las cachés de versiones anteriores se borran.

const CACHE = "zu360-shell-v2";
const SHELL = ["/dashboard", "/manifest.json", "/offline.html"];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE).then((c) => c.addAll(SHELL)).then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k)))
    ).then(() => self.clients.claim())
  );
});

self.addEventListener("fetch", (event) => {
  const { request } = event;
  if (request.method !== "GET") return;

  const url = new URL(request.url);
  if (url.origin !== self.location.origin) return;

  // NUNCA cachear llamadas a la API: los datos SST/salud deben ser frescos.
  if (url.pathname.startsWith("/api") || url.pathname.startsWith("/v1")) {
    return; // deja pasar a la red normalmente
  }

  // Navegación (páginas): network-first, cae al shell/offline si no hay red.
  if (request.mode === "navigate") {
    event.respondWith(
      fetch(request)
        .then((res) => {
          if (res.ok) {
            const copy = res.clone();
            caches.open(CACHE).then((c) => c.put(request, copy));
          }
          return res;
        })
        .catch(() =>
          caches.match(request).then((r) => r || caches.match("/offline.html"))
        )
    );
    return;
  }

  // Archivos del shell: respuesta inmediata desde caché y actualización en
  // segundo plano. Sin copia en caché, red directa (sin guardar nada).
  event.respondWith(
    caches.open(CACHE).then((cache) =>
      cache.match(request).then((cached) => {
        if (!cached) return fetch(request);
        const refresh = fetch(request, { cache: "no-cache" })
          .then((res) => { if (res.ok) return cache.put(request, res.clone()); })
          .catch(() => {});
        event.waitUntil(refresh);
        return cached;
      })
    )
  );
});
