"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Gavel, Plus, Trash2, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { committeeAPI, type CommitteeMember } from "@/lib/api";
import { useCommitteeMeetings, useCompanyCommittees, useEmployees, type Committee } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const typeText: Record<string, string> = { copasst: "COPASST", vigia: "Vigía Ocupacional" };
const roleText: Record<string, string> = { representante_empleador: "Rep. empleador", representante_trabajadores: "Rep. trabajadores" };

const freshCommitteeForm = () => ({ type: "vigia", period_start: new Date().toISOString().slice(0, 10), period_end: "", members: [] as CommitteeMember[] });
const freshMeetingForm = () => ({ held_at: new Date().toISOString().slice(0, 10), agenda: "", minutes: "", attendees: [] as string[] });

export default function CommitteesPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: committees, isLoading, isError } = useCompanyCommittees(companyId);
  const { data: employeeResponse } = useEmployees(companyId);
  const employees = employeeResponse?.data || [];

  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshCommitteeForm);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selected, setSelected] = useState<Committee | null>(null);
  const [showMeetingForm, setShowMeetingForm] = useState(false);
  const [meetingForm, setMeetingForm] = useState(freshMeetingForm);

  const { data: meetings } = useCommitteeMeetings(companyId, selected?.id || "");

  const createCommittee = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy || form.members.length < 2) { setError("Se requieren al menos 2 miembros."); return; }
    setBusy(true); setError("");
    try {
      await committeeAPI.create(companyId, form);
      await queryClient.invalidateQueries({ queryKey: ["committees", companyId] });
      setShowForm(false); setForm(freshCommitteeForm());
    } catch { setError("No se pudo crear el comité. Verifica que todos los miembros pertenezcan a la empresa activa."); }
    finally { setBusy(false); }
  };

  const createMeeting = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selected || busy) return;
    setBusy(true); setError("");
    try {
      await committeeAPI.createMeeting(companyId, selected.id, { ...meetingForm, decisions: [] });
      await queryClient.invalidateQueries({ queryKey: ["committee-meetings", companyId, selected.id] });
      setShowMeetingForm(false); setMeetingForm(freshMeetingForm());
    } catch { setError("No se pudo registrar el acta."); }
    finally { setBusy(false); }
  };

  const addMember = () => setForm({ ...form, members: [...form.members, { employee_id: "", role: "representante_trabajadores", is_alternate: false }] });
  const updateMember = (i: number, patch: Partial<CommitteeMember>) => setForm({ ...form, members: form.members.map((m, idx) => (idx === i ? { ...m, ...patch } : m)) });

  return (
    <DashboardShell>
      <PageHeader title="COPASST / Vigía Ocupacional" description="Comité paritario por periodo, con roster de miembros y actas de reunión." action={<Button onClick={() => { setForm(freshCommitteeForm()); setError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13} /> Nuevo comité</Button>} />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar los comités.</p>
      ) : !committees?.length ? (
        <Empty icon={Gavel} title="Sin comité registrado" description="Crea el COPASST o Vigía Ocupacional vigente de esta empresa." />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2">
          {committees.map((c) => (
            <Card key={c.id} className="p-4 cursor-pointer hover:border-zu-accent" onClick={() => setSelected(c)}>
              <div className="mb-2 flex items-center justify-between"><Badge variant="blue">{typeText[c.type] || c.type}</Badge><span className="text-[11px] text-zu-text3">{c.members.length} miembros</span></div>
              <div className="text-[12px] text-zu-text2">{formatDateTime(c.period_start)} — {formatDateTime(c.period_end)}</div>
            </Card>
          ))}
        </div>
      )}

      {showForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-xl max-h-[90vh] overflow-y-auto p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo comité</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createCommittee} className="space-y-3">
              <div className="grid grid-cols-3 gap-3">
                <label className="block text-xs">Tipo<select value={form.type} onChange={(e) => setForm({ ...form, type: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"><option value="vigia">Vigía (&lt;10 trab.)</option><option value="copasst">COPASST (&gt;=10 trab.)</option></select></label>
                <label className="block text-xs">Inicio periodo<input required type="date" value={form.period_start} onChange={(e) => setForm({ ...form, period_start: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
                <label className="block text-xs">Fin periodo<input required type="date" value={form.period_end} onChange={(e) => setForm({ ...form, period_end: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              </div>
              <div>
                <div className="mb-1.5 flex items-center justify-between"><span className="text-xs font-semibold">Miembros</span><button type="button" onClick={addMember} className="text-[11px] font-semibold text-zu-accent">+ Agregar miembro</button></div>
                <div className="space-y-2">
                  {form.members.map((m, i) => (
                    <div key={i} className="flex items-center gap-2 rounded-md border border-zu-border p-2">
                      <select required value={m.employee_id} onChange={(e) => updateMember(i, { employee_id: e.target.value })} className="flex-1 rounded border border-zu-border px-2 py-1.5 text-[11px]">
                        <option value="">Trabajador…</option>
                        {employees.map((emp) => <option key={emp.id} value={emp.id}>{emp.first_name} {emp.last_name}</option>)}
                      </select>
                      <select value={m.role} onChange={(e) => updateMember(i, { role: e.target.value })} className="rounded border border-zu-border px-2 py-1.5 text-[11px]">
                        <option value="representante_trabajadores">Rep. trabajadores</option>
                        <option value="representante_empleador">Rep. empleador</option>
                      </select>
                      <button type="button" onClick={() => setForm({ ...form, members: form.members.filter((_, idx) => idx !== i) })} className="text-red-500"><Trash2 size={14} /></button>
                    </div>
                  ))}
                </div>
              </div>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear comité"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {selected && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-2xl max-h-[90vh] overflow-y-auto p-5">
            <div className="mb-4 flex items-start justify-between">
              <div><h2 className="text-sm font-bold">{typeText[selected.type]}</h2><p className="mt-1 text-[11px] text-zu-text3">{formatDateTime(selected.period_start)} — {formatDateTime(selected.period_end)}</p></div>
              <button onClick={() => setSelected(null)} aria-label="Cerrar"><X size={18} /></button>
            </div>
            <h3 className="mb-2 text-xs font-bold">Miembros</h3>
            <div className="mb-4 space-y-1">
              {selected.members.map((m, i) => (
                <div key={i} className="flex items-center justify-between rounded-md border border-zu-border px-3 py-1.5 text-[11px]">
                  <span>{employees.find((e) => e.id === m.employee_id)?.first_name || m.employee_id}</span>
                  <Badge variant="gray">{roleText[m.role] || m.role}</Badge>
                </div>
              ))}
            </div>
            <div className="mb-2 flex items-center justify-between"><h3 className="text-xs font-bold">Actas</h3><Button size="sm" variant="secondary" onClick={() => { setMeetingForm(freshMeetingForm()); setShowMeetingForm(true); }}><Plus size={12} /> Nueva acta</Button></div>
            {!meetings?.length ? <p className="text-[11px] text-zu-text3">Sin actas registradas.</p> : (
              <div className="space-y-2">
                {meetings.map((m) => (
                  <Card key={m.id} className="p-3">
                    <div className="mb-1 flex items-center justify-between"><span className="text-[11px] font-semibold">{formatDateTime(m.held_at)}</span></div>
                    <p className="text-[11px] text-zu-text2"><strong>Agenda:</strong> {m.agenda}</p>
                    <p className="mt-1 text-[11px] text-zu-text2"><strong>Acta:</strong> {m.minutes}</p>
                  </Card>
                ))}
              </div>
            )}
          </Card>
        </div>
      )}

      {showMeetingForm && selected && (
        <div className="fixed inset-0 z-[110] flex items-center justify-center bg-black/50 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nueva acta</h2><button onClick={() => setShowMeetingForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createMeeting} className="space-y-3">
              <label className="block text-xs">Fecha<input required type="date" value={meetingForm.held_at} onChange={(e) => setMeetingForm({ ...meetingForm, held_at: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Agenda<input required value={meetingForm.agenda} onChange={(e) => setMeetingForm({ ...meetingForm, agenda: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Acta / minuta<textarea required value={meetingForm.minutes} onChange={(e) => setMeetingForm({ ...meetingForm, minutes: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" rows={3} /></label>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowMeetingForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Registrar acta"}</Button></div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
