"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ShieldCheck, Users, HardHat, TrendingUp, Leaf } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";
import { authAPI, companyAPI, type GoogleSignupPublicConfig } from "@/lib/api";
import { useAuthStore, type CompanyContext, type User } from "@/lib/store";
import GoogleSignup from "./google-signup";

const PILLARS = [
  { icon: ShieldCheck, top: "PREVENCIÓN", bottom: "QUE PROTEGE" },
  { icon: Users, top: "PERSONAS", bottom: "QUE IMPORTAN" },
  { icon: HardHat, top: "CULTURA", bottom: "QUE TRANSFORMA" },
  { icon: TrendingUp, top: "GESTIÓN", bottom: "QUE GENERA VALOR" },
  { icon: Leaf, top: "SOSTENIBILIDAD", bottom: "QUE TRASCIENDE" },
];

export default function LoginPage() {
  const router = useRouter();
  const setSession = useAuthStore((state) => state.setSession);
  const logout = useAuthStore((state) => state.logout);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [googleConfig, setGoogleConfig] = useState<GoogleSignupPublicConfig | null>(null);

  useEffect(() => {
    let active = true;
    authAPI.googleConfig().then(({ data }) => {
      if (active) setGoogleConfig(data);
    }).catch(() => {
      if (active) setGoogleConfig(null);
    });
    return () => { active = false; };
  }, []);

  const handleLogin = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!email.trim() || !password) {
      setError("Ingresa el correo y la contraseña.");
      return;
    }
    setLoading(true);
    setError("");
    logout();
    try {
      const loginResponse = await authAPI.login(email.trim(), password);
      const { access_token: token, refresh_token: refreshToken } = loginResponse.data;
      if (!token || !refreshToken) throw new Error("invalid_login_response");
      useAuthStore.getState().setTokens(token, refreshToken);

      const [profileResponse, companyResponse] = await Promise.all([authAPI.me(), companyAPI.list()]);
      const profile = profileResponse.data;
      const companies = companyResponse.data.data || [];
      if (!profile.id || !profile.company_id || companies.length === 0) throw new Error("session_unavailable");
      const active = companies.find((company) => company.id === profile.company_id);
      if (!active) throw new Error("company_membership_unavailable");
      const user: User = {
        id: profile.id, email: profile.email, role: profile.role,
        firstName: profile.first_name, lastName: profile.last_name,
        tenantId: profile.tenant_id, companyId: profile.company_id,
        isSynthetic: profile.is_synthetic,
      };
      setSession(user, token, refreshToken, companies as CompanyContext[], active as CompanyContext);
      router.replace("/dashboard");
    } catch (cause) {
      logout();
      const status = (cause as { response?: { status?: number } })?.response?.status;
      setError(status === 403 ? "La cuenta no tiene acceso habilitado." : status === 401 ? "Correo o contraseña incorrectos." : "No fue posible iniciar sesión. Revisa la conexión o contacta al administrador.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex">
      <div className="hidden lg:flex flex-col justify-between w-[54%] relative overflow-hidden p-14"
        style={{ background: "linear-gradient(155deg, #12171F 0%, #1A2129 55%, #0E1218 100%)" }}>
        <div className="absolute top-[10%] right-[8%] w-[380px] h-[380px] rounded-full" style={{ background: "radial-gradient(circle, rgba(139,195,74,0.22) 0%, transparent 60%)" }} />
        <div className="absolute inset-0 opacity-[0.04]" style={{ backgroundImage: "radial-gradient(circle, #fff 1px, transparent 1px)", backgroundSize: "28px 28px" }} />
        <div className="relative z-10"><ZULogo size={44} dark showText /></div>
        <div className="relative z-10">
          <div className="w-16 h-[3px] rounded-full mb-6" style={{ background: "linear-gradient(90deg,#8BC34A,transparent)" }} />
          <h1 className="text-white text-[38px] font-extrabold leading-[1.1] tracking-tight mb-4">Prevenimos hoy,<br /><span style={{ color: "#8BC34A" }}>protegemos siempre.</span></h1>
          <p className="text-[15px] leading-relaxed max-w-[420px]" style={{ color: "#9BA3AE" }}>Gestión SST con aislamiento por empresa y herramientas de apoyo para profesionales.</p>
        </div>
        <div className="relative z-10 grid grid-cols-5 gap-3 pt-8 border-t" style={{ borderColor: "rgba(255,255,255,0.08)" }}>
          {PILLARS.map((p) => <div key={p.top} className="flex flex-col items-center text-center"><div className="w-10 h-10 rounded-xl flex items-center justify-center mb-2" style={{ background: "rgba(139,195,74,0.12)", border: "1px solid rgba(139,195,74,0.25)" }}><p.icon size={17} style={{ color: "#8BC34A" }} /></div><div className="text-white text-[9.5px] font-bold tracking-wide leading-tight">{p.top}</div><div className="text-[8px] font-semibold tracking-wide mt-0.5" style={{ color: "#8BC34A" }}>{p.bottom}</div></div>)}
        </div>
      </div>
      <div className="flex-1 flex items-center justify-center p-8 bg-zu-bg">
        <div className="w-full max-w-[380px]">
          <div className="lg:hidden mb-8 flex justify-center"><ZULogo size={48} showText /></div>
          {(googleConfig === null || googleConfig.synthetic_pilot) && <div className="mb-4 rounded-lg border border-amber-300 bg-amber-50 px-3 py-2 text-[11px] font-semibold text-amber-900">{googleConfig === null ? "ENTORNO NO VERIFICADO — no ingreses datos personales hasta confirmar la configuración." : "PILOTO SINTÉTICO — no ingresar datos reales de trabajadores."}</div>}
          <h2 className="text-[26px] font-extrabold text-zu-text1 tracking-tight mb-1">Bienvenido</h2>
          <p className="text-[13px] text-zu-text2 mb-8">Ingresa a tu plataforma de gestión SST</p>
          <GoogleSignup config={googleConfig} />
          {typeof window !== "undefined" && new URLSearchParams(window.location.search).has("expired") && <div role="status" className="mb-3 rounded-md bg-amber-50 p-2 text-[12px] text-amber-900">Tu sesión expiró. Vuelve a iniciar sesión.</div>}
          <form onSubmit={handleLogin} className="space-y-4">
            <div><label htmlFor="email" className="text-[12px] font-semibold text-zu-text2 mb-1.5 block">Correo electrónico</label><input id="email" type="email" autoComplete="username" required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="nombre@empresa.com" className="w-full px-4 py-2.5 rounded-lg border border-zu-border bg-white text-[13px] outline-none focus:border-zu-accent transition" /></div>
            <div><label htmlFor="password" className="text-[12px] font-semibold text-zu-text2 mb-1.5 block">Contraseña</label><input id="password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Tu contraseña" className="w-full px-4 py-2.5 rounded-lg border border-zu-border bg-white text-[13px] outline-none focus:border-zu-accent transition" /></div>
            {error && <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700">{error}</div>}
            <button type="submit" disabled={loading} className="w-full py-2.5 rounded-lg text-white font-semibold text-[13px] transition hover:opacity-90 disabled:opacity-60" style={{ background: "linear-gradient(135deg, #8BC34A, #6BA31A)" }}>{loading ? "Validando…" : "Ingresar a ZU 360"}</button>
          </form>
          <p className="text-[11px] text-zu-text3 text-center mt-8">ZU Seguridad Integral 360 · Colombia{googleConfig?.synthetic_pilot ? " · piloto sintético" : ""}</p>
        </div>
      </div>
    </div>
  );
}
