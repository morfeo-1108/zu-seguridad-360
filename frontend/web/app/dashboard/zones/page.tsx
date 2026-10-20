"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Layers, Plus, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { zoneAPI } from "@/lib/api";
import { useProjects } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const zoneTypeLabel: Record<string, string> = {
  alturas: "Trabajo en alturas", espacio_confinado: "Espacio confinado", soldadura: "Soldadura",
  electrico: "Eléctrico", excavacion: "Excavación", quimicos: "Químicos", general: "General",
};
const riskColor: Record<string, "red" | "amber" | "green"> = { alto: "red", medio: "amber", bajo: "green" };

export default function ZonesPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: projects, isLoading, isError } = useProjects(companyId);
  const [showProjectForm, setShowProjectForm] = useState(false);
  const [projectForm, setProjectForm] = useState({ name: "", address: "" });
  const [zoneFormFor, setZoneFormFor] = useState<string | null>(null);
  const [zoneForm, setZoneForm] = useState({ name: "", type: "general" });
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["projects", companyId] });

  const createProject = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setError("");
    try {
      await zoneAPI.createProject(companyId, projectForm);
      await invalidate(); setShowProjectForm(false); setProjectForm({ name: "", address: "" });
    } catch { setError("No se pudo crear el proyecto."); }
    finally { setBusy(false); }
  };

  const createZone = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || !zoneFormFor || busy) return;
    setBusy(true); setError("");
    try {
      await zoneAPI.createZone(companyId, zoneFormFor, zoneForm);
      await invalidate(); setZoneFormFor(null); setZoneForm({ name: "", type: "general" });
    } catch { setError("No se pudo crear la zona."); }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader
        title="Zonas y protocolos"
        description="Catálogo de EPP requerido por tipo de zona — reconocimiento por GPS/marcador visual, visión por dron y biometría siguen apagados."
        action={<Button onClick={() => { setProjectForm({ name: "", address: "" }); setError(""); setShowProjectForm(true); }} disabled={!companyId}><Plus size={13} /> Nuevo proyecto</Button>}
      />

      <Card className="mb-5 border-amber-300 bg-amber-50 p-3.5">
        <div className="flex items-start gap-2.5">
          <AlertTriangle size={16} className="mt-0.5 shrink-0 text-amber-800" />
          <p className="text-xs leading-relaxed text-amber-900">
            <span className="font-bold">Borrador normativo, no validado profesionalmente.</span> Las normas citadas para trabajo en alturas y espacios confinados aún dicen &quot;pendiente de validación&quot; a propósito — no las uses como única fuente de cumplimiento sin que un profesional SST las confirme contra la norma vigente.
          </p>
        </div>
      </Card>

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar los proyectos de la empresa activa.</p>
      ) : !projects?.length ? (
        <Empty icon={Layers} title="Sin proyectos registrados" description="Crea el primer proyecto/obra para definir sus zonas." />
      ) : (
        <div className="space-y-4">
          {projects.map((project) => (
            <Card key={project.id} className="p-4">
              <div className="mb-3 flex items-center justify-between">
                <div><h2 className="text-sm font-bold">{project.name}</h2>{project.address && <p className="text-[11px] text-zu-text3">{project.address}</p>}</div>
                <Button size="sm" variant="secondary" onClick={() => { setZoneForm({ name: "", type: "general" }); setError(""); setZoneFormFor(project.id); }}>+ Zona</Button>
              </div>
              {!project.zones?.length ? (
                <p className="py-3 text-center text-xs text-zu-text3">Sin zonas definidas.</p>
              ) : (
                <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
                  {project.zones.map((zone) => (
                    <div key={zone.id} className="rounded-md border border-zu-border p-3">
                      <div className="mb-1.5 flex items-center justify-between gap-2">
                        <p className="text-xs font-semibold">{zone.name}</p>
                        <Badge variant={riskColor[zone.risk_class] || "gray"}>riesgo {zone.risk_class}</Badge>
                      </div>
                      <p className="mb-2 text-[11px] text-zu-text3">{zoneTypeLabel[zone.type] || zone.type} · {zone.norm_ref}</p>
                      <div className="flex flex-wrap gap-1">
                        {zone.checklist.map((item) => (
                          <Badge key={item.id} variant={item.mandatory ? "blue" : "gray"}>{item.code}</Badge>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </Card>
          ))}
        </div>
      )}

      {showProjectForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo proyecto</h2><button onClick={() => setShowProjectForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createProject} className="space-y-3">
              <label className="block text-xs">Nombre<input required value={projectForm.name} onChange={(e) => setProjectForm({ ...projectForm, name: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Dirección<input value={projectForm.address} onChange={(e) => setProjectForm({ ...projectForm, address: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowProjectForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {zoneFormFor && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nueva zona</h2><button onClick={() => setZoneFormFor(null)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createZone} className="space-y-3">
              <label className="block text-xs">Nombre<input required value={zoneForm.name} onChange={(e) => setZoneForm({ ...zoneForm, name: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Tipo
                <select value={zoneForm.type} onChange={(e) => setZoneForm({ ...zoneForm, type: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  {Object.entries(zoneTypeLabel).map(([value, label]) => <option key={value} value={value}>{label}</option>)}
                </select>
              </label>
              <p className="text-[10px] text-zu-text3">El checklist de EPP se siembra automáticamente según el tipo de zona.</p>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setZoneFormFor(null)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear"}</Button></div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
