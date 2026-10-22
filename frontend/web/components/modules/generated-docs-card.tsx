"use client";
import { useEffect, useState } from "react";
import { Download, FilePlus2, ShieldCheck } from "lucide-react";
import { Badge, Button, Card, Spinner } from "@/components/ui";
import { generatedDocAPI, type GeneratedDocRecord, type TemplateRecord } from "@/lib/api";

// Documents generated from templates (documents service): each generation
// is a new version with author, date and SHA-256, downloadable later. The
// company's identity on the document comes from Core, not from this form.
export function GeneratedDocsCard() {
  const [templates, setTemplates] = useState<TemplateRecord[]>([]);
  const [docs, setDocs] = useState<GeneratedDocRecord[] | null>(null);
  const [tpl, setTpl] = useState("politica_sst");
  const [content, setContent] = useState("");
  const [busy, setBusy] = useState(false);
  const [msg, setMsg] = useState("");

  const load = async () => setDocs((await generatedDocAPI.list()).data.data);
  useEffect(() => {
    generatedDocAPI.templates().then((r) => setTemplates(r.data.templates)).catch(() => {});
    load().catch(() => setDocs([]));
  }, []);
  const kind = templates.find((t) => t.key === tpl)?.kind;

  const generate = async () => {
    setBusy(true); setMsg("");
    try {
      const r = await generatedDocAPI.generate(tpl, content ? { content } : {});
      setMsg(`Generado: versión ${r.data.version}.`);
      await load();
    } catch { setMsg("No fue posible generar el documento."); } finally { setBusy(false); }
  };
  const download = async (d: GeneratedDocRecord) => {
    const r = await generatedDocAPI.download(d.id);
    const a = document.createElement("a");
    a.href = URL.createObjectURL(r.data);
    a.download = `${d.template_key}-v${d.version}.pdf`;
    a.click();
    URL.revokeObjectURL(a.href);
  };

  return (
    <Card className="mt-5 p-4">
      <h2 className="mb-1 text-sm font-bold">Documentos generados</h2>
      <p className="mb-3 text-[12px] text-zu-text2">Cada generación queda como versión nueva, con autor, fecha y huella SHA‑256.</p>
      <div className="mb-4 flex flex-wrap items-end gap-2">
        <label className="text-[12px] font-semibold">Plantilla
          <select value={tpl} onChange={(e) => setTpl(e.target.value)} className="mt-1 block rounded-lg border border-zu-border bg-white px-3 py-2 text-[13px]">
            {templates.map((t) => <option key={t.key} value={t.key}>{t.title}{t.kind === "generica" ? " (genérica)" : ""}</option>)}
          </select>
        </label>
        <Button onClick={() => void generate()} disabled={busy}><FilePlus2 size={13} /> {busy ? "Generando…" : "Generar PDF"}</Button>
      </div>
      {kind === "generica" && (
        <label className="mb-4 block text-[12px] font-semibold">Contenido del documento
          <textarea value={content} onChange={(e) => setContent(e.target.value)} rows={4} className="mt-1 block w-full rounded-lg border border-zu-border px-3 py-2 text-[13px]" placeholder="Esta plantilla no redacta contenido: escriba aquí el texto del documento." />
        </label>
      )}
      {msg && <p role="status" className="mb-3 text-xs text-zu-text2">{msg}</p>}
      {docs === null ? <Spinner size={16} /> : docs.length === 0 ? <p className="text-xs text-zu-text3">Aún no hay documentos generados para esta empresa.</p> : (
        <ul className="divide-y divide-zu-border">
          {docs.map((d) => (
            <li key={d.id} className="flex flex-wrap items-center justify-between gap-2 py-2.5">
              <div className="min-w-0">
                <p className="text-[13px] font-semibold">{d.title} <Badge variant="gray">v{d.version}</Badge></p>
                <p className="flex items-center gap-1 text-[10.5px] text-zu-text3"><ShieldCheck size={11} /> SHA‑256 {d.sha256.slice(0, 16)}… · {new Date(d.created_at).toLocaleString("es-CO")}</p>
              </div>
              <Button size="sm" variant="secondary" onClick={() => void download(d)}><Download size={12} /> PDF</Button>
            </li>
          ))}
        </ul>
      )}
    </Card>
  );
}
