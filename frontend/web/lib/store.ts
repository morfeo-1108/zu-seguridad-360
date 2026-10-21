import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface User {
  id: string;
  email: string;
  role: string;
  firstName: string;
  lastName: string;
  tenantId: string;
  companyId: string;
  isSynthetic: boolean;
}

export interface CompanyContext {
  id: string;
  name: string;
  nit: string;
  status?: string;
  is_synthetic?: boolean;
}

interface AuthStore {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  activeCompany: CompanyContext | null;
  companies: CompanyContext[];
  isAuthenticated: boolean;
  setTokens: (token: string, refreshToken?: string) => void;
  setSession: (user: User, token: string, refreshToken: string, companies: CompanyContext[], activeCompany: CompanyContext | null) => void;
  setUser: (user: User) => void;
  setCompanies: (companies: CompanyContext[]) => void;
  setActiveCompany: (company: CompanyContext) => void;
  logout: () => void;
}

function storeTokens(token: string, refreshToken?: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem("zu_token", token);
  if (refreshToken) localStorage.setItem("zu_refresh", refreshToken);
}

function clearZUTokens() {
  if (typeof window === "undefined") return;
  localStorage.removeItem("zu_token");
  localStorage.removeItem("zu_refresh");
  localStorage.removeItem("zu_tenant");
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      refreshToken: null,
      activeCompany: null,
      companies: [],
      isAuthenticated: false,
      setTokens: (token, refreshToken) => {
        storeTokens(token, refreshToken);
        set((state) => ({ token, refreshToken: refreshToken ?? state.refreshToken, isAuthenticated: true }));
      },
      setSession: (user, token, refreshToken, companies, activeCompany) => {
        storeTokens(token, refreshToken);
        set({ user, token, refreshToken, companies, activeCompany, isAuthenticated: true });
      },
      setUser: (user) => set({ user }),
      setCompanies: (companies) => set({ companies }),
      setActiveCompany: (activeCompany) => set({ activeCompany }),
      logout: () => {
        clearZUTokens();
        set({ user: null, token: null, refreshToken: null, activeCompany: null, companies: [], isAuthenticated: false });
      },
    }),
    {
      name: "zu-auth",
      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
        activeCompany: state.activeCompany,
        companies: state.companies,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);

interface Project { id: string; name: string; company_name: string; zones_count: number; status: string; }
interface ProjectStore {
  activeProject: Project | null;
  projects: Project[];
  setActiveProject: (p: Project | null) => void;
  setProjects: (ps: Project[]) => void;
}

export const useProjectStore = create<ProjectStore>()(
  persist(
    (set) => ({
      activeProject: null,
      projects: [],
      setActiveProject: (activeProject) => set({ activeProject }),
      setProjects: (projects) => set({ projects }),
    }),
    { name: "zu-active-project" }
  )
);
