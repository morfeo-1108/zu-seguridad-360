"use client";
import Link from "next/link";
import { Building2, Users, Activity, ArrowRight, ShieldCheck, BellRing, FlaskConical, Briefcase } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, Badge, KPICard, Spinner } from "@/components/ui";
import { useAlerts, useCompanies, useEmployees, useCompanyIncidents } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { REAL_MODULES } from "@/lib/modules";
import { formatDateTime } from "@/lib/utils";

const alertCategoryLabel: Record<string, string> = {
  medical_exam: "Examen médico", epp: "EPP", risk_review: "Revisión de riesgo", legal_requirement: "Requisito legal",
};

function greeting() {
  const h = new Date().getHours();
  if (h < 12) return "Buenos días";
  if (h < 19) return "Buenas tardes";
  return "Buenas noches";
}

export default function DashboardPage() {
  const user = useAuthStore((state) => state.user);
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const { data: companies, isLoading: companiesLoading, isError: companiesError } = useCompanies();
  const companyId = activeCompany?.id || "";
  const isSynthetic = Boolean(user?.isSynthetic || activeCompany?.is_synthetic);
  // Worker and incident records are disabled for real accounts (legal gate).
  const workerDataId = isSynthetic ? companyId : "";
  const { data: employeeResponse, isLoading: employeesLoading, isError: employeesError } = useEmployees(workerDataId);
  const { data: incidents, isLoading: incidentsLoading, isError: incidentsError } = useCompanyIncidents(workerDataId);
  const { data: alerts, isLoading: alertsLoading } = useAlerts(companyId);
  const employees = employeeResponse?.data || [];
  const syntheticIncidents = incidents || [];
  const today = new Date().toLocaleDateString("es-CO", { weekday: "long", day: "numeric", month: "long" });

  return (
    <DashboardShell>
      <div className="mb-6 flex items-start justify-between gap-4 pb-5 border-b border-zu-border/70">
        <div className="min-w-0">
          <p className="text-[12px] font-semibold text-zu-text3 capitalize">{today}</p>
          <h1 className="mt-0.5 text-[24px] font-extrabold tracking-tight text-zu-text1">
            {greeting()}, {user?.firstName || "profesional"}
          </h1>
          <p className="mt-1.5 text-[13px] text-zu-text2">
            Esto es lo que necesita tu atención en <span className="font-semibold text-zu-text1">{activeCompany?.name || "tu cartera"}</span> hoy.
          </p>
        </div>
        <div className="shrink-0 flex flex-col items-end gap-2">
          <Link href="/dashboard/portfolio" className="inline-flex items-center gap-1.5 rounded-lg border border-zu-border bg-white px-3.5 py-2 text-[12px] font-bold text-zu-text1 shadow-zu transition-colors hover:border-zu-accent hover:text-zu-accent">
            <Briefcase size={13} /> Ver portafolio completo <ArrowRight size={13} />
          </Link>
          {isSynthetic && (
            <div className="flex items-center gap-1.5 rounded-full border border-amber-200 bg-amber-50 px-3 py-1.5 text-[11px] font-bold text-amber-800">
              <FlaskConical size={12} /> Datos sintéticos de demostración
            </div>
          )}
        </div>
      </div>

      <div className="mb-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <KPICard label="Empresas en cartera" value={companiesLoading ? "…" : companies?.length ?? 0} icon={Building2} sub="autorizadas por membership"/>
        <KPICard label="Trabajadores registrados" value={!isSynthetic ? "—" : employeesLoading ? "…" : employees.length} icon={Users} sub={isSynthetic ? activeCompany?.name || "selecciona una empresa" : "se habilita tras la revisión legal"}/>
        <KPICard label="Incidentes registrados" value={!isSynthetic ? "—" : incidentsLoading ? "…" : syntheticIncidents.length} icon={Activity} sub={isSynthetic ? "empresa activa" : "se habilita tras la revisión legal"}/>
        <KPICard label="Estado del entorno" value={isSynthetic ? "Piloto" : "Activo"} icon={ShieldCheck} valueColor={isSynthetic ? "var(--amber)" : "var(--accent)"} sub={isSynthetic ? "aislado de datos reales" : "cuenta en producción"}/>
      </div>
      {(companiesError || employeesError || incidentsError) && <Card className="mb-4 border-red-200 bg-red-50 p-3 text-xs text-red-800">Una o más consultas a Core no están disponibles. Los conteos marcados “—” no se sustituyen por datos demo.</Card>}

      <Card className="mb-5 p-4">
        <div className="mb-3 flex items-center justify-between"><h2 className="flex items-center gap-1.5 text-sm font-bold"><BellRing size={14} className="text-amber-600" /> Alertas y vencimientos</h2></div>
        {alertsLoading ? <Spinner /> : !alerts?.length ? (
          <p className="py-3 text-center text-xs text-zu-text3">Sin alertas pendientes en los próximos 30 días para la empresa activa.</p>
        ) : (
          <div className="space-y-1.5">
            {alerts.map((a, i) => (
              <div key={i} className={`flex items-center justify-between rounded-md px-3 py-2 text-[11px] ${a.severity === "critical" ? "bg-red-50" : "bg-amber-50"}`}>
                <div>
                  <span className={`font-semibold ${a.severity === "critical" ? "text-red-900" : "text-amber-900"}`}>{a.title}</span>
                  <span className="ml-2 text-zu-text3">{a.detail}</span>
                </div>
                <Badge variant={a.severity === "critical" ? "red" : "amber"}>{alertCategoryLabel[a.category] || a.category} · {formatDateTime(a.due_at)}</Badge>
              </div>
            ))}
          </div>
        )}
      </Card>

      {!isSynthetic ? (
        <Card className="p-4">
          <h2 className="mb-3 text-sm font-bold">Módulos de {activeCompany?.name || "la empresa"}</h2>
          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 xl:grid-cols-3">{REAL_MODULES.map(({ href, title, description, icon: Icon }) => <Link key={href} href={href} className="flex items-start gap-2 rounded-md border border-zu-border p-3 transition hover:border-zu-accent"><Icon size={14} className="mt-0.5 text-zu-accent"/><span><span className="block text-xs font-semibold">{title}</span><span className="block text-[11px] text-zu-text3">{description}</span></span></Link>)}</div>
          <p className="mt-3 text-[11px] text-zu-text3">Trabajadores, incidentes, capacitaciones y exámenes médicos siguen deshabilitados para datos reales mientras se completa la revisión legal de tratamiento de datos personales.</p>
        </Card>
      ) : <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card className="p-4">
          <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Trabajadores recientes</h2><Link href="/dashboard/employees" className="text-xs font-semibold text-zu-accent">Gestionar <ArrowRight size={12} className="inline"/></Link></div>
          {employeesLoading ? <Spinner/> : employees.length ? <div className="divide-y divide-zu-border">{employees.slice(0, 5).map((employee) => <div key={employee.id} className="flex items-center justify-between py-2"><div><p className="text-xs font-semibold">{employee.first_name} {employee.last_name}</p><p className="font-mono text-[10px] text-zu-text3">{employee.document_number}</p></div><Badge variant="amber">SINTÉTICO</Badge></div>)}</div> : <p className="py-5 text-center text-xs text-zu-text3">No hay trabajadores para la empresa activa. <Link href="/dashboard/employees" className="font-semibold text-zu-accent">Crear uno</Link></p>}
        </Card>
        <Card className="p-4">
          <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Incidentes recientes</h2><Link href="/dashboard/incidents" className="text-xs font-semibold text-zu-accent">Ver módulo <ArrowRight size={12} className="inline"/></Link></div>
          {incidentsLoading ? <Spinner/> : syntheticIncidents.length ? <div className="divide-y divide-zu-border">{syntheticIncidents.slice(0, 5).map((incident) => <div key={incident.id} className="py-2"><div className="flex items-center justify-between gap-3"><p className="truncate text-xs font-semibold">{incident.description}</p><Badge variant="amber">SINTÉTICO</Badge></div><p className="mt-1 text-[10px] text-zu-text3">{incident.location} · {incident.status} · {new Date(incident.occurred_at).toLocaleDateString("es-CO")}</p></div>)}</div> : <p className="py-5 text-center text-xs text-zu-text3">No hay incidentes registrados.</p>}
        </Card>
      </div>}
      <div className="mt-4 flex items-center gap-2 text-[11px] text-zu-text3"><Building2 size={13}/> {isSynthetic ? "El selector superior cambia de empresa mediante un nuevo token firmado y membership comprobada por Core." : "Cambia de empresa con el selector de arriba; el servidor confirma tu acceso a cada una."}</div>
    </DashboardShell>
  );
}
