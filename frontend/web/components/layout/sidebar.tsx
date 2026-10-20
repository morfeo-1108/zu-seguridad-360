"use client";
import Link from "next/link";
import { ZULogo } from "@/components/brand/logo";
import { usePathname } from "next/navigation";
import { useAuthStore } from "@/lib/store";
import { LayoutDashboard, Building2, Users, GraduationCap, Stethoscope, ShieldAlert, AlertTriangle, FileText, Cpu, Plane, Inbox, Bot, DollarSign, Settings, ChevronRight, Layers, Network, RefreshCw, HardHat, ClipboardCheck, Gavel, BarChart3, Scale } from "lucide-react";

const NAV_GROUPS = [
  { label: null, items: [{ href: "/dashboard", label: "Dashboard", icon: LayoutDashboard }] },
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
    { href: "/dashboard/autopilot", label: "Autopilot", icon: Cpu },
    { href: "/dashboard/drones", label: "Flota Drones", icon: Plane },
    { href: "/dashboard/zones", label: "Zonas & Validación", icon: Layers },
    { href: "/dashboard/attention", label: "Inbox Maestro", icon: Inbox },
    { href: "/dashboard/siso", label: "SISO-Bot IA", icon: Bot },
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
    <aside className="w-[228px] h-screen bg-[#FAFAF8] border-r border-zu-border flex flex-col flex-shrink-0 overflow-hidden">
      <div className="h-14 flex items-center gap-2.5 px-3.5 border-b border-zu-border">
        <ZULogo size={34} />
        <div>
          <div className="font-extrabold text-[14px] tracking-tight leading-none">ZU <span className="text-zu-accent">360</span></div>
          <div className="text-[9px] text-zu-text3 font-semibold tracking-wider mt-1">SEGURIDAD 360 · SST</div>
        </div>
      </div>
      <nav className="flex-1 overflow-y-auto py-2 px-1.5">
        {NAV_GROUPS.map((group, gi) => (
          <div key={gi}>
            {group.label && <div className="px-2.5 pt-3 pb-1 text-[10px] font-bold text-zu-text3 tracking-widest">{group.label}</div>}
            {group.items.map((item) => {
              const Icon = item.icon;
              const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
              return (
                <Link key={item.href} href={item.href}
                  className={`flex items-center gap-2 px-2.5 py-[5px] rounded-md mb-0.5 text-[12.5px] transition-colors ${
                    isActive ? "bg-[#EFF6FF] text-zu-accent font-semibold" : "text-zu-text2 hover:bg-zu-surface2"
                  }`}>
                  <Icon size={14} className={isActive ? "text-zu-accent" : "text-zu-text3"} />
                  <span className="flex-1 truncate">{item.label}</span>
                </Link>
              );
            })}
          </div>
        ))}
      </nav>
      <div className="border-t border-zu-border p-2.5">
        <div className="flex items-center gap-2 px-1 py-1.5 rounded-md cursor-pointer hover:bg-zu-surface2 transition-colors">
          <div className="w-7 h-7 rounded-full bg-gradient-to-br from-indigo-400 to-purple-500 flex items-center justify-center text-white text-[10px] font-bold">
            {user?.firstName?.[0] || "P"}{user?.lastName?.[0] || "S"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="text-[12px] font-semibold truncate">{user?.firstName || "Profesional"} {user?.lastName || "SST"}</div>
            <div className="text-[10px] text-zu-text3 truncate">{user?.role || "profesional_sst"}</div>
          </div>
          <ChevronRight size={12} className="text-zu-text3" />
        </div>
      </div>
    </aside>
  );
}
