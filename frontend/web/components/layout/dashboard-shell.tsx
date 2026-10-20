"use client";
import { Sidebar } from "./sidebar";
import { InstallButton } from "@/components/pwa/install-button";
import { Header } from "./header";
import { useAuthStore } from "@/lib/store";

export function DashboardShell({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((state) => state.user);
  const activeCompany = useAuthStore((state) => state.activeCompany);
  // El banner solo aplica a cuentas/empresas realmente marcadas sintéticas
  // (semilla del piloto). Antes estaba fijo sin condición — con el
  // registro real por Google activo (sprint "piloto con profesionales
  // reales"), mostrárselo a una cuenta real sería incorrecto, no solo
  // un detalle cosmético.
  const isSynthetic = Boolean(user?.isSynthetic || activeCompany?.is_synthetic);

  return (
    <div className="flex h-screen overflow-hidden bg-zu-bg">
      <Sidebar />
      <div className="flex flex-col flex-1 overflow-hidden">
        {isSynthetic && (
          <div role="alert" className="flex min-h-8 items-center justify-center gap-2 border-b border-amber-300 bg-amber-100 px-3 text-center text-[11px] font-bold tracking-wide text-amber-950">
            <span>ENTORNO SINTÉTICO — NO USAR DATOS REALES DE TRABAJADORES</span>
            <span className="hidden sm:inline font-normal">· empresas y registros marcados como datos de prueba</span>
          </div>
        )}
        <Header />
        <main className="flex-1 overflow-auto p-6 animate-fade-in">{children}</main>
      </div>
    </div>
  );
}
