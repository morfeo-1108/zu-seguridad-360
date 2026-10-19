"use client";
import { useState } from "react";
import { Building2, Users, FileText, AlertTriangle, GraduationCap, Stethoscope, Plane, DollarSign, LogOut } from "lucide-react";

const NAV = [
  { id: "overview", label: "Dashboard SST", icon: Building2 },
  { id: "employees", label: "Empleados", icon: Users },
  { id: "documents", label: "Documentos", icon: FileText },
  { id: "incidents", label: "Incidentes", icon: AlertTriangle },
  { id: "training", label: "Capacitaciones", icon: GraduationCap },
  { id: "medical", label: "Ex. Médicos", icon: Stethoscope },
  { id: "drone-reports", label: "Reportes Dron", icon: Plane },
  { id: "billing", label: "Facturación", icon: DollarSign },
];

export default function PortalPage() {
  const [active, setActive] = useState("overview");

  return (
    <div className="min-h-screen flex" style={{ background: "#F7F6F3" }}>
      <aside className="w-56 bg-white border-r border-gray-200 flex flex-col">
        <div className="h-14 px-4 flex items-center gap-2 border-b border-gray-200">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
            <span className="text-white font-bold text-sm">ZU</span>
          </div>
          <div>
            <div className="text-[13px] font-bold">Portal Cliente</div>
            <div className="text-[10px] text-gray-500">Construcciones Andina</div>
          </div>
        </div>
        <nav className="flex-1 p-2">
          {NAV.map((item) => {
            const Icon = item.icon;
            const isActive = active === item.id;
            return (
              <button key={item.id} onClick={() => setActive(item.id)}
                className={`w-full flex items-center gap-2 px-3 py-2 rounded-md mb-0.5 text-[12.5px] transition-colors ${
                  isActive ? "bg-blue-50 text-blue-700 font-semibold" : "text-gray-600 hover:bg-gray-100"
                }`}>
                <Icon size={14} />
                <span>{item.label}</span>
              </button>
            );
          })}
        </nav>
        <div className="p-3 border-t border-gray-200">
          <div className="flex items-center gap-2 mb-2">
            <div className="w-7 h-7 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white text-[10px] font-bold">JR</div>
            <div>
              <div className="text-[12px] font-semibold">Juan Restrepo</div>
              <div className="text-[10px] text-gray-500">Rep. Legal</div>
            </div>
          </div>
          <button className="w-full text-[11px] text-gray-600 hover:text-gray-800 flex items-center gap-1.5 px-2 py-1.5 rounded-md hover:bg-gray-100">
            <LogOut size={11} /> Cerrar sesión
          </button>
        </div>
      </aside>

      <main className="flex-1 p-6 overflow-auto">
        {active === "overview" && (
          <>
            <div className="mb-5">
              <h1 className="text-xl font-bold mb-1">Mi SG-SST</h1>
              <p className="text-[12px] text-gray-600">Construcciones Andina S.A.S — NIT 900.234.567-1 — ARL Sura Clase IV</p>
            </div>
            <div className="grid grid-cols-4 gap-3 mb-5">
              {[
                { label: "Cumplimiento Res. 0312", value: "64%", color: "#D97706" },
                { label: "Empleados Activos", value: "87", color: "#1C1C1A" },
                { label: "IF Acum. 12 meses", value: "4.2", color: "#1C1C1A" },
                { label: "Incidentes Año", value: "3", color: "#DC2626" },
              ].map((kpi) => (
                <div key={kpi.label} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="text-[11px] text-gray-500 mb-1">{kpi.label}</div>
                  <div className="text-2xl font-bold" style={{ color: kpi.color, fontFamily: "monospace" }}>{kpi.value}</div>
                </div>
              ))}
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h3 className="font-semibold text-[13px] mb-3">Cumplimiento Estándares Mínimos</h3>
                {[
                  { label: "Política SST", done: true },
                  { label: "Plan Anual de Trabajo", done: true },
                  { label: "Matriz de Riesgos GTC 45", done: true },
                  { label: "Programa Capacitaciones", done: false },
                  { label: "Indicadores SST", done: true },
                  { label: "Auditoría Anual", done: false },
                  { label: "Investigación Incidentes", done: false },
                ].map((s) => (
                  <div key={s.label} className="flex items-center gap-2 py-1.5 border-b border-gray-100 last:border-0">
                    <div className={`w-4 h-4 rounded ${s.done ? "bg-green-500" : "bg-gray-200"} flex items-center justify-center`}>
                      {s.done && <span className="text-white text-xs">✓</span>}
                    </div>
                    <span className="text-[12px]">{s.label}</span>
                  </div>
                ))}
              </div>
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h3 className="font-semibold text-[13px] mb-3">Documentos Pendientes de tu firma</h3>
                <div className="space-y-2">
                  {[
                    { title: "Política SST — Versión 2026", urgency: "Vence en 3 días" },
                    { title: "Plan Anual de Trabajo 2026", urgency: "Vence en 7 días" },
                    { title: "Reglamento de Higiene y Seguridad", urgency: "Vence en 10 días" },
                  ].map((d, i) => (
                    <div key={i} className="flex items-center justify-between p-2.5 bg-amber-50 border border-amber-200 rounded-md">
                      <div>
                        <div className="text-[12px] font-medium">{d.title}</div>
                        <div className="text-[10px] text-amber-700">{d.urgency}</div>
                      </div>
                      <button className="px-3 py-1 bg-blue-600 text-white text-[11px] font-semibold rounded-md hover:bg-blue-700 transition-colors">
                        Firmar
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </>
        )}
        {active !== "overview" && (
          <div className="bg-white border border-gray-200 rounded-lg p-12 text-center">
            <div className="text-gray-500 text-sm">Módulo: {NAV.find(n => n.id === active)?.label}</div>
            <div className="text-[12px] text-gray-400 mt-2">Conecta con el API core para datos reales</div>
          </div>
        )}
      </main>
    </div>
  );
}
