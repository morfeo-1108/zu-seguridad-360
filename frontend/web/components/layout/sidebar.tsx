"use client";
import Link from "next/link";
import { ZULogo } from "@/components/brand/logo";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/lib/store";
import { LayoutDashboard, Building2, Users, GraduationCap, Stethoscope, ShieldAlert, AlertTriangle, FileText, Cpu, Plane, Inbox, Bot, DollarSign, Settings, ChevronRight, Layers, Network, RefreshCw, HardHat, ClipboardCheck, Gavel, BarChart3, Scale, Briefcase, Sunrise } from "lucide-react";

const NAV_GROUPS = [
  { label: null, items: [
    { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { href: "/dashboard/portfolio", label: "Portafolio", icon: Briefcase },
  ] },
  { label: "GESTIÓN", items: [
    { href: "/dashboard/companies", label: "Empresas", icon: Building2 },
    { href: "/dashboard/employees", label: "Empleados", icon: Users },
  ]},
  { label: "SISTEMA DE GESTIÓN", items: [
    { href: "/dashboard/management", label: "Ciclo PHVA & ISO", icon: RefreshCw },
  ] },
  { label: "SST OPERATIVO", items: [
    { href: "/dashboard/training", label: "Capacitaciones", icon: GraduationCap },
    { href: "/dashboard/medical", label: "Exámenes Médicos", icon: Stethoscope },
    { href: "/dashboard/risks", label: "Riesgos GTC 45", icon: ShieldAlert },
    { href: "/dashboard/epp", label: "EPP", icon: HardHat },
    { href: "/dashboard/inspections", label: "Inspecciones", icon: ClipboardCheck },
    { href: "/dashboard/committees", label: "COPASST / Vigía", icon: Gavel },
    { href: "/dashboard/indicators", label: "Indicadores IF/IS", icon: BarChart3 },
    { href: "/dashboard/legal-matrix", label: "Matriz legal", icon: Scale },
    { href: "/dashboard/incidents", label: "Incidentes & ATEL", icon: AlertTriangle },
    { href: "/dashboard/documents", label: "Documental", icon: FileText },
  ]},
  { label: "AUTOMATIZACIÓN", items: [
    { href: "/dashboard/autopilot", label: "Briefing diario", icon: Sunrise },
    { href: "/dashboard/drones", label: "Flota Drones", icon: Plane },
    { href: "/dashboard/zones", label: "Zonas & Validación", icon: Layers },
    { href: "/dashboard/attention", label: "Inbox Maestro", icon: Inbox },
    { href: "/dashboard/siso", label: "Consulta normativa", icon: Scale },
    { href: "/dashboard/agents", label: "Centro de Agentes", icon: Network },
  ]},
  { label: "NEGOCIO", items: [
    { href: "/dashboard/billing", label: "Facturación", icon: DollarSign },
    { href: "/dashboard/settings", label: "Configuración", icon: Settings },
  ]},
];

export function Sidebar() {
  const pathname = usePathname();
  const { user } = useAuthStore();
  return (
    <aside className="w-[240px] h-screen bg-[#0E1319] border-r border-white/[0.06] flex flex-col flex-shrink-0 overflow-hidden">
      <div className="h-16 flex items-center gap-2.5 px-5 border-b border-white/[0.06]">
        <ZULogo size={30} dark />
        <div>
          <div className="font-extrabold text-[14.5px] tracking-tight leading-none text-white">ZU <span className="text-zu-lime">360</span></div>
          <div className="text-[9px] text-white/35 font-bold tracking-widest mt-1">SEGURIDAD INTEGRAL</div>
        </div>
      </div>
      <nav className="flex-1 overflow-y-auto py-4 px-3">
        {NAV_GROUPS.map((group, gi) => (
          <div key={gi} className={gi > 0 ? "mt-5" : ""}>
            {group.label && <div className="px-2.5 pb-1.5 text-[10px] font-bold text-white/30 tracking-widest">{group.label}</div>}
            {group.items.map((item) => {
              const Icon = item.icon;
              const isActive = item.href === "/dashboard" ? pathname === item.href : pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <Link key={item.href} href={item.href}
                  className={`relative flex items-center gap-2.5 px-2.5 py-[8px] rounded-lg mb-0.5 text-[13px] transition-colors ${
                    isActive ? "bg-zu-lime/[0.12] text-zu-lime font-bold" : "text-white/55 font-medium hover:bg-white/[0.05] hover:text-white/85"
                  }`}>
                  {isActive && <span className="absolute left-0 top-1.5 bottom-1.5 w-[3px] rounded-full bg-zu-lime" />}
                  <Icon size={15} className={isActive ? "text-zu-lime" : "text-white/35"} />
                  <span className="flex-1 truncate">{item.label}</span>
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
      <div className="border-t border-white/[0.06] p-3">
        <div className="flex items-center gap-2 px-2 py-2 rounded-lg cursor-pointer hover:bg-white/[0.05] transition-colors">
          <div className="w-[30px] h-[30px] rounded-full bg-gradient-to-br from-zu-lime to-zu-accent flex items-center justify-center text-[#0B0F0C] text-[10px] font-bold shadow-sm">
            {user?.firstName?.[0] || "P"}{user?.lastName?.[0] || "S"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-[12.5px] font-bold truncate text-white">{user?.firstName || "Profesional"} {user?.lastName || "SST"}</div>
            <div className="text-[10px] text-white/35 truncate font-medium">{user?.role || "profesional_sst"}</div>
          </div>
          <ChevronRight size={12} className="text-white/30" />
        </div>
      </div>
    </aside>
  );
}
