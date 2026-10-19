"use client";
import { useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, Building2, GraduationCap, ShieldAlert, Users } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, KPICard, PageHeader, Spinner } from "@/components/ui";
import { authAPI } from "@/lib/api";
import { useCompanies, useCompany, useCompanyIncidents, useCompanyTrainings, useEmployees } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

export default function CompanyDetailPage() {
  const params = useParams();
  const companyId = typeof params.id === "string" ? params.id : "";
  const router = useRouter();
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const user = useAuthStore((state) => state.user);
  const { data: companies, isLoading: membershipsLoading } = useCompanies();
  const authorizedCompany = companies?.find((company) => company.id === companyId && company.is_synthetic === true);
  const scopeMatches = activeCompany?.id === companyId && !!authorizedCompany;
  const scopedCompanyId = scopeMatches ? companyId : "";
  const { data: company, isLoading: companyLoading, isError: companyError } = useCompany(scopedCompanyId);
  const { data: employees, isLoading: employeesLoading, isError: employeesError } = useEmployees(scopedCompanyId);
  const { data: incidents, isLoading: incidentsLoading, isError: incidentsError } = useCompanyIncidents(scopedCompanyId);
  const { data: trainings, isLoading: trainingsLoading, isError: trainingsError } = useCompanyTrainings(scopedCompanyId);
  const [switching, setSwitching] = useState(false);
  const [switchError, setSwitchError] = useState("");
  const hasError = companyError || employeesError || incidentsError || trainingsError;
  const loading = membershipsLoading || (scopeMatches && (companyLoading || employeesLoading || incidentsLoading || trainingsLoading));

  const activateCompany = async () => {
    if (!authorizedCompany || switching) return;
    setSwitching(true); setSwitchError("");
    try {
      const response = await authAPI.switchCompany(companyId);
      if (response.data.active_company_id !== companyId) throw new Error("company scope mismatch");
      const auth = useAuthStore.getState();
      auth.setTokens(response.data.access_token, response.data.refresh_token);
      auth.setActiveCompany(authorizedCompany);
      if (user) auth.setUser({ ...user, companyId });
      queryClient.clear();
    } catch { setSwitchError("Core no confirmó una membership activa; no se cargaron datos de esta compañía."); }
    finally { setSwitching(false); }
  };

  return <DashboardShell>
    <PageHeader title="Detalle de empresa" description="Resumen de registros sintéticos consultados desde Core; sin puntajes ni afirmaciones normativas." action={<Button variant="secondary" onClick={() => router.push("/dashboard/companies")}><ArrowLeft size={13}/> Portafolio</Button>} />
    {loading ? <Spinner size={24}/> : !authorizedCompany ? <Empty icon={ShieldAlert} title="Empresa no autorizada" description="No hay membership sintética disponible para la empresa solicitada."/> : !scopeMatches ? <Card className="max-w-xl border-amber-300 bg-amber-50 p-5"><div className="flex items-start gap-3"><Building2 size={18} className="mt-0.5 text-amber-800"/><div><h2 className="text-sm font-bold text-amber-950">Esta empresa no es la activa del token</h2><p className="mt-1 text-xs text-amber-900">Antes de leer datos, Core emitirá un nuevo JWT sólo si la membership sigue vigente.</p>{switchError && <p role="alert" className="mt-2 text-xs text-red-700">{switchError}</p>}<Button className="mt-4" onClick={() => void activateCompany()} disabled={switching}>{switching ? "Validando…" : "Activar empresa y cargar"}</Button></div></div></Card> : hasError || !company ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar un resumen completo desde Core. No se reemplazan errores por métricas de demostración.</p> : !company.is_synthetic ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">La empresa no está clasificada como sintética; no se muestran sus datos.</p> : <>
      <div className="mb-4 flex items-center gap-2"><Badge variant="amber">SINTÉTICA</Badge><Badge variant="gray">ID reservado {company.nit}</Badge><Badge variant="gray">Estado {company.status || "sin estado"}</Badge></div>
      <Card className="mb-5 border-amber-200 bg-amber-50 p-4"><h2 className="text-sm font-bold text-amber-950">{company.name}</h2><p className="mt-1 text-[11px] text-amber-900">El contexto activo coincide con el company_id del token. Los endpoints verifican tenant, empresa y membership en servidor.</p></Card>
      <div className="mb-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4"><KPICard label="Perfiles sintéticos" value={employees?.total ?? "—"} icon={Users} sub="desde Core · compañía activa"/><KPICard label="Incidentes sintéticos" value={incidents?.length ?? "—"} icon={ShieldAlert} sub="sin lesiones ni reporte externo"/><KPICard label="Cursos de prueba" value={trainings?.length ?? "—"} icon={GraduationCap} sub="sin validez normativa"/><KPICard label="Cumplimiento SST" value="No calculado" icon={Building2} sub="sin puntuación ni dictamen legal"/></div>
      <div className="grid grid-cols-1 gap-3 md:grid-cols-3"><Link href="/dashboard/employees"><Card className="p-4 transition hover:shadow-zu-md"><h3 className="text-sm font-semibold">Personas de prueba</h3><p className="mt-1 text-[11px] text-zu-text3">Consultar, crear, editar y eliminar perfiles sintéticos.</p></Card></Link><Link href="/dashboard/incidents"><Card className="p-4 transition hover:shadow-zu-md"><h3 className="text-sm font-semibold">Incidentes sintéticos</h3><p className="mt-1 text-[11px] text-zu-text3">Crear, investigar y cerrar ejercicios sin lesión.</p></Card></Link><Link href="/dashboard/training"><Card className="p-4 transition hover:shadow-zu-md"><h3 className="text-sm font-semibold">Capacitaciones de prueba</h3><p className="mt-1 text-[11px] text-zu-text3">Programar curso y registrar asistencia inválida.</p></Card></Link></div>
      {hasError && <p role="alert" className="mt-4 text-xs text-red-700">Una o más cifras no están disponibles; revisa los servicios de Core.</p>}
    </>}
  </DashboardShell>;
}
