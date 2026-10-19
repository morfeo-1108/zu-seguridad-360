"use client";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, UserRound } from "lucide-react";
import { DashboardShell } from "@/components/layout/dashboard-shell";
import { Badge, Button, Card, Empty, PageHeader, Spinner } from "@/components/ui";
import { useEmployee } from "@/lib/hooks";
import { useAuthStore } from "@/lib/store";

export default function EmployeeProfilePage() {
  const router = useRouter();
  const params = useParams();
  const employeeId = typeof params.id === "string" ? params.id : "";
  const activeCompany = useAuthStore((state) => state.activeCompany);
  const companyId = activeCompany?.id || "";
  const { data: employee, isLoading, isError } = useEmployee(companyId, employeeId);

  return <DashboardShell>
    <PageHeader title="Perfil de prueba" description="Sólo datos sintéticos devueltos por Core bajo el scope de empresa del JWT." action={<Button variant="secondary" onClick={() => router.push("/dashboard/employees")}><ArrowLeft size={13}/> Volver a personas</Button>} />
    {!companyId ? <Empty icon={UserRound} title="Sin empresa activa" description="Activa una empresa con membership desde el encabezado."/> : isLoading ? <Spinner size={24}/> : isError || !employee ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">No se encontró el perfil dentro de la empresa activa. La ruta no cambia el scope firmado en el token.</p> : !employee.is_synthetic ? <p role="alert" className="rounded-md border border-red-200 bg-red-50 p-4 text-sm text-red-800">Core devolvió un perfil no marcado como sintético; por seguridad no se muestra ningún campo.</p> : <>
      <div className="mb-4 flex items-center gap-2"><Badge variant="amber">SINTÉTICO</Badge><Badge variant={employee.status === "active" ? "green" : "gray"}>{employee.status}</Badge></div>
      <Card className="max-w-2xl p-5"><div className="mb-4 flex items-center gap-3"><div className="flex h-12 w-12 items-center justify-center rounded-full bg-amber-100 text-amber-900"><UserRound size={22}/></div><div><h2 className="text-base font-bold">{employee.first_name} {employee.last_name}</h2><p className="text-[11px] text-zu-text3">ID de prueba: {employee.document_number}</p></div></div>
        <dl className="grid grid-cols-1 gap-4 text-xs sm:grid-cols-2"><div><dt className="text-zu-text3">Tipo de identificador</dt><dd>{employee.document_type}</dd></div><div><dt className="text-zu-text3">Correo reservado</dt><dd>{employee.email || "No registrado"}</dd></div><div><dt className="text-zu-text3">Cargo de demostración</dt><dd>{employee.position}</dd></div><div><dt className="text-zu-text3">Área</dt><dd>{employee.department || "No asignada"}</dd></div></dl>
      </Card><p className="mt-3 max-w-2xl text-[11px] leading-relaxed text-zu-text3">Esta ficha no consulta ni contiene exámenes médicos, restricciones, biometría, GPS, evidencias o actividad de agentes. Para editar el cargo/área sintéticos, usa el módulo de personas.</p>
    </>}
  </DashboardShell>;
}
