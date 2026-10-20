"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import { Building, ChevronDown, Check, LogOut, ShieldAlert } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { authAPI } from "@/lib/api";
import { useAuthStore, type User } from "@/lib/store";

export function Header() {
  const [open, setOpen] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();
  const queryClient = useQueryClient();
  const { activeCompany, companies, user } = useAuthStore();

  const switchCompany = async (company: (typeof companies)[number]) => {
    if (company.id === activeCompany?.id || busy) return;
    setBusy(true);
    setError("");
    try {
      const response = await authAPI.switchCompany(company.id);
      const current = useAuthStore.getState();
      current.setTokens(response.data.access_token, response.data.refresh_token);
      current.setActiveCompany(company);
      if (current.user) current.setUser({ ...current.user, companyId: company.id } as User);
      queryClient.clear();
      setOpen(false);
    } catch {
      setError("No fue posible cambiar de empresa. La membresía se valida en el servidor.");
    } finally {
      setBusy(false);
    }
  };

  const logout = async () => {
    try { await authAPI.logout(); } catch { /* local sign-out must still complete */ }
    useAuthStore.getState().logout();
    queryClient.clear();
    router.replace("/auth/login");
  };

  return (
    <header className="h-14 border-b border-zu-border bg-white flex items-center px-5 gap-3.5 flex-shrink-0">
      <div className="flex items-center gap-2 pr-3.5 border-r border-zu-border">
        <ZULogo size={24} />
        <span className="hidden font-extrabold text-[13px] tracking-tight sm:inline">ZU <span className="text-zu-accent">360</span></span>
      </div>
      <div className="relative">
        <button aria-label="Seleccionar empresa activa" aria-expanded={open} onClick={() => setOpen(!open)} disabled={busy}
          className="flex items-center gap-2.5 px-3 py-[7px] rounded-lg border border-zu-border bg-zu-surface2 hover:bg-white transition-colors disabled:opacity-60">
          <span className="flex h-6 w-6 items-center justify-center rounded-md bg-zu-accent/10"><Building size={12} className="text-zu-accent" /></span>
          <div className="text-left"><div className="text-[12px] font-bold leading-none">{activeCompany?.name || "Sin empresa activa"}</div><div className="text-[9px] text-zu-text3 leading-none mt-0.5 font-medium">{activeCompany?.nit ? `NIT ${activeCompany.nit}` : "Memberships asignadas"}</div></div>
          <ChevronDown size={12} className="text-zu-text3" />
        </button>
        {open && <div className="absolute top-full mt-1.5 left-0 w-72 bg-white border border-zu-border rounded-xl shadow-zu-md z-50 py-1.5">
          <div className="px-3 py-2 text-[10px] font-bold text-zu-text3 tracking-widest">EMPRESAS AUTORIZADAS</div>
          {companies.map((company) => <button key={company.id} onClick={() => void switchCompany(company)} disabled={busy} className="w-full flex items-center gap-2 px-3 py-2 hover:bg-zu-surface2 transition-colors text-left disabled:opacity-50">
            <div className="flex-1 min-w-0"><div className="text-[12px] font-semibold truncate">{company.name}</div><div className="text-[10px] text-zu-text3">NIT {company.nit}</div></div>
            {activeCompany?.id === company.id && <Check size={13} className="text-zu-accent" />}
          </button>)}
          {companies.length === 0 && <p className="px-3 py-3 text-xs text-zu-text3">No hay memberships disponibles.</p>}
          {error && <p role="alert" className="px-3 py-2 text-[11px] text-red-700">{error}</p>}
        </div>}
      </div>
      {user?.isSynthetic && (
        <div className="hidden items-center gap-1.5 rounded-full bg-amber-50 px-2.5 py-1 text-[10.5px] font-semibold text-amber-800 sm:flex">
          <ShieldAlert size={11} /> Piloto sintético
        </div>
      )}
      <div className="flex-1" />
      <div className="flex items-center gap-3">
        <span className="hidden md:inline text-[11.5px] font-medium text-zu-text2">{user ? `${user.firstName} ${user.lastName}` : "Sesión verificada"}</span>
        {user?.isSynthetic && <span className="rounded-full bg-amber-50 border border-amber-200 px-2 py-0.5 text-[10px] font-bold text-amber-800">SINTÉTICO</span>}
        <div className="w-px h-4 bg-zu-border" />
        <button onClick={() => void logout()} className="inline-flex items-center gap-1.5 text-[11.5px] font-semibold text-zu-text2 hover:text-red-700 transition-colors" aria-label="Cerrar sesión"><LogOut size={14}/>Salir</button>
      </div>
    </header>
  );
}
