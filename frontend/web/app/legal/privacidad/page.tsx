import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";

export const metadata = { title: "Aviso de privacidad · ZU 360" };

export default function PrivacidadPage() {
  return (
    <div className="min-h-screen bg-[#0B0F0C] text-white">
      <header className="border-b border-white/[0.06]">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-6 py-5">
          <Link href="/"><ZULogo size={28} dark showText /></Link>
          <Link href="/" className="inline-flex items-center gap-1.5 text-[12px] font-semibold text-white/50 hover:text-white transition">
            <ArrowLeft size={14} /> Volver al inicio
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-3xl px-6 py-12">
        <div className="mb-8 rounded-lg border border-amber-500/25 bg-amber-500/[0.06] px-4 py-3 text-[12.5px] leading-relaxed text-amber-300">
          <strong>Versión piloto — v1.0-piloto, {"{"}fecha de publicación: 2026-10-20{"}"}.</strong> Este aviso cubre
          la fase de prueba técnica controlada de ZU Seguridad Integral 360. Está sujeto a revisión por el equipo
          legal/de privacidad antes del lanzamiento público; si el tratamiento de datos cambia materialmente, se
          publicará una nueva versión y se solicitará tu aceptación antes de continuar usando el sistema.
        </div>

        <h1 className="text-[28px] font-extrabold tracking-tight mb-2">Aviso de privacidad</h1>
        <p className="text-[13px] text-white/40 mb-10">Última actualización: 20 de octubre de 2026 · Versión 1.0-piloto</p>

        <div className="space-y-8 text-[14px] leading-relaxed text-white/75">
          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">1. Responsable del tratamiento</h2>
            <p>
              Nova Digital Solutions es responsable del tratamiento de los datos personales que recolecta ZU
              Seguridad Integral 360 durante esta fase piloto. Contacto:{" "}
              <a href="mailto:novadigitalsolutions2025@gmail.com" className="font-semibold text-zu-lime hover:underline">
                novadigitalsolutions2025@gmail.com
              </a>.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">2. Qué datos recolectamos</h2>
            <ul className="list-disc space-y-1.5 pl-5">
              <li><strong>De tu cuenta de Google:</strong> nombre, correo electrónico y foto de perfil básica, entregados por Google al autenticarte — no accedemos a tu contraseña de Google ni a otra información de tu cuenta.</li>
              <li><strong>De la empresa que registras durante el piloto:</strong> razón social, NIT, código CIIU, ARL y datos del representante legal, declarados por ti. Durante esta fase estos datos <strong>no se verifican</strong> contra RUES, DIAN, ARL ni otra fuente oficial.</li>
              <li><strong>Datos técnicos de uso:</strong> registros de inicio de sesión y actividad dentro del sistema, para poder diagnosticar errores durante la prueba.</li>
            </ul>
            <p className="mt-2">
              Mientras dure el piloto, los módulos que tratarían datos reales de trabajadores (exámenes médicos,
              biometría, fotografías identificables de personal) permanecen deshabilitados — no se recolecta ese
              tipo de información todavía.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">3. Finalidad del tratamiento</h2>
            <ul className="list-disc space-y-1.5 pl-5">
              <li>Permitirte crear una cuenta y acceder al sistema durante la fase de prueba técnica.</li>
              <li>Evaluar el funcionamiento técnico del sistema antes de su lanzamiento público.</li>
              <li>Comunicarnos contigo sobre la prueba (soporte, cambios relevantes, cierre del piloto).</li>
            </ul>
            <p className="mt-2">
              No usamos estos datos para calcular clasificaciones de riesgo reales, tomar decisiones de cumplimiento
              normativo, ni para entrenar modelos de inteligencia artificial.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">4. Base legal del tratamiento</h2>
            <p>
              El tratamiento se basa en tu consentimiento libre, previo e informado, otorgado al aceptar este aviso
              durante el registro, conforme a la Ley 1581 de 2012 y sus decretos reglamentarios.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">5. Tus derechos (habeas data)</h2>
            <p>Como titular de tus datos personales, tienes derecho a:</p>
            <ul className="mt-2 list-disc space-y-1.5 pl-5">
              <li>Conocer, actualizar y rectificar tus datos personales.</li>
              <li>Solicitar prueba de la autorización otorgada.</li>
              <li>Ser informado sobre el uso que se les ha dado.</li>
              <li>Revocar la autorización y/o solicitar la supresión de tus datos, cuando no exista un deber legal o contractual que impida su eliminación.</li>
              <li>Acceder de forma gratuita a tus datos personales que hayan sido objeto de tratamiento.</li>
            </ul>
            <p className="mt-2">
              Para ejercer cualquiera de estos derechos, escríbenos a{" "}
              <a href="mailto:novadigitalsolutions2025@gmail.com" className="font-semibold text-zu-lime hover:underline">
                novadigitalsolutions2025@gmail.com
              </a>.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">6. Con quién compartimos tus datos</h2>
            <p>
              Durante esta fase, tus datos se comparten únicamente con los proveedores estrictamente necesarios
              para operar el piloto: Google (para la autenticación de tu cuenta) y nuestro proveedor de
              infraestructura y red (Cloudflare, para servir el sitio de forma segura). No vendemos tus datos ni los
              compartimos con terceros para fines comerciales o publicitarios.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">7. Seguridad</h2>
            <p>
              Toda comunicación con el sistema viaja cifrada (HTTPS). Los datos de la fase piloto se mantienen
              aislados del tratamiento de datos reales de trabajadores, que permanece deshabilitado mientras dure
              esta fase.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">8. Conservación</h2>
            <p>
              Conservamos los datos de tu cuenta de prueba mientras dure el piloto. Al finalizar esta fase, si no
              continúas hacia el lanzamiento formal del servicio, los datos ingresados durante la prueba se
              eliminarán conforme al procedimiento que se comunique al cierre del piloto.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">9. Vigencia y cambios</h2>
            <p>
              Este aviso aplica mientras dure la fase piloto. Antes del lanzamiento público, este documento será
              revisado formalmente y, si cambia de forma material, se publicará una nueva versión y se solicitará
              tu aceptación para continuar usando el sistema.
            </p>
          </section>
        </div>
      </main>
    </div>
  );
}
