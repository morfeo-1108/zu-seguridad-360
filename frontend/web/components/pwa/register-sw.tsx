"use client";
import { useEffect } from "react";

// Registra el Service Worker para habilitar la PWA (instalable + offline shell).
// Solo en producción y si el navegador lo soporta.
export function RegisterSW() {
  useEffect(() => {
    if (typeof window === "undefined") return;
    if (!("serviceWorker" in navigator)) return;
    if (process.env.NODE_ENV !== "production") return;

    const onLoad = () => {
      navigator.serviceWorker.register("/sw.js").catch((err) => {
        console.warn("SW registro falló:", err);
      });
    };
    // Si la página ya terminó de cargar antes de hidratar, "load" no vuelve a
    // dispararse: registrar de inmediato.
    if (document.readyState === "complete") {
      onLoad();
      return;
    }
    window.addEventListener("load", onLoad);
    return () => window.removeEventListener("load", onLoad);
  }, []);

  return null;
}
