"use client";
import { useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowRight, Check, Lock, FileCheck2, AlarmClock, Sunrise, HeartPulse, Scale, Minus } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { Reveal, spotlightHandler } from "@/components/landing/reveal";
import { ScrubVideo } from "@/components/landing/scrub-video";
import { SmoothScroll } from "@/components/landing/smooth-scroll";
import { useAuthStore } from "@/lib/store";
import { BEFORE_AFTER, DIFFERENTIATORS, ROADMAP, TARGETS } from "@/lib/landing-claims";
import { NORMS_INDEXED } from "@/lib/normative-corpus";

// Three.js and the particle canvas load after hydration, off the critical path.
const RefractionSphere = dynamic(() => import("@/components/landing/refraction-sphere").then((m) => m.RefractionSphere), { ssr: false });
const Particles = dynamic(() => import("@/components/landing/particles").then((m) => m.Particles), { ssr: false });

const ICONS: Record<string, typeof Lock> = {
  aislamiento: Lock, trazabilidad: FileCheck2, plazos: AlarmClock, briefing: Sunrise, datos_salud: HeartPulse, motor_normativo: Scale,
};

const VIDEO = {
  wide: { webm: "/videos/hero-scrub.webm", mp4: "/videos/hero-scrub.mp4" },
  narrow: { webm: "/videos/hero-scrub-540.webm", mp4: "/videos/hero-scrub-540.mp4" },
};

const PROFILES = {
  profesional: {
    label: "Profesional SST independiente",
    points: [
      "Atiende varias empresas desde una sola cuenta, cambiando de empresa sin cerrar sesión.",
      "Deja de transcribir hallazgos: el plan de acción sale de la inspección.",
      "Deja de vigilar plazos de memoria: el vencimiento le llega como alerta.",
      "Empieza el día con un resumen de pendientes de toda su cartera.",
      "Genera documentos del SG-SST con versión y huella, listos para una auditoría.",
    ],
  },
  empresa: {
    label: "Empresa cliente",
    points: [
      "Su información vive aislada de la de las demás empresas del profesional.",
      "Cada documento, inspección y aprobación queda con autor y fecha.",
      "La matriz legal, el COPASST, el EPP y el ciclo PHVA en un mismo lugar.",
      "Los datos de salud de sus trabajadores no se procesan hasta cerrar sus controles de seguridad.",
    ],
  },
} as const;

export default function HomePage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const heroRef = useRef<HTMLElement>(null);
  const [scrolled, setScrolled] = useState(false);
  const [profile, setProfile] = useState<keyof typeof PROFILES>("profesional");

  useEffect(() => {
    if (isAuthenticated) router.replace("/dashboard");
  }, [isAuthenticated, router]);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  if (isAuthenticated) return null;

  const verified = DIFFERENTIATORS.filter((d) => d.status === "verificado");
  const upcoming = [...DIFFERENTIATORS.filter((d) => d.status !== "verificado"), ...ROADMAP];

  return (
    <div className="min-h-screen bg-[#0B0F0C] text-white">
      <SmoothScroll />
      <a href="#contenido" className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[60] focus:rounded-lg focus:bg-zu-lime focus:px-4 focus:py-2 focus:text-[#0B0F0C]">
        Saltar al contenido
      </a>

      <header className={`fixed inset-x-0 top-0 z-50 transition-colors duration-300 ${scrolled ? "border-b border-white/[0.06] bg-[#0B0F0C]/90 backdrop-blur-md" : "bg-transparent"}`}>
        <div className="mx-auto flex max-w-7xl items-center justify-between px-4 py-4 sm:px-6 lg:px-10">
          <ZULogo size={34} dark showText />
          <nav aria-label="Secciones" className="hidden items-center gap-8 text-[13px] font-medium text-white/75 md:flex">
            <a href="#diferencias" className="transition hover:text-white">Qué cambia</a>
            <a href="#perfiles" className="transition hover:text-white">Para quién</a>
            <a href="#normativa" className="transition hover:text-white">Normativa</a>
            <Link href="/planes" className="transition hover:text-white">Planes</Link>
          </nav>
          <Link href="/auth/login" className="rounded-lg border border-white/20 bg-white/5 px-4 py-2.5 text-[13px] font-bold text-white backdrop-blur-sm transition hover:border-zu-lime hover:bg-zu-lime hover:text-[#0B0F0C] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-zu-lime sm:px-5">
            Iniciar sesión
          </Link>
        </div>
      </header>

      {/* ── Hero: video scrubeado por scroll + esfera ── */}
      <section ref={heroRef} className="relative h-[100svh] min-h-[600px] overflow-hidden">
        <ScrubVideo triggerRef={heroRef} poster="/videos/hero-poster.webp" sources={VIDEO} className="absolute inset-0" />
        <div className="absolute inset-0 bg-gradient-to-b from-black/80 via-black/55 to-[#0B0F0C]" />
        <div className="absolute inset-0 bg-gradient-to-r from-black/75 via-black/30 to-transparent" />
        <Particles className="pointer-events-none absolute inset-0 h-full w-full" />
        <RefractionSphere className="pointer-events-none absolute -right-[18%] top-[6%] h-[46vw] w-[70vw] opacity-60 md:right-0 md:top-0 md:h-full md:w-[46%] md:opacity-100" />

        <div className="relative z-10 mx-auto flex h-full max-w-7xl items-center px-4 sm:px-6 lg:px-10">
          <div className="max-w-2xl pt-16 animate-hero-in">
            <p className="mb-6 inline-flex items-center gap-2 rounded-full border border-white/15 bg-black/30 px-4 py-1.5 text-[12px] font-semibold tracking-wide text-zu-lime backdrop-blur-sm">
              <span className="h-1.5 w-1.5 rounded-full bg-zu-lime" aria-hidden />
              SEGURIDAD Y SALUD EN EL TRABAJO · COLOMBIA
            </p>
            <h1 className="text-[38px] font-extrabold leading-[1.05] tracking-tight sm:text-[54px] lg:text-[64px]">
              Lleve todas sus empresas
              <br />
              <span className="text-zu-lime">sin perder el criterio</span> en ninguna.
            </h1>
            <p className="mt-6 max-w-xl text-[17px] leading-relaxed text-white/80 sm:text-[19px]">
              La plataforma para el profesional SST que gestiona varias empresas: cada hallazgo con plazo,
              cada documento con versión y huella, y un resumen de su cartera cada mañana. Las decisiones siguen siendo suyas.
            </p>
            <div className="mt-9 flex flex-wrap items-center gap-4">
              <Link href="/auth/login" className="group inline-flex items-center gap-2 rounded-full bg-zu-lime px-7 py-3.5 text-[15px] font-bold text-[#0B0F0C] shadow-[0_8px_30px_rgba(139,195,74,0.35)] transition hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white">
                Ingresar a la plataforma
                <ArrowRight size={16} className="transition group-hover:translate-x-1" aria-hidden />
              </Link>
              <a href="#diferencias" className="rounded-full px-2 py-2 text-[14px] font-semibold text-white/85 transition hover:text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-white">
                Ver qué cambia ↓
              </a>
            </div>
          </div>
        </div>
      </section>

      <main id="contenido">
        {/* ── Antes / después ── */}
        <section id="diferencias" className="border-t border-white/[0.06] bg-[#0B0F0C] py-20 sm:py-24">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-10">
            <Reveal className="mb-12 max-w-2xl">
              <p className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Qué cambia en su día</p>
              <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[38px]">Del control manual a un sistema que le avisa.</h2>
            </Reveal>
            <Reveal>
              <div className="overflow-hidden rounded-2xl border border-white/[0.08]">
                <div className="grid grid-cols-1 bg-white/[0.03] text-[12px] font-bold uppercase tracking-wider text-white/60 sm:grid-cols-2">
                  <div className="hidden px-6 py-4 sm:block">Sin ZU 360</div>
                  <div className="hidden px-6 py-4 text-zu-lime sm:block">Con ZU 360</div>
                </div>
                {BEFORE_AFTER.map((r) => (
                  <div key={r.after} className="grid grid-cols-1 border-t border-white/[0.06] sm:grid-cols-2">
                    <p className="flex gap-3 px-6 py-4 text-[14px] leading-relaxed text-white/60"><Minus size={16} className="mt-0.5 shrink-0 text-white/35" aria-hidden /><span><span className="sr-only">Sin ZU 360: </span>{r.before}</span></p>
                    <p className="flex gap-3 px-6 pb-5 pt-0 text-[14px] leading-relaxed text-white sm:py-4"><Check size={16} className="mt-0.5 shrink-0 text-zu-lime" aria-hidden /><span><span className="sr-only">Con ZU 360: </span>{r.after}</span></p>
                  </div>
                ))}
              </div>
            </Reveal>
            <Reveal className="mt-10">
              <div className="grid gap-4 sm:grid-cols-3">
                {TARGETS.map((t) => (
                  <div key={t.metric} className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-5">
                    <p className="text-[28px] font-extrabold text-white">{t.value}</p>
                    <p className="mt-1 text-[14px] text-white/75">{t.metric}</p>
                    <p className="mt-2 text-[12px] font-semibold uppercase tracking-wide text-amber-300/90">{t.note}</p>
                  </div>
                ))}
              </div>
              <p className="mt-4 max-w-3xl text-[13px] leading-relaxed text-white/55">
                Estas tres cifras son metas del diseño del producto, no resultados medidos. Publicaremos las cifras reales cuando el piloto con profesionales SST las haya medido.
              </p>
            </Reveal>
          </div>
        </section>

        {/* ── Diferenciadores verificados ── */}
        <section className="border-t border-white/[0.06] bg-[#0D1210] py-20 sm:py-24">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-10">
            <Reveal className="mb-12 max-w-2xl">
              <p className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Lo que ya funciona</p>
              <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[38px]">Cada afirmación de esta página está probada en el sistema.</h2>
            </Reveal>
            <ul className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
              {verified.map((d, i) => {
                const Icon = ICONS[d.id] ?? Check;
                return (
                  <li key={d.id}>
                    <Reveal delay={(i % 3) * 80} className="h-full">
                      <div onMouseMove={spotlightHandler} className="spotlight-card h-full rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6">
                        <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-zu-lime/10 text-zu-lime"><Icon size={20} aria-hidden /></div>
                        <h3 className="mb-2 text-[16px] font-bold">{d.title}</h3>
                        <p className="text-[14px] leading-relaxed text-white/70">{d.body}</p>
                        {d.proof && <p className="mt-4 border-l-2 border-zu-lime/60 pl-3 text-[13px] leading-relaxed text-white/60">{d.proof}</p>}
                      </div>
                    </Reveal>
                  </li>
                );
              })}
            </ul>
          </div>
        </section>

        {/* ── Beneficios por perfil ── */}
        <section id="perfiles" className="border-t border-white/[0.06] bg-[#0B0F0C] py-20 sm:py-24">
          <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-10">
            <Reveal className="mb-10 max-w-2xl">
              <p className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Para quién</p>
              <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[38px]">Lo que gana cada lado.</h2>
            </Reveal>
            <div role="tablist" aria-label="Perfil" className="mb-6 inline-flex rounded-full border border-white/10 bg-white/[0.03] p-1">
              {(Object.keys(PROFILES) as (keyof typeof PROFILES)[]).map((k) => (
                <button
                  key={k} role="tab" id={`tab-${k}`} aria-selected={profile === k} aria-controls={`panel-${k}`}
                  onClick={() => setProfile(k)}
                  onKeyDown={(e) => { if (e.key === "ArrowRight" || e.key === "ArrowLeft") { const next = k === "profesional" ? "empresa" : "profesional"; setProfile(next); document.getElementById(`tab-${next}`)?.focus(); } }}
                  tabIndex={profile === k ? 0 : -1}
                  className={`rounded-full px-5 py-2.5 text-[14px] font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-zu-lime ${profile === k ? "bg-zu-lime text-[#0B0F0C]" : "text-white/75 hover:text-white"}`}
                >
                  {PROFILES[k].label}
                </button>
              ))}
            </div>
            <div role="tabpanel" id={`panel-${profile}`} aria-labelledby={`tab-${profile}`} className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 sm:p-8">
              <ul className="space-y-4">
                {PROFILES[profile].points.map((p) => (
                  <li key={p} className="flex gap-3 text-[15px] leading-relaxed text-white/85"><Check size={18} className="mt-0.5 shrink-0 text-zu-lime" aria-hidden />{p}</li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        {/* ── Normativa y estado del proyecto ── */}
        <section id="normativa" className="border-t border-white/[0.06] bg-[#0D1210] py-20 sm:py-24">
          <div className="mx-auto grid max-w-7xl gap-12 px-4 sm:px-6 lg:grid-cols-2 lg:px-10">
            <Reveal>
              <p className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Marco normativo</p>
              <h2 className="text-[28px] font-extrabold tracking-tight sm:text-[34px]">Normas vigentes en la base de consulta</h2>
              <p className="mt-3 text-[14px] leading-relaxed text-white/65">
                Textos oficiales, artículo por artículo, con control de vigencia: cuando una norma fue derogada, se indica por cuál.
              </p>
              <ul className="mt-6 grid gap-2 sm:grid-cols-2">
                {NORMS_INDEXED.filter((n) => n.status !== "derogada").map((n) => (
                  <li key={n.name} className="rounded-xl border border-white/[0.07] bg-white/[0.02] px-4 py-3">
                    <p className="text-[14px] font-semibold">{n.name}</p>
                    <p className="text-[12px] text-white/55">{n.topic}</p>
                  </li>
                ))}
              </ul>
            </Reveal>
            <Reveal delay={100}>
              <p className="mb-3 text-[12px] font-bold uppercase tracking-widest text-amber-300">Estado del proyecto</p>
              <h2 className="text-[28px] font-extrabold tracking-tight sm:text-[34px]">Un sistema nuevo, en piloto, y lo decimos.</h2>
              <p className="mt-3 text-[14px] leading-relaxed text-white/70">
                ZU 360 está en piloto con profesionales SST en Colombia desde octubre de 2026. No tenemos diez años en el mercado;
                tenemos un sistema que hace lo que dice esta página, y una lista honesta de lo que aún no hace:
              </p>
              <ul className="mt-6 space-y-3">
                {upcoming.map((r) => (
                  <li key={r.id} className="rounded-xl border border-amber-400/15 bg-amber-400/[0.04] px-4 py-3">
                    <p className="text-[14px] font-semibold">{r.title} <span className="ml-1 text-[11px] font-bold uppercase tracking-wide text-amber-300">próximamente</span></p>
                    <p className="mt-1 text-[13px] leading-relaxed text-white/60">{r.body}</p>
                  </li>
                ))}
              </ul>
            </Reveal>
          </div>
        </section>

        <section className="border-t border-white/[0.06] bg-[#0B0F0C] py-20 sm:py-24">
          <Reveal className="mx-auto max-w-3xl px-4 text-center sm:px-6">
            <h2 className="text-[28px] font-extrabold tracking-tight sm:text-[36px]">Pruébelo con una de sus empresas.</h2>
            <p className="mx-auto mt-4 max-w-md text-[15px] text-white/65">Ingrese con su cuenta de Google o con correo y contraseña. El precio de cada plan es público.</p>
            <div className="mt-8 flex flex-wrap justify-center gap-4">
              <Link href="/auth/login" className="group inline-flex items-center gap-2 rounded-full bg-zu-lime px-7 py-3.5 text-[15px] font-bold text-[#0B0F0C] transition hover:bg-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white">
                Ingresar <ArrowRight size={16} className="transition group-hover:translate-x-1" aria-hidden />
              </Link>
              <Link href="/planes" className="inline-flex items-center rounded-full border border-white/20 px-7 py-3.5 text-[15px] font-semibold text-white transition hover:border-white/50 focus-visible:outline focus-visible:outline-2 focus-visible:outline-white">
                Ver planes y precios
              </Link>
            </div>
          </Reveal>
        </section>
      </main>

      <footer className="border-t border-white/[0.06] bg-[#0B0F0C] px-4 py-8 text-[12px] text-white/50 sm:px-6 lg:px-10">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-4">
          <span>ZU Seguridad Integral 360 · Colombia</span>
          <nav aria-label="Legal" className="flex gap-6">
            <Link href="/legal/terminos" className="hover:text-white">Términos</Link>
            <Link href="/legal/privacidad" className="hover:text-white">Privacidad</Link>
          </nav>
        </div>
      </footer>
    </div>
  );
}
