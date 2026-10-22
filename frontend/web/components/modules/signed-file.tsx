"use client";
import { useEffect, useState } from "react";
import { fileLinkAPI } from "@/lib/api";

// Stored files are served only through short-lived signed links (media
// service). Legacy URLs pointing at the internal storage host are reported as
// unavailable instead of rendering a broken link.
function useSigned(url: string) {
  const [signed, setSigned] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);
  useEffect(() => {
    let alive = true;
    if (!url.includes("/api/media/files/")) { setFailed(true); return; }
    fileLinkAPI.sign(url).then((r) => alive && setSigned(r.data.url)).catch(() => alive && setFailed(true));
    return () => { alive = false; };
  }, [url]);
  return { signed, failed };
}

export function SignedImage({ url, alt, className }: { url: string; alt: string; className?: string }) {
  const { signed, failed } = useSigned(url);
  if (failed) return <div className={`${className} flex items-center justify-center bg-zu-surface2 text-[10px] text-zu-text3`}>Archivo no disponible</div>;
  if (!signed) return <div className={`${className} animate-pulse bg-zu-surface2`} />;
  return <img src={signed} alt={alt} className={className} loading="lazy" />;
}

export function SignedLink({ url, children, className }: { url: string; children: React.ReactNode; className?: string }) {
  const [busy, setBusy] = useState(false);
  const open = async () => {
    setBusy(true);
    try { window.open((await fileLinkAPI.sign(url)).data.url, "_blank", "noopener"); } catch { alert("No fue posible abrir el archivo."); } finally { setBusy(false); }
  };
  if (!url.includes("/api/media/files/")) return <span className={`${className} cursor-not-allowed opacity-60`} title="Archivo cargado antes de la corrección de almacenamiento; vuelva a subirlo.">{children}</span>;
  return <button type="button" onClick={() => void open()} disabled={busy} className={className}>{children}</button>;
}
