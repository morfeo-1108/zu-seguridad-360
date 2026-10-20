"use client";
import { useParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { CheckCircle2, ShieldAlert, XCircle, FlaskConical } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import api from "@/lib/api";

interface CertificateVerification {
  folio: string;
  employee_name: string;
  training_title: string;
  category: string;
  issued_at: string;
  expires_at?: string;
  is_valid: boolean;
  is_synthetic: boolean;
  revoked_at?: string;
  validity_notice: string;
}

export default function VerifyCertificatePage() {
  const params = useParams<{ folio: string }>();
  const folio = params.folio;

  const { data, isLoading, isError } = useQuery({
    queryKey: ["public-verify", folio],
    queryFn: async () => (await api.get<{ found: boolean; certificate: CertificateVerification }>(`/api/core/public/certificates/${folio}/verify`)).data.certificate,
    enabled: !!folio,
    retry: 1,
  });

  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-[#0B0D10] px-4 py-10 text-white">
      <div className="mb-6 flex items-center gap-2.5">
        <ZULogo size={34} dark />
        <span className="text-sm font-bold tracking-tight">ZU <span className="text-lime-400">360</span> · Verificación pública</span>
      </div>

      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-white/[0.04] p-6 backdrop-blur">
        {isLoading ? (
          <p className="py-8 text-center text-sm text-white/60">Verificando folio…</p>
        ) : isError || !data ? (
          <div className="flex flex-col items-center gap-3 py-6 text-center">
            <XCircle size={40} className="text-red-400" />
            <h1 className="text-base font-bold">Folio no encontrado</h1>
            <p className="text-sm text-white/60">El código <span className="font-mono">{folio}</span> no corresponde a ningún certificado emitido por ZU360.</p>
          </div>
        ) : (
          <div>
            <div className="mb-4 flex flex-col items-center gap-2 text-center">
              {data.is_synthetic ? (
                <FlaskConical size={40} className="text-amber-400" />
              ) : data.is_valid ? (
                <CheckCircle2 size={40} className="text-lime-400" />
              ) : (
                <ShieldAlert size={40} className="text-red-400" />
              )}
              <h1 className="text-lg font-bold">{data.training_title}</h1>
              <p className="text-sm text-white/70">{data.employee_name}</p>
            </div>

            <div className={`mb-4 rounded-lg px-3 py-2.5 text-xs leading-relaxed ${data.is_synthetic ? "bg-amber-400/10 text-amber-200" : data.is_valid ? "bg-lime-400/10 text-lime-200" : "bg-red-400/10 text-red-200"}`}>
              {data.validity_notice}
            </div>

            <dl className="space-y-2 text-xs">
              <div className="flex justify-between border-b border-white/10 pb-2"><dt className="text-white/50">Folio</dt><dd className="font-mono">{data.folio}</dd></div>
              <div className="flex justify-between border-b border-white/10 pb-2"><dt className="text-white/50">Categoría</dt><dd>{data.category || "—"}</dd></div>
              <div className="flex justify-between border-b border-white/10 pb-2"><dt className="text-white/50">Emitido</dt><dd>{new Date(data.issued_at).toLocaleDateString("es-CO")}</dd></div>
              {data.expires_at && <div className="flex justify-between border-b border-white/10 pb-2"><dt className="text-white/50">Vence</dt><dd>{new Date(data.expires_at).toLocaleDateString("es-CO")}</dd></div>}
              {data.revoked_at && <div className="flex justify-between border-b border-white/10 pb-2"><dt className="text-white/50">Revocado</dt><dd>{new Date(data.revoked_at).toLocaleDateString("es-CO")}</dd></div>}
            </dl>
          </div>
        )}
      </div>
      <p className="mt-6 max-w-md text-center text-[11px] text-white/35">Esta verificación consulta directamente la base de datos de ZU360 por folio; no depende de la palabra del portador del certificado.</p>
    </main>
  );
}
