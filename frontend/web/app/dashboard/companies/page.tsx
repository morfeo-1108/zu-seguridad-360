"use client";
import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { Building2, Edit, Plus, Search, Trash2, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, Badge, PageHeader, Button, Spinner, Empty } from "@/components/ui";
import { authAPI, companyAPI } from "@/lib/api";
import { useCompanies, type Company } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { statusBadge } from "@/lib/utils";

const freshForm = () => ({
  name: "ZU360 PILOTO SINTÉTICO - ",
  nit: `TEST-NIT-${Math.random().toString(36).slice(2, 8).toUpperCase()}`,
  arl: "ARL SINTÉTICA",
  legal_rep: "Persona sintética representante",
  legal_rep_email: "representante@zu360.test",
});

type CompanyForm = ReturnType<typeof freshForm>;

export default function CompaniesPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: companies, isLoading, isError } = useCompanies();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const [query, setQuery] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Company | null>(null);
  const [form, setForm] = useState<CompanyForm>(freshForm);
  const [busy, setBusy] = useState(false);
  const [openingId, setOpeningId] = useState("");
  const [error, setError] = useState("");
  const filtered = useMemo(() => companies?.filter((company) => !query || company.name.toLowerCase().includes(query.toLowerCase()) || company.nit.toLowerCase().includes(query.toLowerCase())), [companies, query]);

  const activateCompany = async (company: Company) => {
    if (company.is_synthetic !== true) throw new Error("company is not classified synthetic");
    if (useAuthStore.getState().activeCompany?.id === company.id) return;
    const response = await authAPI.switchCompany(company.id);
    if (response.data.active_company_id !== company.id) throw new Error("active company mismatch");
    const auth = useAuthStore.getState();
    auth.setTokens(response.data.access_token, response.data.refresh_token);
    auth.setActiveCompany(company);
    if (auth.user) auth.setUser({ ...auth.user, companyId: company.id });
    queryClient.clear();
  };

  const activateAndOpen = async (company: Company) => {
    if (openingId || busy) return;
    setOpeningId(company.id); setError("");
    try { await activateCompany(company); router.push(`/dashboard/companies/${company.id}`); }
    catch { setError("No se pudo activar la compañía. Core debe confirmar la membership antes de emitir el token."); }
    finally { setOpeningId(""); }
  };

  const openEdit = async (company: Company) => {
    if (openingId || busy) return;
    if (company.is_synthetic !== true) { setError("Core no marcó esta compañía como sintética; no se edita desde el piloto."); return; }
    setOpeningId(company.id); setError("");
    try {
      await activateCompany(company);
      setEditing(company);
      setForm({ name: company.name, nit: company.nit, arl: company.arl || "ARL SINTÉTICA", legal_rep: company.legal_rep || "Persona sintética representante", legal_rep_email: company.legal_rep_email || "representante@zu360.test" });
      setShowForm(true);
    } catch { setError("No se pudo activar la compañía; Core debe confirmar la membership antes de editar."); }
    finally { setOpeningId(""); }
  };

  const saveCompany = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (busy) return;
    if (!form.name.trim().toUpperCase().startsWith("ZU360 PILOTO SINTÉTICO") || !form.nit.trim().toUpperCase().startsWith("TEST-NIT-") || !form.legal_rep.trim().toLowerCase().startsWith("persona sintética") || (!form.legal_rep_email.toLowerCase().endsWith("@zu360.test") && !form.legal_rep_email.toLowerCase().endsWith(".invalid"))) {
      setError("Nombre, identificador, representante y correo deben usar los marcadores sintéticos reservados."); return;
    }
    setBusy(true); setError("");
    try {
      if (editing) {
        const response = await companyAPI.update(editing.id, { name: form.name.trim(), arl: form.arl, legal_rep: form.legal_rep.trim(), legal_rep_email: form.legal_rep_email.trim() });
        if (response.data.is_synthetic !== true) throw new Error("Core returned a non-synthetic company");
        useAuthStore.getState().setActiveCompany(response.data);
      } else {
        const created = await companyAPI.create({
          name: form.name.trim(), nit: form.nit.trim().toUpperCase(), ciiu: "TEST", arl: "ARL SINTÉTICA",
          legal_rep: form.legal_rep.trim(), legal_rep_email: form.legal_rep_email.trim(),
          legal_rep_phone: "", address: "", city: "", department: "",
        });
        const company = created.data;
        if (company.is_synthetic !== true) throw new Error("Core did not classify the company as synthetic");
        await activateCompany(company);
      }
      const refreshed = await companyAPI.list();
      useAuthStore.getState().setCompanies(refreshed.data.data);
      await queryClient.invalidateQueries({ queryKey: ["companies"] });
      setShowForm(false); setEditing(null); setForm(freshForm());
      if (!editing) router.push(`/dashboard/companies/${useAuthStore.getState().activeCompany?.id}`);
    } catch (cause) {
      const status = (cause as { response?: { status?: number } })?.response?.status;
      setError(status === 403 ? "Core denegó la operación por falta de membership/permiso." : status === 409 ? "El identificador de prueba ya existe o no cumple las reglas." : "No se pudo guardar la compañía sintética. Revisa los datos reservados y el estado de Core.");
    } finally { setBusy(false); }
  };

  const archiveCompany = async (company: Company) => {
    if (!company.is_synthetic || busy || openingId) return;
    if (activeCompany?.id === company.id) { setError("Cambia primero a otra empresa. No se archiva la compañía activa del token."); return; }
    if (!window.confirm(`¿Archivar la empresa sintética “${company.name}”? Se realizará una baja lógica y dejará de aparecer en el portafolio.`)) return;
    setBusy(true); setError("");
    try {
      await companyAPI.delete(company.id);
      const refreshed = await companyAPI.list();
      useAuthStore.getState().setCompanies(refreshed.data.data);
      await queryClient.invalidateQueries({ queryKey: ["companies"] });
    } catch { setError("No se pudo archivar. Core valida el scope activo, la membership y la clasificación sintética."); }
    finally { setBusy(false); }
  };

  const openCreate = () => { setEditing(null); setForm(freshForm()); setError(""); setShowForm(true); };

  return (
    <DashboardShell>
      <PageHeader title="Empresas autorizadas" description={`${companies?.length ?? 0} memberships visibles para la cuenta autenticada.`} action={<Button onClick={openCreate}><Plus size={13}/> Nueva empresa sintética</Button>} />
      {error && <p role="alert" className="mb-3 rounded-md border border-red-200 bg-red-50 p-3 text-xs text-red-800">{error}</p>}
      <div className="mb-4 flex items-center gap-2"><div className="relative max-w-md flex-1"><Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-zu-text3"/><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Buscar por nombre o ID de prueba…" className="w-full rounded-md border border-zu-border bg-white py-2 pl-8 pr-3 text-xs outline-none focus:border-zu-accent"/></div></div>
      {isLoading ? <Spinner size={24}/> : isError ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar el portafolio autorizado desde Core.</p> : !filtered?.length ? <Empty icon={Building2} title="No hay empresas autorizadas" description={query ? "Sin coincidencias." : "Crea una compañía ficticia para probar el portafolio multiempresa."} action={!query ? <Button onClick={openCreate}><Plus size={13}/> Nueva empresa sintética</Button> : undefined}/> : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">{filtered.map((company) => <Card key={company.id} className="p-4 transition hover:shadow-zu-md">
          <div className="mb-3 flex items-start justify-between gap-3"><div className="min-w-0"><div className="truncate text-[13px] font-semibold">{company.name}</div><div className="font-mono text-[11px] text-zu-text3">ID de prueba: {company.nit}</div></div><Badge variant={statusBadge(company.status || "unknown")}>{company.status || "sin estado"}</Badge></div>
          <div className="flex flex-wrap items-center gap-2"><Badge variant={company.is_synthetic ? "amber" : "red"}>{company.is_synthetic ? "SINTÉTICA" : "NO CLASIFICADA"}</Badge><Badge variant="gray">CIIU {company.ciiu || "sin dato"}</Badge><Badge variant="gray">Riesgo SST: no calculado</Badge></div>
          <p className="mt-3 text-[11px] text-zu-text3">{openingId === company.id ? "Validando membership y activando token…" : activeCompany?.id === company.id ? "Empresa activa en el token." : "Membership autorizada; activa con un nuevo token para abrir."}</p>
          <div className="mt-3 flex flex-wrap gap-2"><Button size="sm" onClick={() => void activateAndOpen(company)} disabled={!!openingId || busy}><Building2 size={12}/> Abrir</Button><Button size="sm" variant="secondary" onClick={() => void openEdit(company)} disabled={!!openingId || busy || !company.is_synthetic}><Edit size={12}/> Editar</Button><Button size="sm" variant="danger" onClick={() => void archiveCompany(company)} disabled={!!openingId || busy || !company.is_synthetic || activeCompany?.id === company.id}><Trash2 size={12}/> Archivar</Button></div>
        </Card>)}</div>
      )}
      {showForm && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-lg p-5">
        <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">{editing ? "Editar empresa sintética" : "Nueva compañía ficticia"}<span className="ml-2"><Badge variant="amber">SINTÉTICA</Badge></span></h2><button onClick={() => { setShowForm(false); setEditing(null); }} aria-label="Cerrar formulario"><X size={18}/></button></div>
        <p className="mb-4 rounded-md bg-amber-50 p-2 text-[11px] text-amber-900">Nombre, identificador, representante y correo deben ser ficticios. Ubicación, teléfono, clasificación y documentos reales no se habilitan.</p>
        <form onSubmit={saveCompany} className="space-y-3"><label className="block text-xs">Nombre sintético<input required minLength={3} maxLength={255} value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label><label className="block text-xs">Identificador reservado<input required disabled={!!editing} value={form.nit} onChange={(event) => setForm({ ...form, nit: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2 font-mono disabled:bg-zu-surface2"/></label><label className="block text-xs">Representante ficticio<input required value={form.legal_rep} onChange={(event) => setForm({ ...form, legal_rep: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label><label className="block text-xs">Correo reservado de prueba<input required type="email" value={form.legal_rep_email} onChange={(event) => setForm({ ...form, legal_rep_email: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label><div className="grid grid-cols-2 gap-3"><label className="block text-xs">CIIU de prueba<input disabled value="TEST" className="mt-1 w-full rounded border border-zu-border bg-zu-surface2 px-3 py-2"/></label><label className="block text-xs">ARL ficticia<input disabled value="ARL SINTÉTICA" className="mt-1 w-full rounded border border-zu-border bg-zu-surface2 px-3 py-2"/></label></div>
          {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
          <div className="flex justify-end gap-2"><Button variant="secondary" onClick={() => { setShowForm(false); setEditing(null); }}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : editing ? "Guardar cambios" : "Crear y activar"}</Button></div>
        </form>
      </Card></div>}
    </DashboardShell>
  );
}
