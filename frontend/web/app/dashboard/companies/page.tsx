"use client";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { Building2, Edit, Plus, Search, Trash2, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, Badge, PageHeader, Button, Spinner, Empty } from "@/components/ui";
import { authAPI, companyAPI } from "@/lib/api";
import { useCompanies, type Company } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { statusBadge } from "@/lib/utils";

const syntheticForm = () => ({
  name: "ZU360 PILOTO SINTÉTICO - ",
  nit: `TEST-NIT-${Math.random().toString(36).slice(2, 8).toUpperCase()}`,
  ciiu: "TEST",
  arl: "ARL SINTÉTICA",
  legal_rep: "Persona sintética representante",
  legal_rep_email: "representante@zu360.test",
  legal_rep_phone: "", address: "", city: "", department: "",
});

const realForm = (legalRep = "", legalRepEmail = "") => ({
  name: "", nit: "", ciiu: "", arl: "",
  legal_rep: legalRep, legal_rep_email: legalRepEmail,
  legal_rep_phone: "", address: "", city: "", department: "",
});

type CompanyForm = ReturnType<typeof realForm>;

export default function CompaniesPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: companies, isLoading, isError } = useCompanies();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const user = useAuthStore((state) => state.user);
  // Synthetic pilot accounts keep the reserved test markers; real accounts
  // register real companies.
  const synthetic = user?.isSynthetic === true;
  const ownerName = [user?.firstName, user?.lastName].filter(Boolean).join(" ");
  const freshForm = () => (synthetic ? syntheticForm() : realForm(ownerName, user?.email || ""));
  const [query, setQuery] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Company | null>(null);
  const [form, setForm] = useState<CompanyForm>(freshForm);
  const [busy, setBusy] = useState(false);
  const [openingId, setOpeningId] = useState("");
  const [error, setError] = useState("");
  const filtered = useMemo(() => companies?.filter((company) => !query || company.name.toLowerCase().includes(query.toLowerCase()) || company.nit.toLowerCase().includes(query.toLowerCase())), [companies, query]);
  const isFirstCompany = !isLoading && !isError && (companies?.length ?? 0) === 0;
  const sameMode = (company: Company) => (company.is_synthetic === true) === synthetic;

  // A brand-new account lands here: open the registration form right away.
  useEffect(() => {
    if (isFirstCompany && !synthetic) { setEditing(null); setForm(freshForm()); setShowForm(true); }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isFirstCompany, synthetic]);

  const activateCompany = async (company: Company) => {
    if (!sameMode(company)) throw new Error("company classification does not match the account");
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
    catch { setError("No se pudo abrir la empresa. El servidor debe confirmar tu acceso antes de activarla."); }
    finally { setOpeningId(""); }
  };

  const openEdit = async (company: Company) => {
    if (openingId || busy) return;
    if (!sameMode(company)) { setError("Esta empresa no se puede editar desde esta cuenta."); return; }
    setOpeningId(company.id); setError("");
    try {
      await activateCompany(company);
      setEditing(company);
      setForm({
        ...freshForm(), name: company.name, nit: company.nit, ciiu: company.ciiu || "", arl: company.arl || "",
        legal_rep: company.legal_rep || "", legal_rep_email: company.legal_rep_email || "",
        legal_rep_phone: company.legal_rep_phone || "", address: company.address || "",
        city: company.city || "", department: company.department || "",
      });
      setShowForm(true);
    } catch { setError("No se pudo activar la empresa; el servidor debe confirmar tu acceso antes de editar."); }
    finally { setOpeningId(""); }
  };

  const validate = () => {
    if (synthetic) {
      if (!form.name.trim().toUpperCase().startsWith("ZU360 PILOTO SINTÉTICO") || !form.nit.trim().toUpperCase().startsWith("TEST-NIT-") || !form.legal_rep.trim().toLowerCase().startsWith("persona sintética") || (!form.legal_rep_email.toLowerCase().endsWith("@zu360.test") && !form.legal_rep_email.toLowerCase().endsWith(".invalid"))) {
        return "Nombre, identificador, representante y correo deben usar los marcadores sintéticos reservados.";
      }
      return "";
    }
    if (form.name.trim().length < 3) return "Escribe la razón social de la empresa.";
    if (!/^[0-9][0-9.\-]{4,19}$/.test(form.nit.trim())) return "El NIT debe tener solo números, puntos o guion (por ejemplo 900.123.456-7).";
    if (!/^[A-Za-z0-9]{1,10}$/.test(form.ciiu.trim())) return "El código CIIU debe tener entre 1 y 10 letras o números.";
    if (!form.arl.trim()) return "Indica la ARL de la empresa.";
    if (!form.legal_rep.trim()) return "Indica el representante legal.";
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.legal_rep_email.trim())) return "Escribe un correo válido para el representante legal.";
    return "";
  };

  const saveCompany = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (busy) return;
    const invalid = validate();
    if (invalid) { setError(invalid); return; }
    setBusy(true); setError("");
    try {
      if (editing) {
        const response = await companyAPI.update(editing.id, synthetic
          ? { name: form.name.trim(), arl: form.arl, legal_rep: form.legal_rep.trim(), legal_rep_email: form.legal_rep_email.trim() }
          : {
              name: form.name.trim(), arl: form.arl.trim(), legal_rep: form.legal_rep.trim(),
              legal_rep_email: form.legal_rep_email.trim(), legal_rep_phone: form.legal_rep_phone.trim(),
              address: form.address.trim(),
            });
        if (!sameMode(response.data)) throw new Error("unexpected company classification");
        useAuthStore.getState().setActiveCompany(response.data);
      } else {
        const created = await companyAPI.create(synthetic
          ? {
              name: form.name.trim(), nit: form.nit.trim().toUpperCase(), ciiu: "TEST", arl: "ARL SINTÉTICA",
              legal_rep: form.legal_rep.trim(), legal_rep_email: form.legal_rep_email.trim(),
              legal_rep_phone: "", address: "", city: "", department: "",
            }
          : {
              name: form.name.trim(), nit: form.nit.trim(), ciiu: form.ciiu.trim().toUpperCase(), arl: form.arl.trim(),
              legal_rep: form.legal_rep.trim(), legal_rep_email: form.legal_rep_email.trim(),
              legal_rep_phone: form.legal_rep_phone.trim(), address: form.address.trim(),
              city: form.city.trim(), department: form.department.trim(),
            });
        const company = created.data;
        if (!sameMode(company)) throw new Error("unexpected company classification");
        await activateCompany(company);
      }
      const refreshed = await companyAPI.list();
      useAuthStore.getState().setCompanies(refreshed.data.data);
      await queryClient.invalidateQueries({ queryKey: ["companies"] });
      setShowForm(false); setEditing(null); setForm(freshForm());
      if (!editing) router.push(`/dashboard/companies/${useAuthStore.getState().activeCompany?.id}`);
    } catch (cause) {
      const status = (cause as { response?: { status?: number } })?.response?.status;
      setError(status === 402
        ? "Tu plan no permite más empresas. Elige o amplía tu plan en Facturación para registrar otra."
        : status === 403
          ? "No tienes permiso para esta operación."
          : status === 409
            ? (synthetic ? "El identificador de prueba ya existe o no cumple las reglas." : "Ya existe una empresa registrada con ese NIT, o algún dato no es válido.")
            : "No se pudo guardar la empresa. Revisa los datos e inténtalo de nuevo.");
    } finally { setBusy(false); }
  };

  const archiveCompany = async (company: Company) => {
    if (!sameMode(company) || busy || openingId) return;
    if (activeCompany?.id === company.id) { setError("Cambia primero a otra empresa. No se puede archivar la empresa activa."); return; }
    if (!window.confirm(`¿Archivar la empresa “${company.name}”? Se realizará una baja lógica y dejará de aparecer en el portafolio.`)) return;
    setBusy(true); setError("");
    try {
      await companyAPI.delete(company.id);
      const refreshed = await companyAPI.list();
      useAuthStore.getState().setCompanies(refreshed.data.data);
      await queryClient.invalidateQueries({ queryKey: ["companies"] });
    } catch { setError("No se pudo archivar la empresa. Inténtalo de nuevo."); }
    finally { setBusy(false); }
  };

  const openCreate = () => { setEditing(null); setForm(freshForm()); setError(""); setShowForm(true); };
  const closeForm = () => { setShowForm(false); setEditing(null); };
  const newLabel = synthetic ? "Nueva empresa sintética" : "Registrar empresa";
  const idLabel = synthetic ? "ID de prueba" : "NIT";
  const input = "mt-1 w-full rounded border border-zu-border px-3 py-2";
  const field = (label: string, key: keyof CompanyForm, props: { required?: boolean; type?: string; maxLength?: number; placeholder?: string; disabled?: boolean } = {}) => (
    <label className="block text-xs">{label}<input value={form[key]} onChange={(event) => setForm({ ...form, [key]: event.target.value })} maxLength={props.maxLength ?? 255} className={`${input} disabled:bg-zu-surface2`} {...props}/></label>
  );

  return (
    <DashboardShell>
      <PageHeader
        title={synthetic ? "Empresas autorizadas" : "Tus empresas"}
        description={synthetic ? `${companies?.length ?? 0} memberships visibles para la cuenta autenticada.` : `${companies?.length ?? 0} empresas en tu cuenta.`}
        action={<Button onClick={openCreate}><Plus size={13}/> {newLabel}</Button>}
      />
      {error && !showForm && <p role="alert" className="mb-3 rounded-md border border-red-200 bg-red-50 p-3 text-xs text-red-800">{error}</p>}
      {isFirstCompany && !synthetic && (
        <Card className="mb-4 p-5">
          <h2 className="text-sm font-bold">Bienvenido{ownerName ? `, ${user?.firstName}` : ""}: registra tu primera empresa</h2>
          <p className="mt-1 text-xs text-zu-text2">Todo en ZU 360 se organiza por empresa. Registra la primera para empezar; luego puedes agregar más desde aquí.</p>
          <div className="mt-3"><Button onClick={openCreate}><Plus size={13}/> Registrar empresa</Button></div>
        </Card>
      )}
      <div className="mb-4 flex items-center gap-2"><div className="relative max-w-md flex-1"><Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-zu-text3"/><input value={query} onChange={(event) => setQuery(event.target.value)} placeholder={`Buscar por nombre o ${idLabel}…`} className="w-full rounded-md border border-zu-border bg-white py-2 pl-8 pr-3 text-xs outline-none focus:border-zu-accent"/></div></div>
      {isLoading ? <Spinner size={24}/> : isError ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar tus empresas.</p> : !filtered?.length ? <Empty icon={Building2} title={synthetic ? "No hay empresas autorizadas" : "Aún no tienes empresas"} description={query ? "Sin coincidencias." : synthetic ? "Crea una compañía ficticia para probar el portafolio multiempresa." : "Registra la primera empresa que acompañas."} action={!query ? <Button onClick={openCreate}><Plus size={13}/> {newLabel}</Button> : undefined}/> : (
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">{filtered.map((company) => <Card key={company.id} className="p-4 transition hover:shadow-zu-md">
          <div className="mb-3 flex items-start justify-between gap-3"><div className="min-w-0"><div className="truncate text-[13px] font-semibold">{company.name}</div><div className="font-mono text-[11px] text-zu-text3">{idLabel}: {company.nit}</div></div><Badge variant={statusBadge(company.status || "unknown")}>{company.status || "sin estado"}</Badge></div>
          <div className="flex flex-wrap items-center gap-2">{company.is_synthetic && <Badge variant="amber">SINTÉTICA</Badge>}<Badge variant="gray">CIIU {company.ciiu || "sin dato"}</Badge><Badge variant="gray">Riesgo SST: no calculado</Badge></div>
          <p className="mt-3 text-[11px] text-zu-text3">{openingId === company.id ? "Activando empresa…" : activeCompany?.id === company.id ? "Empresa activa." : "Ábrela para trabajar en ella."}</p>
          <div className="mt-3 flex flex-wrap gap-2"><Button size="sm" onClick={() => void activateAndOpen(company)} disabled={!!openingId || busy}><Building2 size={12}/> Abrir</Button><Button size="sm" variant="secondary" onClick={() => void openEdit(company)} disabled={!!openingId || busy || !sameMode(company)}><Edit size={12}/> Editar</Button><Button size="sm" variant="danger" onClick={() => void archiveCompany(company)} disabled={!!openingId || busy || !sameMode(company) || activeCompany?.id === company.id}><Trash2 size={12}/> Archivar</Button></div>
        </Card>)}</div>
      )}
      {showForm && <div className="fixed inset-0 z-[100] flex items-center justify-center overflow-y-auto bg-black/40 p-4"><Card className="my-auto w-full max-w-lg p-5">
        <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">{editing ? "Editar empresa" : synthetic ? "Nueva compañía ficticia" : "Registrar empresa"}{synthetic && <span className="ml-2"><Badge variant="amber">SINTÉTICA</Badge></span>}</h2><button onClick={closeForm} aria-label="Cerrar formulario"><X size={18}/></button></div>
        {synthetic
          ? <p className="mb-4 rounded-md bg-amber-50 p-2 text-[11px] text-amber-900">Nombre, identificador, representante y correo deben ser ficticios. Ubicación, teléfono, clasificación y documentos reales no se habilitan.</p>
          : <p className="mb-4 text-[11px] text-zu-text2">Los datos se declaran y no se verifican contra RUES, DIAN ni la ARL. La clase de riesgo queda pendiente de confirmación.</p>}
        <form onSubmit={saveCompany} className="space-y-3">
          {synthetic ? <>
            {field("Nombre sintético", "name", { required: true })}
            {field("Identificador reservado", "nit", { required: true, disabled: !!editing })}
            {field("Representante ficticio", "legal_rep", { required: true })}
            {field("Correo reservado de prueba", "legal_rep_email", { required: true, type: "email" })}
            <div className="grid grid-cols-2 gap-3"><label className="block text-xs">CIIU de prueba<input disabled value="TEST" className={`${input} bg-zu-surface2`}/></label><label className="block text-xs">ARL ficticia<input disabled value="ARL SINTÉTICA" className={`${input} bg-zu-surface2`}/></label></div>
          </> : <>
            {field("Razón social", "name", { required: true })}
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {field("NIT", "nit", { required: true, maxLength: 20, placeholder: "900.123.456-7", disabled: !!editing })}
              {field("Código CIIU", "ciiu", { required: true, maxLength: 10, placeholder: "4111", disabled: !!editing })}
              {field("ARL", "arl", { required: true, maxLength: 100 })}
              {field("Representante legal", "legal_rep", { required: true })}
              {field("Correo del representante", "legal_rep_email", { required: true, type: "email" })}
              {field("Teléfono (opcional)", "legal_rep_phone", { maxLength: 20 })}
              {field("Ciudad (opcional)", "city", { maxLength: 100, disabled: !!editing })}
              {field("Departamento (opcional)", "department", { maxLength: 100, disabled: !!editing })}
            </div>
            {field("Dirección (opcional)", "address")}
          </>}
          {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
          <div className="flex justify-end gap-2"><Button variant="secondary" onClick={closeForm}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : editing ? "Guardar cambios" : synthetic ? "Crear y activar" : "Registrar empresa"}</Button></div>
        </form>
      </Card></div>}
    </DashboardShell>
  );
}
