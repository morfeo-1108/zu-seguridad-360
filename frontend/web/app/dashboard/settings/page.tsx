"use client";
import { useEffect, useState } from "react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Button, Card, PageHeader, Spinner } from "@/components/ui";
import { paymentsAPI, type BillingProfileRecord } from "@/lib/api";
import { useAuthStore } from "@/lib/store";

const input = "w-full rounded-lg border border-zu-border bg-white px-3 py-2 text-[13px] focus:border-zu-accent focus:outline-none focus:ring-2 focus:ring-zu-accent/20";
const empty: BillingProfileRecord = { id_type: "13", id_number: "", legal_name: "", email: "", address: "", city: "" };

export default function SettingsPage() {
  const user = useAuthStore((s) => s.user);
  const tenantId = user?.tenantId || "";
  const [profile, setProfile] = useState<BillingProfileRecord>(empty);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<{ ok: boolean; text: string } | null>(null);

  useEffect(() => {
    if (!tenantId) return;
    paymentsAPI.getProfile(tenantId)
      .then((r) => setProfile({ ...empty, ...r.data }))
      .catch(() => setProfile({ ...empty, email: user?.email || "" }))
      .finally(() => setLoading(false));
  }, [tenantId, user?.email]);

  const save = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true); setMsg(null);
    try {
      const r = await paymentsAPI.saveProfile(tenantId, profile);
      setProfile({ ...empty, ...r.data });
      setMsg({ ok: true, text: r.data.dv ? `Guardado. Dígito de verificación calculado: ${r.data.dv}.` : "Guardado." });
    } catch (err: unknown) {
      const data = (err as { response?: { data?: unknown } })?.response?.data;
      const detail = typeof data === "string" ? data : (data as { message?: string } | undefined)?.message;
      setMsg({ ok: false, text: detail || "No se pudo guardar. Revise los campos." });
    } finally { setSaving(false); }
  };

  const set = (k: keyof BillingProfileRecord) => (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => setProfile({ ...profile, [k]: e.target.value });

  return (
    <DashboardShell>
      <PageHeader title="Configuración" description="Datos de su cuenta y datos para la factura electrónica de su suscripción." />
      <div className="grid gap-5 lg:grid-cols-[1fr_1.4fr]">
        <Card className="p-5">
          <h2 className="mb-3 text-sm font-bold">Cuenta</h2>
          <dl className="space-y-2 text-[13px]">
            <div><dt className="text-[11px] text-zu-text3">Nombre</dt><dd className="font-semibold">{user ? `${user.firstName} ${user.lastName}` : "—"}</dd></div>
            <div><dt className="text-[11px] text-zu-text3">Correo</dt><dd className="font-semibold">{user?.email || "—"}</dd></div>
            <div><dt className="text-[11px] text-zu-text3">Rol</dt><dd className="font-semibold">{user?.role || "—"}</dd></div>
          </dl>
          <p className="mt-4 text-[11px] leading-relaxed text-zu-text3">El nombre y el correo vienen de su cuenta de acceso; no se editan aquí.</p>
        </Card>
        <Card className="p-5">
          <h2 className="mb-1 text-sm font-bold">Perfil fiscal (adquirente)</h2>
          <p className="mb-4 text-[12px] leading-relaxed text-zu-text2">A nombre de quién se emite la factura electrónica de su suscripción. Sin estos datos la factura no puede emitirse.</p>
          {loading ? <Spinner size={18} /> : (
            <form onSubmit={(e) => void save(e)} className="grid gap-3 sm:grid-cols-2">
              <label className="text-[12px] font-semibold">Tipo de documento
                <select className={`${input} mt-1`} value={profile.id_type} onChange={set("id_type")}>
                  <option value="13">Cédula de ciudadanía</option>
                  <option value="31">NIT</option>
                </select>
              </label>
              <label className="text-[12px] font-semibold">Número {profile.id_type === "31" ? "(sin DV)" : ""}
                <input className={`${input} mt-1`} inputMode="numeric" pattern="[0-9]*" required value={profile.id_number} onChange={set("id_number")} />
              </label>
              <label className="text-[12px] font-semibold sm:col-span-2">{profile.id_type === "31" ? "Razón social" : "Nombre completo"}
                <input className={`${input} mt-1`} required value={profile.legal_name} onChange={set("legal_name")} />
              </label>
              <label className="text-[12px] font-semibold">Correo de facturación
                <input type="email" className={`${input} mt-1`} required value={profile.email} onChange={set("email")} />
              </label>
              <label className="text-[12px] font-semibold">Ciudad
                <input className={`${input} mt-1`} value={profile.city || ""} onChange={set("city")} />
              </label>
              <label className="text-[12px] font-semibold sm:col-span-2">Dirección
                <input className={`${input} mt-1`} value={profile.address || ""} onChange={set("address")} />
              </label>
              <div className="flex items-center gap-3 sm:col-span-2">
                <Button type="submit" disabled={saving}>{saving ? "Guardando…" : "Guardar perfil fiscal"}</Button>
                {msg && <p role="status" className={`text-[12px] ${msg.ok ? "text-emerald-700" : "text-red-700"}`}>{msg.text}</p>}
              </div>
            </form>
          )}
        </Card>
      </div>
    </DashboardShell>
  );
}
