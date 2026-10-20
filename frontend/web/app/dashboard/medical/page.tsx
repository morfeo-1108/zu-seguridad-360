"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, CalendarClock, Plus, Stethoscope, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { medicalAPI } from "@/lib/api";
import { useEmployees, useMedicalExpiring, useMedicalPendingReview, type MedicalExam } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const aptitudeLabel: Record<string, string> = { apto: "Apto", apto_con_restriccion: "Apto con restricción", no_apto: "No apto", pendiente_revision: "Pendiente" };
const aptitudeColor: Record<string, "green" | "amber" | "red" | "gray"> = { apto: "green", apto_con_restriccion: "amber", no_apto: "red", pendiente_revision: "gray" };

const freshForm = () => ({ employee_id: "", type: "ingreso", ips_name: "", scheduled_at: new Date().toISOString().slice(0, 10) });

export default function MedicalPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: pending, isLoading, isError } = useMedicalPendingReview(companyId);
  const { data: expiring } = useMedicalExpiring(companyId);
  const { data: employeeResponse } = useEmployees(companyId);
  const employees = employeeResponse?.data || [];
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);
  const [reviewing, setReviewing] = useState<MedicalExam | null>(null);
  const [aptitude, setAptitude] = useState("apto");
  const [notes, setNotes] = useState("");

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["medical-pending", companyId] });
    queryClient.invalidateQueries({ queryKey: ["medical-expiring", companyId] });
  };

  const scheduleExam = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setFormError("");
    try {
      await medicalAPI.schedule(companyId, { ...form, scheduled_at: new Date(form.scheduled_at).toISOString() });
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch { setFormError("No se pudo programar el examen. Verifica trabajador, tipo e IPS."); }
    finally { setBusy(false); }
  };

  const submitReview = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!reviewing || busy) return;
    setBusy(true);
    try {
      await medicalAPI.review(companyId, reviewing.id, { aptitude_result: aptitude, review_notes: notes });
      await invalidate(); setReviewing(null); setNotes("");
    } catch { /* el botón permanece visible para reintentar */ }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader
        title="Exámenes médicos ocupacionales"
        description="Programación y revisión de aptitud por un profesional; no se ejecuta OCR ni se infiere diagnóstico automáticamente."
        action={<Button onClick={() => { setForm(freshForm()); setFormError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13} /> Programar examen</Button>}
      />

      {!!expiring?.length && (
        <Card className="mb-4 border-amber-300 bg-amber-50 p-3">
          <div className="flex items-start gap-2"><CalendarClock size={16} className="mt-0.5 text-amber-800" /><p className="text-xs text-amber-900"><span className="font-bold">{expiring.length}</span> examen(es) vencen en los próximos 30 días.</p></div>
        </Card>
      )}

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar los exámenes pendientes de revisión.</p>
      ) : !pending?.length ? (
        <Empty icon={Stethoscope} title="Sin exámenes pendientes de revisión" description="Programa el primer examen médico ocupacional." />
      ) : (
        <Card className="overflow-hidden p-0">
          <div className="overflow-x-auto">
            <table className="w-full border-collapse">
              <thead><tr>{["Tipo", "IPS", "Programado", "Resultado", "Acciones"].map((l) => <Th key={l}>{l}</Th>)}</tr></thead>
              <tbody>
                {pending.map((exam) => (
                  <Tr key={exam.id}>
                    <Td className="text-[12px] font-semibold capitalize">{exam.type}</Td>
                    <Td className="text-[12px]">{exam.ips_name || "—"}</Td>
                    <Td className="text-[11px]">{exam.scheduled_at ? formatDateTime(exam.scheduled_at) : "—"}</Td>
                    <Td><Badge variant={aptitudeColor[exam.aptitude_result]}>{aptitudeLabel[exam.aptitude_result]}</Badge></Td>
                    <Td><Button size="sm" variant="secondary" onClick={() => { setReviewing(exam); setAptitude("apto"); setNotes(""); }}>Revisar</Button></Td>
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
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Programar examen médico</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={scheduleExam} className="space-y-3">
              <label className="block text-xs">Trabajador
                <select required value={form.employee_id} onChange={(e) => setForm({ ...form, employee_id: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  <option value="" disabled>Elegir…</option>
                  {employees.map((e) => <option key={e.id} value={e.id}>{e.first_name} {e.last_name}</option>)}
                </select>
              </label>
              <label className="block text-xs">Tipo
                <select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  <option value="ingreso">Ingreso</option><option value="periodico">Periódico</option><option value="retiro">Retiro</option>
                </select>
              </label>
              <label className="block text-xs">IPS<input required value={form.ips_name} onChange={(e) => setForm({ ...form, ips_name: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Fecha<input required type="date" value={form.scheduled_at} onChange={(e) => setForm({ ...form, scheduled_at: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              {formError && <p role="alert" className="text-xs text-red-700">{formError}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Programar"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {reviewing && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Revisar aptitud</h2><button onClick={() => setReviewing(null)} aria-label="Cerrar"><X size={18} /></button></div>
            <p className="mb-3 text-[11px] text-zu-text3">{reviewing.type} · {reviewing.ips_name}</p>
            <form onSubmit={submitReview} className="space-y-3">
              <label className="block text-xs">Resultado
                <select value={aptitude} onChange={(e) => setAptitude(e.target.value)} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  <option value="apto">Apto</option><option value="apto_con_restriccion">Apto con restricción</option><option value="no_apto">No apto</option>
                </select>
              </label>
              <label className="block text-xs">Notas de revisión<textarea value={notes} onChange={(e) => setNotes(e.target.value)} rows={3} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setReviewing(null)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Guardar revisión"}</Button></div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
