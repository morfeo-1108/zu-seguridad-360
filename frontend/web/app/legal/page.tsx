import Link from "next/link";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";

export const metadata = { title: "Documentos legales · ZU 360" };

const DOCS = [
  {
    href: "/legal/terminos",
    title: "Términos del servicio",
    description: "Condiciones de uso de la plataforma durante la fase piloto.",
  },
  {
    href: "/legal/privacidad",
    title: "Aviso de privacidad",
    description: "Cómo tratamos los datos personales, conforme a la Ley 1581 de 2012.",
  },
];

export default function LegalPage() {
  return (
    <div className="min-h-screen bg-[#0B0F0C] text-white">
      <header className="border-b border-white/[0.06]">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-6 py-5">
          <Link href="/"><ZULogo size={28} dark showText /></Link>
          <Link href="/" className="inline-flex items-center gap-1.5 text-[12px] font-semibold text-white/50 hover:text-white transition">
            <ArrowLeft size={14} /> Volver al inicio
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-6 py-12">
        <h1 className="text-[28px] font-extrabold tracking-tight mb-2">Documentos legales</h1>
        <p className="text-[13px] text-white/40 mb-10">Versión 1.0-piloto · Última actualización: 20 de octubre de 2026</p>

        <ul className="space-y-3">
          {DOCS.map((d) => (
            <li key={d.href}>
              <Link
                href={d.href}
                className="flex items-center justify-between gap-4 rounded-xl border border-white/[0.08] bg-white/[0.02] px-5 py-4 transition hover:border-white/20 hover:bg-white/[0.04]"
              >
                <span>
                  <span className="block text-[15px] font-bold">{d.title}</span>
                  <span className="mt-0.5 block text-[13px] text-white/55">{d.description}</span>
                </span>
                <ChevronRight size={18} className="shrink-0 text-white/40" />
              </Link>
            </li>
          ))}
        </ul>
      </main>
    </div>
  );
}
