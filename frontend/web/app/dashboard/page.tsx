"use client";
import Link from "next/link";
import { Building2, Users, Activity, AlertTriangle, ArrowRight, CircleCheck } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, Badge, KPICard, PageHeader, Spinner } from "@/components/ui";
import { useCompanies, useEmployees, useCompanyIncidents } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

export default function DashboardPage() {
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const { data: companies, isLoading: companiesLoading, isError: companiesError } = useCompanies();
  const companyId = activeCompany?.id || "";
  const { data: employeeResponse, isLoading: employeesLoading, isError: employeesError } = useEmployees(companyId);
  const { data: incidents, isLoading: incidentsLoading, isError: incidentsError } = useCompanyIncidents(companyId);
  const employees = employeeResponse?.data || [];
  const syntheticIncidents = incidents || [];

  return (
    <DashboardShell>
      <PageHeader title="Panel del piloto" description="Estado de datos sintéticos en la empresa activa; sin métricas inventadas ni conexiones a módulos de alto riesgo." />
      <Card className="mb-5 border-amber-300 bg-amber-50 p-4">
        <div className="flex items-start gap-3"><AlertTriangle size={18} className="mt-0.5 text-amber-800"/><div><h2 className="text-sm font-bold text-amber-950">Resumen de prueba · {activeCompany?.name || "sin empresa activa"}</h2><p className="mt-1 text-xs leading-relaxed text-amber-900">Esta pantalla muestra únicamente registros marcados como sintéticos. No se ejecutan OCR, biometría, IA externa, drones, alertas regulatorias ni cálculos de cumplimiento.</p></div></div>
      </Card>
      <div className="mb-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
        <KPICard label="Empresas con membership" value={companiesLoading ? "…" : companies?.length ?? 0} icon={Building2} sub="lista autorizada por Core"/>
        <KPICard label="Perfiles sintéticos" value={employeesLoading ? "…" : employees.length} icon={Users} sub={activeCompany?.name || "selecciona una empresa"}/>
        <KPICard label="Incidentes de demostración" value={incidentsLoading ? "…" : syntheticIncidents.length} icon={Activity} sub="empresa activa · sin reporte externo"/>
        <KPICard label="Estado del entorno" value="Sintético" icon={CircleCheck} valueColor="var(--amber)" sub="no habilitado para datos reales"/>
      </div>
      {(companiesError || employeesError || incidentsError) && <Card className="mb-4 border-red-200 bg-red-50 p-3 text-xs text-red-800">Una o más consultas a Core no están disponibles. Los conteos marcados “—” no se sustituyen por datos demo.</Card>}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card className="p-4">
          <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Personas de prueba</h2><Link href="/dashboard/employees" className="text-xs font-semibold text-zu-accent">Gestionar <ArrowRight size={12} className="inline"/></Link></div>
          {employeesLoading ? <Spinner/> : employees.length ? <div className="divide-y divide-zu-border">{employees.slice(0, 5).map((employee) => <div key={employee.id} className="flex items-center justify-between py-2"><div><p className="text-xs font-semibold">{employee.first_name} {employee.last_name}</p><p className="font-mono text-[10px] text-zu-text3">{employee.document_number}</p></div><Badge variant="amber">SINTÉTICO</Badge></div>)}</div> : <p className="py-5 text-center text-xs text-zu-text3">No hay perfiles para la empresa activa. <Link href="/dashboard/employees" className="font-semibold text-zu-accent">Crear uno</Link></p>}
        </Card>
        <Card className="p-4">
          <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Incidentes de demostración</h2><Link href="/dashboard/incidents" className="text-xs font-semibold text-zu-accent">Ver módulo <ArrowRight size={12} className="inline"/></Link></div>
          {incidentsLoading ? <Spinner/> : syntheticIncidents.length ? <div className="divide-y divide-zu-border">{syntheticIncidents.slice(0, 5).map((incident) => <div key={incident.id} className="py-2"><div className="flex items-center justify-between gap-3"><p className="truncate text-xs font-semibold">{incident.description}</p><Badge variant="amber">SINTÉTICO</Badge></div><p className="mt-1 text-[10px] text-zu-text3">{incident.location} · {incident.status} · {new Date(incident.occurred_at).toLocaleDateString("es-CO")}</p></div>)}</div> : <p className="py-5 text-center text-xs text-zu-text3">No hay incidentes de prueba registrados.</p>}
        </Card>
      </div>
      <div className="mt-4 flex items-center gap-2 text-[11px] text-zu-text3"><Building2 size={13}/> El selector superior cambia de empresa mediante un nuevo token firmado y membership comprobada por Core.</div>
    </DashboardShell>
  );
}
