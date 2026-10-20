"use client";
import { useQuery } from "@tanstack/react-query";
import { alertAPI, committeeAPI, companyAPI, employeeAPI, eppAPI, incidentAPI, inspectionAPI, legalMatrixAPI, portfolioAPI, riskAPI, trainingAPI, type AlertAPIRecord, type CommitteeAPIRecord, type CompanyAPIRecord, type EmployeeAPIRecord, type EPPItemAPIRecord, type IncidentAPIRecord, type InspectionAPIRecord, type LegalRequirementAPIRecord, type PortfolioCompanyRecord, type RiskEntryAPIRecord, type TrainingAPIRecord } from "./api";

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
