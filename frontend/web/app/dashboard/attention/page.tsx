"use client";
import { useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { AlertTriangle, CheckCircle2, Inbox } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { attentionAPI } from "@/lib/api";
import { useAttentionInbox, type AttentionItem } from "@/lib/hooks";
import { formatDateTime } from "@/lib/utils";

const urgencyColor: Record<string, "red" | "amber" | "blue" | "gray"> = { critical: "red", high: "amber", medium: "blue", low: "gray" };
const typeLabel: Record<string, string> = { incident: "Incidente", medical: "Examen médico", document: "Documento" };

export default function AttentionPage() {
  const queryClient = useQueryClient();
  const { data: items, isLoading, isError } = useAttentionInbox();
  const [busyId, setBusyId] = useState("");

  const acknowledge = async (item: AttentionItem) => {
    if (busyId) return;
    setBusyId(item.id);
    try {
      await attentionAPI.acknowledge(item.id);
      await queryClient.invalidateQueries({ queryKey: ["attention-inbox"] });
    } catch { /* el botón permanece visible para reintentar */ }
    finally { setBusyId(""); }
  };

  return (
    <DashboardShell>
      <PageHeader title="Bandeja de atención" description="Pendientes de toda tu cartera en un solo lugar, ordenados por urgencia — incidentes sin investigar, exámenes sin revisar, documentos sin firmar." />

      {isLoading ? <Spinner size={24} /> : isError ? (
        <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar la bandeja de atención.</p>
      ) : !items?.length ? (
        <Empty icon={CheckCircle2} title="Sin pendientes" description="No hay incidentes, exámenes ni documentos esperando acción en tu cartera." />
      ) : (
        <div className="space-y-2">
          {items.map((item) => (
            <Card key={item.id} className="flex items-center justify-between gap-3 p-4">
              <div className="flex min-w-0 items-start gap-3">
                {item.urgency === "critical" && <AlertTriangle size={16} className="mt-0.5 shrink-0 text-red-600" />}
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="truncate text-sm font-semibold">{item.title}</p>
                    <Badge variant={urgencyColor[item.urgency]}>{item.urgency}</Badge>
                  </div>
                  <p className="mt-0.5 truncate text-xs text-zu-text3">{item.company_name} · {typeLabel[item.type] || item.type} · {formatDateTime(item.created_at)}</p>
                  {item.description && <p className="mt-1 text-xs text-zu-text2">{item.description}</p>}
                </div>
              </div>
              <div className="flex shrink-0 items-center gap-2">
                <Link href={item.action_url} className="text-xs font-semibold text-zu-accent hover:underline">Ir al módulo</Link>
                <Button size="sm" variant="secondary" onClick={() => void acknowledge(item)} disabled={busyId === item.id}>
                  {busyId === item.id ? "…" : "Descartar"}
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}
      <div className="mt-4 flex items-center gap-2 text-[11px] text-zu-text3"><Inbox size={13} /> Esta bandeja cruza todas las empresas de tu cartera — el servidor filtra por tu tenant verificado en el token, no por un parámetro de la URL.</div>
    </DashboardShell>
  );
}
