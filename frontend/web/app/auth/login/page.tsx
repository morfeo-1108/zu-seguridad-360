"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ShieldCheck, Users, HardHat, TrendingUp, Leaf, ArrowLeft } from "lucide-react";
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
    <div className="min-h-screen flex bg-[#0B0F0C]">
      {/* ── Panel de marca ── */}
      <div className="hidden lg:flex flex-col justify-between w-[52%] relative overflow-hidden p-14">
        <div className="absolute inset-0 bg-gradient-to-br from-[#0E1410] via-[#0B0F0C] to-[#0B0F0C]" />
        <div className="absolute -left-32 top-[8%] h-[420px] w-[420px] animate-float-slow rounded-full bg-zu-lime/10 blur-[110px]" />
        <div className="absolute -right-10 bottom-[5%] h-[360px] w-[360px] animate-float-slower rounded-full bg-emerald-500/10 blur-[110px]" />
        <div className="absolute inset-0 opacity-[0.04]" style={{ backgroundImage: "radial-gradient(circle, #fff 1px, transparent 1px)", backgroundSize: "28px 28px" }} />

        <div className="relative z-10 flex items-center justify-between">
          <ZULogo size={40} dark showText />
          <Link href="/" className="inline-flex items-center gap-1.5 text-[12px] font-semibold text-white/50 hover:text-white transition">
            <ArrowLeft size={14} /> Volver al inicio
          </Link>
        </div>

        <div className="relative z-10 animate-hero-in">
          <div className="w-16 h-[3px] rounded-full mb-6 bg-gradient-to-r from-zu-lime to-transparent" />
          <h1 className="text-white text-[38px] font-extrabold leading-[1.1] tracking-tight mb-4">
            Prevenimos hoy,
            <br />
            <span className="bg-gradient-to-r from-zu-lime to-emerald-400 bg-clip-text text-transparent">protegemos siempre.</span>
          </h1>
          <p className="text-[15px] leading-relaxed max-w-[420px] text-white/55">
            Gestión SST con aislamiento real por empresa y control de acceso verificado en
            cada solicitud, no solo en esta pantalla.
          </p>
        </div>

        <div className="relative z-10 grid grid-cols-5 gap-3 pt-8 border-t border-white/[0.08]">
          {PILLARS.map((p) => (
            <div key={p.top} className="flex flex-col items-center text-center">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center mb-2 bg-zu-lime/10 border border-zu-lime/25">
                <p.icon size={17} className="text-zu-lime" />
              </div>
              <div className="text-white text-[9.5px] font-bold tracking-wide leading-tight">{p.top}</div>
              <div className="text-[8px] font-semibold tracking-wide mt-0.5 text-zu-lime">{p.bottom}</div>
            </div>
          ))}
        </div>
      </div>

      {/* ── Formulario ── */}
      <div className="flex-1 flex items-center justify-center p-8">
        <div className="w-full max-w-[380px] animate-fade-in">
          <div className="lg:hidden mb-8 flex items-center justify-between">
            <ZULogo size={44} dark showText />
            <Link href="/" className="text-[12px] font-semibold text-white/50 hover:text-white transition">← Inicio</Link>
          </div>
          {(googleConfig === null || googleConfig.synthetic_pilot) && (
            <div className="mb-4 rounded-lg border border-amber-500/25 bg-amber-500/[0.06] px-3 py-2 text-[11px] font-semibold text-amber-300">
              {googleConfig === null ? "ENTORNO NO VERIFICADO — no ingreses datos personales hasta confirmar la configuración." : "PILOTO SINTÉTICO — no ingresar datos reales de trabajadores."}
            </div>
          )}
          <h2 className="text-[26px] font-extrabold text-white tracking-tight mb-1">Bienvenido de nuevo</h2>
          <p className="text-[13px] text-white/50 mb-8">Ingresa a tu plataforma de gestión SST</p>
          <GoogleSignup config={googleConfig} />
          {typeof window !== "undefined" && new URLSearchParams(window.location.search).has("expired") && (
            <div role="status" className="mb-3 rounded-md bg-amber-500/[0.08] border border-amber-500/20 p-2 text-[12px] text-amber-300">Tu sesión expiró. Vuelve a iniciar sesión.</div>
          )}
          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label htmlFor="email" className="text-[12px] font-semibold text-white/60 mb-1.5 block">Correo electrónico</label>
              <input id="email" type="email" autoComplete="username" required value={email} onChange={(e) => setEmail(e.target.value)} placeholder="nombre@empresa.com"
                className="w-full px-4 py-2.5 rounded-lg border border-white/10 bg-white/[0.04] text-white text-[13px] outline-none focus:border-zu-lime transition placeholder:text-white/30" />
            </div>
            <div>
              <label htmlFor="password" className="text-[12px] font-semibold text-white/60 mb-1.5 block">Contraseña</label>
              <input id="password" type="password" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Tu contraseña"
                className="w-full px-4 py-2.5 rounded-lg border border-white/10 bg-white/[0.04] text-white text-[13px] outline-none focus:border-zu-lime transition placeholder:text-white/30" />
            </div>
            {error && <div role="alert" className="rounded-md border border-red-500/25 bg-red-500/[0.08] px-3 py-2 text-[12px] text-red-300">{error}</div>}
            <button type="submit" disabled={loading}
              className="w-full py-2.5 rounded-lg text-[#0B0F0C] font-bold text-[13px] transition hover:bg-white disabled:opacity-60 bg-zu-lime">
              {loading ? "Validando…" : "Ingresar a ZU 360"}
            </button>
          </form>
          <p className="text-[11px] text-white/35 text-center mt-8">ZU Seguridad Integral 360 · Colombia{googleConfig?.synthetic_pilot ? " · piloto sintético" : ""}</p>
        </div>
      </div>
    </div>
  );
}
