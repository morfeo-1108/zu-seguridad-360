"use client";
import { FormEvent, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { AlertTriangle, HardHat, Plus, X } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner, Th, Td, Tr } from "@/components/ui";
import { eppAPI } from "@/lib/api";
import { useEmployees, useEPPExpiring, useEPPItems } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";
import { formatDateTime } from "@/lib/utils";

const freshItemForm = () => ({ name: "", category: "", lifespan_days: 365 });
const freshDeliveryForm = () => ({ employee_id: "", item_id: "", quantity: 1, condition: "nuevo" });

export default function EPPPage() {
  const queryClient = useQueryClient();
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: items, isLoading: itemsLoading, isError: itemsError } = useEPPItems(companyId);
  const { data: expiring, isLoading: expiringLoading } = useEPPExpiring(companyId);
  const { data: employeeResponse } = useEmployees(companyId);
  const employees = employeeResponse?.data || [];

  const [showItemForm, setShowItemForm] = useState(false);
  const [itemForm, setItemForm] = useState(freshItemForm);
  const [showDeliveryForm, setShowDeliveryForm] = useState(false);
  const [deliveryForm, setDeliveryForm] = useState(freshDeliveryForm);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["epp-items", companyId] });
    queryClient.invalidateQueries({ queryKey: ["epp-expiring", companyId] });
  };

  const createItem = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy) return;
    setBusy(true); setError("");
    try {
      await eppAPI.createItem(companyId, { ...itemForm, lifespan_days: Number(itemForm.lifespan_days) });
      await invalidate(); setShowItemForm(false); setItemForm(freshItemForm());
    } catch { setError("No se pudo crear el elemento de protección."); }
    finally { setBusy(false); }
  };

  const createDelivery = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!companyId || busy || !deliveryForm.employee_id || !deliveryForm.item_id) return;
    setBusy(true); setError("");
    try {
      await eppAPI.createDelivery(companyId, { ...deliveryForm, quantity: Number(deliveryForm.quantity) });
      await invalidate(); setShowDeliveryForm(false); setDeliveryForm(freshDeliveryForm());
    } catch { setError("No se pudo registrar la entrega. Verifica que el trabajador y el elemento pertenezcan a la empresa activa."); }
    finally { setBusy(false); }
  };

  return (
    <DashboardShell>
      <PageHeader
        title="Elementos de protección personal (EPP)"
        description="Catálogo por empresa y entregas por trabajador, con vencimiento calculado por periodicidad."
        action={
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => { setItemForm(freshItemForm()); setError(""); setShowItemForm(true); }} disabled={!companyId}><Plus size={13} /> Elemento</Button>
            <Button onClick={() => { setDeliveryForm(freshDeliveryForm()); setError(""); setShowDeliveryForm(true); }} disabled={!companyId || !items?.length}><Plus size={13} /> Entrega</Button>
          </div>
        }
      />

      {!companyId ? <Empty icon={AlertTriangle} title="Sin empresa activa" description="Selecciona una empresa con membership desde el encabezado." /> : (
        <>
          <Card className="mb-5 p-4">
            <h3 className="mb-3 text-[12px] font-bold text-zu-text1">Vencen en los próximos 30 días</h3>
            {expiringLoading ? <Spinner size={18} /> : !expiring?.length ? (
              <p className="text-[12px] text-zu-text3">Sin renovaciones próximas.</p>
            ) : (
              <div className="space-y-1.5">
                {expiring.map((d) => (
                  <div key={d.id} className="flex items-center justify-between rounded-md bg-amber-50 px-3 py-1.5 text-[11px]">
                    <span className="font-semibold text-amber-900">{d.item?.name || "Elemento"}</span>
                    <Badge variant="amber">Vence {d.expires_at ? formatDateTime(d.expires_at) : "—"}</Badge>
                  </div>
                ))}
              </div>
            )}
          </Card>

          <h3 className="mb-3 text-[12px] font-bold text-zu-text1">Catálogo de la empresa activa</h3>
          {itemsLoading ? <Spinner size={24} /> : itemsError ? (
            <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se pudo cargar el catálogo EPP.</p>
          ) : !items?.length ? (
            <Empty icon={HardHat} title="Sin elementos en el catálogo" description="Registra el primer elemento de protección para esta empresa." />
          ) : (
            <Card className="overflow-hidden p-0">
              <div className="overflow-x-auto">
                <table className="w-full border-collapse">
                  <thead><tr>{["Nombre", "Categoría", "Periodicidad de renovación"].map((l) => <Th key={l}>{l}</Th>)}</tr></thead>
                  <tbody>
                    {items.map((item) => (
                      <Tr key={item.id}>
                        <Td className="text-[12px] font-semibold">{item.name}</Td>
                        <Td className="text-[12px]">{item.category}</Td>
                        <Td className="text-[12px]">{item.lifespan_days} días</Td>
                      </Tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </Card>
          )}
        </>
      )}

      {showItemForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nuevo elemento de protección</h2><button onClick={() => setShowItemForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createItem} className="space-y-3">
              <label className="block text-xs">Nombre<input required value={itemForm.name} onChange={(e) => setItemForm({ ...itemForm, name: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Categoría<input required value={itemForm.category} onChange={(e) => setItemForm({ ...itemForm, category: e.target.value })} placeholder="Alturas, visual, respiratoria, manos…" className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <label className="block text-xs">Periodicidad de renovación (días)<input required type="number" min={1} value={itemForm.lifespan_days} onChange={(e) => setItemForm({ ...itemForm, lifespan_days: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowItemForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Crear elemento"}</Button></div>
            </form>
          </Card>
        </div>
      )}

      {showDeliveryForm && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/40 p-4">
          <Card className="w-full max-w-md p-5">
            <div className="mb-3 flex items-center justify-between"><h2 className="text-sm font-bold">Nueva entrega</h2><button onClick={() => setShowDeliveryForm(false)} aria-label="Cerrar"><X size={18} /></button></div>
            <form onSubmit={createDelivery} className="space-y-3">
              <label className="block text-xs">Trabajador
                <select required value={deliveryForm.employee_id} onChange={(e) => setDeliveryForm({ ...deliveryForm, employee_id: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  <option value="">Selecciona…</option>
                  {employees.map((emp) => <option key={emp.id} value={emp.id}>{emp.first_name} {emp.last_name}</option>)}
                </select>
              </label>
              <label className="block text-xs">Elemento
                <select required value={deliveryForm.item_id} onChange={(e) => setDeliveryForm({ ...deliveryForm, item_id: e.target.value })} className="mt-1 w-full rounded border border-zu-border px-3 py-2">
                  <option value="">Selecciona…</option>
                  {(items || []).map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}
                </select>
              </label>
              <label className="block text-xs">Cantidad<input required type="number" min={1} value={deliveryForm.quantity} onChange={(e) => setDeliveryForm({ ...deliveryForm, quantity: Number(e.target.value) })} className="mt-1 w-full rounded border border-zu-border px-3 py-2" /></label>
              <p className="text-[10px] text-zu-text3">La confirmación queda registrada con tu usuario y la hora actual; no es una firma digital capturada.</p>
              {error && <p role="alert" className="text-xs text-red-700">{error}</p>}
              <div className="flex justify-end gap-2 pt-1"><Button variant="secondary" onClick={() => setShowDeliveryForm(false)}>Cancelar</Button><Button type="submit" disabled={busy}>{busy ? "Guardando…" : "Confirmar entrega"}</Button></div>
            </form>
          </Card>
        </div>
      )}
    </DashboardShell>
  );
}
