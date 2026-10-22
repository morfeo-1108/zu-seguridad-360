"use client";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { Bell } from "lucide-react";
import { notificationAPI } from "@/lib/api";
import { useNotifications } from "@/lib/hooks";

// In-app inbox (notifications service). Scope is the signed-in user and
// tenant from the token; tenant-wide notices show to every user of it.
export function NotificationBell() {
  const qc = useQueryClient();
  const { data } = useNotifications();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  const unread = data?.unread ?? 0;

  useEffect(() => {
    if (!open) return;
    const onDoc = (e: MouseEvent) => { if (!ref.current?.contains(e.target as Node)) setOpen(false); };
    const onKey = (e: KeyboardEvent) => { if (e.key === "Escape") setOpen(false); };
    document.addEventListener("mousedown", onDoc);
    document.addEventListener("keydown", onKey);
    return () => { document.removeEventListener("mousedown", onDoc); document.removeEventListener("keydown", onKey); };
  }, [open]);

  const markRead = async (id: string) => { await notificationAPI.read(id); void qc.invalidateQueries({ queryKey: ["notifications"] }); };
  const markAll = async () => { await notificationAPI.readAll(); void qc.invalidateQueries({ queryKey: ["notifications"] }); };

  return (
    <div className="relative" ref={ref}>
      <button onClick={() => setOpen(!open)} aria-expanded={open} aria-label={`Notificaciones${unread ? `, ${unread} sin leer` : ""}`}
        className="relative flex h-8 w-8 items-center justify-center rounded-lg text-zu-text2 transition-colors hover:bg-zu-surface2 hover:text-zu-text1">
        <Bell size={16} />
        {unread > 0 && <span className="absolute -right-0.5 -top-0.5 min-w-[16px] rounded-full bg-red-600 px-1 text-center text-[9.5px] font-bold leading-4 text-white">{unread > 99 ? "99+" : unread}</span>}
      </button>
      {open && (
        <div role="dialog" aria-label="Notificaciones" className="absolute right-0 top-full z-50 mt-1.5 w-[min(92vw,360px)] rounded-xl border border-zu-border bg-white py-1.5 shadow-zu-md">
          <div className="flex items-center justify-between px-3 py-2">
            <span className="text-[10px] font-bold tracking-widest text-zu-text3">NOTIFICACIONES</span>
            {unread > 0 && <button onClick={() => void markAll()} className="text-[11px] font-semibold text-zu-accent hover:underline">Marcar todas como leídas</button>}
          </div>
          <ul className="max-h-[60vh] overflow-y-auto">
            {(data?.items ?? []).map((n) => (
              <li key={n.id} className={`border-t border-zu-border px-3 py-2.5 ${n.read_at ? "opacity-60" : ""}`}>
                <div className="flex items-start justify-between gap-2">
                  <p className="text-[12.5px] font-semibold leading-snug">{n.title}</p>
                  {!n.read_at && <button onClick={() => void markRead(n.id)} className="shrink-0 text-[10.5px] font-semibold text-zu-text3 hover:text-zu-accent">Leída</button>}
                </div>
                <p className="mt-0.5 text-[11.5px] leading-relaxed text-zu-text2">{n.message}</p>
                <div className="mt-1 flex items-center justify-between text-[10px] text-zu-text3">
                  <span>{new Date(n.created_at).toLocaleString("es-CO")}</span>
                  {n.link && <Link href={n.link} onClick={() => { setOpen(false); if (!n.read_at) void markRead(n.id); }} className="font-semibold text-zu-accent hover:underline">Ver</Link>}
                </div>
              </li>
            ))}
            {(data?.items ?? []).length === 0 && <li className="px-3 py-4 text-center text-xs text-zu-text3">Sin notificaciones.</li>}
          </ul>
        </div>
      )}
    </div>
  );
}
