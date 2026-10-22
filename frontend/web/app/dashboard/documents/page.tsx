"use client";
import { ChangeEvent, FormEvent, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, FileText, Image as ImageIcon, Trash2, Upload } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { documentAPI, mediaAPI } from "@/lib/api";
import { SignedImage, SignedLink } from "@/components/modules/signed-file";
import { GeneratedDocsCard } from "@/components/modules/generated-docs-card";
import { useDocuments } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const CATEGORIES = ["politicas", "procedimientos", "formatos", "actas", "contratos", "otro"];
const isImage = (format: string) => ["jpg", "jpeg", "png", "webp"].includes(format.toLowerCase());

export default function DocumentsPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: documents, isLoading, isError } = useDocuments(companyId);
  const { data: photos } = useDocuments(companyId, "foto");
  const [uploading, setUploading] = useState(false);
  const [title, setTitle] = useState("");
  const [category, setCategory] = useState(CATEGORIES[0]);
  const [error, setError] = useState("");
  const docInput = useRef<HTMLInputElement>(null);
  const photoInput = useRef<HTMLInputElement>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["documents", companyId] });

  const uploadAndRegister = async (file: File, cat: string, docTitle: string) => {
    const upload = await mediaAPI.upload(file, cat === "foto" ? "fotos" : "documents");
    const format = file.name.split(".").pop() || "pdf";
    await documentAPI.create(companyId, {
      title: docTitle, category: cat, format,
      file_url: upload.data.url, file_size: upload.data.size, checksum_sha: upload.data.checksum_sha,
    });
  };

  const submitDocument = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const file = docInput.current?.files?.[0];
    if (!file || !companyId || uploading) return;
    setUploading(true); setError("");
    try {
      await uploadAndRegister(file, category, title || file.name);
      await invalidate(); setTitle(""); if (docInput.current) docInput.current.value = "";
    } catch { setError("No se pudo subir el documento."); }
    finally { setUploading(false); }
  };

  const uploadPhoto = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file || !companyId || uploading) return;
    setUploading(true); setError("");
    try {
      await uploadAndRegister(file, "foto", file.name);
      await invalidate();
    } catch { setError("No se pudo subir la fotografía."); }
    finally { setUploading(false); if (photoInput.current) photoInput.current.value = ""; }
  };

  const remove = async (id: string) => {
    try { await documentAPI.remove(companyId, id); await invalidate(); } catch { /* el botón permanece para reintentar */ }
  };

  const nonPhotos = (documents || []).filter((d) => d.category !== "foto");

  return (
    <DashboardShell>
      <PageHeader title="Documentos y fotografías" description="Gestión documental de la empresa activa — subida real a almacenamiento de objetos, con checksum SHA-256 por archivo." />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : (
        <>
          <Card className="mb-5 p-4">
            <h2 className="mb-3 flex items-center gap-1.5 text-sm font-bold"><FileText size={14} /> Subir documento</h2>
            <form onSubmit={submitDocument} className="flex flex-wrap items-end gap-2">
              <label className="text-xs">Título<input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Opcional — usa el nombre del archivo" className="mt-1 block w-56 rounded border border-zu-border px-3 py-2" /></label>
              <label className="text-xs">Categoría
                <select value={category} onChange={(e) => setCategory(e.target.value)} className="mt-1 block rounded border border-zu-border px-3 py-2">
                  {CATEGORIES.map((c) => <option key={c} value={c}>{c}</option>)}
                </select>
              </label>
              <input ref={docInput} type="file" accept=".pdf,.docx,.xlsx,.pptx,.csv,.jpg,.jpeg,.png,.webp" required className="text-xs" />
              <Button type="submit" disabled={uploading}><Upload size={13} /> {uploading ? "Subiendo…" : "Subir"}</Button>
            </form>
            {error && <p role="alert" className="mt-2 text-xs text-red-700">{error}</p>}
          </Card>

          <Card className="mb-5 p-4">
            <div className="mb-3 flex items-center justify-between">
              <h2 className="flex items-center gap-1.5 text-sm font-bold"><ImageIcon size={14} /> Fotografías de la empresa</h2>
              <label className="cursor-pointer text-xs font-semibold text-zu-accent">
                + Subir foto
                <input ref={photoInput} type="file" accept=".jpg,.jpeg,.png,.webp" onChange={uploadPhoto} className="hidden" />
              </label>
            </div>
            {!photos?.length ? (
              <p className="py-4 text-center text-xs text-zu-text3">Sin fotografías registradas.</p>
            ) : (
              <div className="grid grid-cols-2 gap-2 sm:grid-cols-4 lg:grid-cols-6">
                {photos.map((photo) => (
                  <div key={photo.id} className="group relative aspect-square overflow-hidden rounded-md border border-zu-border">
                    <SignedImage url={photo.file_url} alt={photo.title} className="h-full w-full object-cover" />
                    <button onClick={() => void remove(photo.id)} className="absolute right-1 top-1 rounded bg-black/60 p-1 text-white opacity-0 transition group-hover:opacity-100"><Trash2 size={12} /></button>
                  </div>
                ))}
              </div>
            )}
          </Card>

          {isLoading ? <Spinner size={24} /> : isError ? (
            <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudieron cargar los documentos.</p>
          ) : !nonPhotos.length ? (
            <Empty icon={FileText} title="Sin documentos" description="Sube el primer documento de la empresa." />
          ) : (
            <Card className="overflow-hidden p-0">
              <div className="divide-y divide-zu-border">
                {nonPhotos.map((doc) => (
                  <div key={doc.id} className="flex items-center justify-between gap-3 p-3">
                    <div className="min-w-0">
                      <SignedLink url={doc.file_url} className="truncate text-left text-xs font-semibold text-zu-accent hover:underline">{doc.title}</SignedLink>
                      <p className="text-[10px] text-zu-text3">{(doc.file_size / 1024).toFixed(0)} KB · {formatDateTime(doc.created_at)}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant="gray">{doc.category}</Badge>
                      <button onClick={() => void remove(doc.id)} className="text-red-500"><Trash2 size={14} /></button>
                    </div>
                  </div>
                ))}
              </div>
            </Card>
          )}
        </>
      )}
      <GeneratedDocsCard />
    </DashboardShell>
  );
}
