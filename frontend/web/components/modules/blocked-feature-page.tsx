"use client";
import Link from "next/link";
import { ArrowLeft, LockKeyhole, ShieldAlert } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Card, PageHeader } from "@/components/ui";

export function BlockedFeaturePage({ title, reason }: { title: string; reason: string }) {
  return (
    <DashboardShell>
      <PageHeader title={title} description="Función no habilitada en el piloto local sintético." />
      <Card className="mx-auto mt-8 max-w-2xl border-amber-300 bg-amber-50 p-6">
        <div className="flex items-start gap-3">
          <LockKeyhole size={21} className="mt-0.5 shrink-0 text-amber-800" aria-hidden="true" />
          <div>
            <div className="mb-1 flex items-center gap-2 text-sm font-bold text-amber-950"><ShieldAlert size={15}/> MÓDULO BLOQUEADO</div>
            <p className="text-sm leading-relaxed text-amber-950">{reason}</p>
            <p className="mt-3 text-xs text-amber-900">No se envió información a servicios externos ni se simuló un resultado exitoso. El acceso sólo se habilitará después de implementar controles y pruebas específicos para este módulo.</p>
          </div>
        </div>
        <div className="mt-5"><Link href="/dashboard" className="inline-flex items-center gap-1.5 rounded-md border border-zu-border bg-white px-4 py-1.5 text-xs font-semibold text-zu-text1 transition-colors hover:bg-zu-surface2"><ArrowLeft size={13}/> Volver al panel</Link></div>
      </Card>
    </DashboardShell>
  );
}
