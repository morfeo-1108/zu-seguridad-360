Para incorporar las nuevas capacidades estratégicas que hemos definido (el **módulo de conocimiento y aprendizaje autónomo por retroalimentación**, la **ejecución proactiva de agentes de IA**, el soporte omnicanal para **Web, Google Play y App Store**, y los tres planes comerciales orientados a profesionales), se debe expandir y robustecer tu **Checklist de Preparación para Producción**.

Aquí tienes el checklist completo actualizado, integrando de manera estricta los nuevos requerimientos técnicos y de seguridad:

---

# Checklist de preparación para producción — ZU Seguridad 360 (Versión Ampliada y Mejorada)

**Regla de salida:** no habilitar clientes reales mientras quede abierto algún riesgo crítico de build, aislamiento entre clientes, autenticación, protección de datos o validación de agentes autónomos. Si una función no supera su gate, debe quedar **deshabilitada** y no anunciarse como disponible.

## 1. Alcance y producto

* [ ] Definir qué módulos se lanzan en la primera versión: núcleo SST, portal web, aplicaciones móviles (iOS / Android), billing, IA asistida, biometría, drones, etc.
* [ ] Clasificar cada función como **productiva**, **beta**, **demo** o **deshabilitada**.
* [ ] Eliminar de los flujos de producción datos de demostración, mensajes de “simulación” y confirmaciones que no correspondan a una operación persistida.
* [ ] Deshabilitar por defecto IA avanzada, biometría, visión y drones hasta que superen sus verificaciones específicas.
* [ ] Definir roles y responsabilidades: profesional SST (usuario principal), empleador, trabajador, IPS, auditor, contratista y administrador.

## 2. Build y calidad del software

* [ ] Corregir referencias y dependencias que impiden compilar los servicios, incluidos los problemas identificados en `core` y `billing`.
* [ ] Generar, revisar y versionar los archivos de dependencias requeridos —por ejemplo, `go.sum`— y fijar versiones de herramientas, librerías e imágenes.
* [ ] Hacer que CI compile **todos** los servicios desde un entorno limpio.
* [ ] Ejecutar y exigir pruebas unitarias, de integración y de contrato; ningún job de seguridad o pruebas debe poder fallar y dejar verde el pipeline.
* [ ] Añadir pruebas E2E para los flujos SST que se ofrecerán: autenticación, creación y consulta de registros, revisión, exportación y límites del plan.
* [ ] Probar instalación desde cero y actualización desde una versión anterior.
* [ ] No publicar una imagen que no haya sido construida y probada por el pipeline.

## 3. Autenticación, permisos y aislamiento multiempresa

* [ ] Proteger **cada endpoint** con autenticación; los servicios internos también deben validar identidad de servicio.
* [ ] Aplicar autorización por rol, acción, objeto, empresa y tenant; utilizar denegación por defecto.
* [ ] Derivar el tenant de la sesión y de la membresía validada en el servidor, no confiar en un tenant enviado libremente por body, URL o header.
* [ ] Validar que cada relación sea legítima: `tenant → empresa → empleado/proyecto/zona/recurso`.
* [ ] Corregir y volver a probar, como mínimo, exportaciones, evidencias, datos médicos, PHVA, zonas, proyectos, billing, atención y SLA.
* [ ] Ejecutar pruebas negativas con **dos tenants** y UUIDs conocidos: cada acceso cruzado debe ser rechazado.
* [ ] Incorporar controles en la base de datos —RLS o defensa equivalente, claves foráneas y restricciones de pertenencia cuando correspondan— además del control en la aplicación.
* [ ] Evitar que endpoints públicos de certificados, reportes o descargas entreguen datos personales innecesarios.

## 4. Datos SST y archivos sensibles

* [ ] Definir qué datos se almacenan —incluidos datos médicos, fotos, ubicación y evidencias—, para qué finalidad, quién accede y durante cuánto tiempo.
* [ ] Aplicar privilegio mínimo y separar los permisos de trabajador, empleador, profesional SST, IPS y auditor.
* [ ] Cifrar datos y respaldos; proteger claves y credenciales con un gestor de secretos.
* [ ] Asegurar el servicio de archivos con autenticación y autorización por tenant/empresa.
* [ ] Limitar tamaño y tipo de archivo, verificar contenido real, analizar archivos con una herramienta antimalware operativa y controlar quién descarga cada objeto.
* [ ] Usar enlaces de descarga temporales y autorizados; definir expiración y eliminación.
* [ ] Implementar registro de acceso, cambios, revisiones, exportaciones y borrados de información sensible.
* [ ] Definir procesos probados de retención, exportación y borrado.
* [ ] Revisar con asesoría competente los requisitos de privacidad, datos sensibles, firma y SST aplicables a los países donde se ofrecerá el servicio.

## 5. Login, portal y aplicaciones (Web, Google Play y App Store)

* [ ] Conectar web, portal y móvil a autenticación real; probar inicio, renovación, cierre y recuperación de cuenta.
* [ ] Proteger el almacenamiento de sesión: evitar que tokens de larga duración queden expuestos en almacenamiento accesible a scripts o aplicaciones.
* [ ] Sustituir fixtures y datos fijos por datos reales obtenidos del backend.
* [ ] Verificar que cada botón y confirmación persista el cambio y muestre claramente errores y estados pendientes.
* [ ] Probar formularios, permisos, validación, subida de evidencia y descarga desde roles distintos.
* [ ] Garantizar la correcta compilación y despliegue de las aplicaciones nativas para **Google Play (Android)** y **App Store (iOS)** con sus respectivas políticas de privacidad y permisos de cámara/ubicación.
* [ ] Probar rigurosamente la operación móvil sin conexión (*offline-first*), almacenamiento local cifrado, reintentos y sincronización idempotente de inspecciones en campo.
* [ ] Revisar accesibilidad básica y que errores, carga y confirmaciones sean comprensibles.

## 6. Planes, licencias y facturación (Estructura B2P)

* [ ] Definir catálogo versionado de planes adaptados al profesional SST:
* **Plan Básico ($70 USD/mes):** Hasta 10 empresas gestionadas, núcleo SST y Resolución 0312 básica.
* **Plan Pro ($130 USD/mes):** Hasta 25 empresas gestionadas, alertas tempranas, EPP y exámenes médicos.
* **Plan Enterprise ($300 USD/mes):** Empresas ilimitadas, IA avanzada habilitada y módulo de drones/campo.


* [ ] Aplicar entitlements y cuotas en el backend (ej. control estricto del límite de empresas por tenant), no solo en la interfaz.
* [ ] Probar alta, cambio de plan, exceso de cuota, período de gracia, suspensión, cancelación y reactivación.
* [ ] Hacer que el estado de suscripción determine efectivamente el acceso al producto.
* [ ] Proteger billing y comprobar en servidor la relación entre tenant y empresa.
* [ ] Implementar confirmación de pago verificable e idempotente; probar webhooks, duplicados, reintentos y reconciliación.
* [ ] No marcar una factura como pagada sin evidencia del proceso correspondiente.
* [ ] Verificar la vigencia de la licencia profesional SST y su vinculación legal con las empresas atendidas.

## 7. Despliegue y seguridad operacional

* [ ] Retirar secretos y contraseñas de archivos, imágenes y repositorios; rotar cualquier valor que pueda haberse reutilizado.
* [ ] Configurar TLS en el acceso externo y mantener bases de datos, colas, almacenamiento y paneles administrativos en redes privadas.
* [ ] Corregir y validar Compose, Kubernetes, Terraform, gateway y scripts de instalación; verificar que rutas y puertos coincidan.
* [ ] Ejecutar migraciones versionadas con historial, control de concurrencia y errores visibles; no iniciar tráfico antes de verificar el esquema.
* [ ] Añadir readiness y health checks por servicio y dependencia.
* [ ] Configurar backups cifrados y probar restauraciones. Documentar objetivos de pérdida y tiempo de recuperación (RPO/RTO).
* [ ] Implementar monitoreo, alertas, logs con identificadores de solicitud y métricas de errores, colas, almacenamiento y uso por tenant.
* [ ] Definir procedimientos para incidentes, degradación del servicio, restauración y comunicación a clientes.
* [ ] Fijar imágenes y acciones por versión inmutable; generar inventario de dependencias y bloquear vulnerabilidades graves según la política acordada.

## 8. Gates de mayor riesgo (IA, Conocimiento Autónomo, Biometría y Drones)

### Módulo de IA, Conocimiento Autónomo (*Knowledge Core*) y OCR

* [ ] Aplicar autenticación y autorización por tenant/objeto en todas las rutas, incluidos WebSockets y motores de inferencia.
* [ ] Garantizar el aislamiento estricto de la memoria vectorial (*RAG*) por empresa: **ningún dato confidencial de un tenant puede cruzar o entrenar modelos de otro tenant**.
* [ ] Validar en el servidor el circuito de retroalimentación (*Feedback Learning*): registrar las correcciones del profesional sobre los borradores de la IA de forma segura y anonimizada para las mejoras globales.
* [ ] Bloquear resultados de OCR médico y salidas automatizadas sensibles hasta que una persona autorizada (el profesional SST) los revise y firme digitalmente.
* [ ] El consentimiento de tratamiento de datos debe verificarse en el servidor, estar asociado a persona, finalidad y versión, y admitir revocación y borrado.
* [ ] Impedir que modos demo, modelos faltantes o resultados de confianza no calibrada generen evidencia o decisiones legales.
* [ ] Identificar las transferencias a proveedores externos de IA (ej. Anthropic), minimizarlas y controlarlas mediante tokens y caché en Redis.
* [ ] Mostrar las salidas de la IA como asistencia o borrador con diseño editorial impecable en PDF, exigiendo revisión profesional antes de usarlas operativamente.

### Drones y Operación en Campo

* [ ] No habilitar vuelos reales hasta probar la integración con el hardware objetivo.
* [ ] Autenticar dispositivos y broker; proteger comunicaciones y limitar permisos por topic.
* [ ] Calcular preflight con datos confiables en el servidor; no aceptar del cliente una aprobación de vuelo como autoridad.
* [ ] Implementar confirmación del dispositivo para cada comando crítico, junto con protección contra repetición y expiración.
* [ ] Probar pérdida de enlace, batería baja, GPS degradado, geocerca, reinicio y aterrizaje de emergencia.
* [ ] Persistir misión, decisiones, comandos y confirmaciones para auditoría.
* [ ] Mantener el módulo apagado si falla una dependencia crítica o no existe confirmación del hardware.

---

## Criterio final de **GO / NO-GO**

Antes de habilitar producción, exigir evidencia documentada de que:

* [ ] Todos los servicios del alcance se compilan y despliegan desde un entorno limpio en el VPS.
* [ ] Las pruebas negativas confirman aislamiento absoluto entre tenants y permisos por rol/objeto.
* [ ] Los flujos ofrecidos funcionan con datos reales de prueba, no con mocks.
* [ ] Las aplicaciones móviles compilan de forma limpia para distribución en App Store y Google Play.
* [ ] Las migraciones, backups y restauraciones están probados.
* [ ] Los secretos, permisos, alertas y pasarelas de pago están configurados y operativos.
* [ ] Los planes ($70, $130 y $300 USD) aplican sus límites de empresas y accesos en el backend.
* [ ] El motor de conocimiento y cada módulo de IA o drones tiene aprobación independiente; de lo contrario, permanece desactivado.


> **Nota de estado sobre precios:** los valores USD 70/130/300 de esta versión provienen de la propuesta inicial y son sólo hipótesis de trabajo. No son precios aprobados, cotizados ni validados con clientes. Antes de publicarlos hay que definir moneda/facturación, impuestos, costos variables (IA/almacenamiento/soporte), límites y disposición a pagar; las pruebas actuales no deben cobrar.
