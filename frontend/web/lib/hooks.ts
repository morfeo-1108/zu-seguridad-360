"use client";
import { useQuery } from "@tanstack/react-query";
import { alertAPI, attentionAPI, billingAPI, briefingAPI, notificationAPI, onboardingAPI, slaAPI, checklistTemplateAPI, committeeAPI, companyAPI, documentAPI, employeeAPI, eppAPI, incidentAPI, inspectionAPI, legalMatrixAPI, managementAPI, medicalAPI, portfolioAPI, riskAPI, trainingAPI, zoneAPI, type AlertAPIRecord, type AttentionItemAPIRecord, type ChecklistTemplateAPIRecord, type CommitteeAPIRecord, type CompanyAPIRecord, type DocumentAPIRecord, type EmployeeAPIRecord, type EPPItemAPIRecord, type IncidentAPIRecord, type InspectionAPIRecord, type LegalRequirementAPIRecord, type ManagementSystemAPIRecord, type MedicalExamAPIRecord, type PlanInfo, type PortfolioCompanyRecord, type ProjectAPIRecord, type RiskEntryAPIRecord, type SubscriptionAPIRecord, type TrainingAPIRecord } from "./api";

export type Company = CompanyAPIRecord;
export type Employee = EmployeeAPIRecord;
export type Incident = IncidentAPIRecord;
export type Training = TrainingAPIRecord;
export type RiskEntry = RiskEntryAPIRecord;

export function useCompanies() {
  return useQuery({ queryKey: ["companies"], queryFn: async () => (await companyAPI.list()).data.data, staleTime: 30_000, retry: 1 });
}

export function useCompany(id: string) {
  return useQuery({ queryKey: ["company", id], queryFn: async () => (await companyAPI.get(id)).data, enabled: !!id, retry: 1 });
}

export function useEmployees(companyId: string) {
  return useQuery({ queryKey: ["employees", companyId], queryFn: async () => (await employeeAPI.list(companyId)).data, enabled: !!companyId, retry: 1 });
}

export function useEmployee(companyId: string, employeeId: string) {
  return useQuery({ queryKey: ["employee", companyId, employeeId], queryFn: async () => (await employeeAPI.get(companyId, employeeId)).data, enabled: !!companyId && !!employeeId, retry: 1 });
}

export function useCompanyIncidents(companyId: string) {
  return useQuery({ queryKey: ["incidents", companyId], queryFn: async () => (await incidentAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useCompanyTrainings(companyId: string) {
  return useQuery({ queryKey: ["trainings", companyId], queryFn: async () => (await trainingAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useCompanyRisks(companyId: string) {
  return useQuery({ queryKey: ["risks", companyId], queryFn: async () => (await riskAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useRiskHeatmap(companyId: string) {
  return useQuery({ queryKey: ["risks-heatmap", companyId], queryFn: async () => (await riskAPI.heatmap(companyId)).data, enabled: !!companyId, retry: 1 });
}

export function useEPPItems(companyId: string) {
  return useQuery({ queryKey: ["epp-items", companyId], queryFn: async () => (await eppAPI.listItems(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useEPPExpiring(companyId: string) {
  return useQuery({ queryKey: ["epp-expiring", companyId], queryFn: async () => (await eppAPI.listExpiring(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type EPPItem = EPPItemAPIRecord;
export type Inspection = InspectionAPIRecord;

export function useCompanyInspections(companyId: string) {
  return useQuery({ queryKey: ["inspections", companyId], queryFn: async () => (await inspectionAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type ChecklistTemplate = ChecklistTemplateAPIRecord;

export function useChecklistTemplates(companyId: string) {
  return useQuery({ queryKey: ["checklist-templates", companyId], queryFn: async () => (await checklistTemplateAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type MedicalExam = MedicalExamAPIRecord;

export function useMedicalPendingReview(companyId: string) {
  return useQuery({ queryKey: ["medical-pending", companyId], queryFn: async () => (await medicalAPI.pendingReview(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useMedicalExpiring(companyId: string) {
  return useQuery({ queryKey: ["medical-expiring", companyId], queryFn: async () => (await medicalAPI.expiring(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useMedicalByEmployee(companyId: string, employeeId: string) {
  return useQuery({ queryKey: ["medical-employee", companyId, employeeId], queryFn: async () => (await medicalAPI.byEmployee(companyId, employeeId)).data.data, enabled: !!companyId && !!employeeId, retry: 1 });
}

export type ManagementSystem = ManagementSystemAPIRecord;

export function useActiveManagementSystem(companyId: string) {
  return useQuery({ queryKey: ["management-active", companyId], queryFn: async () => (await managementAPI.getActive(companyId)).data, enabled: !!companyId, retry: false });
}

export type Committee = CommitteeAPIRecord;

export function useCompanyCommittees(companyId: string) {
  return useQuery({ queryKey: ["committees", companyId], queryFn: async () => (await committeeAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export function useCommitteeMeetings(companyId: string, committeeId: string) {
  return useQuery({ queryKey: ["committee-meetings", companyId, committeeId], queryFn: async () => (await committeeAPI.listMeetings(companyId, committeeId)).data.data, enabled: !!companyId && !!committeeId, retry: 1 });
}

export type LegalRequirement = LegalRequirementAPIRecord;

export function useLegalMatrix(companyId: string) {
  return useQuery({ queryKey: ["legal-matrix", companyId], queryFn: async () => (await legalMatrixAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type AppAlert = AlertAPIRecord;

export function useAlerts(companyId: string) {
  return useQuery({ queryKey: ["alerts", companyId], queryFn: async () => (await alertAPI.list(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type PortfolioCompany = PortfolioCompanyRecord;

export function usePortfolio() {
  return useQuery({ queryKey: ["portfolio"], queryFn: async () => (await portfolioAPI.summary()).data.data, staleTime: 30_000, retry: 1 });
}

export type AttentionItem = AttentionItemAPIRecord;

export function useAttentionInbox() {
  return useQuery({ queryKey: ["attention-inbox"], queryFn: async () => (await attentionAPI.inbox()).data.data, staleTime: 15_000, retry: 1 });
}

export type Project = ProjectAPIRecord;

export function useProjects(companyId: string) {
  return useQuery({ queryKey: ["projects", companyId], queryFn: async () => (await zoneAPI.listProjects(companyId)).data.data, enabled: !!companyId, retry: 1 });
}

export type ZuDocument = DocumentAPIRecord;

export function useDocuments(companyId: string, category?: string) {
  return useQuery({ queryKey: ["documents", companyId, category], queryFn: async () => (await documentAPI.list(companyId, category)).data.data, enabled: !!companyId, retry: 1 });
}

export type Plan = PlanInfo;
export type Subscription = SubscriptionAPIRecord;

export function usePlans() {
  return useQuery({ queryKey: ["billing-plans"], queryFn: async () => (await billingAPI.listPlans()).data.plans, staleTime: 5 * 60_000, retry: 1 });
}

export function useSubscription(tenantId: string) {
  return useQuery({
    queryKey: ["subscription", tenantId], enabled: !!tenantId, retry: false,
    queryFn: async () => {
      try { return (await billingAPI.getSubscription(tenantId)).data; }
      catch { return null; }
    },
  });
}

export function useNotifications() {
  return useQuery({ queryKey: ["notifications"], queryFn: async () => (await notificationAPI.list()).data, refetchInterval: 60_000, retry: 1 });
}

export function useSLASummary() {
  return useQuery({ queryKey: ["sla-summary"], queryFn: async () => (await slaAPI.summary()).data, retry: 1 });
}

export function useOnboardingProgress(companyId: string) {
  return useQuery({ queryKey: ["onboarding", companyId], queryFn: async () => (await onboardingAPI.progress()).data, enabled: !!companyId, retry: 1 });
}

export function useBriefing() {
  return useQuery({
    queryKey: ["briefing"], retry: false,
    queryFn: async () => {
      try { return (await briefingAPI.today()).data; }
      catch { return (await briefingAPI.refresh()).data; } // first visit: build it now
    },
  });
}
