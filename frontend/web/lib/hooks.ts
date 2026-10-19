"use client";
import { useQuery } from "@tanstack/react-query";
import { companyAPI, employeeAPI, incidentAPI, trainingAPI, type CompanyAPIRecord, type EmployeeAPIRecord, type IncidentAPIRecord, type TrainingAPIRecord } from "./api";

export type Company = CompanyAPIRecord;
export type Employee = EmployeeAPIRecord;
export type Incident = IncidentAPIRecord;
export type Training = TrainingAPIRecord;

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
