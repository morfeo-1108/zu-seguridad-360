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
  legal_rep_phone?: string;
  address?: string;
  city?: string;
  department?: string;
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

export interface RiskEntryAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  matrix_version: number;
  process: string;
  area: string;
  hazard_type: string;
  hazard_source: string;
  health_effect: string;
  workers_exposed: number;
  nd: number;
  ne: number;
  np: number;
  nc: number;
  nr: number;
  acceptability: "aceptable" | "no_aceptable" | "critico";
  current_control: string;
  created_at: string;
}

export interface RiskHeatmap {
  critical: RiskEntryAPIRecord[];
  not_acceptable: RiskEntryAPIRecord[];
  acceptable: RiskEntryAPIRecord[];
  summary: { total: number; critical: number; not_acceptable: number; acceptable: number };
}

export interface CreateRiskPayload {
  process: string;
  area: string;
  hazard_type: string;
  hazard_source: string;
  health_effect: string;
  workers_exposed: number;
  nd: number;
  ne: number;
  nc: number;
  current_control?: string;
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
  qr_code: string;
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
  googleComplete: (payload: { registration_token: string; company?: GoogleInitialCompany; terms_accepted: boolean; privacy_notice_acknowledged: boolean }) => api.post<LoginResponse>("/api/core/auth/google/complete", payload),
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

export const riskAPI = {
  list: (companyId: string) => api.get<{ data: RiskEntryAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/risks`),
  heatmap: (companyId: string) => api.get<RiskHeatmap>(`/api/core/companies/${companyId}/risks/heatmap`),
  create: (companyId: string, data: CreateRiskPayload) => api.post<RiskEntryAPIRecord>(`/api/core/companies/${companyId}/risks`, data),
};

export interface EPPItemAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  name: string;
  category: string;
  lifespan_days: number;
}

export interface EPPDeliveryAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  employee_id: string;
  item_id: string;
  quantity: number;
  condition: string;
  delivered_at: string;
  expires_at: string | null;
  confirmed_by: string | null;
  confirmed_at: string | null;
  returned_at: string | null;
  return_reason: string;
  item?: EPPItemAPIRecord;
}

export const eppAPI = {
  listItems: (companyId: string) => api.get<{ data: EPPItemAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/epp/items`),
  createItem: (companyId: string, data: { name: string; category: string; lifespan_days: number }) =>
    api.post<EPPItemAPIRecord>(`/api/core/companies/${companyId}/epp/items`, data),
  listExpiring: (companyId: string) => api.get<{ data: EPPDeliveryAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/epp/deliveries/expiring`),
  listByEmployee: (companyId: string, employeeId: string) =>
    api.get<{ data: EPPDeliveryAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/epp/deliveries/employee/${employeeId}`),
  createDelivery: (companyId: string, data: { employee_id: string; item_id: string; quantity: number; condition?: string }) =>
    api.post<EPPDeliveryAPIRecord>(`/api/core/companies/${companyId}/epp/deliveries`, data),
  returnDelivery: (companyId: string, deliveryId: string, reason: string) =>
    api.put<EPPDeliveryAPIRecord>(`/api/core/companies/${companyId}/epp/deliveries/${deliveryId}/return`, { reason }),
};

export interface ChecklistItem { code: string; description: string; compliant: boolean; notes?: string }
export interface ActionPlanItem { item_code: string; description: string; responsible: string; due_date: string | null; status: string }

export interface InspectionAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  title: string;
  area: string;
  inspector_id: string;
  performed_at: string;
  checklist: ChecklistItem[];
  compliance_pct: number;
  status: "completada" | "aprobada";
  approved_by: string | null;
  approved_at: string | null;
  action_plan: ActionPlanItem[];
  created_at: string;
}

export const inspectionAPI = {
  list: (companyId: string) => api.get<{ data: InspectionAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/inspections`),
  get: (companyId: string, id: string) => api.get<InspectionAPIRecord>(`/api/core/companies/${companyId}/inspections/${id}`),
  create: (companyId: string, data: { title: string; area: string; performed_at: string; checklist: ChecklistItem[] }) =>
    api.post<InspectionAPIRecord>(`/api/core/companies/${companyId}/inspections`, data),
  approve: (companyId: string, id: string) => api.put<InspectionAPIRecord>(`/api/core/companies/${companyId}/inspections/${id}/approve`, {}),
};

export interface ChecklistTemplateItem { code: string; description: string }
export interface ChecklistTemplateAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  name: string;
  category: string;
  items: ChecklistTemplateItem[];
  created_by: string;
  created_at: string;
}

export const checklistTemplateAPI = {
  list: (companyId: string) => api.get<{ data: ChecklistTemplateAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/checklist-templates`),
  create: (companyId: string, data: { name: string; category: string; items: ChecklistTemplateItem[] }) =>
    api.post<ChecklistTemplateAPIRecord>(`/api/core/companies/${companyId}/checklist-templates`, data),
  remove: (companyId: string, id: string) => api.delete(`/api/core/companies/${companyId}/checklist-templates/${id}`),
};

export interface CommitteeMember { employee_id: string; role: string; is_alternate: boolean }
export interface CommitteeAPIRecord {
  id: string; tenant_id: string; company_id: string;
  type: "copasst" | "vigia"; period_start: string; period_end: string; members: CommitteeMember[];
}
export interface CommitteeMeetingAPIRecord {
  id: string; committee_id: string; held_at: string; agenda: string; minutes: string;
  attendees: string[]; decisions: { topic: string; vote_for: number; vote_against: number; result: string }[];
  recorded_by: string;
}

export const committeeAPI = {
  list: (companyId: string) => api.get<{ data: CommitteeAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/committees`),
  create: (companyId: string, data: { type: string; period_start: string; period_end: string; members: CommitteeMember[] }) =>
    api.post<CommitteeAPIRecord>(`/api/core/companies/${companyId}/committees`, data),
  listMeetings: (companyId: string, committeeId: string) =>
    api.get<{ data: CommitteeMeetingAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/committees/${committeeId}/meetings`),
  createMeeting: (companyId: string, committeeId: string, data: { held_at: string; agenda: string; minutes: string; attendees: string[]; decisions: unknown[] }) =>
    api.post<CommitteeMeetingAPIRecord>(`/api/core/companies/${companyId}/committees/${committeeId}/meetings`, data),
};

export interface MonthlyIndicators {
  period: string;
  active_employees: number;
  standard_monthly_hours: number;
  estimated_hht: number;
  hht_is_estimated: boolean;
  work_accidents: number;
  lost_days: number;
  frequency_index: number;
  severity_index: number;
  formula_note: string;
  absenteeism_note: string;
}

export const indicatorAPI = {
  monthly: (companyId: string, year: number, month: number, standardHours = 220) =>
    api.get<MonthlyIndicators>(`/api/core/companies/${companyId}/indicators/monthly`, { params: { year, month, standard_hours: standardHours } }),
};

export interface LegalRequirementAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  norm_reference: string;
  title: string;
  applicable: boolean;
  compliance_status: "cumple" | "parcial" | "no_cumple" | "pendiente_evaluacion";
  evidence_ref: string;
  notes: string;
  reviewed_by: string | null;
  reviewed_at: string | null;
}

export const legalMatrixAPI = {
  list: (companyId: string) => api.get<{ data: LegalRequirementAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/legal-matrix`),
  create: (companyId: string, data: { norm_reference: string; title: string; applicable: boolean }) =>
    api.post<LegalRequirementAPIRecord>(`/api/core/companies/${companyId}/legal-matrix`, data),
  updateCompliance: (companyId: string, id: string, data: { compliance_status: string; evidence_ref?: string; notes?: string }) =>
    api.put<LegalRequirementAPIRecord>(`/api/core/companies/${companyId}/legal-matrix/${id}/compliance`, data),
};

export interface AlertAPIRecord {
  category: string;
  severity: "critical" | "warning";
  title: string;
  detail: string;
  due_at: string;
  reference_id: string;
}

export const alertAPI = {
  list: (companyId: string) => api.get<{ data: AlertAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/alerts`),
};

export interface PortfolioCompanyRecord {
  company_id: string;
  name: string;
  is_synthetic: boolean;
  active_employees: number;
  open_incidents: number;
  critical_risks: number;
  pending_legal_requirements: number;
}

export const portfolioAPI = {
  summary: () => api.get<{ data: PortfolioCompanyRecord[] }>(`/api/core/portfolio`),
};

export interface MedicalExamAPIRecord {
  id: string;
  employee_id: string;
  type: "ingreso" | "periodico" | "retiro";
  ips_name: string;
  scheduled_at: string | null;
  performed_at: string | null;
  expires_at: string | null;
  aptitude_result: "apto" | "apto_con_restriccion" | "no_apto" | "pendiente_revision";
  restrictions: string[];
  review_notes: string;
  created_at: string;
}

export const medicalAPI = {
  pendingReview: (companyId: string) => api.get<{ data: MedicalExamAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/medical-exams/pending-review`),
  expiring: (companyId: string) => api.get<{ data: MedicalExamAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/medical-exams/expiring`),
  byEmployee: (companyId: string, employeeId: string) => api.get<{ data: MedicalExamAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/medical-exams/employee/${employeeId}`),
  schedule: (companyId: string, data: { employee_id: string; type: string; ips_name: string; scheduled_at: string }) =>
    api.post<MedicalExamAPIRecord>(`/api/core/companies/${companyId}/medical-exams`, data),
  review: (companyId: string, id: string, data: { aptitude_result: string; restrictions?: string[]; review_notes?: string; expires_at?: string }) =>
    api.put<MedicalExamAPIRecord>(`/api/core/companies/${companyId}/medical-exams/${id}/review`, data),
};

export interface ComplianceStandardAPIRecord {
  id: string;
  framework: string;
  clause: string;
  title: string;
  phase: string;
  linked_module: string;
  status: string;
  system_assessment: string;
}

export interface ManagementSystemAPIRecord {
  id: string;
  company_id: string;
  year: number;
  current_phase: "planear" | "hacer" | "verificar" | "actuar";
  status: string;
  standards: ComplianceStandardAPIRecord[];
}

export const managementAPI = {
  getActive: (companyId: string) => api.get<ManagementSystemAPIRecord>(`/api/core/companies/${companyId}/phva/systems/active`),
  initSystem: (companyId: string, data: { year: number }) => api.post<ManagementSystemAPIRecord>(`/api/core/companies/${companyId}/phva/systems`, data),
  assess: (companyId: string, systemId: string) => api.post(`/api/core/companies/${companyId}/phva/systems/${systemId}/assess`, {}),
  score: (companyId: string, systemId: string) => api.get<{ compliance_score: number | null; validation_coverage: number; total_standards: number; human_validated: number; note: string }>(`/api/core/companies/${companyId}/phva/systems/${systemId}/score`),
  advancePhase: (companyId: string, systemId: string) => api.post<{ current_phase: string }>(`/api/core/companies/${companyId}/phva/systems/${systemId}/advance-phase`, {}),
  validateStandard: (companyId: string, standardId: string, status: string) =>
    api.put(`/api/core/companies/${companyId}/phva/standards/${standardId}/validate`, { status }),
};

export const trainingAPI = {
  list: (companyId: string) => api.get<{ data: TrainingAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/trainings`),
  create: (companyId: string, data: CreateTrainingPayload) => api.post<TrainingAPIRecord>(`/api/core/companies/${companyId}/trainings`, data),
  complete: (companyId: string, id: string, employeeIds: string[]) => api.post<{ simulation: true; certificates_created: number; certificates: SyntheticCertificateRecord[]; notice: string }>(`/api/core/companies/${companyId}/trainings/${id}/complete`, { employee_ids: employeeIds }),
};

export interface AttentionItemAPIRecord {
  id: string;
  tenant_id: string;
  company_id: string;
  company_name: string;
  type: "incident" | "medical" | "document";
  title: string;
  description: string;
  urgency: "critical" | "high" | "medium" | "low";
  priority: number;
  action_url: string;
  created_at: string;
  sla_breach_at: string | null;
}

export const attentionAPI = {
  inbox: () => api.get<{ data: AttentionItemAPIRecord[]; total: number }>(`/api/attention/inbox`),
  briefing: () => api.get<{ date: string; total_pending: number; critical_count: number; top_items: AttentionItemAPIRecord[]; generated_at: string }>(`/api/attention/briefing`),
  acknowledge: (itemId: string) => api.post(`/api/attention/inbox/${itemId}/acknowledge`, {}),
};

export interface ZoneChecklistItemAPIRecord {
  id: string;
  zone_id: string;
  code: string;
  description: string;
  type: string;
  mandatory: boolean;
  detection_class: string;
  norm_ref: string;
  severity: string;
  order: number;
}

export interface ZoneAPIRecord {
  id: string;
  project_id: string;
  name: string;
  type: string;
  norm_ref: string;
  risk_class: string;
  checklist: ZoneChecklistItemAPIRecord[];
}

export interface ProjectAPIRecord {
  id: string;
  company_id: string;
  name: string;
  address: string;
  status: string;
  zones: ZoneAPIRecord[];
}

export const zoneAPI = {
  listProjects: (companyId: string) => api.get<{ data: ProjectAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/projects`),
  createProject: (companyId: string, data: { name: string; address: string }) =>
    api.post<ProjectAPIRecord>(`/api/core/companies/${companyId}/projects`, data),
  createZone: (companyId: string, projectId: string, data: { name: string; type: string }) =>
    api.post<ZoneAPIRecord>(`/api/core/companies/${companyId}/projects/${projectId}/zones`, data),
};

export interface MediaUploadResponse {
  file_id: string;
  url: string;
  s3_key: string;
  size: number;
  mime_type: string;
  checksum_sha: string;
  worm: boolean;
}

export const mediaAPI = {
  upload: (file: File, prefix: string) => {
    const form = new FormData();
    form.append("file", file);
    form.append("prefix", prefix);
    return api.post<MediaUploadResponse>(`/api/media/upload`, form, { headers: { "Content-Type": "multipart/form-data" } });
  },
};

export interface DocumentAPIRecord {
  id: string;
  company_id: string;
  title: string;
  category: string;
  format: string;
  status: string;
  file_url: string;
  file_size: number;
  checksum_sha: string;
  created_at: string;
}

export const documentAPI = {
  list: (companyId: string, category?: string) =>
    api.get<{ data: DocumentAPIRecord[]; total: number }>(`/api/core/companies/${companyId}/documents${category ? `?category=${category}` : ""}`),
  create: (companyId: string, data: { title: string; category: string; format: string; file_url: string; file_size: number; checksum_sha: string }) =>
    api.post<DocumentAPIRecord>(`/api/core/companies/${companyId}/documents`, data),
  remove: (companyId: string, id: string) => api.delete(`/api/core/companies/${companyId}/documents/${id}`),
};

export interface PlanInfo {
  PriceMonthly: number;
  FounderPriceMonthly: number;
  MaxCompanies: number;
  MaxProfessionalUsers: number;
  MaxWorkers: number;
}

export interface SubscriptionAPIRecord {
  id: string;
  tenant_id: string;
  plan_code: string;
  price_monthly: number;
  max_companies: number;
  max_professional_users: number;
  max_workers: number;
  is_founder: boolean;
  founder_price_until: string | null;
  status: string;
  current_period_end: string;
}

export const billingAPI = {
  listPlans: () => api.get<{ plans: Record<string, PlanInfo> }>(`/api/billing/plans`),
  getSubscription: (tenantId: string) => api.get<SubscriptionAPIRecord>(`/api/billing/tenants/${tenantId}/subscription`),
  canAddCompany: (tenantId: string) => api.get<{ allowed: boolean; current_count: number; max_companies: number }>(`/api/billing/tenants/${tenantId}/can-add-company`),
  subscribe: (tenantId: string, planCode: string, wantsFounder: boolean) =>
    api.post<SubscriptionAPIRecord>(`/api/billing/tenants/${tenantId}/subscription`, { plan_code: planCode, wants_founder: wantsFounder }),
};
