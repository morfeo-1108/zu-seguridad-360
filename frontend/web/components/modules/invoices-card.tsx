"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { CreditCard, FileText, RefreshCcw } from "lucide-react";
import { Badge, Button, Card, Spinner } from "@/components/ui";
import { paymentsAPI, type ElectronicInvoiceRecord, type InvoiceRecord } from "@/lib/api";

const cop = (c: number) => `$${Math.round(c / 100).toLocaleString("es-CO")}`;
const statusBadge: Record<InvoiceRecord["status"], { label: string; v: "amber" | "green" | "red" | "gray" }> = {
  pending: { label: "Pendiente de pago", v: "amber" }, paid: { label: "Pagada", v: "green" },
  overdue: { label: "Vencida", v: "red" }, voided: { label: "Anulada", v: "gray" },
};
const errText = (e: unknown) => {
  const d = (e as { response?: { data?: unknown; status?: number } })?.response;
  return typeof d?.data === "string" && d.data ? d.data : d?.status === 503 ? "Los pagos en línea aún no están habilitados." : "No fue posible completar la operación.";
};

// Invoices of the tenant: issue the current period, pay it through Wompi Web
// Checkout (card data never touches ZU 360), and see the electronic invoice.
export function InvoicesCard({ tenantId, hasSubscription }: { tenantId: string; hasSubscription: boolean }) {
  const [invoices, setInvoices] = useState<InvoiceRecord[] | null>(null);
  const [einv, setEinv] = useState<Record<string, ElectronicInvoiceRecord>>({});
  const [busy, setBusy] = useState("");
  const [msg, setMsg] = useState("");

  const load = async () => {
    const r = await paymentsAPI.invoices(tenantId);
    setInvoices(r.data.data);
    for (const inv of r.data.data.filter((i) => i.status === "paid")) {
      paymentsAPI.electronic(tenantId, inv.id).then((e) => setEinv((m) => ({ ...m, [inv.id]: e.data }))).catch(() => {});
    }
  };
  useEffect(() => { if (tenantId) void load().catch(() => setInvoices([])); }, [tenantId]); // eslint-disable-line react-hooks/exhaustive-deps

  const issue = async () => {
    setBusy("issue"); setMsg("");
    try { await paymentsAPI.issue(tenantId); await load(); } catch (e) { setMsg(errText(e)); } finally { setBusy(""); }
  };
  const pay = async (id: string) => {
    setBusy(id); setMsg("");
    try { window.location.assign((await paymentsAPI.checkout(tenantId, id)).data.checkout_url); }
    catch (e) { setMsg(errText(e)); setBusy(""); }
  };
  const retry = async (id: string) => {
    setBusy(`e-${id}`);
    try { const e = await paymentsAPI.retryElectronic(tenantId, id); setEinv((m) => ({ ...m, [id]: e.data })); } catch (e) { setMsg(errText(e)); } finally { setBusy(""); }
  };

  return (
    <Card className="mb-5 p-4">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <h2 className="text-sm font-bold">Facturas</h2>
        {hasSubscription && <Button size="sm" variant="secondary" onClick={() => void issue()} disabled={busy === "issue"}><FileText size={12} /> Generar factura del periodo</Button>}
      </div>
      {invoices === null ? <Spinner size={18} /> : invoices.length === 0 ? (
        <p className="text-xs text-zu-text3">{hasSubscription ? "Aún no hay facturas." : "Elija un plan para generar su primera factura."}</p>
      ) : (
        <ul className="divide-y divide-zu-border">
          {invoices.map((inv) => {
            const e = einv[inv.id];
            return (
              <li key={inv.id} className="flex flex-wrap items-center justify-between gap-3 py-3">
                <div>
                  <p className="text-[13px] font-semibold">{inv.dian_number} · {new Date(inv.period_start).toLocaleDateString("es-CO", { month: "long", year: "numeric" })}</p>
                  <p className="text-[11px] text-zu-text3">Subtotal {cop(inv.subtotal)} + IVA {cop(inv.iva)} = <b>{cop(inv.total)}</b> · vence {new Date(inv.due_at).toLocaleDateString("es-CO")}</p>
                  {e && (
                    <p className="mt-1 text-[11px] text-zu-text2">
                      Factura electrónica {e.number || ""}: <b>{e.status.replaceAll("_", " ")}</b>{e.environment === "2" ? " (ambiente de pruebas DIAN)" : ""}.{" "}
                      {e.status === "sin_perfil_fiscal" && <><Link href="/dashboard/settings" className="font-semibold text-zu-accent hover:underline">Complete su perfil fiscal</Link> y <button className="font-semibold text-zu-accent hover:underline" onClick={() => void retry(inv.id)} disabled={busy === `e-${inv.id}`}>reintente <RefreshCcw size={10} className="inline" /></button>.</>}
                    </p>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Badge variant={statusBadge[inv.status].v}>{statusBadge[inv.status].label}</Badge>
                  {(inv.status === "pending" || inv.status === "overdue") && (
                    <Button size="sm" onClick={() => void pay(inv.id)} disabled={busy === inv.id}><CreditCard size={12} /> {busy === inv.id ? "Abriendo…" : "Pagar con Wompi"}</Button>
                  )}
                </div>
              </li>
            );
          })}
        </ul>
      )}
      {msg && <p role="alert" className="mt-2 text-xs text-red-700">{msg}</p>}
    </Card>
  );
}
