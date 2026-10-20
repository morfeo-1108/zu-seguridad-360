"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Plus, Scale, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { legalMatrixAPI } from "@/lib/api";
import { useLegalMatrix, type LegalRequirement } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const statusText: Record<string, string> = { cumple: "Cumple", parcial: "Parcial", no_cumple: "No cumple", pendiente_evaluacion: "Pendiente" };
const statusColor: Record<string, "green" | "amber" | "red" | "gray"> = { cumple: "green", parcial: "amber", no_cumple: "red", pendiente_evaluacion: "gray" };

const freshForm = () => ({ norm_reference: "", title: "", applicable: true });

export default function LegalMatrixPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: requirements, isLoading, isError } = useLegalMatrix(companyId);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [editing, setEditing] = useState<LegalRequirement | null>(null);
  const [complianceStatus, setComplianceStatus] = useState("cumple");
  const [evidenceRef, setEvidenceRef] = useState("");

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["legal-matrix", companyId] });

  const createRequirement = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setError("");
    try {
      await legalMatrixAPI.create(companyId, form);
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch { setError("No se pudo registrar el requisito legal."); }
    finally { setBusy(false); }
  };

  const openEdit = (req: LegalRequirement) => {
    setEditing(req); setComplianceStatus(req.compliance_status); setEvidenceRef(req.evidence_ref || "");
  };

  const saveCompliance = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!editing || busy) return;
    setBusy(true); setError("");
    try {
      await legalMatrixAPI.updateCompliance(companyId, editing.id, { compliance_status: complianceStatus, evidence_ref: evidenceRef });
      await invalidate(); setEditing(null);
    } catch { setError("No se pudo actualizar el cumplimiento."); }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Matriz legal" description="Requisitos normativos aplicables y su estado de cumplimiento, revisado por un profesional." action={<Button onClick={() => { setForm(freshForm()); setError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13} /> Nuevo requisito</Button>} />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar la matriz legal.</p>
      ) : !requirements?.length ? (
        <Empty icon={Scale} title="Sin requisitos registrados" description="Registra el primer requisito normativo aplicable a esta empresa." />
      ) : (
        <Card className="overflow-hidden p-0">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead><tr>{["Norma", "Título", "Aplica", "Cumplimiento", "Evidencia", "Acciones"].map((l) => <Th key={l}>{l}</Th>)}</tr></thead>
              <tbody>
                {requirements.map((req) => (
                  <Tr key={req.id}>
                    <Td className="text-[12px] font-semibold">{req.norm_reference}</Td>
                    <Td className="max-w-xs text-[12px]">{req.title}</Td>
                    <Td><Badge variant={req.applicable ? "blue" : "gray"}>{req.applicable ? "Sí" : "No"}</Badge></Td>
                    <Td><Badge variant={statusColor[req.compliance_status]}>{statusText[req.compliance_status]}</Badge></Td>
                    <Td className="max-w-[160px] truncate text-[11px] text-zu-text2">{req.evidence_ref || "—"}</Td>
                    <Td><Button size="sm" variant="secondary" onClick={() => openEdit(req)}>Evaluar</Button></Td>
                  </Tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {showForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo requisito legal</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createRequirement} className="space-y-3">
              <label className="block text-xs">Norma / artículo<input required value={form.norm_reference} onChange={(e) => setForm({ ...form, norm_reference: e.target.value })} placeholder="Decreto 1072 de 2015, Art. 2.2.4.6.8" className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Título / descripción<input required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="flex items-center gap-2 text-xs"><input type="checkbox" checked={form.applicable} onChange={(e) => setForm({ ...form, applicable: e.target.checked })} /> Aplica a esta empresa</label>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Registrar"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {editing && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Evaluar cumplimiento</h2><button onClick={() => setEditing(null)} aria-label="Cerrar"><X size={18} /></button></div>
            <p className="mb-3 text-[11px] text-zu-text3">{editing.norm_reference} — {editing.title}</p>
            <form onSubmit={saveCompliance} className="space-y-3">
              <label className="block text-xs">Estado<select value={complianceStatus} onChange={(e) => setComplianceStatus(e.target.value)} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                <option value="cumple">Cumple</option><option value="parcial">Parcial</option><option value="no_cumple">No cumple</option><option value="pendiente_evaluacion">Pendiente</option>
              </select></label>
              <label className="block text-xs">Evidencia<input value={evidenceRef} onChange={(e) => setEvidenceRef(e.target.value)} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setEditing(null)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Guardar"}</Button></div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
