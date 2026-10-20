"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Building2, CheckCircle2, Scale, ShieldAlert, Users } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { authAPI } from "@/lib/api";
import { usePortfolio } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

export default function PortfolioPage() {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: portfolio, isLoading, isError } = usePortfolio();
  const { activeCompany, companies } = useAuthStore();
  const [switchingId, setSwitchingId] = useState("");

  const totalEmployees = portfolio?.reduce((sum, c) => sum + c.active_employees, 0) ?? 0;
  const totalOpenIncidents = portfolio?.reduce((sum, c) => sum + c.open_incidents, 0) ?? 0;
  const totalCriticalRisks = portfolio?.reduce((sum, c) => sum + c.critical_risks, 0) ?? 0;
  const totalPendingLegal = portfolio?.reduce((sum, c) => sum + c.pending_legal_requirements, 0) ?? 0;

  const openCompany = async (companyId: string) => {
    if (companyId === activeCompany?.id || switchingId) return;
    const target = companies.find((c) => c.id === companyId);
    if (!target) return;
    setSwitchingId(companyId);
    try {
      const response = await authAPI.switchCompany(companyId);
      const current = useAuthStore.getState();
      current.setTokens(response.data.access_token, response.data.refresh_token);
      current.setActiveCompany(target);
      queryClient.clear();
      router.push("/dashboard");
    } finally {
      setSwitchingId("");
    }
  };

  return (
    <DashboardShell>
      <PageHeader title="Panel de portafolio" description="Vista consolidada de todas las empresas con las que tienes membership, calculada en un único endpoint a nivel de tenant — no son estimaciones del cliente." />

      {isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar el panel de portafolio.</p>
      ) : !portfolio?.length ? (
        <Empty icon={Building2} title="Sin empresas" description="Todavía no tienes membership en ninguna empresa." />
      ) : (
        <>
          <div className="mb-5 grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            <Card className="p-4"><div className="flex items-center gap-2 text-zu-text3"><Building2 size={14} /><span className="text-xs font-semibold">Empresas</span></div><p className="mt-1 text-2xl font-bold">{portfolio.length}</p></Card>
            <Card className="p-4"><div className="flex items-center gap-2 text-zu-text3"><Users size={14} /><span className="text-xs font-semibold">Trabajadores activos</span></div><p className="mt-1 text-2xl font-bold">{totalEmployees}</p></Card>
            <Card className="p-4"><div className="flex items-center gap-2 text-zu-text3"><AlertTriangle size={14} /><span className="text-xs font-semibold">Incidentes abiertos</span></div><p className={`mt-1 text-2xl font-bold ${totalOpenIncidents ? "text-amber-700" : ""}`}>{totalOpenIncidents}</p></Card>
            <Card className="p-4"><div className="flex items-center gap-2 text-zu-text3"><ShieldAlert size={14} /><span className="text-xs font-semibold">Riesgos críticos</span></div><p className={`mt-1 text-2xl font-bold ${totalCriticalRisks ? "text-red-700" : ""}`}>{totalCriticalRisks}</p></Card>
          </div>

          <div className="grid grid-cols-1 gap-3 lg:grid-cols-2">
            {portfolio.map((c) => (
              <Card key={c.company_id} className="p-4">
                <div className="mb-3 flex items-start justify-between gap-3">
                  <div>
                    <h2 className="text-sm font-bold">{c.name}</h2>
                    {c.is_synthetic && <Badge variant="amber">SINTÉTICO</Badge>}
                  </div>
                  {c.company_id === activeCompany?.id ? (
                    <Badge variant="blue"><CheckCircle2 size={11} className="mr-1 inline" />Activa</Badge>
                  ) : (
                    <button onClick={() => void openCompany(c.company_id)} disabled={!!switchingId} className="rounded-md border border-zu-border px-2.5 py-1 text-[11px] font-semibold text-zu-accent hover:bg-zu-surface2 disabled:opacity-50">
                      {switchingId === c.company_id ? "Cambiando…" : "Abrir"}
                    </button>
                  )}
                </div>
                <div className="grid grid-cols-2 gap-2 text-[11px] sm:grid-cols-4">
                  <div className="rounded-md bg-zu-surface2 px-2 py-1.5"><p className="text-zu-text3">Activos</p><p className="text-sm font-bold">{c.active_employees}</p></div>
                  <div className="rounded-md bg-zu-surface2 px-2 py-1.5"><p className="text-zu-text3">Incidentes</p><p className={`text-sm font-bold ${c.open_incidents ? "text-amber-700" : ""}`}>{c.open_incidents}</p></div>
                  <div className="rounded-md bg-zu-surface2 px-2 py-1.5"><p className="text-zu-text3">Riesgo crítico</p><p className={`text-sm font-bold ${c.critical_risks ? "text-red-700" : ""}`}>{c.critical_risks}</p></div>
                  <div className="rounded-md bg-zu-surface2 px-2 py-1.5"><p className="flex items-center gap-1 text-zu-text3"><Scale size={10} /> Legal pend.</p><p className={`text-sm font-bold ${c.pending_legal_requirements ? "text-amber-700" : ""}`}>{c.pending_legal_requirements}</p></div>
                </div>
              </Card>
            ))}
          </div>
          {totalPendingLegal > 0 && <p className="mt-4 text-[11px] text-zu-text3">{totalPendingLegal} requisito(s) legal(es) sin evaluar en todo el portafolio. Revísalos en la matriz legal de cada empresa.</p>}
        </>
      )}
    </DashboardShell>
  );
}
