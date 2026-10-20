"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, ClipboardCheck, Eye, Plus, Trash2, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { inspectionAPI, type ChecklistItem } from "@/lib/api";
import { useCompanyInspections, type Inspection } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const freshChecklistRow = (): ChecklistItem => ({ code: "", description: "", compliant: true, notes: "" });
const freshForm = () => ({ title: "", area: "", performed_at: new Date().toISOString().slice(0, 10), checklist: [freshChecklistRow()] });

export default function InspectionsPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: inspections, isLoading, isError } = useCompanyInspections(companyId);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selected, setSelected] = useState<Inspection | null>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["inspections", companyId] });

  const updateRow = (index: number, patch: Partial<ChecklistItem>) => {
    setForm((prev) => ({ ...prev, checklist: prev.checklist.map((row, i) => (i === index ? { ...row, ...patch } : row)) }));
  };

  const createInspection = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setFormError("");
    try {
      await inspectionAPI.create(companyId, form);
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch { setFormError("No se pudo registrar la inspección. Verifica que cada ítem tenga código y descripción."); }
    finally { setBusy(false); }
  };

  const approve = async (inspection: Inspection) => {
    try {
      const response = await inspectionAPI.approve(companyId, inspection.id);
      await invalidate(); setSelected(response.data);
    } catch { /* el botón de aprobar permanece visible para reintentar */ }
  };

  return (
    <DashboardShell>
      <PageHeader
        title="Inspecciones"
        description="Checklist estructurado; el plan de acción se genera automáticamente de los ítems no conformes."
        action={<Button onClick={() => { setForm(freshForm()); setFormError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13} /> Nueva inspección</Button>}
      />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar las inspecciones de la empresa activa.</p>
      ) : !inspections?.length ? (
        <Empty icon={ClipboardCheck} title="Sin inspecciones registradas" description="Registra la primera inspección con su checklist." />
      ) : (
        <Card className="overflow-hidden p-0">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead><tr>{["Título", "Área", "Fecha", "Cumplimiento", "Estado", "Acciones"].map((l) => <Th key={l}>{l}</Th>)}</tr></thead>
              <tbody>
                {inspections.map((insp) => (
                  <Tr key={insp.id}>
                    <Td className="text-[12px] font-semibold">{insp.title}</Td>
                    <Td className="text-[12px]">{insp.area}</Td>
                    <Td className="text-[11px]">{formatDateTime(insp.performed_at)}</Td>
                    <Td><Badge variant={insp.compliance_pct >= 80 ? "green" : insp.compliance_pct >= 50 ? "amber" : "red"}>{insp.compliance_pct.toFixed(0)}%</Badge></Td>
                    <Td><Badge variant={insp.status === "aprobada" ? "green" : "blue"}>{insp.status === "aprobada" ? "Aprobada" : "Completada"}</Badge></Td>
                    <Td><Button size="sm" variant="secondary" onClick={() => setSelected(insp)}><Eye size={12} /> Ver</Button></Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {showForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nueva inspección</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createInspection} className="space-y-3">
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
                <label className="block text-xs">Título<input required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
                <label className="block text-xs">Área<input required value={form.area} onChange={(e) => setForm({ ...form, area: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
                <label className="block text-xs">Fecha<input required type="date" value={form.performed_at} onChange={(e) => setForm({ ...form, performed_at: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              </div>
              <div>
                <div className="mb-1.5 flex items-center justify-between">
                  <span className="text-xs font-semibold text-zu-text1">Checklist</span>
                  <button type="button" onClick={() => setForm({ ...form, checklist: [...form.checklist, freshChecklistRow()] })} className="text-[11px] font-semibold text-zu-accent">+ Agregar ítem</button>
                </div>
                <div className="space-y-2">
                  {form.checklist.map((row, i) => (
                    <div key={i} className="flex items-center gap-2 rounded-md border border-zu-border p-2">
                      <input required placeholder="Código" value={row.code} onChange={(e) => updateRow(i, { code: e.target.value })} className="w-20 rounded border border-zu-border px-2 py-1.5 text-[11px]" />
                      <input required placeholder="Descripción del ítem" value={row.description} onChange={(e) => updateRow(i, { description: e.target.value })} className="flex-1 rounded border border-zu-border px-2 py-1.5 text-[11px]" />
                      <label className="flex items-center gap-1 text-[10px] font-semibold text-zu-text2">
                        <input type="checkbox" checked={row.compliant} onChange={(e) => updateRow(i, { compliant: e.target.checked })} /> Conforme
                      </label>
                      <button type="button" onClick={() => setForm({ ...form, checklist: form.checklist.filter((_, idx) => idx !== i) })} disabled={form.checklist.length === 1} className="text-red-500 disabled:opacity-30"><Trash2 size={14} /></button>
                    </div>
                  ))}
                </div>
              </div>
              <p className="text-[10px] text-zu-text3">El % de cumplimiento y el plan de acción (uno por ítem no conforme) los calcula el servidor.</p>
              {formError && <p role="alert" className="text-xs text-red-700">{formError}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Registrar inspección"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto p-5">
            <div className="mb-4 flex items-start justify-between gap-3">
              <div><h2 className="text-sm font-bold">{selected.title}</h2><p className="mt-1 text-[11px] text-zu-text3">{selected.area} · {formatDateTime(selected.performed_at)}</p></div>
              <button onClick={() => setSelected(null)} aria-label="Cerrar"><X size={18} /></button>
            </div>
            <div className="mb-4 flex gap-2"><Badge variant={selected.status === "aprobada" ? "green" : "blue"}>{selected.status === "aprobada" ? "Aprobada" : "Completada"}</Badge><Badge variant={selected.compliance_pct >= 80 ? "green" : "amber"}>{selected.compliance_pct.toFixed(0)}% cumplimiento</Badge></div>
            <h3 className="mb-2 text-xs font-bold text-zu-text1">Checklist</h3>
            <div className="mb-4 space-y-1">
              {selected.checklist.map((item) => (
                <div key={item.code} className="flex items-center justify-between rounded-md border border-zu-border px-3 py-1.5 text-[11px]">
                  <span>{item.code} — {item.description}</span>
                  <Badge variant={item.compliant ? "green" : "red"}>{item.compliant ? "Conforme" : "No conforme"}</Badge>
                </div>
              ))}
            </div>
            {selected.action_plan.length > 0 && (
              <>
                <h3 className="mb-2 text-xs font-bold text-zu-text1">Plan de acción (generado de los hallazgos)</h3>
                <div className="mb-4 space-y-1">
                  {selected.action_plan.map((a) => (
                    <div key={a.item_code} className="rounded-md bg-amber-50 px-3 py-1.5 text-[11px] text-amber-900">{a.description} <span className="font-semibold">({a.status})</span></div>
                  ))}
                </div>
              </>
            )}
            {selected.status !== "aprobada" && <div className="flex justify-end"><Button size="sm" onClick={() => void approve(selected)}>Aprobar inspección</Button></div>}
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
