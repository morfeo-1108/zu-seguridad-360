"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Plus, ShieldAlert, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { riskAPI } from "@/lib/api";
import { useCompanyRisks, useRiskHeatmap } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const acceptabilityText: Record<string, string> = { aceptable: "Aceptable", no_aceptable: "No aceptable", critico: "Crítico" };
const acceptabilityColor: Record<string, "green" | "amber" | "red"> = { aceptable: "green", no_aceptable: "amber", critico: "red" };

const freshForm = () => ({
  process: "", area: "", hazard_type: "", hazard_source: "", health_effect: "",
  workers_exposed: 1, nd: 2, ne: 2, nc: 25, current_control: "",
});

export default function RisksPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: risks, isLoading, isError } = useCompanyRisks(companyId);
  const { data: heatmap } = useRiskHeatmap(companyId);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["risks", companyId] });
    queryClient.invalidateQueries({ queryKey: ["risks-heatmap", companyId] });
  };

  const createRisk = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setFormError("");
    try {
      await riskAPI.create(companyId, { ...form, workers_exposed: Number(form.workers_exposed), nd: Number(form.nd), ne: Number(form.ne), nc: Number(form.nc) });
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch {
      setFormError("No se pudo registrar el riesgo. Revisa que ND (1-10), NE (1-4) y NC (10-100) estén en rango GTC-45.");
    } finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader
        title="Matriz de riesgos — GTC-45 / IPEVR"
        description="Nivel de riesgo (NR) calculado por el servidor: ND × NE = NP; NP × NC = NR."
        action={<Button onClick={() => { setForm(freshForm()); setFormError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13} /> Nuevo riesgo</Button>}
      />

      {!companyId ? (
        <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." />
      ) : (
        <>
          {heatmap && (
            <div className="mb-5 grid grid-cols-3 gap-3">
              <Card className="p-4"><div className="text-[11px] font-semibold text-zu-text3">Crítico</div><div className="mt-1 text-2xl font-extrabold text-red-600">{heatmap.summary.critical}</div></Card>
              <Card className="p-4"><div className="text-[11px] font-semibold text-zu-text3">No aceptable</div><div className="mt-1 text-2xl font-extrabold text-amber-600">{heatmap.summary.not_acceptable}</div></Card>
              <Card className="p-4"><div className="text-[11px] font-semibold text-zu-text3">Aceptable</div><div className="mt-1 text-2xl font-extrabold text-green-600">{heatmap.summary.acceptable}</div></Card>
            </div>
          )}

          {isLoading ? <Spinner size={24} /> : isError ? (
            <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar la matriz de riesgos de la empresa activa.</p>
          ) : !risks?.length ? (
            <Empty icon={ShieldAlert} title="Sin riesgos registrados" description="Registra el primer peligro para iniciar la matriz GTC-45 de esta empresa." />
          ) : (
            <Card className="overflow-hidden p-0">
              <div className="overflow-x-auto">
                <table className="w-full border-collapse">
                  <thead><tr>{["Proceso / área", "Peligro", "ND", "NE", "NP", "NC", "NR", "Aceptabilidad", "Control actual"].map((label) => <Th key={label}>{label}</Th>)}</tr></thead>
                  <tbody>
                    {risks.map((r) => (
                      <Tr key={r.id}>
                        <Td className="text-[11px]"><div className="font-semibold">{r.process}</div><div className="text-zu-text3">{r.area}</div></Td>
                        <Td className="max-w-xs text-[11px]"><div className="font-semibold">{r.hazard_type}</div><div className="text-zu-text3">{r.hazard_source}</div></Td>
                        <Td className="text-center text-[11px]">{r.nd}</Td>
                        <Td className="text-center text-[11px]">{r.ne}</Td>
                        <Td className="text-center text-[11px] font-semibold">{r.np}</Td>
                        <Td className="text-center text-[11px]">{r.nc}</Td>
                        <Td className="text-center text-[12px] font-extrabold">{r.nr}</Td>
                        <Td><Badge variant={acceptabilityColor[r.acceptability] || "gray"}>{acceptabilityText[r.acceptability] || r.acceptability}</Badge></Td>
                        <Td className="max-w-[160px] truncate text-[11px] text-zu-text2">{r.current_control || "—"}</Td>
                      </Tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <div className="border-t border-zu-border p-3 text-[11px] text-zu-text3">{risks.length} entradas en la matriz de la empresa activa, ordenadas por NR descendente.</div>
            </Card>
          )}
        </>
      )}

      {showForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-xl p-5 max-h-[90vh] overflow-y-auto">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="text-sm font-bold">Nuevo riesgo (metodología GTC-45)</h2>
              <button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18} /></button>
            </div>
            <form onSubmit={createRisk} className="space-y-3">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <label className="block text-xs">Proceso<input required value={form.process} onChange={(e) => setForm({ ...form, process: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
                <label className="block text-xs">Área<input required value={form.area} onChange={(e) => setForm({ ...form, area: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              </div>
              <label className="block text-xs">Tipo de peligro<input required value={form.hazard_type} onChange={(e) => setForm({ ...form, hazard_type: e.target.value })} placeholder="Físico, químico, locativo, biomecánico…" className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Fuente / descripción del peligro<input required value={form.hazard_source} onChange={(e) => setForm({ ...form, hazard_source: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Efecto posible en la salud<input required value={form.health_effect} onChange={(e) => setForm({ ...form, health_effect: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Trabajadores expuestos<input required type="number" min={0} value={form.workers_exposed} onChange={(e) => setForm({ ...form, workers_exposed: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <div className="grid grid-cols-3 gap-3">
                <label className="block text-xs">ND (1-10)<input required type="number" min={1} max={10} value={form.nd} onChange={(e) => setForm({ ...form, nd: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /><span className="mt-1 block text-[10px] text-zu-text3">Nivel de deficiencia</span></label>
                <label className="block text-xs">NE (1-4)<input required type="number" min={1} max={4} value={form.ne} onChange={(e) => setForm({ ...form, ne: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /><span className="mt-1 block text-[10px] text-zu-text3">Nivel de exposición</span></label>
                <label className="block text-xs">NC (10-100)<input required type="number" min={10} max={100} value={form.nc} onChange={(e) => setForm({ ...form, nc: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /><span className="mt-1 block text-[10px] text-zu-text3">Nivel de consecuencia</span></label>
              </div>
              <label className="block text-xs">Control actual<input value={form.current_control} onChange={(e) => setForm({ ...form, current_control: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <p className="text-[10px] text-zu-text3">NP (ND×NE) y NR (NP×NC) los calcula el servidor con la escala de aceptabilidad GTC-45 — no se envían desde este formulario.</p>
              {formError && <p role="alert" className="text-xs text-red-700">{formError}</p>}
              <div className="flex justify-end gap-2 pt-1">
                <Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button>
                <Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Registrar riesgo"}</Button>
              </div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
