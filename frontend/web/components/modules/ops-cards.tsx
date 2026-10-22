"use client";
import Link from "next/link";
import { AlarmClock, CheckCircle2, Circle, Lock } from "lucide-react";
import { Card, Spinner } from "@/components/ui";
import { useOnboardingProgress, useSLASummary } from "@/lib/hooks";

export function SLACard() {
  const { data, isLoading, isError } = useSLASummary();
  return (
    <Card className="p-4">
      <h2 className="mb-2 flex items-center gap-2 text-sm font-bold"><AlarmClock size={14} className="text-zu-accent" /> Plazos de atención</h2>
      {isLoading ? <Spinner size={16} /> : isError || !data ? <p className="text-xs text-zu-text3">No disponible en este momento.</p> : (
        <>
          <div className="grid grid-cols-3 gap-2 text-center">
            <div><p className="text-xl font-bold">{data.open}</p><p className="text-[10.5px] text-zu-text3">abiertos</p></div>
            <div><p className={`text-xl font-bold ${data.breached_open ? "text-red-600" : ""}`}>{data.breached_open}</p><p className="text-[10.5px] text-zu-text3">vencidos</p></div>
            <div><p className="text-xl font-bold">{data.compliance_pct_last_30d === null ? "—" : `${Math.round(data.compliance_pct_last_30d)}%`}</p><p className="text-[10.5px] text-zu-text3">a tiempo (30 días)</p></div>
          </div>
          <p className="mt-3 text-[11px] text-zu-text3">Cada inspección con hallazgos abre un plazo de revisión de 72 horas.{data.compliance_pct_last_30d === null ? " Aún no hay plazos cerrados para calcular el cumplimiento." : ""}</p>
        </>
      )}
    </Card>
  );
}

export function OnboardingCard({ companyId }: { companyId: string }) {
  const { data, isLoading, isError } = useOnboardingProgress(companyId);
  if (!companyId) return null;
  return (
    <Card className="p-4">
      <h2 className="mb-1 text-sm font-bold">Puesta en marcha del SG‑SST</h2>
      {isLoading ? <Spinner size={16} /> : isError || !data ? <p className="text-xs text-zu-text3">No disponible en este momento.</p> : (
        <>
          <p className="mb-3 text-[11px] text-zu-text3">
            {data.done} de {data.total} pasos con registro real
            {data.completed_at ? ` · completado en ${(data.elapsed_hours / 24).toFixed(1)} días` : ` · ${(data.elapsed_hours / 24).toFixed(1)} días desde el inicio`}
          </p>
          <div className="mb-3 h-1.5 overflow-hidden rounded-full bg-zu-surface2"><div className="h-full rounded-full bg-zu-accent" style={{ width: `${(data.done / Math.max(1, data.total)) * 100}%` }} /></div>
          <ul className="space-y-1.5">
            {data.steps.map((s) => (
              <li key={s.key} className="flex items-start gap-2 text-[12px]">
                {s.status === "done" ? <CheckCircle2 size={14} className="mt-0.5 shrink-0 text-emerald-600" /> : s.status === "blocked" ? <Lock size={14} className="mt-0.5 shrink-0 text-zu-text3" /> : <Circle size={14} className="mt-0.5 shrink-0 text-zu-text3" />}
                <span className={s.status === "done" ? "text-zu-text2" : ""}>
                  {s.status === "pending" ? <Link href={s.link} className="font-semibold hover:text-zu-accent hover:underline">{s.title}</Link> : <span className="font-semibold">{s.title}</span>}
                  {s.status === "blocked" && <span className="block text-[10.5px] text-zu-text3">{s.blocked_reason}</span>}
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
    </Card>
  );
}
