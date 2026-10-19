import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "./store";

interface RetriableRequest extends InternalAxiosRequestConfig {
  _zuRetried?: boolean;
}

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || "",
  headers: { "Content-Type": "application/json" },
  timeout: 30000,
});

api.interceptors.request.use((config) => {
  if (typeof window !== "undefined") {
    const token = localStorage.getItem("zu_token");
    if (token) config.headers.Authorization = `Bearer ${token}`;
    else delete config.headers.Authorization;
    // Tenant/company scope must come only from verified JWT claims, never headers.
    delete config.headers["X-Tenant-ID"];
    delete config.headers["X-Company-ID"];
  }
  return config;
});

let refreshInFlight: Promise<string> | null = null;
api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetriableRequest | undefined;
    const url = original?.url || "";
    const isAuthRoute = url.includes("/auth/");
    if (error.response?.status !== 401 || !original || original._zuRetried || isAuthRoute || typeof window === "undefined") {
      return Promise.reject(error);
    }

    const refresh = localStorage.getItem("zu_refresh");
    if (!refresh) return Promise.reject(error);
    original._zuRetried = true;

    try {
      if (!refreshInFlight) {
        refreshInFlight = axios.post("/api/core/auth/refresh", { refresh_token: refresh }, { timeout: 15000 })
          .then((response) => {
            const accessToken = response.data?.access_token as string | undefined;
            const refreshToken = response.data?.refresh_token as string | undefined;
            if (!accessToken) throw new Error("invalid refresh response");
            useAuthStore.getState().setTokens(accessToken, refreshToken);
            return accessToken;
          })
          .finally(() => { refreshInFlight = null; });
      }
      const newAccessToken = await refreshInFlight;
      original.headers.Authorization = `Bearer ${newAccessToken}`;
      return api(original);
    } catch {
      useAuthStore.getState().logout();
      if (!window.location.pathname.startsWith("/auth/login")) window.location.assign("/auth/login?expired=1");
      return Promise.reject(error);
    }
  }
);

export default api;

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: "Bearer";
  expires_in: number;
}

export interface CurrentUserResponse {
  id: string;
  email: string;
  role: string;
  first_name: string;
  last_name: string;
  tenant_id: string;
  company_id: string;
  is_synthetic: boolean;
}

export interface CompanyAPIRecord {
  id: string;
  name: string;
  nit: string;
  ciiu?: string;
  risk_class?: string;
  arl?: string;
  legal_rep?: string;
  legal_rep_email?: string;
  employee_count?: number;
  risk_score?: number;
  status?: string;
  compliance?: number;
  plan_id?: string;
  is_synthetic?: boolean;
}

export interface EmployeeAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  is_synthetic: boolean;
  first_name: string;
  last_name: string;
  document_type: string;
  document_number: string;
  email: string;
  position: string;
  department: string;
  status: string;
}

export interface IncidentAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  is_synthetic: boolean;
  type: string;
  severity: string;
  status: string;
  occurred_at: string;
  location: string;
  description: string;
  injured_count: number;
  legal_hold: boolean;
  closed_at?: string | null;
  closed_by?: string | null;
}

export interface CreateIncidentPayload {
  type: "incidente";
  occurred_at: string;
  location: string;
  description: string;
  injured_count: 0;
  employee_id?: string;
}

export interface TrainingAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  is_synthetic: boolean;
  title: string;
  norm_ref: string;
  category: string;
  duration_hours: number;
  validity_months: number;
  instructor_name: string;
  scheduled_at?: string | null;
  completed_at?: string | null;
  status: string;
  attendee_count: number;
}

export interface CreateTrainingPayload {
  title: string;
  norm_ref: "";
  category: string;
  duration_hours: number;
  validity_months: number;
  instructor_name: string;
  scheduled_at: string;
  evaluation_min_score: 0;
}

export interface SyntheticCertificateRecord {
  id: string;
  folio: string;
  employee_id: string;
  training_id: string;
  company_id: string;
  is_synthetic: boolean;
  score: number;
  issued_at: string;
  expires_at: string;
  is_valid: boolean;
}

export interface FurATPreview {
  status: "simulated_preview";
  synthetic: true;
  official: false;
  notice: string;
}

export interface GoogleSignupPublicConfig {
  enabled: boolean;
  synthetic_pilot: boolean;
  client_id?: string;
  terms_url?: string;
  privacy_url?: string;
  terms_version?: string;
  privacy_version?: string;
}

export interface GoogleInitialCompany {
  name: string;
  nit: string;
  ciiu: string;
  arl: string;
  legal_rep: string;
  legal_rep_email: string;
}

export interface GoogleRegistrationStart {
  registration_required: true;
  registration_token: string;
  profile: { email: string; first_name: string; last_name: string };
}

export const authAPI = {
  login: (email: string, password: string) => api.post<LoginResponse>("/api/core/auth/login", { email, password }),
  googleConfig: () => api.get<GoogleSignupPublicConfig>("/api/core/auth/google/config"),
  googleChallenge: () => api.post<{ nonce: string }>("/api/core/auth/google/challenge", {}),
  googleStart: (credential: string, nonce: string) => api.post<LoginResponse | GoogleRegistrationStart>("/api/core/auth/google/start", { credential, nonce }),
  googleComplete: (payload: { registration_token: string; company: GoogleInitialCompany; terms_accepted: boolean; privacy_notice_acknowledged: boolean }) => api.post<LoginResponse>("/api/core/auth/google/complete", payload),
  me: () => api.get<CurrentUserResponse>("/api/core/auth/me"),
  switchCompany: (companyId: string) => api.post<LoginResponse & { active_company_id: string }>("/api/core/auth/switch-company", { company_id: companyId }),
  logout: () => api.post("/api/core/auth/logout"),
};

export const companyAPI = {
  list: () => api.get<{ data: CompanyAPIRecord[]; total: number }>("/api/core/companies"),
  get: (id: string) => api.get<CompanyAPIRecord>(`/api/core/companies/${id}`),
  create: (data: unknown) => api.post<CompanyAPIRecord>("/api/core/companies", data),
  update: (id: string, data: unknown) => api.put<CompanyAPIRecord>(`/api/core/companies/${id}`, data),
  delete: (id: string) => api.delete(`/api/core/companies/${id}`),
  getRiskScore: (id: string) => api.get(`/api/core/companies/${id}/score`),
};

export const employeeAPI = {
  list: (companyId: string, page = 1, limit = 100) => api.get<{ data: EmployeeAPIRecord[]; total: number; page: number; limit: number }>(`/api/core/companies/${companyId}/employees`, { params: { page, limit } }),
  get: (companyId: string, employeeId: string) => api.get<EmployeeAPIRecord>(`/api/core/companies/${companyId}/employees/${employeeId}`),
  create: (companyId: string, data: unknown) => api.post<EmployeeAPIRecord>(`/api/core/companies/${companyId}/employees`, data),
  update: (companyId: string, employeeId: string, data: unknown) => api.put<EmployeeAPIRecord>(`/api/core/companies/${companyId}/employees/${employeeId}`, data),
  delete: (companyId: string, employeeId: string) => api.delete(`/api/core/companies/${companyId}/employees/${employeeId}`),
};

export const incidentAPI = {
  list: (companyId: string) => api.get<{ data: IncidentAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/incidents`),
  get: (companyId: string, id: string) => api.get<IncidentAPIRecord>(`/api/core/companies/${companyId}/incidents/${id}`),
  create: (companyId: string, data: CreateIncidentPayload) => api.post<IncidentAPIRecord>(`/api/core/companies/${companyId}/incidents`, data),
  investigate: (companyId: string, id: string) => api.put<IncidentAPIRecord>(`/api/core/companies/${companyId}/incidents/${id}/investigate`, {}),
  close: (companyId: string, id: string) => api.put<IncidentAPIRecord>(`/api/core/companies/${companyId}/incidents/${id}/close`, {}),
  furatPreview: (companyId: string, id: string) => api.post<FurATPreview>(`/api/core/companies/${companyId}/incidents/${id}/furat`, {}),
};

export const trainingAPI = {
  list: (companyId: string) => api.get<{ data: TrainingAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/trainings`),
  create: (companyId: string, data: CreateTrainingPayload) => api.post<TrainingAPIRecord>(`/api/core/companies/${companyId}/trainings`, data),
  complete: (companyId: string, id: string, employeeIds: string[]) => api.post<{ simulation: true; certificates_created: number; certificates: SyntheticCertificateRecord[]; notice: string }>(`/api/core/companies/${companyId}/trainings/${id}/complete`, { employee_ids: employeeIds }),
};
