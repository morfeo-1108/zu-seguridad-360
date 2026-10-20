"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, Building2, Check, Sparkles, Users } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { billingAPI, type PlanInfo } from "@/lib/api";

const PLAN_ORDER = ["esencial", "profesional", "consultora", "agencia"];
const PLAN_LABEL: Record<string, string> = { esencial: "Esencial", profesional: "Profesional", consultora: "Consultora", agencia: "Agencia SST" };
const PLAN_BLURB: Record<string, string> = {
  esencial: "Para el profesional independiente que está comenzando o maneja pocos clientes.",
  profesional: "El plan central — el que más se ajusta a un profesional con una cartera activa.",
  consultora: "Para pequeños equipos o profesionales con una cartera ya consolidada.",
  agencia: "Para consultoras estructuradas, redes de profesionales o equipos con muchos clientes.",
};
const PLAN_FEATURES: Record<string, string[]> = {
  esencial: ["Dashboard de cartera", "Gestión documental", "Matriz de riesgos GTC-45", "Indicadores básicos", "Reportes por empresa", "Soporte estándar"],
  profesional: ["Todo lo de Esencial", "Cambio rápido entre empresas", "Plantillas reutilizables", "Alertas y vencimientos automáticos", "Indicadores avanzados IF/IS", "Reportes consolidados de cartera", "Soporte prioritario"],
  consultora: ["Todo lo de Profesional", "Panel general de toda la cartera", "Permisos por empresa y por usuario", "Reportes comparativos", "Plantillas propias de la consultora", "Soporte prioritario"],
  agencia: ["Todo lo de Consultora", "Roles avanzados", "Reportes ejecutivos", "Integraciones ampliadas", "Soporte especializado", "Onboarding acompañado"],
};

const cop = (centavos: number) => `$${(centavos / 100).toLocaleString("es-CO")}`;

export default function PlanesPage() {
  const [plans, setPlans] = useState<Record<string, PlanInfo> | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    billingAPI.listPlans().then((r) => setPlans(r.data.plans)).catch(() => setError(true));
  }, []);

  return (
    <div className="min-h-screen bg-[#0B0F0C] text-white">
      <header className="border-b border-white/[0.06]">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-5 lg:px-10">
          <Link href="/"><ZULogo size={30} dark showText /></Link>
          <Link href="/auth/login" className="rounded-lg border border-white/15 bg-white/5 px-5 py-2.5 text-[13px] font-bold text-white backdrop-blur-sm transition hover:border-zu-lime hover:bg-zu-lime hover:text-[#0B0F0C]">
            Iniciar sesión
          </Link>
        </div>
      </header>

      <section className="relative overflow-hidden py-20">
        <div className="pointer-events-none absolute left-1/2 top-0 h-[400px] w-[600px] -translate-x-1/2 rounded-full bg-zu-lime/[0.07] blur-[130px]" />
        <div className="relative mx-auto max-w-4xl px-6 text-center lg:px-10">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-4 py-1.5 text-[12px] font-semibold tracking-wide text-zu-lime">
            <span className="h-1.5 w-1.5 rounded-full bg-zu-lime" />
            PRECIO PÚBLICO, SIN COTIZACIÓN
          </div>
          <h1 className="text-[38px] font-extrabold leading-[1.1] tracking-tight sm:text-[48px]">
            Un plan por profesional, no por empresa.
          </h1>
          <p className="mx-auto mt-5 max-w-2xl text-[16px] leading-relaxed text-white/60">
            Administra toda tu cartera de empresas desde una sola cuenta. El plan define cuántas empresas,
            usuarios profesionales y trabajadores puedes gestionar — no cobramos por cada cliente nuevo como si fuera una cuenta aparte.
          </p>
        </div>
      </section>

      <section className="relative mx-auto max-w-7xl px-6 pb-16 lg:px-10">
        {error ? (
          <p className="text-center text-sm text-white/50">No se pudieron cargar los planes en este momento. Intenta de nuevo en un momento.</p>
        ) : !plans ? (
          <p className="text-center text-sm text-white/40">Cargando planes…</p>
        ) : (
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {PLAN_ORDER.map((code) => {
              const plan = plans[code];
              if (!plan) return null;
              const isAnchor = code === "profesional";
              return (
                <div key={code} className={`relative flex flex-col rounded-2xl border p-6 ${isAnchor ? "border-zu-lime/50 bg-zu-lime/[0.04]" : "border-white/[0.08] bg-white/[0.02]"}`}>
                  {isAnchor && <div className="absolute -top-3 left-6 rounded-full bg-zu-lime px-3 py-1 text-[10px] font-bold tracking-wide text-[#0B0F0C]">MÁS ELEGIDO</div>}
                  <h3 className="text-[17px] font-extrabold">{PLAN_LABEL[code]}</h3>
                  <p className="mt-1 text-[12px] leading-relaxed text-white/45">{PLAN_BLURB[code]}</p>

                  <div className="mt-5 border-t border-white/[0.08] pt-5">
                    <div className="text-[30px] font-extrabold tracking-tight">{cop(plan.PriceMonthly)}<span className="text-[13px] font-normal text-white/40">/mes</span></div>
                    <div className="mt-1.5 flex items-center gap-1.5 text-[11.5px] font-semibold text-emerald-400">
                      <Sparkles size={12} /> Fundador: {cop(plan.FounderPriceMonthly)}/mes por 12 meses
                    </div>
                  </div>

                  <dl className="mt-5 space-y-1.5 border-t border-white/[0.08] pt-5 text-[12.5px]">
                    <div className="flex items-center gap-2 text-white/70"><Building2 size={13} className="text-white/35" /> Hasta {plan.MaxCompanies} empresas</div>
                    <div className="flex items-center gap-2 text-white/70"><Users size={13} className="text-white/35" /> {plan.MaxProfessionalUsers} usuario{plan.MaxProfessionalUsers > 1 ? "s" : ""} profesional{plan.MaxProfessionalUsers > 1 ? "es" : ""}</div>
                    <div className="flex items-center gap-2 text-white/70">Hasta {plan.MaxWorkers.toLocaleString("es-CO")} trabajadores</div>
                  </dl>

                  <ul className="mt-5 flex-1 space-y-2 border-t border-white/[0.08] pt-5">
                    {PLAN_FEATURES[code].map((f) => (
                      <li key={f} className="flex items-start gap-2 text-[12.5px] leading-snug text-white/60">
                        <Check size={14} className="mt-0.5 shrink-0 text-zu-lime" /> {f}
                      </li>
                    ))}
                  </ul>

                  <Link href="/auth/login" className={`mt-6 inline-flex items-center justify-center gap-1.5 rounded-lg px-4 py-2.5 text-[13px] font-bold transition ${isAnchor ? "bg-zu-lime text-[#0B0F0C] hover:bg-white" : "border border-white/15 text-white hover:border-zu-lime hover:bg-zu-lime hover:text-[#0B0F0C]"}`}>
                    Elegir {PLAN_LABEL[code]} <ArrowRight size={14} />
                  </Link>
                </div>
              );
            })}
          </div>
        )}

        <p className="mx-auto mt-10 max-w-2xl text-center text-[12px] text-white/35">
          Precios antes de IVA si aplica tributariamente. El precio fundador está limitado a los primeros 20 profesionales que se suscriban, fijo por 12 meses desde el registro. ¿Necesitas más de 60 empresas o un acuerdo a la medida? <a href="mailto:hola@zu360.co" className="text-zu-lime hover:underline">Escríbenos</a>.
        </p>
      </section>

      <footer className="border-t border-white/[0.06] px-6 py-8 text-center text-[12px] text-white/35 lg:px-10">
        ZU Seguridad Integral 360 · Colombia
      </footer>
    </div>
  );
}
