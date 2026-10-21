import { AlertTriangle, BarChart3, ClipboardCheck, FileText, HardHat, MapPin, RefreshCcw, Scale, Users } from "lucide-react";

// Company modules usable with real (non-synthetic) accounts. Workers,
// incidents, trainings and medical exams stay disabled for real data until
// docs/production/GATES_DATOS_REALES_TRABAJADORES.md is closed.
export const REAL_MODULES = [
  { href: "/dashboard/risks", title: "Matriz de riesgos", description: "Identificación de peligros y valoración de riesgos.", icon: AlertTriangle },
  { href: "/dashboard/inspections", title: "Inspecciones", description: "Listas de chequeo y hallazgos de inspección.", icon: ClipboardCheck },
  { href: "/dashboard/epp", title: "Elementos de protección (EPP)", description: "Inventario y entregas de EPP.", icon: HardHat },
  { href: "/dashboard/legal-matrix", title: "Matriz legal", description: "Requisitos legales aplicables y su seguimiento.", icon: Scale },
  { href: "/dashboard/committees", title: "Comités", description: "COPASST, convivencia y sus reuniones.", icon: Users },
  { href: "/dashboard/management", title: "Ciclo PHVA", description: "Planear, hacer, verificar y actuar del SG-SST.", icon: RefreshCcw },
  { href: "/dashboard/documents", title: "Documentos", description: "Documentos y fotos de la empresa.", icon: FileText },
  { href: "/dashboard/indicators", title: "Indicadores", description: "Indicadores del SG-SST.", icon: BarChart3 },
  { href: "/dashboard/zones", title: "Proyectos y zonas", description: "Sedes, proyectos y zonas de trabajo.", icon: MapPin },
];
