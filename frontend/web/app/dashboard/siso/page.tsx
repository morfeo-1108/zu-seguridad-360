"use client";
import { useState } from "react";
import { AlertTriangle, ExternalLink, Scale, Search } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Button, Card, PageHeader, Spinner } from "@/components/ui";
import { normativeAPI, type NormAnswer } from "@/lib/api";
import { NORMS_INDEXED } from "@/lib/normative-corpus";

const EXAMPLES = [
  "¿Con qué frecuencia debe reunirse el COPASST?",
  "¿Quiénes deben integrar el equipo investigador de un accidente de trabajo?",
  "¿Qué tipos de evaluaciones médicas ocupacionales existen?",
  "¿Cómo se conforma el Comité de Convivencia Laboral?",
];

export default function NormativePage() {
  const [q, setQ] = useState("");
  const [answer, setAnswer] = useState<NormAnswer | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const ask = async (question: string) => {
    if (question.trim().length < 5) return;
    setQ(question); setBusy(true); setError("");
    try { setAnswer((await normativeAPI.query(question.trim())).data); }
    catch { setError("La consulta normativa no está disponible en este momento."); setAnswer(null); }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Consulta normativa" description="Pregunte en lenguaje natural y reciba los artículos vigentes que responden, con su texto literal y el enlace a la fuente oficial. No es un chat que redacta: lo que lee es la norma." />
      <Card className="mb-5 p-4">
        <form onSubmit={(e) => { e.preventDefault(); void ask(q); }} className="flex flex-col gap-2 sm:flex-row">
          <label htmlFor="q" className="sr-only">Pregunta</label>
          <input id="q" value={q} onChange={(e) => setQ(e.target.value)} maxLength={500} placeholder="Ej.: ¿cada cuánto se reúne el COPASST?"
            className="flex-1 rounded-lg border border-zu-border px-3 py-2.5 text-[14px] focus:border-zu-accent focus:outline-none focus:ring-2 focus:ring-zu-accent/20" />
          <Button type="submit" disabled={busy || q.trim().length < 5}><Search size={14} /> {busy ? "Buscando…" : "Consultar"}</Button>
        </form>
        <div className="mt-3 flex flex-wrap gap-2">
          {EXAMPLES.map((e) => <button key={e} onClick={() => void ask(e)} className="rounded-full border border-zu-border px-3 py-1 text-[11.5px] text-zu-text2 transition hover:border-zu-accent hover:text-zu-accent">{e}</button>)}
        </div>
      </Card>

      {busy && <Spinner size={20} />}
      {error && <Card className="p-4 text-sm text-red-700">{error}</Card>}
      {answer && !busy && (
        <div className="space-y-3">
          {answer.warnings.map((w) => (
            <Card key={w} className="flex items-start gap-2 border-amber-300 bg-amber-50 p-3 text-[12.5px] text-amber-900"><AlertTriangle size={15} className="mt-0.5 shrink-0" />{w}</Card>
          ))}
          {answer.sources.map((s, i) => (
            <Card key={s.citation} className="p-4">
              <div className="mb-2 flex flex-wrap items-start justify-between gap-2">
                <div>
                  <p className="text-[11px] font-bold uppercase tracking-wide text-zu-text3">{i === 0 ? "Más pertinente" : `Resultado ${i + 1}`}</p>
                  <h2 className="text-[15px] font-bold">{s.citation}</h2>
                  {s.article_title && <p className="text-[12px] text-zu-text2">{s.article_title}</p>}
                </div>
                <a href={s.source_url} target="_blank" rel="noreferrer" className="inline-flex items-center gap-1 text-[12px] font-semibold text-zu-accent hover:underline">Fuente oficial <ExternalLink size={12} /></a>
              </div>
              <blockquote className="max-h-72 overflow-y-auto whitespace-pre-line border-l-2 border-zu-accent/50 pl-3 text-[13px] leading-relaxed text-zu-text1">{s.text}</blockquote>
            </Card>
          ))}
          <p className="text-[11px] text-zu-text3">{answer.note}</p>
        </div>
      )}

      <Card className="mt-6 p-4">
        <h2 className="mb-2 flex items-center gap-2 text-sm font-bold"><Scale size={14} className="text-zu-accent" /> Normas en la base de consulta</h2>
        <ul className="grid gap-1.5 text-[12px] sm:grid-cols-2">
          {NORMS_INDEXED.map((n) => (
            <li key={n.name} className={n.status === "derogada" ? "text-zu-text3 line-through decoration-zu-text3/50" : ""}>
              <b>{n.name}</b> — {n.topic}{n.derogatedBy ? <span className="no-underline"> (derogada por {n.derogatedBy})</span> : ""}
            </li>
          ))}
        </ul>
        <p className="mt-3 text-[11px] text-zu-text3">La GTC 45 (ICONTEC) no está incluida: es una guía con derechos de autor sin texto público.</p>
      </Card>
    </DashboardShell>
  );
}
