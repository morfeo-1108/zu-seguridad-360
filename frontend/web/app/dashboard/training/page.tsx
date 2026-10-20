"use client";
import { FormEvent, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { CalendarDays, GraduationCap, Plus, Users, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { trainingAPI, type SyntheticCertificateRecord } from "@/lib/api";
import { useCompanyTrainings, useEmployees, type Training } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDate } from "@/lib/utils";

const today = () => new Date().toISOString().slice(0, 10);
const freshForm = () => ({ title: "Demostración de capacitación sintética", category: "Prueba", duration_hours: "1", validity_months: "3", instructor_name: "Sintético Facilitador", scheduled_at: today() });

export default function TrainingPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: trainings, isLoading, isError } = useCompanyTrainings(companyId);
  const { data: employeeResponse, isLoading: employeesLoading, isError: employeesError } = useEmployees(companyId);
  const employees = employeeResponse?.data ?? [];
  const syntheticEmployees = useMemo(() => employees.filter((employee) => employee.is_synthetic), [employees]);
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);
  const [attendanceTraining, setAttendanceTraining] = useState<Training | null>(null);
  const [selectedEmployeeIds, setSelectedEmployeeIds] = useState<string[]>([]);
  const [actionError, setActionError] = useState("");
  const [result, setResult] = useState<{ certificates: SyntheticCertificateRecord[]; notice: string } | null>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["trainings", companyId] });
  const createTraining = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault(); if (!companyId || busy) return;
    const hours = Number(form.duration_hours); const months = Number(form.validity_months);
    if (!Number.isFinite(hours) || !Number.isFinite(months)) { setFormError("La duración y vigencia sintéticas deben ser números."); return; }
    setBusy(true); setFormError("");
    try {
      await trainingAPI.create(companyId, { title: form.title, norm_ref: "", category: form.category, duration_hours: hours, validity_months: months, instructor_name: form.instructor_name, scheduled_at: form.scheduled_at, evaluation_min_score: 0 });
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch { setFormError("No se pudo crear. Usa título/categoría con prefijo de prueba, sin norma, instructor sintético, duración ≤8 h y vigencia ficticia entre 1 y 36 meses."); }
    finally { setBusy(false); }
  };

  const toggleEmployee = (id: string) => setSelectedEmployeeIds((current) => current.includes(id) ? current.filter((value) => value !== id) : [...current, id]);
  const completeAttendance = async () => {
    if (!companyId || !attendanceTraining || selectedEmployeeIds.length === 0 || busy) return;
    setBusy(true); setActionError("");
    try {
      const response = await trainingAPI.complete(companyId, attendanceTraining.id, selectedEmployeeIds);
      setResult({ certificates: response.data.certificates, notice: response.data.notice });
      setAttendanceTraining(null); setSelectedEmployeeIds([]); await invalidate();
    } catch { setActionError("No se pudo registrar la asistencia. Verifica que las personas y el curso sean sintéticos y pertenezcan a la empresa activa."); }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Capacitaciones sintéticas" description="Agenda y asistencia de prueba. Sin afirmaciones de cumplimiento, vigencia oficial ni validación pública." action={<Button onClick={() => { setForm(freshForm()); setFormError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13}/> Crear curso de prueba</Button>} />
      <div className="mb-4 rounded-md border border-amber-200 bg-amber-50 p-3 text-[11px] leading-relaxed text-amber-950"><strong>ENTORNO SINTÉTICO.</strong> Las constancias son inválidas, no tienen QR/URL público y no acreditan capacitación. Las vigencias configuradas sólo son valores de demostración; no se interpretan como norma.</div>
      {!companyId ? <Empty icon={GraduationCap} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado."/> : isLoading ? <Spinner size={24}/> : isError ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar cursos de prueba desde Core.</p> : !trainings?.length ? <Empty icon={GraduationCap} title="Sin capacitaciones sintéticas" description="Crea un curso rotulado como demostración para ejercitar agenda y asistencia."/> : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">{trainings.map((training) => <Card key={training.id} className="p-4">
          <div className="flex items-start justify-between gap-3"><div><h2 className="text-sm font-semibold">{training.title}</h2><p className="mt-1 text-[11px] text-zu-text2">{training.category} · {training.duration_hours} h (prueba)</p></div><Badge variant={training.is_synthetic ? "amber" : "red"}>{training.is_synthetic ? "SINTÉTICO" : "NO CLASIFICADO"}</Badge></div>
          <div className="mt-3 flex flex-wrap gap-x-4 gap-y-1 text-[10px] text-zu-text3"><span><CalendarDays size={11} className="mr-1 inline"/>{training.scheduled_at ? formatDate(training.scheduled_at) : "Sin fecha"}</span><span><Users size={11} className="mr-1 inline"/>{training.attendee_count} asistencias registradas</span><span>Estado: {training.status}</span></div>
          {training.is_synthetic && training.status !== "completed" && <Button size="sm" variant="secondary" className="mt-4" onClick={() => { setAttendanceTraining(training); setSelectedEmployeeIds([]); setActionError(""); }}><Users size={12}/> Registrar asistencia sintética</Button>}
          {training.status === "completed" && <p className="mt-4 text-[10px] text-zu-text3">Curso completado en simulación; no representa una capacitación oficial.</p>}
        </Card>)}</div>
      )}

      {result && <Card className="mt-5 border-amber-300 bg-amber-50 p-4"><div className="mb-1 text-sm font-bold text-amber-950">Asistencia registrada — SIMULACIÓN, NO VÁLIDA</div><p className="text-xs text-amber-950">{result.notice}</p><div className="mt-3 space-y-1">{result.certificates.map((certificate) => <div key={certificate.id} className="flex flex-wrap items-center gap-2 text-[11px]"><Badge variant="amber">SINTÉTICO</Badge><span className="font-mono">{certificate.folio}</span><span>{syntheticEmployees.find((employee) => employee.id === certificate.employee_id)?.first_name || "Perfil de prueba"}</span><Badge variant="red">NO VÁLIDO</Badge>{certificate.qr_code && <a href={certificate.qr_code} target="_blank" rel="noreferrer" className="font-semibold text-zu-accent underline">Ver verificación pública</a>}</div>)}</div><p className="mt-2 text-[10px] text-amber-900">Cada folio es consultable públicamente (sin sesión) en la página de verificación; mostrará claramente que es un dato sintético y no válido.</p></Card>}

      {showForm && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-xl p-5"><div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo curso sintético</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18}/></button></div><form onSubmit={createTraining} className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <label className="text-xs sm:col-span-2">Título<input required minLength={5} value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/><span className="text-[10px] text-zu-text3">Debe iniciar con “Demostración”, “Prueba” o “Sintético”.</span></label>
        <label className="text-xs">Categoría<input required value={form.category} onChange={(event) => setForm({ ...form, category: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
        <label className="text-xs">Instructor de prueba<input value={form.instructor_name} onChange={(event) => setForm({ ...form, instructor_name: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
        <label className="text-xs">Duración (horas de prueba)<input required type="number" min="0.1" max="8" step="0.1" value={form.duration_hours} onChange={(event) => setForm({ ...form, duration_hours: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
        <label className="text-xs">Vigencia ficticia (meses)<input required type="number" min="1" max="36" value={form.validity_months} onChange={(event) => setForm({ ...form, validity_months: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
        <label className="text-xs sm:col-span-2">Fecha de agenda<input required type="date" value={form.scheduled_at} onChange={(event) => setForm({ ...form, scheduled_at: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
        <p className="text-[10px] text-zu-text3 sm:col-span-2">La referencia normativa y evaluación automática se envían vacías/desactivadas.</p>
        {formError && <p role="alert" className="text-xs text-red-700 sm:col-span-2">{formError}</p>}
        <div className="flex justify-end gap-2 sm:col-span-2"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear curso sintético"}</Button></div>
      </form></Card></div>}

      {attendanceTraining && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-xl p-5"><div className="mb-3 flex items-center justify-between"><div><h2 className="text-sm font-bold">Asistencia de demostración</h2><p className="mt-1 text-[11px] text-zu-text2">{attendanceTraining.title}</p></div><button onClick={() => setAttendanceTraining(null)} aria-label="Cerrar"><X size={18}/></button></div><p className="mb-3 text-[10px] text-amber-900">Selecciona únicamente perfiles SINTÉTICOS pertenecientes a la empresa activa. No se evalúa aprobación normativa.</p><div className="max-h-64 space-y-2 overflow-y-auto">{employeesLoading ? <Spinner size={20}/> : employeesError ? <p role="alert" className="text-xs text-red-700">No se pudo cargar la lista de trabajadores de esta empresa.</p> : syntheticEmployees.map((employee) => <label key={employee.id} className="flex items-center gap-2 rounded border border-zu-border p-2 text-xs"><input type="checkbox" checked={selectedEmployeeIds.includes(employee.id)} onChange={() => toggleEmployee(employee.id)}/><span>{employee.first_name} {employee.last_name}</span><span className="ml-auto font-mono text-[10px] text-zu-text3">{employee.document_number}</span><Badge variant="amber">SINTÉTICO</Badge></label>)}{!employeesLoading && !employeesError && !syntheticEmployees.length && <p className="text-xs text-zu-text2">Crea primero al menos un trabajador de prueba.</p>}</div>{actionError && <p role="alert" className="mt-3 text-xs text-red-700">{actionError}</p>}<div className="mt-4 flex justify-end gap-2"><Button variant="secondary" onClick={() => setAttendanceTraining(null)}>Cancelar</Button><Button onClick={() => void completeAttendance()} disabled={busy || employeesLoading || employeesError || selectedEmployeeIds.length === 0}>{busy ? "Registrando…" : `Registrar ${selectedEmployeeIds.length} asistencia(s)`}</Button></div></Card></div>}
    </DashboardShell>
  );
}
