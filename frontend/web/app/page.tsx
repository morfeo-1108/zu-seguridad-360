"use client";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  ShieldCheck, Building2, HardHat, FileCheck2, GraduationCap, FileText,
  Lock, Boxes, ArrowRight, Sparkles, ScrollText, RadioTower, Fingerprint, Eye,
} from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { Reveal, spotlightHandler } from "@/components/landing/reveal";
import { useAuthStore } from "@/lib/store";

const CAPABILITIES = [
  { icon: Building2, title: "Toda tu cartera, un solo lugar", desc: "Administra todas tus empresas cliente desde una sola cuenta, sin mezclar nunca la información de una con otra." },
  { icon: Lock, title: "Control de acceso real", desc: "Cada empresa ve únicamente su propia información. El aislamiento entre clientes está en el diseño del sistema, no en la buena voluntad de quien lo configura." },
  { icon: ScrollText, title: "PHVA sobre 5 marcos normativos", desc: "ISO 45001, 9001, 14001, 31000 y Resolución 0312 de 2019, con el ciclo Planear-Hacer-Verificar-Actuar y aprobación de un profesional en cada fase." },
  { icon: HardHat, title: "Trabajadores y EPP bajo control", desc: "Expedientes, cargos, restricciones médicas y elementos de protección personal, con historial completo y trazable por trabajador." },
  { icon: FileCheck2, title: "Incidentes y FURAT sin fricción", desc: "Investigación de accidentes con trazabilidad completa y el FURAT listo para radicar ante la ARL." },
  { icon: GraduationCap, title: "Capacitación con vencimientos bajo control", desc: "Cursos, asistencia y certificados con alertas automáticas antes de que una vigencia se venza." },
  { icon: FileText, title: "Documentos y certificados verificables", desc: "Generación con plantillas, control de versiones y verificación pública por folio — cualquiera puede confirmar que un certificado es auténtico." },
  { icon: ShieldCheck, title: "Permisos por rol, sin excepciones", desc: "Cada acción queda ligada a un rol y una empresa específicos, con revisión humana obligatoria para cualquier cambio normativo." },
];

const DIFFERENTIATORS = [
  { icon: Lock, title: "Aislamiento verificado, no prometido", desc: "La mayoría de plataformas de SST en Colombia declara gestión multiempresa. La nuestra se probó con intentos reales de acceso cruzado entre cuentas distintas — no es una frase de la página de ventas." },
  { icon: Sparkles, title: "Asistencia con trazabilidad, no una caja negra", desc: "Cuando la IA entra en juego, cita su fuente normativa y pasa por la aprobación de un profesional antes de modificar cualquier registro. Ninguna decisión se toma sola." },
  { icon: FileCheck2, title: "Evidencia con procedencia", desc: "Quién hizo qué, cuándo y sobre cuál empresa — siempre exportable. Ante una auditoría del Ministerio, eso pesa más que una promesa de cumplimiento." },
  { icon: Eye, title: "Precio público, sin letra pequeña", desc: "En un mercado donde ningún proveedor publica su precio, el nuestro está en esta misma página. Sin cotización, sin esperar una llamada de ventas." },
];

const ROADMAP_MODULES = [
  { icon: Sparkles, title: "Asistente normativo con IA", desc: "Consulta en lenguaje natural sobre la normativa colombiana de SST, con la fuente citada en cada respuesta." },
  { icon: RadioTower, title: "Inspección por visión artificial", desc: "Validación fotográfica del uso correcto de EPP por zona de trabajo." },
  { icon: Fingerprint, title: "Control de acceso biométrico", desc: "Identificación de trabajadores en campo, con consentimiento explícito y trazabilidad del dato." },
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

  // The autoplay attribute alone is unreliable across browsers; force
  // playback explicitly once the element exists, and retry on the
  // visibility/canplay events in case the first attempt was blocked before
  // enough of the file had buffered.
  useEffect(() => {
    const el = videoRef.current;
    if (!el) return;
    const tryPlay = () => { el.play().catch(() => { /* retried on canplay/loadeddata below */ }); };
    tryPlay();
    el.addEventListener("canplay", tryPlay);
    el.addEventListener("loadeddata", tryPlay);
    return () => {
      el.removeEventListener("canplay", tryPlay);
      el.removeEventListener("loadeddata", tryPlay);
    };
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
          <Link href="/auth/login" className="rounded-lg border border-white/15 bg-white/5 px-5 py-2.5 text-[13px] font-bold text-white backdrop-blur-sm transition hover:border-zu-lime hover:bg-zu-lime hover:text-[#0B0F0C]">
            Iniciar sesión
          </Link>
        </div>
      </header>

      {/* ── Hero: video de marca a pantalla completa ── */}
      <section className="relative flex min-h-[100vh] items-center overflow-hidden">
        <video
          ref={videoRef}
          autoPlay
          muted
          loop
          playsInline
          preload="auto"
          poster="/videos/hero-poster.jpg"
          onCanPlay={() => setVideoReady(true)}
          onLoadedData={() => setVideoReady(true)}
          className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-700 ${videoReady ? "opacity-100" : "opacity-0"}`}
        >
          <source src="/videos/hero.webm" type="video/webm" />
          <source src="/videos/hero.mp4" type="video/mp4" />
        </video>
        <img
          src="/videos/hero-poster.jpg"
          alt=""
          aria-hidden
          className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-700 ${videoReady ? "opacity-0" : "opacity-100"}`}
        />
        {/* Velo oscuro para legibilidad del texto sobre el video */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/75 via-black/55 to-[#0B0F0C]" />
        <div className="absolute inset-0 bg-gradient-to-r from-black/70 via-black/20 to-black/60" />

        <div className="relative z-10 mx-auto w-full max-w-5xl px-6 pt-28 text-center lg:px-10">
          <div className="animate-hero-in">
            <div className="mb-7 inline-flex items-center gap-2 rounded-full border border-white/15 bg-white/5 px-4 py-1.5 text-[12px] font-semibold tracking-wide text-zu-lime backdrop-blur-sm">
              <span className="h-1.5 w-1.5 rounded-full bg-zu-lime" />
              SOFTWARE DE SEGURIDAD Y SALUD EN EL TRABAJO · COLOMBIA
            </div>
            <h1 className="text-[42px] font-extrabold leading-[1.05] tracking-tight sm:text-[58px] lg:text-[72px]">
              Prevenimos hoy,
              <br />
              <span className="bg-gradient-to-r from-zu-lime to-emerald-400 bg-clip-text text-transparent">
                protegemos siempre.
              </span>
            </h1>
            <p className="mx-auto mt-7 max-w-2xl text-[17px] leading-relaxed text-white/75 sm:text-[20px]">
              La plataforma que un profesional SST necesita para gestionar varias empresas a la vez:
              control de acceso verificado en cada solicitud, trazabilidad completa y un ciclo PHVA
              construido sobre los marcos normativos que ya exige la ley colombiana.
            </p>
            <div className="mx-auto mt-6 flex max-w-2xl flex-wrap items-center justify-center gap-x-6 gap-y-2 text-[12px] font-semibold uppercase tracking-wide text-white/50">
              <span>Resolución 0312</span><span className="text-white/20">·</span>
              <span>Decreto 1072</span><span className="text-white/20">·</span>
              <span>ISO 45001</span><span className="text-white/20">·</span>
              <span>GTC‑45</span>
            </div>
            <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
              <Link
                href="/auth/login"
                className="group inline-flex items-center gap-2 rounded-full bg-zu-lime px-8 py-4 text-[15px] font-bold text-[#0B0F0C] shadow-[0_8px_30px_rgba(139,195,74,0.35)] transition hover:bg-white"
              >
                Ingresar a la plataforma
                <ArrowRight size={16} className="transition group-hover:translate-x-1" />
              </Link>
              <a href="#capacidades" className="text-[14px] font-semibold text-white/80 hover:text-white transition">
                Ver qué incluye ↓
              </a>
            </div>
          </div>
        </div>

        <div className="absolute inset-x-0 bottom-0 h-32 bg-gradient-to-t from-[#0B0F0C] to-transparent" />
      </section>

      {/* ── Por qué somos diferentes ── */}
      <section className="relative border-t border-white/[0.06] bg-[#0B0F0C] py-20">
        <div className="relative mx-auto max-w-7xl px-6 lg:px-10">
          <Reveal className="mb-12 max-w-2xl">
            <div className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">Por qué ZU 360</div>
            <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[36px]">Lo que ningún otro software de SST en Colombia hace igual.</h2>
          </Reveal>
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {DIFFERENTIATORS.map((d, i) => (
              <Reveal key={d.title} delay={i * 80}>
                <div onMouseMove={spotlightHandler} className="spotlight-card h-full rounded-2xl border border-zu-lime/15 bg-gradient-to-b from-zu-lime/[0.05] to-transparent p-6 transition-all duration-300 hover:-translate-y-1">
                  <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-zu-lime/10 text-zu-lime">
                    <d.icon size={20} />
                  </div>
                  <h3 className="mb-2 text-[15px] font-bold text-white">{d.title}</h3>
                  <p className="text-[13px] leading-relaxed text-white/55">{d.desc}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── Capacidades ── */}
      <section id="capacidades" className="relative overflow-hidden border-t border-white/[0.06] bg-[#0B0F0C] py-24">
        <div className="pointer-events-none absolute inset-0 opacity-[0.07]" style={{ backgroundImage: "radial-gradient(circle, #fff 1px, transparent 1px)", backgroundSize: "28px 28px" }} />
        <div className="pointer-events-none absolute -left-32 top-10 h-[420px] w-[420px] animate-float-slow rounded-full bg-zu-lime/10 blur-[110px]" />
        <div className="pointer-events-none absolute -right-24 bottom-0 h-[360px] w-[360px] animate-float-slower rounded-full bg-emerald-500/10 blur-[110px]" />
        <div className="relative mx-auto max-w-7xl px-6 lg:px-10">
          <Reveal className="mb-14 max-w-2xl">
            <div className="mb-3 text-[12px] font-bold uppercase tracking-widest text-zu-lime">La plataforma</div>
            <h2 className="text-[32px] font-extrabold tracking-tight sm:text-[40px]">Construida sobre el rigor que exige la ley, no sobre una lista de funciones.</h2>
            <p className="mt-3 text-[15px] text-white/55">
              Cada módulo nace de un requisito normativo real del sistema de gestión de SST en Colombia — pensado para el día a día de quien administra varias empresas a la vez.
            </p>
          </Reveal>
          <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {CAPABILITIES.map((c, i) => (
              <Reveal key={c.title} delay={(i % 4) * 80}>
                <div onMouseMove={spotlightHandler} className="spotlight-card group h-full rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6 transition-all duration-300 hover:-translate-y-1 hover:border-white/[0.14] hover:bg-white/[0.04] hover:shadow-[0_20px_40px_-20px_rgba(0,0,0,0.5)]">
                  <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-zu-lime/10 text-zu-lime transition group-hover:bg-zu-lime group-hover:text-[#0B0F0C]">
                    <c.icon size={20} />
                  </div>
                  <h3 className="mb-2 text-[15px] font-bold text-white">{c.title}</h3>
                  <p className="text-[13px] leading-relaxed text-white/50">{c.desc}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── Hoja de ruta ── */}
      <section id="piloto" className="relative overflow-hidden border-t border-white/[0.06] bg-[#0D1210] py-24">
        <div className="pointer-events-none absolute inset-0 opacity-[0.06]" style={{ backgroundImage: "radial-gradient(circle, #fff 1px, transparent 1px)", backgroundSize: "28px 28px" }} />
        <div className="pointer-events-none absolute left-1/3 top-0 h-[380px] w-[380px] animate-float-slower rounded-full bg-amber-400/10 blur-[120px]" />
        <div className="relative mx-auto max-w-7xl px-6 lg:px-10">
          <Reveal className="mb-14 flex flex-wrap items-end justify-between gap-6">
            <div className="max-w-xl">
              <div className="mb-3 text-[12px] font-bold uppercase tracking-widest text-amber-400">En desarrollo</div>
              <h2 className="text-[28px] font-extrabold tracking-tight sm:text-[34px]">La siguiente frontera: inteligencia artificial aplicada a la prevención.</h2>
              <p className="mt-3 text-[15px] text-white/55">
                Estamos construyendo estos módulos con el mismo estándar de rigor normativo
                y protección de datos del resto de la plataforma, antes de ponerlos en manos
                de cada profesional.
              </p>
            </div>
            <Boxes size={48} className="shrink-0 text-white/15" />
          </Reveal>
          <div className="grid gap-5 sm:grid-cols-3">
            {ROADMAP_MODULES.map((m, i) => (
              <Reveal key={m.title} delay={i * 90}>
                <div onMouseMove={spotlightHandler} className="spotlight-card amber h-full rounded-2xl border border-amber-500/15 bg-amber-500/[0.04] p-6 transition-all duration-300 hover:-translate-y-1">
                  <div className="mb-4 inline-flex h-11 w-11 items-center justify-center rounded-xl bg-amber-500/10 text-amber-400">
                    <m.icon size={20} />
                  </div>
                  <h3 className="mb-2 text-[15px] font-bold text-white">{m.title}</h3>
                  <p className="text-[13px] leading-relaxed text-white/50">{m.desc}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA final ── */}
      <section className="relative overflow-hidden border-t border-white/[0.06] bg-[#0B0F0C] py-24">
        <div className="pointer-events-none absolute left-1/2 top-1/2 h-[500px] w-[700px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-zu-lime/[0.06] blur-[140px]" />
        <Reveal className="relative mx-auto max-w-4xl px-6 text-center lg:px-10">
          <h2 className="text-[30px] font-extrabold tracking-tight sm:text-[38px]">
            Lleva la gestión de tu cartera a otro nivel.
          </h2>
          <p className="mx-auto mt-4 max-w-md text-[15px] text-white/55">
            Ingresa con tu cuenta de profesional SST y descubre todo lo que ZU 360 puede hacer por tus empresas.
          </p>
          <Link
            href="/auth/login"
            className="group mt-8 inline-flex items-center gap-2 rounded-full bg-zu-lime px-8 py-4 text-[15px] font-bold text-[#0B0F0C] shadow-[0_8px_30px_rgba(139,195,74,0.25)] transition hover:bg-white hover:shadow-[0_8px_40px_rgba(139,195,74,0.4)]"
          >
            Ingresar a ZU 360
            <ArrowRight size={17} className="transition group-hover:translate-x-1" />
          </Link>
        </Reveal>
      </section>

      <footer className="border-t border-white/[0.06] bg-[#0B0F0C] px-6 py-8 text-center text-[12px] text-white/35 lg:px-10">
        ZU Seguridad Integral 360 · Colombia
      </footer>
    </div>
  );
}
