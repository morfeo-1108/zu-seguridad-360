// Registro único de lo que la página pública afirma sobre ZU 360.
//
// Regla (Fase 3.6 del plan): una capacidad solo se publica como hecho si
// está "verificado" — ejecutada de punta a punta contra el sistema real, con
// la evidencia anotada aquí. Lo demás se muestra como hoja de ruta o no se
// muestra. Publicitar algo que el sistema no hace es publicidad engañosa
// (Estatuto del Consumidor, Ley 1480 de 2011; vigilancia de la SIC).
//
// Al cambiar un estado, actualice también la evidencia y la fecha.

export type ClaimStatus = "verificado" | "hoja_de_ruta";

export interface Claim {
  id: string;
  status: ClaimStatus;
  title: string;
  body: string;
  proof?: string; // frase de prueba visible para el lector
  evidence: string; // interno: cómo y cuándo se verificó
}

export const DIFFERENTIATORS: Claim[] = [
  {
    id: "aislamiento",
    status: "verificado",
    title: "Cada empresa cliente ve solo lo suyo",
    body: "Puede llevar varias empresas desde una cuenta sin que la información de una se mezcle con la de otra. El alcance sale de la sesión firmada, no de lo que envíe el navegador.",
    proof: "Probado con intentos reales de acceso cruzado entre cuentas: todos rechazados.",
    evidence: "smoke-company-isolation.py 25/25 (2026-10-19); pruebas de acceso cruzado en notifications, documents, media, SLA y onboarding (2026-10-22): 404/403 en todos los casos.",
  },
  {
    id: "trazabilidad",
    status: "verificado",
    title: "Documentos con versión, autor y huella",
    body: "Cada documento generado queda guardado con su número de versión, quién lo generó, cuándo, y su huella SHA-256. Si alguien altera el archivo, la huella deja de coincidir.",
    proof: "Al descargar, la huella del archivo coincide con la registrada al generarlo.",
    evidence: "documents: politica_sst v1→v2, sha256 de descarga = sha256 almacenado (2026-10-22). Identidad de la empresa tomada de Core, no del formulario.",
  },
  {
    id: "plazos",
    status: "verificado",
    title: "Ningún hallazgo se queda sin dueño",
    body: "Una inspección con hallazgos genera su plan de acción y abre un plazo de revisión. Si vence sin aprobación, la plataforma le avisa con el nombre de la empresa y la inspección.",
    proof: "Plazo vencido detectado y notificado de forma automática en la prueba del 22 de octubre de 2026.",
    evidence: "Core inspection.completed → SLA abre plazo → vencimiento → notifications (in-app) + mail; aprobación cierra el plazo (2026-10-22).",
  },
  {
    id: "briefing",
    status: "verificado",
    title: "Un resumen de su cartera cada mañana",
    body: "Antes de empezar el día recibe los pendientes de todas sus empresas: inspecciones sin aprobar, plazos vencidos, EPP por vencer, COPASST sin reunión del mes, requisitos legales sin revisar. Solo lee sus registros; no cambia nada por usted.",
    proof: "Cada punto indica su fundamento, por ejemplo: reunión mensual del COPASST (Res. 2013 de 1986, art. 7).",
    evidence: "autopilot briefing 06:30 America/Bogota, generado y entregado en la prueba con tenant QA (2026-10-22).",
  },
  {
    id: "datos_salud",
    status: "verificado",
    title: "Datos de salud: cerrados hasta que sea seguro",
    body: "Los módulos con información de salud de los trabajadores permanecen cerrados hasta completar los controles que exige su tratamiento como dato sensible (Ley 1581 de 2012). Preferimos lanzar con menos módulos que con un riesgo para sus trabajadores.",
    evidence: "Compuerta GATES_DATOS_REALES_TRABAJADORES activa en producción: employees/incidents/medical rechazan datos reales (2026-10-22).",
  },
  {
    id: "motor_normativo",
    status: "verificado",
    title: "Consulta normativa con artículo y fuente",
    body: "Pregunte en lenguaje natural y reciba los tres artículos vigentes más pertinentes, con su texto literal y el enlace a la fuente oficial. Si una norma relacionada está derogada, se lo advierte en vez de citarla.",
    proof: "En una prueba con 14 preguntas reales, el artículo correcto apareció entre los tres primeros resultados en 12.",
    evidence: "corpus/eval_questions.json vía API desplegada (2026-10-22): top-1 10/14, top-3 12/14, latencia media 0,41 s; 16 normas, 889 fragmentos. Fallos: plazo de reporte de AT a la ARL; datos sensibles (art. 4 en vez de 5).",
  },
];

export const ROADMAP: Claim[] = [
  {
    id: "pagos",
    status: "hoja_de_ruta",
    title: "Pago en línea de la suscripción",
    body: "Pago con tarjeta, PSE o Nequi a través de Wompi y factura electrónica automática.",
    evidence: "Flujo completo construido y probado contra un Wompi simulado; pendiente aprobación del comercio en Wompi y resolución de facturación DIAN.",
  },
  {
    id: "ia_vision",
    status: "hoja_de_ruta",
    title: "Verificación fotográfica de EPP",
    body: "Validación del uso de elementos de protección por fotografía de la zona de trabajo, con revisión humana.",
    evidence: "Módulo de visión sin activar en producción.",
  },
  {
    id: "drones",
    status: "hoja_de_ruta",
    title: "Rondas de inspección con dron",
    body: "Inspecciones aéreas programadas para instalaciones extensas.",
    evidence: "Servicio de dron bloqueado por diseño: no hay controles de seguridad física ni autorización de operador implementados.",
  },
];

/** Comparativa antes/después: solo filas verificables en el sistema actual. */
export const BEFORE_AFTER: { before: string; after: string }[] = [
  { before: "Cada empresa en su propia carpeta y su propia hoja de cálculo.", after: "Todas sus empresas en una cuenta, con información aislada entre sí." },
  { before: "Los hallazgos de una inspección se copian a mano a un plan de acción.", after: "El plan de acción se genera desde los hallazgos no conformes." },
  { before: "Los plazos se vigilan de memoria o con recordatorios sueltos.", after: "Cada hallazgo abre un plazo y el vencimiento genera una alerta." },
  { before: "Documentos sin control de versiones: nadie sabe cuál es el vigente.", after: "Cada documento con versión, autor, fecha y huella verificable." },
  { before: "El estado de cada empresa se arma revisando una por una.", after: "Un briefing diario con los pendientes de toda la cartera." },
];

/**
 * Las cifras del modelo de negocio (20–50 empresas por profesional, 1–2 h por
 * empresa al mes, onboarding de 2–3 días) son METAS del diseño, aún no
 * medidas en uso real. Se publican como meta, explícitamente.
 */
export const TARGETS = [
  { metric: "Empresas por profesional", value: "20–50", note: "meta del diseño" },
  { metric: "Horas por empresa al mes", value: "1–2", note: "meta del diseño" },
  { metric: "Puesta en marcha de una empresa", value: "2–3 días", note: "la plataforma ya mide la duración real de cada onboarding" },
];
