"use client";
import { FormEvent, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { Edit, Plus, Search, Trash2, UserRound, Users, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { employeeAPI } from "@/lib/api";
import { useCompanies, useEmployees, type Employee } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const freshForm = () => ({
  first_name: "Persona",
  last_name: "Sintético Demo",
  document_number: `SYN-${Math.random().toString(36).slice(2, 10).toUpperCase()}`,
  email: "persona.demo@zu360.test",
  position: "Demostración SST",
  department: "Sintético",
});

type EmployeeForm = ReturnType<typeof freshForm>;

export default function EmployeesPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: companies, isLoading: companiesLoading, isError: companiesError } = useCompanies();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const [search, setSearch] = useState("");
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Employee | null>(null);
  const [form, setForm] = useState<EmployeeForm>(freshForm);
  const [formError, setFormError] = useState("");
  const [busy, setBusy] = useState(false);

  const selectedId = activeCompany?.id || "";
  const selectedCompany = companies?.find((company) => company.id === selectedId);
  const { data: employeeResponse, isLoading, isError } = useEmployees(selectedId);
  const employees = employeeResponse?.data || [];
  const filtered = useMemo(() => employees.filter((employee) => {
    const q = search.trim().toLocaleLowerCase();
    return !q || `${employee.first_name} ${employee.last_name} ${employee.document_number}`.toLocaleLowerCase().includes(q);
  }), [employees, search]);

  const openCreate = () => { setEditing(null); setForm(freshForm()); setFormError(""); setShowForm(true); };
  const openEdit = (employee: Employee) => {
    if (!employee.is_synthetic) return;
    setEditing(employee);
    setForm({ first_name: employee.first_name, last_name: employee.last_name, document_number: employee.document_number, email: employee.email || "", position: employee.position || "Demostración SST", department: employee.department || "Sintético" });
    setFormError(""); setShowForm(true);
  };

  const saveEmployee = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!selectedId || busy) return;
    setBusy(true); setFormError("");
    try {
      if (editing) {
        await employeeAPI.update(selectedId, editing.id, { position: form.position, department: form.department, email: form.email });
      } else {
        await employeeAPI.create(selectedId, {
          ...form, document_type: "SYN", phone: "", contract_type: "service",
        });
      }
      await queryClient.invalidateQueries({ queryKey: ["employees", selectedId] });
      setShowForm(false); setEditing(null); setForm(freshForm());
    } catch (cause) {
      const status = (cause as { response?: { status?: number } })?.response?.status;
      setFormError(status === 403 ? "No tienes membership activa para esta empresa." : status === 404 ? "La API de datos laborales no está habilitada para este entorno sintético." : "No se pudo guardar. Core sólo acepta perfiles y campos explícitamente sintéticos; verifica prefijos y dominio de correo reservado.");
    } finally { setBusy(false); }
  };

  const deleteEmployee = async (employee: Employee) => {
    if (!selectedId || !employee.is_synthetic || !window.confirm(`¿Eliminar el perfil sintético ${employee.first_name} ${employee.last_name}?`)) return;
    try {
      await employeeAPI.delete(selectedId, employee.id);
      await queryClient.invalidateQueries({ queryKey: ["employees", selectedId] });
    } catch { window.alert("No se pudo eliminar el perfil sintético. La empresa y membership se verifican en el servidor."); }
  };

  if (companiesLoading) return <DashboardShell><Spinner size={24} /></DashboardShell>;
  if (companiesError) return <DashboardShell><PageHeader title="Trabajadores sintéticos"/><p role="alert" className="text-sm text-red-700">No se pudieron cargar las empresas autorizadas desde Core.</p></DashboardShell>;

  return (
    <DashboardShell>
      <PageHeader title="Trabajadores sintéticos" description="Sólo perfiles ficticios del entorno aislado; no ingresar información real." action={<Button onClick={openCreate} disabled={!selectedId}><Plus size={13}/> Nuevo perfil de prueba</Button>} />
      <div className="mb-4 flex flex-wrap items-center gap-3"><span className="text-xs text-zu-text2">Empresa activa:</span><span className="max-w-md truncate text-xs font-semibold">{selectedCompany?.name || "sin membership activa"}</span><div className="relative min-w-[220px] flex-1"><Search size={13} className="absolute left-3 top-1/2 -translate-y-1/2 text-zu-text3"/><input value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Buscar persona sintética…" className="w-full rounded-md border border-zu-border bg-white py-2 pl-8 pr-3 text-xs outline-none focus:border-zu-accent"/></div></div>
      {selectedCompany && <div className="mb-3 text-[11px] text-zu-text3">El listado y cada acción usan el company_id firmado en el token activo. Para cambiar scope, usa el selector superior.</div>}
      {!selectedId ? <Empty icon={Users} title="Sin empresas asignadas" description="El usuario autenticado no tiene memberships disponibles."/> : isLoading ? <Spinner size={24}/> : isError ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron consultar perfiles sintéticos para esta empresa.</p> : filtered.length === 0 ? <Empty icon={UserRound} title="Sin perfiles sintéticos" description={search ? "No hay coincidencias." : "Crea un perfil ficticio para probar el flujo de trabajadores."} action={!search ? <Button onClick={openCreate}><Plus size={13}/> Crear persona de prueba</Button> : undefined}/> : (
        <Card className="overflow-hidden p-0"><div className="overflow-x-auto"><table className="w-full border-collapse"><thead><tr className="border-b border-zu-border bg-zu-surface2">{["Persona de prueba", "Identificador reservado", "Cargo de demostración", "Área", "Estado", "Acciones"].map((title) => <Th key={title}>{title}</Th>)}</tr></thead><tbody>
          {filtered.map((employee) => <Tr key={employee.id}><Td><button onClick={() => router.push(`/dashboard/employees/${employee.id}`)} className="text-left font-semibold hover:text-zu-accent">{employee.first_name} {employee.last_name}</button><div className="mt-1"><Badge variant={employee.is_synthetic ? "amber" : "red"}>{employee.is_synthetic ? "SINTÉTICO" : "NO CLASIFICADO"}</Badge></div></Td><Td className="font-mono text-[11px]">{employee.document_number}</Td><Td>{employee.position || "—"}</Td><Td>{employee.department || "—"}</Td><Td><Badge variant={employee.status === "active" ? "green" : "gray"}>{employee.status}</Badge></Td><Td><div className="flex gap-1"><Button variant="secondary" size="sm" disabled={!employee.is_synthetic} onClick={() => openEdit(employee)}><Edit size={12}/> Editar</Button><Button variant="danger" size="sm" disabled={!employee.is_synthetic} onClick={() => void deleteEmployee(employee)}><Trash2 size={12}/> Eliminar</Button></div></Td></Tr>)}
        </tbody></table></div><div className="border-t border-zu-border p-3 text-[11px] text-zu-text3">Mostrando {filtered.length} de {employeeResponse?.total ?? filtered.length} registros sintéticos.</div></Card>
      )}
      {showForm && <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4"><Card className="w-full max-w-lg p-5"><div className="mb-4 flex items-center justify-between"><h2 className="text-sm font-bold">{editing ? "Editar perfil sintético" : "Crear persona ficticia"}<span className="ml-2"><Badge variant="amber">SINTÉTICO</Badge></span></h2><button onClick={() => setShowForm(false)} aria-label="Cerrar formulario"><X size={18}/></button></div>
        <p className="mb-4 rounded-md bg-amber-50 p-2 text-[11px] text-amber-900">No ingreses datos reales. Los identificadores, cargos, áreas y correos deben estar claramente marcados como ficticios.</p>
        <form onSubmit={saveEmployee} className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <label className="text-xs">Nombre<input required disabled={!!editing} value={form.first_name} onChange={(event) => setForm({ ...form, first_name: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2 disabled:bg-zu-surface2"/></label>
          <label className="text-xs">Apellido<input required disabled={!!editing} value={form.last_name} onChange={(event) => setForm({ ...form, last_name: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2 disabled:bg-zu-surface2"/></label>
          <label className="text-xs">ID de prueba<input required disabled={!!editing} value={form.document_number} onChange={(event) => setForm({ ...form, document_number: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2 font-mono disabled:bg-zu-surface2"/></label>
          <label className="text-xs">Correo reservado de prueba<input type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/></label>
          <label className="text-xs">Cargo ficticio<input required value={form.position} onChange={(event) => setForm({ ...form, position: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/><span className="text-[10px] text-zu-text3">Prefijo: Demostración/Prueba/Sintético.</span></label>
          <label className="text-xs">Área ficticia<input value={form.department} onChange={(event) => setForm({ ...form, department: event.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2"/><span className="text-[10px] text-zu-text3">Prefijo: Demostración/Prueba/Sintético.</span></label>
          {formError && <p role="alert" className="text-xs text-red-700 sm:col-span-2">{formError}</p>}
          <div className="flex justify-end gap-2 sm:col-span-2"><Button variant="secondary" onClick={() => { setShowForm(false); setEditing(null); }}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : editing ? "Guardar cambios" : "Guardar perfil sintético"}</Button></div>
        </form>
      </Card></div>}
    </DashboardShell>
  );
}
