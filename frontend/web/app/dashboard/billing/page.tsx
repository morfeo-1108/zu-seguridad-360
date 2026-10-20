"use client";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, Building2, CheckCircle2, Sparkles, Users } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, PageHeader, Spinner } from "@/components/ui";
import { billingAPI } from "@/lib/api";
import { usePlans, useSubscription } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

const planLabel: Record<string, string> = { esencial: "Esencial", profesional: "Profesional", consultora: "Consultora", agencia: "Agencia SST" };
const planOrder = ["esencial", "profesional", "consultora", "agencia"];
const cop = (centavos: number) => `$${(centavos / 100).toLocaleString("es-CO")}`;

export default function BillingPage() {
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const tenantId = user?.tenantId || "";
  const { data: plans, isLoading: plansLoading } = usePlans();
  const { data: subscription, isLoading: subLoading } = useSubscription(tenantId);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  const subscribe = async (planCode: string) => {
    if (busy) return;
    setBusy(planCode); setError("");
    try {
      await billingAPI.subscribe(tenantId, planCode, true);
      await queryClient.invalidateQueries({ queryKey: ["subscription", tenantId] });
    } catch { setError("No se pudo registrar el plan."); }
    finally { setBusy(""); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Facturación" description="Planes por profesional, con cupo de empresas, usuarios y trabajadores administrados. El cobro automático con Wompi sigue en pruebas de sandbox — elegir un plan aquí lo registra, no genera un cargo." />

      <Card className="mb-5 border-amber-300 bg-amber-50 p-3.5">
        <div className="flex items-start gap-2.5">
          <AlertTriangle size={16} className="mt-0.5 shrink-0 text-amber-800" />
          <p className="text-xs leading-relaxed text-amber-900">
            El cupo de <b>empresas</b> ya lo valida el servidor antes de crear una empresa nueva. El cupo de <b>usuarios profesionales</b> y <b>trabajadores</b> se muestra como referencia — todavía no existe en Core un flujo para invitar a otro profesional a tu cuenta ni un límite de trabajadores por cartera, así que esos dos no se hacen cumplir aún.
          </p>
        </div>
      </Card>

      {subLoading ? <Spinner size={20} /> : subscription && (
        <Card className="mb-5 p-4">
          <div className="flex items-center justify-between">
            <div>
              <h2 className="flex items-center gap-2 text-sm font-bold">Plan actual: {planLabel[subscription.plan_code] || subscription.plan_code}
                {subscription.is_founder && <Badge variant="green"><Sparkles size={10} className="mr-1 inline" />Precio fundador</Badge>}
              </h2>
              <p className="mt-1 text-xs text-zu-text3">{cop(subscription.price_monthly)}/mes · hasta {subscription.max_companies} empresas, {subscription.max_professional_users} usuario(s), {subscription.max_workers} trabajadores</p>
              {subscription.is_founder && subscription.founder_price_until && <p className="mt-1 text-[10px] text-zu-text3">Precio fundador vigente hasta {new Date(subscription.founder_price_until).toLocaleDateString("es-CO")}</p>}
            </div>
            <Badge variant={subscription.status === "active" ? "green" : "red"}>{subscription.status}</Badge>
          </div>
        </Card>
      )}

      {plansLoading ? <Spinner size={24} /> : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {planOrder.map((code) => {
            const plan = plans?.[code];
            if (!plan) return null;
            const isCurrent = subscription?.plan_code === code;
            return (
              <Card key={code} className={`p-4 ${isCurrent ? "border-zu-accent ring-1 ring-zu-accent" : ""}`}>
                <div className="mb-2 flex items-center justify-between"><h3 className="text-sm font-bold">{planLabel[code]}</h3>{isCurrent && <Badge variant="blue"><CheckCircle2 size={10} className="mr-1 inline" />Activo</Badge>}</div>
                <p className="text-2xl font-bold">{cop(plan.PriceMonthly)}<span className="text-xs font-normal text-zu-text3">/mes</span></p>
                <p className="mb-3 text-[11px] text-emerald-700">Fundador: {cop(plan.FounderPriceMonthly)}/mes por 12 meses</p>
                <dl className="mb-4 space-y-1 text-[11px] text-zu-text2">
                  <div className="flex items-center gap-1.5"><Building2 size={11} /> Hasta {plan.MaxCompanies} empresas</div>
                  <div className="flex items-center gap-1.5"><Users size={11} /> {plan.MaxProfessionalUsers} usuario(s) profesional(es)</div>
                  <div className="flex items-center gap-1.5">Hasta {plan.MaxWorkers.toLocaleString("es-CO")} trabajadores</div>
                </dl>
                <Button className="w-full" variant={isCurrent ? "secondary" : "primary"} disabled={isCurrent || busy === code} onClick={() => void subscribe(code)}>
                  {busy === code ? "Registrando…" : isCurrent ? "Plan actual" : "Elegir este plan"}
                </Button>
              </Card>
            );
          })}
        </div>
      )}
      {error && <p role="alert" className="mt-3 text-xs text-red-700">{error}</p>}
    </DashboardShell>
  );
}
