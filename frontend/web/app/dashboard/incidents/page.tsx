"use client";
import { FormEvent, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Eye, FileWarning, Plus, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { incidentAPI, type FurATPreview } from "@/lib/api";
import { useCompanyIncidents, type Incident } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const freshForm = () => ({
  occurred_at: new Date().toISOString().slice(0, 10),
  location: "SITIO SINTÉTICO - Laboratorio de pruebas",
  description: "SIMULACIÓN: ejercicio de reporte sintético para validar estados, sin lesión ni persona real.",
});

const statusText: Record<string, string> = { open: "Abierto", under_review: "En revisión", closed: "Cerrado", legal_hold: "Retención legal" };
const statusColor: Record<string, "blue" | "red" | "amber" | "green" | "gray"> = { open: "amber", under_review: "blue", closed: "green", legal_hold: "red" };

export default function IncidentsPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: incidents, isLoading, isError } = useCompanyIncidents(companyId);
  const [filter, setFilter] = useState("all");
  const [showForm, setShowForm] = useState(false);
  const [form, setForm] = useState(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);
  const [selected, setSelected] = useState<Incident | null>(null);
  const [preview, setPreview] = useState<FurATPreview | null>(null);
  const [actionError, setActionError] = useState("");

  const filtered = useMemo(() => (incidents || []).filter((incident) => filter === "all" || incident.status === filter), [incidents, filter]);
  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["incidents", companyId] });

  const createIncident = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setFormError("");
    try {
      await incidentAPI.create(companyId, {
        type: "incidente", occurred_at: form.occurred_at, location: form.location,
        description: form.description, injured_count: 0,
      });
      await invalidate(); setShowForm(false); setForm(freshForm());
    } catch {
      setFormError("No se pudo crear el incidente. Core sólo acepta el prefijo SINTÉTICO/SIMULACIÓN y cero personas lesionadas; revisa membership y modo sintético.");
    } finally { setBusy(false); }
  };

  const investigate = async (incident: Incident) => {
    setBusy(true); setActionError("");
    try {
      const response = await incidentAPI.investigate(companyId, incident.id);
      await invalidate(); setSelected(response.data);
    } catch { setActionError("No se pudo actualizar el estado de prueba; el servidor valida empresa y membership."); }
    finally { setBusy(false); }
  };

  const closeIncident = async (incident: Incident) => {
    if (!incident.is_synthetic || incident.legal_hold || !window.confirm("¿Cerrar este incidente sintético? Esta acción sólo modifica el registro de prueba.")) return;
    setBusy(true); setActionError("");
    try {
      const response = await incidentAPI.close(companyId, incident.id);
      await invalidate(); setSelected(response.data);
    } catch { setActionError("No se pudo cerrar. Puede estar en retención legal o pertenecer a otra compañía."); }
    finally { setBusy(false); }
  };

  const createFuratPreview = async (incident: Incident) => {
    setBusy(true); setActionError(""); setPreview(null);
    try { setPreview((await incidentAPI.furatPreview(companyId, incident.id)).data); }
    catch { setActionError("No se pudo generar la previsualización sintética."); }
    finally { setBusy(false); }
  };

  const openDetail = (incident: Incident) => { setSelected(incident); setPreview(null); setActionError(""); };

  return (
    <DashboardShell>
      <PageHeader title="Incidentes sintéticos" description="Ejercicios sin lesión, GPS, evidencia ni notificación externa." action={<Button onClick={() => { setForm(freshForm()); setFormError(""); setShowForm(true); }} disabled={!companyId}><Plus size={13}/> Crear simulación</Button>} />
      <div className="mb-4 rounded-md border border-amber-200 bg-amber-50 p-3 text-[11px] leading-relaxed text-amber-950">
        <strong>ENTORNO SINTÉTICO.</strong> Sólo se admite un evento “incidente” sin lesionados, con ubicación SINTÉTICA y descripción SIMULACIÓN. La previsualización FURAT no tiene validez y no se transmite a ARL ni autoridad.
      </div>
      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24}/> : isError ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar incidentes sintéticos de la empresa activa.</p> : (
        <>
          <div className="mb-4 flex flex-wrap gap-1.5">
            {[ ["all", "Todos"], ["open", "Abiertos"], ["under_review", "En revisión"], ["closed", "Cerrados"] ].map(([value, label]) => <button key={value} onClick={() => setFilter(value)} className={`rounded-md border px-3 py-1.5 text-[11px] font-semibold ${filter === value ? "border-transparent bg-zu-accent text-white" : "border-zu-border bg-white text-zu-text2 hover:bg-zu-surface2"}`}>{label}</button>)}
          </div>
          {!filtered.length ? <Empty icon={AlertTriangle} title="Sin incidentes sintéticos" description="Crea una simulación para probar el ciclo de reporte e investigación." /> : (
            <Card className="overflow-hidden p-0"><div className="overflow-x-auto"><table className="w-full border-collapse"><thead><tr>{["Fecha", "Tipo", "Ubicación de prueba", "Estado", "Marca", "Acciones"].map((label) => <Th key={label}>{label}</Th>)}</tr></thead><tbody>
              {filtered.map((incident) => <Tr key={incident.id}>
                <Td className="text-[11px]">{formatDateTime(incident.occurred_at)}</Td>
                <Td><span className="text-xs">Incidente sin lesión</span></Td>
                <Td className="max-w-xs truncate text-[11px]">{incident.location}</Td>
                <Td><Badge variant={statusColor[incident.status] || "gray"}>{statusText[incident.status] || incident.status}</Badge></Td>
                <Td><Badge variant={incident.is_synthetic ? "amber" : "red"}>{incident.is_synthetic ? "SINTÉTICO" : "NO CLASIFICADO"}</Badge></Td>
                <Td><Button size="sm" variant="secondary" onClick={() => openDetail(incident)}><Eye size={12}/> Ver</Button></Td>
              </Tr>)}
            </tbody></table></div><div className="border-t border-zu-border p-3 text-[11px] text-zu-text3">{filtered.length} registros sintéticos en la compañía activa.</div></Card>
          )}
        </>
      )}

      {showForm && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-xl p-5">
        <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo ejercicio de incidente</h2><button onClick={() => setShowForm(false)} aria-label="Cerrar"><X size={18}/></button></div>
        <p className="mb-4 text-[11px] text-zu-text2">Datos enviados a Core con `injured_count=0`. No uses nombres, cédulas, contactos ni ubicaciones reales.</p>
        <form onSubmit={createIncident} className="space-y-3">
          <label className="block text-xs">Fecha de simulación<input required type="date" value={form.occurred_at} max={new Date().toISOString().slice(0, 10)} onChange={(event) => setForm({ ...form, occurred_at: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
          <label className="block text-xs">Ubicación ficticia<input required value={form.location} onChange={(event) => setForm({ ...form, location: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/><span className="mt-1 block text-[10px] text-zu-text3">Debe comenzar con “SITIO SINTÉTICO”.</span></label>
          <label className="block text-xs">Descripción<input required minLength={20} value={form.description} onChange={(event) => setForm({ ...form, description: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/><span className="mt-1 block text-[10px] text-zu-text3">Debe comenzar con “SIMULACIÓN:” y no incluir datos personales.</span></label>
          {formError && <p role="alert" className="text-xs text-red-700">{formError}</p>}
          <div className="flex justify-end gap-2"><Button variant="secondary" onClick={() => setShowForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear simulación"}</Button></div>
        </form>
      </Card></div>}

      {selected && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-2xl p-5">
        <div className="mb-4 flex items-start justify-between gap-3"><div><h2 className="text-sm font-bold">Detalle del incidente</h2><p className="mt-1 text-[10px] text-zu-text3 font-mono">{selected.id}</p></div><button onClick={() => { setSelected(null); setPreview(null); setActionError(""); }} aria-label="Cerrar"><X size={18}/></button></div>
        {!selected.is_synthetic ? <p role="alert" className="rounded border border-red-200 bg-red-50 p-3 text-xs text-red-800">El servidor devolvió un registro sin marca sintética; no se muestra su contenido ni se permiten acciones.</p> : <>
          <div className="mb-3 flex flex-wrap gap-2"><Badge variant="amber">SINTÉTICO</Badge><Badge variant={statusColor[selected.status] || "gray"}>{statusText[selected.status] || selected.status}</Badge></div>
          <dl className="grid grid-cols-1 gap-3 text-xs sm:grid-cols-2"><div><dt className="text-zu-text3">Fecha</dt><dd>{formatDateTime(selected.occurred_at)}</dd></div><div><dt className="text-zu-text3">Ubicación</dt><dd>{selected.location}</dd></div><div className="sm:col-span-2"><dt className="text-zu-text3">Descripción</dt><dd>{selected.description}</dd></div></dl>
          <p className="mt-3 text-[10px] text-zu-text3">Lesionados: {selected.injured_count}. No contiene evidencia, GPS ni transmisión externa.</p>
          {selected.closed_at && <div className="mt-2 rounded border border-zu-border bg-zu-surface2 p-2 text-[10px] text-zu-text2"><strong>Auditoría de cierre sintético:</strong> {formatDateTime(selected.closed_at)}{selected.closed_by ? ` · usuario ${selected.closed_by}` : ""}</div>}
          {preview && <div className="mt-4 rounded-md border border-amber-300 bg-amber-50 p-3 text-xs"><div className="mb-1 font-bold text-amber-950"><FileWarning size={13} className="mr-1 inline"/>FURAT PREVIEW — NO VÁLIDO — SINTÉTICO</div><p>{preview.notice}</p></div>}
          {actionError && <p role="alert" className="mt-3 text-xs text-red-700">{actionError}</p>}
          <div className="mt-5 flex flex-wrap justify-end gap-2">
            <Button size="sm" variant="secondary" onClick={() => void createFuratPreview(selected)} disabled={busy}>Previsualizar FURAT</Button>
            {selected.status === "open" && <Button size="sm" variant="secondary" onClick={() => void investigate(selected)} disabled={busy}>Marcar en revisión</Button>}
            {(selected.status === "open" || selected.status === "under_review") && !selected.legal_hold && <Button size="sm" variant="danger" onClick={() => void closeIncident(selected)} disabled={busy}>Cerrar simulación</Button>}
          </div>
        </>}
      </Card></div>}
    </DashboardShell>
  );
}
