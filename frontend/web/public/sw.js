// Service Worker — ZU Seguridad Integral 360 (PWA)
// Estrategia: app-shell cacheado + network-first para datos frescos.
// Objetivo: que la app abra aunque la red esté intermitente, sin servir
// datos sensibles obsoletos (las respuestas de API NO se cachean).

const CACHE = "zu360-shell-v1";
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

  // NUNCA cachear llamadas a la API: los datos SST/salud deben ser frescos.
  if (url.pathname.startsWith("/api") || url.pathname.startsWith("/v1")) {
    return; // deja pasar a la red normalmente
  }

  // Navegación (páginas): network-first, cae al shell/offline si no hay red.
  if (request.mode === "navigate") {
    event.respondWith(
      fetch(request)
        .then((res) => {
          const copy = res.clone();
          caches.open(CACHE).then((c) => c.put(request, copy));
          return res;
        })
        .catch(() =>
          caches.match(request).then((r) => r || caches.match("/offline.html"))
        )
    );
    return;
  }

  // Assets estáticos: cache-first (rápido y estable).
  event.respondWith(
    caches.match(request).then((cached) => cached || fetch(request))
  );
});
