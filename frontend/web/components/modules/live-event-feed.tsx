"use client";
import { Activity, ShieldAlert } from "lucide-react";

export function LiveEventFeed({ compact = false }: { projectId?: string; compact?: boolean }) {
  return <div className={compact ? "" : "rounded-lg border border-zu-border bg-white p-3"}>
    <div className="flex items-center gap-2"><Activity size={14} className="text-zu-text3"/><span className="text-[12px] font-semibold">Eventos externos deshabilitados</span></div>
    <p className="mt-2 flex items-start gap-1.5 text-[10px] leading-relaxed text-zu-text3"><ShieldAlert size={12} className="mt-0.5 shrink-0"/>El feed de agentes no abre conexiones SSE ni consulta endpoints de telemetría durante el piloto sintético.</p>
  </div>;
}
