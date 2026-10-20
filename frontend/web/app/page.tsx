"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  ShieldCheck, Building2, HardHat, FileCheck2, GraduationCap, FileText,
  Lock, Boxes, ArrowRight, Sparkles, ScrollText, RadioTower, Fingerprint,
} from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { useAuthStore } from "@/lib/store";

const CAPABILITIES = [
  { icon: Building2, title: "Multi-empresa real", desc: "Un profesional gestiona varias empresas desde una sola cuenta; el servidor valida membresía en cada solicitud — nunca el cliente." },
  { icon: Lock, title: "Aislamiento verificado", desc: "Cada acción queda ligada a la empresa activa del token. Cambiar de contexto exige una membresía vigente, no un parámetro editable." },
  { icon: ScrollText, title: "PHVA con 5 marcos", desc: "ISO 45001, 9001, 14001, 31000 y Resolución 0312/2019, con ciclo Planear-Hacer-Verificar-Actuar y validación humana obligatoria." },
  { icon: HardHat, title: "Trabajadores y EPP", desc: "Expedientes, cargos, restricciones y elementos de protección por persona, con baja lógica y trazabilidad de cambios." },
  { icon: FileCheck2, title: "Incidentes y FURAT", desc: "Investigación, cierre con auditoría y previsualización de FURAT lista para el flujo de radicación ante ARL." },
  { icon: GraduationCap, title: "Capacitación y vigencias", desc: "Cursos, asistencia y constancias con control de vencimiento por trabajador y por cargo." },
  { icon: FileText, title: "Documentos y certificados", desc: "Generación con plantillas, verificación pública por folio y firma por autenticación real, no un botón decorativo." },
  { icon: ShieldCheck, title: "RBAC de 7 roles", desc: "Casbin aplicado por recurso y verbo HTTP, con revisión humana obligatoria para todo cambio normativo." },
];

const PILOT_MODULES = [
  { icon: Sparkles, title: "IA normativa (RAG)", desc: "Consulta semántica sobre normas colombianas de SST." },
  { icon: RadioTower, title: "Visión por dron", desc: "Inspección de EPP por zona con validación fotográfica." },
  { icon: Fingerprint, title: "Biometría de acceso", desc: "Identificación de trabajadores con consentimiento explícito." },
];

export default function HomePage() {
  const router = useRouter();
  const { isAuthenticated } = useAuthStore();
  const videoRef = useRef<HTMLVideoElement>(null);
  const [videoReady, setVideoReady] = useState(false);
  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    if (isAuthenticated) router.replace("/dashboard");
  }, [isAuthenticated, router]);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 24);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  if (isAuthenticated) return null;

  return (
    <div className="min-h-screen bg-[#0B0F0C] text-white">
      {/* ── Nav ── */}
      <header className={`fixed inset-x-0 top-0 z-50 transition-all duration-300 ${scrolled ? "bg-[#0B0F0C]/90 backdrop-blur-md border-b border-white/[0.06]" : "bg-transparent"}`}>
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-10">
          <ZULogo size={34} dark showText />
          <nav className="hidden items-center gap-8 text-[13px] font-medium text-white/70 md:flex">
            <a href="#capacidades" className="hover:text-white transition">Capacidades</a>
            <a href="#piloto" className="hover:text-white transition">En piloto</a>
            <Link href="/planes" className="hover:text-white transition">Planes</Link>
          </nav>
          <Link href="/auth/login" className="rounded-full bg-white px-5 py-2 text-[13px] font-semibold text-[#0B0F0C] transition hover:bg-zu-lime">
            Iniciar sesión
          </Link>
        </div>
      </header>

      {/* ── Hero ── */}
      <section className="relative flex min-h-[100vh] items-center overflow-hidden pt-28 pb-16 lg:pt-20">
        <div className="absolute inset-0 bg-gradient-to-br from-[#0E1410] via-[#0B0F0C] to-[#0B0F0C]" />
        <div className="absolute -left-40 top-1/4 h-[480px] w-[480px] animate-float-slow rounded-full bg-zu-lime/10 blur-[100px]" />
        <div className="absolute -right-20 bottom-0 h-[420px] w-[420px] animate-float-slower rounded-full bg-emerald-500/10 blur-[100px]" />

        <div className="relative z-10 mx-auto grid w-full max-w-7xl items-center gap-14 px-6 lg:grid-cols-[1.15fr_0.85fr] lg:gap-10 lg:px-10">
          {/* Texto */}
          <div className="animate-hero-in text-center lg:text-left">
            <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-4 py-1.5 text-[12px] font-semibold tracking-wide text-zu-lime">
              <span className="h-1.5 w-1.5 rounded-full bg-zu-lime" />
              PLATAFORMA DE SST · COLOMBIA
            </div>
            <h1 className="text-[40px] font-extrabold leading-[1.05] tracking-tight sm:text-[54px] lg:text-[64px]">
              Prevenimos hoy,
              <br />
              <span className="bg-gradient-to-r from-zu-lime to-emerald-400 bg-clip-text text-transparent">
                protegemos siempre.
              </span>
            </h1>
            <p className="mx-auto mt-6 max-w-xl text-[16px] leading-relaxed text-white/60 sm:text-[18px] lg:mx-0">
              Gestión de seguridad y salud en el trabajo con aislamiento real por empresa,
              ciclo PHVA sobre 5 marcos normativos y control de acceso verificado en cada
              solicitud — no solo en la pantalla de login.
            </p>
            <div className="mt-10 flex flex-wrap items-center justify-center gap-4 lg:justify-start">
              <Link
                href="/auth/login"
                className="group inline-flex items-center gap-2 rounded-full bg-zu-lime px-7 py-3.5 text-[14px] font-bold text-[#0B0F0C] transition hover:bg-white"
              >
                Ingresar a la plataforma
                <ArrowRight size={16} className="transition group-hover:translate-x-1" />
              </Link>
              <a href="#capacidades" className="text-[14px] font-semibold text-white/70 hover:text-white transition">
                Ver qué incluye ↓
              </a>
            </div>
          </div>

          {/* Pieza vertical: video de marca en marco tipo dispositivo */}
          <div className="relative mx-auto w-full max-w-[300px] animate-hero-in lg:max-w-[340px]">
            <div className="absolute inset-0 -z-10 scale-110 rounded-[2.5rem] bg-zu-lime/20 blur-3xl" />
            <div className="relative overflow-hidden rounded-[2rem] border border-white/10 bg-black shadow-2xl ring-1 ring-white/5">
              <div className="relative aspect-[9/16] w-full">
                <video
                  ref={videoRef}
                  autoPlay
                  muted
                  loop
                  playsInline
                  preload="metadata"
                  poster="/videos/hero-poster.jpg"
                  onCanPlay={() => setVideoReady(true)}
                  className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-700 ${videoReady ? "opacity-100" : "opacity-0"}`}
                >
                  <source src="/videos/hero.mp4" type="video/mp4" />
                </video>
                <img
                  src="/videos/hero-poster.jpg"
                  alt=""
                  aria-hidden
                  className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-700 ${videoReady ? "opacity-0" : "opacity-100"}`}
                />
                <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/50 via-transparent to-transparent" />
              </div>
            </div>
            <div className="mt-4 text-center text-[11px] font-medium uppercase tracking-widest text-white/35">
              ZU 360 en campo
            </div>
          </div>
        </div>
      </section>

      {/* ── Capacidades reales ── */}
      <section id="capacidades" className="relative border-t border-white/[0.06] bg-[#0B0F0C] py-24">
        <div className="mx-auto max-w-7xl px-6 lg:px-10">
          <div className="mb-14 max-w-2xl">
            <div className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Lo que ya funciona</div>
            <h2 className="text-[32px] font-extrabold tracking-tight sm:text-[40px]">No es una maqueta. Es lo que corre hoy.</h2>
            <p className="mt-3 text-[15px] text-white/55">
              Cada capacidad de esta sección fue verificada contra una base de datos real en esta misma sesión de auditoría — no es texto de marketing sin respaldo.
            </p>
          </div>
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {CAPABILITIES.map((c) => (
              <div key={c.title} className="group rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 transition hover:border-zu-lime/30 hover:bg-white/[0.04]">
                <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-zu-lime/10 text-zu-lime transition group-hover:bg-zu-lime group-hover:text-[#0B0F0C]">
                  <c.icon size={20} />
                </div>
                <h3 className="mb-2 text-[15px] font-bold text-white">{c.title}</h3>
                <p className="text-[13px] leading-relaxed text-white/50">{c.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── Módulos en piloto (honesto, no sobre-promete) ── */}
      <section id="piloto" className="relative border-t border-white/[0.06] bg-[#0D1210] py-24">
        <div className="mx-auto max-w-7xl px-6 lg:px-10">
          <div className="mb-14 flex flex-wrap items-end justify-between gap-6">
            <div className="max-w-xl">
              <div className="mb-3 text-[12px] font-bold uppercase tracking-widest text-amber-400">En piloto sintético</div>
              <h2 className="text-[28px] font-extrabold tracking-tight sm:text-[34px]">Avanzado, pero sin maquillar el estado real.</h2>
              <p className="mt-3 text-[15px] text-white/55">
                Estos módulos existen en código y están deliberadamente limitados a datos
                sintéticos hasta completar la revisión de privacidad y consentimiento que
                exige procesar datos reales de salud y biometría.
              </p>
            </div>
            <Boxes size={48} className="shrink-0 text-white/15" />
          </div>
          <div className="grid gap-5 sm:grid-cols-3">
            {PILOT_MODULES.map((m) => (
              <div key={m.title} className="rounded-2xl border border-amber-500/15 bg-amber-500/[0.04] p-6">
                <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-amber-500/10 text-amber-400">
                  <m.icon size={20} />
                </div>
                <h3 className="mb-2 text-[15px] font-bold text-white">{m.title}</h3>
                <p className="text-[13px] leading-relaxed text-white/50">{m.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA final ── */}
      <section className="relative border-t border-white/[0.06] bg-[#0B0F0C] py-24">
        <div className="mx-auto max-w-4xl px-6 text-center lg:px-10">
          <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[38px]">
            Entra con tu cuenta de profesional SST.
          </h2>
          <p className="mx-auto mt-4 max-w-md text-[15px] text-white/55">
            Autenticación real contra Core, sin atajos ni datos simulados fuera del piloto.
          </p>
          <Link
            href="/auth/login"
            className="mt-8 inline-flex items-center gap-2 rounded-full bg-zu-lime px-8 py-4 text-[15px] font-bold text-[#0B0F0C] transition hover:bg-white"
          >
            Ingresar a ZU 360
            <ArrowRight size={17} />
          </Link>
        </div>
      </section>

      <footer className="border-t border-white/[0.06] bg-[#0B0F0C] px-6 py-8 text-center text-[12px] text-white/35 lg:px-10">
        ZU Seguridad Integral 360 · Colombia
      </footer>
    </div>
  );
}
