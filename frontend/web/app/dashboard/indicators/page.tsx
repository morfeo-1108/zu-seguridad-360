"use client";
import { useEffect, useState } from "react";
import { AlertTriangle, BarChart3, Info } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { indicatorAPI, type MonthlyIndicators } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

const now = new Date();

export default function IndicatorsPage() {
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const [year, setYear] = useState(now.getFullYear());
  const [month, setMonth] = useState(now.getMonth() + 1);
  const [standardHours, setStandardHours] = useState(220);
  const [data, setData] = useState<MonthlyIndicators | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!companyId) return;
    setLoading(true); setError("");
    indicatorAPI.monthly(companyId, year, month, standardHours)
      .then((res) => setData(res.data))
      .catch(() => setError("No se pudieron calcular los indicadores de este periodo."))
      .finally(() => setLoading(false));
  }, [companyId, year, month, standardHours]);

  return (
    <DashboardShell>
      <PageHeader title="Indicadores IF / IS" description="Índice de Frecuencia y Severidad, calculados por el servidor a partir de incidentes reales." />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : (
        <>
          <Card className="mb-5 flex flex-wrap items-end gap-4 p-4">
            <label className="text-xs">Año<input type="number" value={year} onChange={(e) => setYear(Number(e.target.value))} className="mt-1 block w-24 rounded border border-zu-border px-3 py-2" /></label>
            <label className="text-xs">Mes<select value={month} onChange={(e) => setMonth(Number(e.target.value))} className="mt-1 block w-32 rounded border border-zu-border px-3 py-2">
              {Array.from({ length: 12 }, (_, i) => i + 1).map((m) => <option key={m} value={m}>{new Date(2000, m - 1).toLocaleDateString("es-CO", { month: "long" })}</option>)}
            </select></label>
            <label className="text-xs">Horas mensuales estándar por trabajador<input type="number" min={1} value={standardHours} onChange={(e) => setStandardHours(Number(e.target.value))} className="mt-1 block w-40 rounded border border-zu-border px-3 py-2" /></label>
          </Card>

          {loading ? <Spinner size={24} /> : error ? (
            <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">{error}</p>
          ) : !data ? (
            <Empty icon={BarChart3} title="Sin datos" description="Selecciona un periodo." />
          ) : (
            <>
              <div className="mb-5 grid gap-3 sm:grid-cols-2">
                <Card className="p-5">
                  <div className="text-[11px] font-semibold text-zu-text3">Índice de Frecuencia (IF)</div>
                  <div className="mt-1 text-3xl font-extrabold text-zu-text1">{data.frequency_index.toFixed(2)}</div>
                  <div className="mt-1 text-[11px] text-zu-text3">{data.work_accidents} accidente(s) de trabajo en {data.period}</div>
                </Card>
                <Card className="p-5">
                  <div className="text-[11px] font-semibold text-zu-text3">Índice de Severidad (IS)</div>
                  <div className="mt-1 text-3xl font-extrabold text-zu-text1">{data.severity_index.toFixed(2)}</div>
                  <div className="mt-1 text-[11px] text-zu-text3">{data.lost_days} día(s) perdidos en {data.period}</div>
                </Card>
              </div>
              <Card className="p-4">
                <div className="mb-2 flex items-center gap-1.5 text-[11px] font-bold text-amber-700"><Info size={13} /> Cómo se calculó</div>
                <p className="text-[11px] leading-relaxed text-zu-text2">{data.formula_note}</p>
                <p className="mt-2 text-[11px] leading-relaxed text-zu-text2">HHT estimado: {data.active_employees} trabajadores activos × {data.standard_monthly_hours} h = <strong>{data.estimated_hht.toLocaleString("es-CO")} horas-hombre</strong> (estimado, no medido).</p>
                <p className="mt-2 text-[11px] leading-relaxed text-zu-text3">{data.absenteeism_note}</p>
              </Card>
            </>
          )}
        </>
      )}
    </DashboardShell>
  );
}
