import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { ZULogo } from "@/components/brand/logo";

export const metadata = { title: "Términos del servicio · ZU 360" };

export default function TerminosPage() {
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
          <strong>Versión piloto — v1.0-piloto, {"{"}fecha de publicación: 2026-10-20{"}"}.</strong> Este documento
          rige la fase de prueba técnica controlada de ZU Seguridad Integral 360, previa a su lanzamiento público.
          Está sujeto a revisión y actualización por el equipo legal/de privacidad antes de abrir el servicio al
          público general; si eso ocurre, se te pedirá aceptar la versión actualizada antes de continuar usando el
          sistema.
        </div>

        <h1 className="text-[28px] font-extrabold tracking-tight mb-2">Términos del servicio</h1>
        <p className="text-[13px] text-white/40 mb-10">Última actualización: 20 de octubre de 2026 · Versión 1.0-piloto</p>

        <div className="space-y-8 text-[14px] leading-relaxed text-white/75">
          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">1. Quiénes somos</h2>
            <p>
              ZU Seguridad Integral 360 ("el sistema", "la plataforma") es operado por Nova Digital Solutions
              ("nosotros", "el operador"). Puedes contactarnos en cualquier momento al correo indicado en la
              sección 10 de este documento.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">2. Naturaleza piloto del servicio</h2>
            <p>
              Estás accediendo a una <strong>fase de prueba técnica controlada</strong>, anterior al lanzamiento
              público del sistema. Esto implica:
            </p>
            <ul className="mt-2 list-disc space-y-1.5 pl-5">
              <li>Las funcionalidades, la disponibilidad y el comportamiento del sistema pueden cambiar sin aviso previo mientras dure esta fase.</li>
              <li>No se garantiza continuidad del servicio, respaldo permanente de la información ni disponibilidad ininterrumpida.</li>
              <li>Los datos de empresas y trabajadores que ingreses durante esta fase son para fines de prueba técnica — <strong>no deben usarse como base para decisiones reales de cumplimiento en Seguridad y Salud en el Trabajo</strong> hasta que el sistema sea lanzado formalmente y así se comunique expresamente.</li>
              <li>La información declarada sobre tu empresa (NIT, CIIU, ARL, representante legal) no se verifica contra fuentes oficiales durante esta fase.</li>
            </ul>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">3. Registro y cuenta</h2>
            <p>
              El acceso durante esta fase se realiza mediante autenticación con tu cuenta de Google. Al registrarte,
              declaras que la información de empresa que proporcionas es veraz dentro del contexto de esta prueba y
              aceptas que puede tratarse como información ilustrativa, no verificada, mientras dure el piloto. Eres
              responsable de mantener la confidencialidad de tu sesión.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">4. Uso aceptable</h2>
            <p>No está permitido durante esta fase ni en ninguna otra:</p>
            <ul className="mt-2 list-disc space-y-1.5 pl-5">
              <li>Usar el sistema para fines ilegales o para vulnerar derechos de terceros.</li>
              <li>Intentar acceder sin autorización a cuentas, datos o infraestructura que no te corresponden.</li>
              <li>Compartir tus credenciales de acceso con terceros no autorizados.</li>
              <li>Ingresar datos reales de trabajadores o información médica/biométrica real — estos módulos permanecen deshabilitados durante esta fase.</li>
            </ul>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">5. Propiedad intelectual</h2>
            <p>
              El software, el diseño, las marcas y el contenido de ZU Seguridad Integral 360 son propiedad de Nova
              Digital Solutions o de sus licenciantes. El acceso durante el piloto no otorga ningún derecho de
              propiedad ni licencia más allá del uso previsto para la prueba.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">6. Limitación de responsabilidad</h2>
            <p>
              Dado el carácter piloto del servicio, Nova Digital Solutions no garantiza la exactitud, integridad ni
              disponibilidad continua del sistema durante esta fase, y no será responsable por decisiones tomadas
              con base en datos ingresados o generados durante el piloto. Esta limitación no aplica a conductas
              dolosas o gravemente negligentes del operador.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">7. Terminación</h2>
            <p>
              Puedes dejar de usar el sistema en cualquier momento. Durante esta fase piloto, el operador puede
              suspender o cerrar cuentas de prueba, o finalizar el piloto en su totalidad, notificando a los
              participantes con antelación razonable cuando sea posible.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">8. Cambios a estos términos</h2>
            <p>
              Podemos actualizar este documento durante el piloto o al momento del lanzamiento público. Los cambios
              materiales se comunicarán a los usuarios registrados y, cuando corresponda, se solicitará una nueva
              aceptación antes de continuar usando el sistema.
            </p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">9. Ley aplicable</h2>
            <p>Estos términos se rigen por las leyes de la República de Colombia.</p>
          </section>

          <section>
            <h2 className="text-[16px] font-bold text-white mb-2">10. Contacto</h2>
            <p>
              Para preguntas sobre estos términos, escríbenos a{" "}
              <a href="mailto:novadigitalsolutions2025@gmail.com" className="font-semibold text-zu-lime hover:underline">
                novadigitalsolutions2025@gmail.com
              </a>.
            </p>
          </section>
        </div>
      </main>
    </div>
  );
}
