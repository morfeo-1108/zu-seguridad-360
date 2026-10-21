"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { authAPI, companyAPI, type GoogleSignupPublicConfig, type LoginResponse } from "@/lib/api";
import { useAuthStore, type CompanyContext, type User } from "@/lib/store";

declare global {
  interface Window {
    google?: {
      accounts: {
        id: {
          initialize: (options: {
            client_id: string;
            nonce: string;
            callback: (response: { credential?: string }) => void;
            auto_select?: boolean;
            use_fedcm_for_prompt?: boolean;
          }) => void;
          renderButton: (parent: HTMLElement, options: Record<string, string | number | boolean>) => void;
          cancel: () => void;
        };
      };
    };
  }
}

interface Props {
  config: GoogleSignupPublicConfig | null;
}

function loadGoogleIdentity(): Promise<void> {
  if (window.google?.accounts.id) return Promise.resolve();
  return new Promise((resolve, reject) => {
    const existing = document.getElementById("google-identity-services") as HTMLScriptElement | null;
    if (existing) {
      existing.addEventListener("load", () => resolve(), { once: true });
      existing.addEventListener("error", () => reject(new Error("google_script_unavailable")), { once: true });
      return;
    }
    const script = document.createElement("script");
    script.id = "google-identity-services";
    script.src = "https://accounts.google.com/gsi/client";
    script.async = true;
    script.defer = true;
    script.onload = () => resolve();
    script.onerror = () => reject(new Error("google_script_unavailable"));
    document.head.appendChild(script);
  });
}

export default function GoogleSignup({ config }: Props) {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);
  const nonceRef = useRef("");
  const setupRef = useRef<(() => Promise<void>) | null>(null);
  const [registration, setRegistration] = useState<{ token: string; email: string; firstName: string; lastName: string } | null>(null);
  const [termsAccepted, setTermsAccepted] = useState(false);
  const [privacyAcknowledged, setPrivacyAcknowledged] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");

  const finishSession = useCallback(async (tokens: LoginResponse) => {
    if (!tokens.access_token || !tokens.refresh_token) throw new Error("invalid_google_session");
    useAuthStore.getState().logout();
    useAuthStore.getState().setTokens(tokens.access_token, tokens.refresh_token);
    const [profileResponse, companyResponse] = await Promise.all([authAPI.me(), companyAPI.list()]);
    const profile = profileResponse.data;
    const companies = companyResponse.data.data || [];
    if (!profile.id) throw new Error("session_unavailable");
    // A new account has no company yet: it is registered from the dashboard.
    const active = companies.find((item) => item.id === profile.company_id) ?? null;
    if (profile.company_id && !active) throw new Error("company_membership_unavailable");
    const user: User = {
      id: profile.id,
      email: profile.email,
      role: profile.role,
      firstName: profile.first_name,
      lastName: profile.last_name,
      tenantId: profile.tenant_id,
      companyId: active?.id ?? "",
      isSynthetic: profile.is_synthetic,
    };
    useAuthStore.getState().setSession(user, tokens.access_token, tokens.refresh_token, companies as CompanyContext[], active as CompanyContext | null);
    router.replace(companies.length === 0 ? "/dashboard/companies" : "/dashboard");
  }, [router]);

  const setupButton = useCallback(async () => {
    if (!config?.enabled || !config.client_id || !containerRef.current) return;
    setError("");
    try {
      const [challenge] = await Promise.all([authAPI.googleChallenge(), loadGoogleIdentity()]);
      const nonce = challenge.data.nonce;
      if (!nonce || !window.google?.accounts.id || !containerRef.current) throw new Error("google_setup_unavailable");
      nonceRef.current = nonce;
      const parent = containerRef.current;
      parent.replaceChildren();
      window.google.accounts.id.initialize({
        client_id: config.client_id,
        nonce,
        auto_select: false,
        callback: async ({ credential }) => {
          if (!credential) {
            setError("Google no devolvió una credencial válida. Inténtalo de nuevo.");
            void setupRef.current?.();
            return;
          }
          setBusy(true);
          setError("");
          try {
            const response = await authAPI.googleStart(credential, nonceRef.current);
            const payload = response.data;
            if ("registration_required" in payload && payload.registration_required) {
              setRegistration({
                token: payload.registration_token,
                email: payload.profile.email,
                firstName: payload.profile.first_name,
                lastName: payload.profile.last_name,
              });
            } else {
              await finishSession(payload as LoginResponse);
            }
          } catch (cause) {
            const status = (cause as { response?: { status?: number } })?.response?.status;
            setError(status === 409
              ? "Este correo ya está asociado a una cuenta. Inicia sesión con el método existente o contacta soporte."
              : status === 403
                ? "La cuenta no está habilitada. Contacta soporte."
                : "No se pudo verificar el inicio con Google. Vuelve a intentarlo.");
            void setupRef.current?.();
          } finally {
            setBusy(false);
          }
        },
      });
      window.google.accounts.id.renderButton(parent, {
        type: "standard", theme: "outline", size: "large", text: "continue_with",
        shape: "rectangular", logo_alignment: "left", width: 360, locale: "es",
      });
    } catch {
      setError("El acceso con Google no está disponible en este momento.");
    }
  }, [config, finishSession]);

  useEffect(() => {
    setupRef.current = setupButton;
    void setupButton();
    return () => { setupRef.current = null; };
  }, [setupButton]);

  const submitRegistration = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!registration || !termsAccepted || !privacyAcknowledged) {
      setError("Debes aceptar los términos y confirmar que leíste el aviso de privacidad.");
      return;
    }
    setBusy(true);
    setError("");
    try {
      const response = await authAPI.googleComplete({
        registration_token: registration.token,
        terms_accepted: termsAccepted,
        privacy_notice_acknowledged: privacyAcknowledged,
      });
      await finishSession(response.data);
    } catch (cause) {
      const status = (cause as { response?: { status?: number } })?.response?.status;
      if (status === 401 || status === 409) {
        const message = status === 409
          ? "Este correo ya tiene una cuenta. Inicia nuevamente o accede con la cuenta ya registrada."
          : "La validación de Google expiró. Inicia el proceso otra vez.";
        setRegistration(null);
        setTermsAccepted(false);
        setPrivacyAcknowledged(false);
        void setupRef.current?.().then(() => setError(message));
      } else {
        setError("No se pudo completar el registro. Inténtalo nuevamente.");
      }
    } finally {
      setBusy(false);
    }
  };

  const cancelRegistration = () => {
    setRegistration(null);
    setTermsAccepted(false);
    setPrivacyAcknowledged(false);
    setError("");
    void setupRef.current?.();
  };

  if (!config?.enabled || !config.client_id || !config.terms_url || !config.privacy_url) return null;

  if (registration) {
    return (
      <section className="mb-6 rounded-xl border border-zu-border bg-white p-4" aria-labelledby="google-registration-title">
        <h3 id="google-registration-title" className="text-[15px] font-bold text-zu-text1">Crea tu cuenta</h3>
        <p className="mt-1 text-[12px] text-zu-text2">Google verificó <b>{registration.firstName} {registration.lastName}</b> ({registration.email}). Después de crear la cuenta registras tus empresas desde el panel.</p>
        <form onSubmit={submitRegistration} className="mt-4 space-y-3">
          <label className="flex items-start gap-2 text-[12px] text-zu-text2">
            <input type="checkbox" checked={termsAccepted} onChange={(event) => setTermsAccepted(event.target.checked)} className="mt-0.5" />
            <span>Acepto los <a href={config.terms_url} target="_blank" rel="noreferrer" className="font-semibold text-zu-accent underline">Términos del servicio</a> (versión {config.terms_version || "configurada"}).</span>
          </label>
          <label className="flex items-start gap-2 text-[12px] text-zu-text2">
            <input type="checkbox" checked={privacyAcknowledged} onChange={(event) => setPrivacyAcknowledged(event.target.checked)} className="mt-0.5" />
            <span>Confirmo que leí el <a href={config.privacy_url} target="_blank" rel="noreferrer" className="font-semibold text-zu-accent underline">Aviso de privacidad</a> (versión {config.privacy_version || "configurada"}).</span>
          </label>
          {error && <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-[12px] text-red-700">{error}</div>}
          <div className="flex gap-2 pt-1">
            <button type="button" onClick={cancelRegistration} disabled={busy} className="flex-1 rounded-lg border border-zu-border py-2.5 text-[12px] font-semibold text-zu-text2 disabled:opacity-50">Volver</button>
            <button type="submit" disabled={busy || !termsAccepted || !privacyAcknowledged} className="flex-1 rounded-lg bg-zu-accent py-2.5 text-[12px] font-semibold text-white disabled:opacity-50">{busy ? "Creando cuenta…" : "Crear cuenta"}</button>
          </div>
        </form>
      </section>
    );
  }

  return (
    <section className="mb-5">
      <div ref={containerRef} className="flex min-h-10 justify-center" aria-label="Acceder o registrarse con Google" />
      {busy && <p className="mt-2 text-center text-[11px] text-white/40">Verificando cuenta…</p>}
      {error && <div role="alert" className="mt-3 rounded-md border border-red-500/25 bg-red-500/[0.08] px-3 py-2 text-[12px] text-red-300">{error}</div>}
      <div className="my-5 flex items-center gap-3 text-[10px] uppercase tracking-wider text-white/35"><span className="h-px flex-1 bg-white/10" /><span>o continúa con tu correo</span><span className="h-px flex-1 bg-white/10" /></div>
    </section>
  );
}

