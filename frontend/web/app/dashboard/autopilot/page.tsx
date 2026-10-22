"use client";
import { useState } from "react";
import Link from "next/link";
import { useQueryClient } from "@tanstack/react-query";
import { RefreshCcw, Sunrise, CheckCircle2, ArrowRight } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { briefingAPI, type BriefingItem } from "@/lib/api";
import { useBriefing } from "@/lib/hooks";

const severity: Record<BriefingItem["severity"], { label: string; variant: "red" | "amber" | "gray" }> = {
  high: { label: "Prioridad alta", variant: "red" },
  medium: { label: "Prioridad media", variant: "amber" },
  low: { label: "Seguimiento", variant: "gray" },
};

export default function BriefingPage() {
  const qc = useQueryClient();
  const { data, isLoading, isError } = useBriefing();
  const [busy, setBusy] = useState(false);

  const refresh = async () => {
    setBusy(true);
    try { qc.setQueryData(["briefing"], (await briefingAPI.refresh()).data); } finally { setBusy(false); }
  };

  const byCompany = new Map<string, BriefingItem[]>();
  for (const it of data?.items ?? []) byCompany.set(it.company_name, [...(byCompany.get(it.company_name) ?? []), it]);

  return (
    <DashboardShell>
      <PageHeader
        title="Briefing diario"
        description="Pendientes de toda su cartera, calculados desde sus registros. Se envía cada mañana a las 6:30 (lunes a sábado). Es solo lectura: no crea ni modifica nada."
        action={<Button variant="secondary" onClick={() => void refresh()} disabled={busy}><RefreshCcw size={13} className={busy ? "animate-spin" : ""} /> Actualizar ahora</Button>}
      />
      {isLoading ? <Spinner size={22} /> : isError || !data ? (
        <Card className="p-4 text-sm text-red-700" >No fue posible calcular el briefing. Intente de nuevo en unos minutos.</Card>
      ) : data.items.length === 0 ? (
        <Empty icon={CheckCircle2} title="Sin pendientes" description={`Revisamos ${data.companies} empresa(s) y no hay pendientes detectados en sus registros.`} />
      ) : (
        <div className="space-y-4">
          <p className="text-xs text-zu-text3">
            <Sunrise size={12} className="mr-1 inline" />
            {data.items.length} pendiente(s) en {byCompany.size} de {data.companies} empresa(s) · actualizado {new Date(data.generated_at).toLocaleString("es-CO")}
          </p>
          {[...byCompany.entries()].map(([company, items]) => (
            <Card key={company} className="p-4">
              <h2 className="mb-3 text-sm font-bold">{company}</h2>
              <ul className="divide-y divide-zu-border">
                {items.map((it) => (
                  <li key={it.kind} className="flex flex-wrap items-start justify-between gap-3 py-3">
                    <div className="min-w-0">
                      <p className="text-[13px] font-semibold">{it.text}</p>
                      {it.basis && <p className="mt-0.5 text-[11px] text-zu-text3">Fundamento: {it.basis}</p>}
                    </div>
                    <div className="flex items-center gap-3">
                      <Badge variant={severity[it.severity].variant}>{severity[it.severity].label}</Badge>
                      <Link href={it.link} className="inline-flex items-center gap-1 text-[12px] font-semibold text-zu-accent hover:underline">Atender <ArrowRight size={12} /></Link>
                    </div>
                  </li>
                ))}
              </ul>
            </Card>
          ))}
        </div>
      )}
    </DashboardShell>
  );
}
