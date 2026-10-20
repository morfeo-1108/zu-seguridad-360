"use client";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, CheckCircle2, RefreshCw } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { managementAPI } from "@/lib/api";
import { useActiveManagementSystem } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const phaseLabel: Record<string, string> = { planear: "Planear", hacer: "Hacer", verificar: "Verificar", actuar: "Actuar" };
const phaseOrder = ["planear", "hacer", "verificar", "actuar"];
const statusColor: Record<string, "green" | "amber" | "red" | "gray"> = { cumple: "green", no_cumple: "red", pendiente: "gray", parcial: "amber" };

export default function ManagementPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: system, isLoading, isError } = useActiveManagementSystem(companyId);
  const [busy, setBusy] = useState(false);
  const [score, setScore] = useState<{ compliance_score: number | null; validation_coverage: number; total_standards: number; human_validated: number; note: string } | null>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["management-active", companyId] });

  const initSystem = async () => {
    if (!companyId || busy) return;
    setBusy(true);
    try { await managementAPI.initSystem(companyId, { year: new Date().getFullYear() }); await invalidate(); }
    catch { /* el botón permanece visible para reintentar */ }
    finally { setBusy(false); }
  };

  const advancePhase = async () => {
    if (!system || busy) return;
    setBusy(true);
    try { await managementAPI.advancePhase(companyId, system.id); await invalidate(); }
    catch { /* el botón permanece visible para reintentar */ }
    finally { setBusy(false); }
  };

  const assess = async () => {
    if (!system || busy) return;
    setBusy(true);
    try { await managementAPI.assess(companyId, system.id); await invalidate(); }
    catch { /* el botón permanece visible para reintentar */ }
    finally { setBusy(false); }
  };

  const loadScore = async () => {
    if (!system) return;
    try { setScore((await managementAPI.score(companyId, system.id)).data); } catch { /* no bloquea el resto de la vista */ }
  };

  const validateStandard = async (standardId: string, status: "cumple" | "no_cumple") => {
    if (busy) return;
    setBusy(true);
    try { await managementAPI.validateStandard(companyId, standardId, status); await invalidate(); }
    catch { /* el botón permanece visible para reintentar */ }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Sistema de gestión PHVA" description="El sistema propone el cumplimiento a partir de los módulos; un profesional SST valida cada estándar — nada legalmente relevante se cierra solo." />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : isLoading ? <Spinner size={24} /> : isError || !system ? (
        <Empty icon={RefreshCw} title="Sin ciclo PHVA iniciado" description="Inicializa el ciclo del año en curso sobre Resolución 0312 e ISO 45001." action={<Button onClick={() => void initSystem()} disabled={busy}>{busy ? "Creando…" : "Iniciar ciclo " + new Date().getFullYear()}</Button>} />
      ) : (
        <>
          <Card className="mb-5 p-4">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div><h2 className="text-sm font-bold">Ciclo {system.year}</h2><p className="text-[11px] text-zu-text3">{system.standards.length} estándares cargados</p></div>
              <div className="flex items-center gap-2">
                <Button size="sm" variant="secondary" onClick={() => void assess()} disabled={busy}>Proponer cumplimiento</Button>
                <Button size="sm" variant="secondary" onClick={() => void loadScore()}>Ver score</Button>
                {phaseOrder.indexOf(system.current_phase) < 3 && <Button size="sm" onClick={() => void advancePhase()} disabled={busy}>Avanzar fase <RefreshCw size={12} /></Button>}
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              {phaseOrder.map((p, i) => (
                <div key={p} className={`flex-1 rounded-md px-2 py-2 text-center text-[11px] font-semibold ${p === system.current_phase ? "bg-zu-accent text-white" : i < phaseOrder.indexOf(system.current_phase) ? "bg-emerald-100 text-emerald-800" : "bg-zu-surface2 text-zu-text3"}`}>{phaseLabel[p]}</div>
              ))}
            </div>
            {score && (
              <div className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-4">
                <div className="rounded-md bg-zu-surface2 px-3 py-2 text-center"><p className="text-[10px] text-zu-text3">Score</p><p className="text-lg font-bold">{score.compliance_score != null ? score.compliance_score.toFixed(0) + "%" : "—"}</p></div>
                <div className="rounded-md bg-zu-surface2 px-3 py-2 text-center"><p className="text-[10px] text-zu-text3">Cobertura validada</p><p className="text-lg font-bold">{score.validation_coverage.toFixed(0)}%</p></div>
                <div className="rounded-md bg-zu-surface2 px-3 py-2 text-center"><p className="text-[10px] text-zu-text3">Estándares</p><p className="text-lg font-bold">{score.total_standards}</p></div>
                <div className="rounded-md bg-zu-surface2 px-3 py-2 text-center"><p className="text-[10px] text-zu-text3">Validados</p><p className="text-lg font-bold">{score.human_validated}</p></div>
              </div>
            )}
            {score && <p className="mt-2 text-[10px] text-zu-text3">{score.note}</p>}
          </Card>

          <Card className="overflow-hidden p-0">
            <div className="divide-y divide-zu-border">
              {system.standards.map((std) => (
                <div key={std.id} className="flex items-center justify-between gap-3 p-3">
                  <div className="min-w-0">
                    <p className="truncate text-xs font-semibold">{std.clause} — {std.title}</p>
                    <p className="text-[10px] text-zu-text3">{std.framework} · fase {phaseLabel[std.phase] || std.phase} · módulo {std.linked_module}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant={statusColor[std.status] || "gray"}>{std.status}</Badge>
                    {std.status === "pendiente" || std.status === "parcial" ? (
                      <>
                        <Button size="sm" variant="secondary" onClick={() => void validateStandard(std.id, "cumple")} disabled={busy}><CheckCircle2 size={12} /> Validar</Button>
                        <Button size="sm" variant="secondary" onClick={() => void validateStandard(std.id, "no_cumple")} disabled={busy}>No cumple</Button>
                      </>
                    ) : null}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </>
      )}
    </DashboardShell>
  );
}
