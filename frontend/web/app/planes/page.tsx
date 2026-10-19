// NOTA LEGAL INTERNA (no visible al usuario): confirmar con abogado si ZU como persona
// jurídica requiere licencia SST propia (Res. 908/2025, licencia por sede) o si basta
// la licencia personal de la profesional al frente. El modelo funciona en ambos casos.

import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Planes — ZU Seguridad Integral 360",
  description: "Gestión de SG-SST con profesional licenciado, plataforma inteligente y visitas presenciales. Planes para pymes y empresas de alto riesgo.",
};

// ── Datos de planes, derivados de la normativa real (Res. 0312/2019 y 908/2025) ──
// El perfil del profesional y la frecuencia de visita responden a tamaño × riesgo.

const PYME_PLANS = [
  {
    name: "Esencial",
    target: "1–10 trabajadores · Riesgo I–II–III",
    tagline: "Mantenimiento de SG-SST ya implementado",
    price: "$390.000",
    priceNote: "/ mes + IVA",
    professional: "Técnico/Tecnólogo SST licenciado",
    visits: "1 visita presencial / mes",
    hours: "Hasta 8 horas profesional / mes",
    highlight: false,
    features: [
      "Mantenimiento y gestión de tu SG-SST existente",
      "Plataforma completa con alertas de vencimientos",
      "Actualización documental y control de versiones",
      "Autoevaluación estándares mínimos (7 estándares)",
      "Firma y aval de profesional licenciado",
      "Capacitaciones e-learning",
      "Soporte por chat SISO-Bot IA",
    ],
  },
  {
    name: "Profesional",
    target: "11–50 trabajadores · Riesgo I–II–III",
    price: "$690.000",
    priceNote: "/ mes + IVA",
    professional: "Profesional SST licenciado",
    visits: "2 visitas presenciales / mes",
    hours: "Hasta 16 horas profesional / mes",
    highlight: true,
    features: [
      "Todo lo del plan Esencial",
      "Autoevaluación de 21 estándares mínimos",
      "Investigación de incidentes con IA asistida",
      "Matriz de riesgos GTC 45 con mapa de calor",
      "Indicadores IF, IS, ILI en tiempo real",
      "Ciclo PHVA completo con validación profesional",
      "Gestión de exámenes médicos con OCR",
      "Portal para la empresa cliente",
    ],
  },
];

const ENTERPRISE_PLANS = [
  {
    name: "Corporativo",
    target: "+50 trabajadores · Cualquier riesgo",
    price: "Desde $1.490.000",
    priceNote: "/ mes + IVA · según sedes",
    professional: "Profesional SST con posgrado + licencia",
    visits: "Visitas semanales o según cronograma",
    hours: "Bolsa de horas configurable",
    highlight: true,
    features: [
      "Los 60 estándares mínimos completos",
      "Profesional dedicado con posgrado en SST",
      "Múltiples sedes / obras bajo un contrato",
      "Cumplimiento ISO 45001 · 9001 · 14001 · 31000",
      "Vigilancia normativa con aprobación experta",
      "Analítica predictiva por Cognición IA",
      "Exportación y portabilidad de datos",
      "SLA de respuesta garantizado",
    ],
  },
  {
    name: "Alto Riesgo",
    target: "Riesgo IV–V · Construcción, minería, industria",
    price: "Desde $2.190.000",
    priceNote: "/ mes + IVA · según obra",
    professional: "Profesional SST con posgrado + licencia",
    visits: "Presencia reforzada + drones de inspección",
    hours: "Bolsa de horas ampliada",
    highlight: false,
    features: [
      "Todo lo del plan Corporativo",
      "Inspección con drones y visión por IA*",
      "Validación de EPP por trabajador (biometría)*",
      "Trabajo en alturas y espacios confinados",
      "Investigación de ATEL prioritaria",
      "Evidencia fotográfica georreferenciada",
      "Plan de emergencias y análisis Bow-Tie",
      "Acompañamiento en visitas del Ministerio/ARL",
    ],
  },
];

function Check() {
  return (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" style={{ flexShrink: 0, marginTop: 2 }}>
      <path d="M20 6L9 17l-5-5" stroke="#6BA31A" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function PlanCard({ plan }: { plan: any }) {
  return (
    <div className={`rounded-2xl border p-6 flex flex-col ${plan.highlight ? "border-zu-accent shadow-lg bg-white relative" : "border-zu-border bg-white"}`}>
      {plan.highlight && (
        <div className="absolute -top-3 left-6 bg-zu-accent text-white text-[10px] font-bold px-3 py-1 rounded-full tracking-wide">
          MÁS CONTRATADO
        </div>
      )}
      <h3 className="text-[19px] font-extrabold text-zu-graphite">{plan.name}</h3>
      <p className="text-[12px] text-zu-text2 mt-1 mb-1">{plan.target}</p>
      {plan.tagline && <p className="text-[10.5px] text-zu-accent font-semibold mb-3">{plan.tagline}</p>}

      <div className="mb-4 pb-4 border-b border-zu-border">
        <span className="text-[28px] font-black text-zu-graphite tracking-tight">{plan.price}</span>
        <span className="text-[12px] text-zu-text3 ml-1.5">{plan.priceNote}</span>
      </div>

      <div className="space-y-2 mb-4 pb-4 border-b border-zu-border">
        <div className="flex items-center gap-2 text-[12px]">
          <span className="text-zu-text3 w-20 flex-shrink-0">Profesional</span>
          <span className="font-semibold text-zu-text1">{plan.professional}</span>
        </div>
        <div className="flex items-center gap-2 text-[12px]">
          <span className="text-zu-text3 w-20 flex-shrink-0">Visitas</span>
          <span className="font-semibold text-zu-accent">{plan.visits}</span>
        </div>
        <div className="flex items-center gap-2 text-[12px]">
          <span className="text-zu-text3 w-20 flex-shrink-0">Horas</span>
          <span className="font-semibold text-zu-text1">{plan.hours}</span>
        </div>
      </div>

      <ul className="space-y-2 flex-1">
        {plan.features.map((f: string, i: number) => (
          <li key={i} className="flex gap-2 text-[12.5px] text-zu-text2 leading-snug">
            <Check /> {f}
          </li>
        ))}
      </ul>

      <button className={`mt-6 w-full py-2.5 rounded-lg font-semibold text-[13px] transition ${plan.highlight ? "text-white hover:opacity-90" : "text-zu-accent border border-zu-accent hover:bg-zu-surface2"}`}
        style={plan.highlight ? { background: "linear-gradient(135deg,#8BC34A,#6BA31A)" } : {}}>
        Solicitar cotización
      </button>
    </div>
  );
}

export default function PlanesPage() {
  return (
    <div className="min-h-screen bg-zu-bg">
      {/* Hero */}
      <header className="relative overflow-hidden" style={{ background: "linear-gradient(155deg,#12171F,#1A2129 60%,#0E1218)" }}>
        <div className="absolute top-0 right-[10%] w-[400px] h-[400px] rounded-full" style={{ background: "radial-gradient(circle,rgba(139,195,74,0.18),transparent 62%)" }} />
        <div className="max-w-6xl mx-auto px-6 py-16 relative z-10">
          <div className="flex items-center gap-2.5 mb-10">
            <div className="w-9 h-9 rounded-lg bg-[#0E1319] flex items-center justify-center font-black text-[15px]">
              <span style={{ color: "#8BC34A" }}>Z</span><span style={{ color: "#C4CAD2" }}>U</span>
            </div>
            <span className="text-white font-bold text-[14px]">ZU 360 <span style={{ color: "#8BC34A" }}>·</span> SEGURIDAD</span>
          </div>
          <div className="inline-block text-[11px] font-bold tracking-widest text-[#8BC34A] border border-[#8BC34A]/30 rounded-full px-4 py-1.5 mb-5">
            GESTIÓN INTEGRAL DE SG-SST
          </div>
          <h1 className="text-white text-[40px] font-black leading-tight tracking-tight max-w-3xl mb-4">
            Tu SG-SST cumplido, con <span style={{ color: "#8BC34A" }}>profesional licenciado</span> y tecnología que trabaja por él.
          </h1>
          <p className="text-[15px] text-[#9BA3AE] max-w-2xl leading-relaxed">
            Combinamos un profesional en SST con licencia vigente, visitas presenciales según tu nivel de riesgo, y una plataforma con inteligencia artificial que automatiza la gestión. Cumplimiento real, no solo papeles.
          </p>
        </div>
      </header>

      {/* Cómo funciona */}
      <section className="max-w-6xl mx-auto px-6 py-14">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
          {[
            { n: "01", t: "Diagnóstico", d: "Evaluamos tu empresa por número de trabajadores y nivel de riesgo para asignar el plan y el perfil profesional que la ley exige." },
            { n: "02", t: "Profesional + plataforma", d: "Un profesional SST licenciado firma y avala, mientras la plataforma automatiza documentos, alertas e indicadores." },
            { n: "03", t: "Visitas y cumplimiento", d: "Visitas presenciales según tu riesgo, evidencia digital y tu SG-SST siempre listo para una auditoría del Ministerio." },
          ].map((s) => (
            <div key={s.n} className="bg-white border border-zu-border rounded-xl p-5">
              <div className="text-[22px] font-black text-zu-accent font-mono mb-2">{s.n}</div>
              <h3 className="font-bold text-[15px] text-zu-graphite mb-1">{s.t}</h3>
              <p className="text-[12.5px] text-zu-text2 leading-relaxed">{s.d}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Segmento PYME */}
      <section className="max-w-6xl mx-auto px-6 pb-8">
        <div className="flex items-center gap-3 mb-6">
          <div className="h-px flex-1 bg-zu-border" />
          <h2 className="text-[13px] font-bold tracking-widest text-zu-text2">PYMES · 1 A 50 TRABAJADORES · RIESGO I–III</h2>
          <div className="h-px flex-1 bg-zu-border" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {PYME_PLANS.map((p) => <PlanCard key={p.name} plan={p} />)}
        </div>
      </section>

      {/* Segmento Empresa grande / alto riesgo */}
      <section className="max-w-6xl mx-auto px-6 py-8">
        <div className="flex items-center gap-3 mb-6">
          <div className="h-px flex-1 bg-zu-border" />
          <h2 className="text-[13px] font-bold tracking-widest text-zu-text2">EMPRESAS +50 TRAB. · RIESGO IV–V · ALTO CUMPLIMIENTO</h2>
          <div className="h-px flex-1 bg-zu-border" />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
          {ENTERPRISE_PLANS.map((p) => <PlanCard key={p.name} plan={p} />)}
        </div>
        <p className="text-[11px] text-zu-text3 mt-4 text-center">
          * Módulo de drones y biometría sujeto a habilitación regulatoria (Aerocivil) y disponibilidad por zona.<br/>Precios de referencia 2026, no incluyen IVA. El plan Esencial cubre el mantenimiento de un SG-SST ya implementado; el diseño e implementación inicial desde cero se cotiza como servicio aparte. El valor final se confirma en la cotización según sedes, ciudad y alcance.
        </p>
      </section>

      {/* Franja de confianza normativa */}
      <section className="max-w-6xl mx-auto px-6 py-12">
        <div className="rounded-2xl p-8 text-white" style={{ background: "linear-gradient(135deg,#12171F,#1A2129)" }}>
          <h2 className="text-[22px] font-extrabold mb-5">Cumplimiento respaldado por la norma</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {[
              { t: "Res. 0312/2019", d: "Estándares mínimos según tamaño y riesgo" },
              { t: "Res. 908/2025", d: "Profesional con licencia SST vigente y validada" },
              { t: "Decreto 1072/2015", d: "SG-SST y ciclo PHVA" },
              { t: "Ley 1581/2012", d: "Protección de datos de salud" },
            ].map((n) => (
              <div key={n.t}>
                <div className="text-[13px] font-bold text-[#8BC34A] font-mono">{n.t}</div>
                <div className="text-[11.5px] text-[#9BA3AE] mt-1 leading-snug">{n.d}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA final */}
      <section className="max-w-6xl mx-auto px-6 pb-20 text-center">
        <h2 className="text-[26px] font-extrabold text-zu-graphite mb-3">¿No sabes qué plan necesita tu empresa?</h2>
        <p className="text-[14px] text-zu-text2 mb-6 max-w-xl mx-auto">
          Cuéntanos cuántos trabajadores tienes y tu actividad económica. Te decimos exactamente qué exige la ley y qué plan te corresponde.
        </p>
        <button className="px-8 py-3 rounded-lg text-white font-bold text-[14px]" style={{ background: "linear-gradient(135deg,#8BC34A,#6BA31A)" }}>
          Solicitar diagnóstico gratuito
        </button>
      </section>

      <footer className="border-t border-zu-border py-8 text-center text-[11px] text-zu-text3">
        ZU Seguridad Integral 360 · Colombia 2026 · Gestión profesional de SG-SST
      </footer>
    </div>
  );
}
