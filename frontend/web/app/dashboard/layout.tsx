"use client";
import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { authAPI, companyAPI } from "@/lib/api";
import { useAuthStore, type CompanyContext, type User } from "@/lib/store";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const [ready, setReady] = useState(false);
  const [error, setError] = useState("");

  const verifySession = async () => {
    setError("");
    const token = typeof window !== "undefined" ? localStorage.getItem("zu_token") : null;
    const refreshToken = typeof window !== "undefined" ? localStorage.getItem("zu_refresh") : null;
    if (!token || !refreshToken) {
      useAuthStore.getState().logout();
      router.replace("/auth/login");
      return;
    }
    try {
      const [profileResponse, companiesResponse] = await Promise.all([authAPI.me(), companyAPI.list()]);
      const profile = profileResponse.data;
      const companies = companiesResponse.data.data || [];
      if (!profile.id) throw new Error("No hay una sesión activa disponible.");
      if (companies.length === 0) {
        // New account: everything else is company-scoped, so the professional
        // starts by registering their first company.
        const user: User = {
          id: profile.id, email: profile.email, role: profile.role,
          firstName: profile.first_name, lastName: profile.last_name,
          tenantId: profile.tenant_id, companyId: "", isSynthetic: profile.is_synthetic,
        };
        useAuthStore.getState().setSession(user, token, refreshToken, [], null);
        if (pathname !== "/dashboard/companies") router.replace("/dashboard/companies");
        setReady(true);
        return;
      }
      let active = companies.find((company) => company.id === profile.company_id);
      let accessToken = token;
      let currentRefresh = refreshToken;
      if (!active) {
        const first = companies[0];
        const switched = await authAPI.switchCompany(first.id);
        accessToken = switched.data.access_token;
        currentRefresh = switched.data.refresh_token;
        useAuthStore.getState().setTokens(accessToken, currentRefresh);
        active = first;
      }
      const user: User = {
        id: profile.id,
        email: profile.email,
        role: profile.role,
        firstName: profile.first_name,
        lastName: profile.last_name,
        tenantId: profile.tenant_id,
        companyId: active.id,
        isSynthetic: profile.is_synthetic,
      };
      useAuthStore.getState().setSession(user, accessToken, currentRefresh, companies as CompanyContext[], active as CompanyContext);
      setReady(true);
    } catch (cause) {
      const status = (cause as { response?: { status?: number } })?.response?.status;
      if (status === 401 || status === 403) {
        useAuthStore.getState().logout();
        router.replace("/auth/login");
      } else {
        setError("No se pudo verificar tu sesión. Revisa tu conexión e inténtalo de nuevo.");
        setReady(false);
      }
    }
  };

  useEffect(() => { void verifySession(); }, []);

  // Without a company every other screen is unusable; keep the professional on
  // the companies page until the first one is registered.
  const companyCount = useAuthStore((state) => state.companies.length);
  const needsFirstCompany = ready && companyCount === 0 && pathname !== "/dashboard/companies";
  useEffect(() => {
    if (needsFirstCompany) router.replace("/dashboard/companies");
  }, [needsFirstCompany, router]);

  if (!ready || needsFirstCompany) {
    return <div className="min-h-screen flex items-center justify-center bg-zu-bg p-6">
      <div className="max-w-md text-center text-sm text-zu-text2">
        {error ? <><p role="alert" className="mb-3">{error}</p><button onClick={() => void verifySession()} className="rounded-md bg-zu-accent px-4 py-2 font-semibold text-white">Reintentar</button></> : "Verificando sesión con Core…"}
      </div>
    </div>;
  }
  return <>{children}</>;
}
