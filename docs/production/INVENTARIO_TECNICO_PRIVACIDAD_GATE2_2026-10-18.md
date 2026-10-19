# Inventario técnico preliminar de privacidad — Gate 2

**Corte:** 18 de octubre de 2026. **Estado:** borrador derivado del código; incompleto y no aprobado. No es inventario legal definitivo, evaluación de impacto ni concepto jurídico. El sistema continúa en **NO-GO para datos personales reales**.

## Categorías que aparecen en el modelo

- **Identidad y vínculo laboral:** `Employee` contiene nombres, identificación, correo, teléfono, cargo y datos de relación laboral, además de restricciones activas y metadatos. El seed del piloto no crea trabajadores; las APIs y `EmployeeService` quedan cerrados en este build.
- **Salud ocupacional:** `MedicalExam` contiene tipo y fechas del examen, IPS/profesional de salud, resultado de aptitud, restricciones, notas de revisión, URL/documento y campos OCR. El servicio falla cerrado antes de DB o mensajería; rutas y OCR AI están bloqueados.
- **Accidentes e incidentes:** `Incident` puede contener vínculo al trabajador, descripción libre, localización/GPS, parte del cuerpo, naturaleza de lesión, datos FURAT, evidencia y retención legal. Las rutas Core y `IncidentService` están cerrados para el piloto.
- **Capacitación y certificados:** `Training`/`Certificate` vinculan trabajador, empresa, asistencia, puntaje, instructor, folio, fechas y vigencia. Rutas y `TrainingService` están cerrados; no se calculan ni emiten certificados con reglas de vigencia no aprobadas.
- **Fotografías, biometría y ubicación:** `EvidenceRecord` incluye URL/checksum de foto, empleado, GPS, metadatos de detección y tipo de evidencia; AI tiene enrolamiento biométrico e inspección por zona. Las rutas de evidencia no se exponen y todo `/v1/zone-inspection` y `/v1/ocr` permanece denegado.
- **Documentos, certificados y firmas:** los modelos contemplan archivos, firmantes, códigos QR, firmas, metadatos y referencias normativas. Documents registraba generación, reportes y verificación sin autorización propia; ahora sólo permite `/health`.
- **Consentimientos y auditoría:** `ConsentRecord` puede contener categoría, finalidad, método, evidencia, IP y versión de política. Captura y validación están bloqueadas porque el flujo disponible no demuestra consentimiento informado, finalidad ni actor. La revocación también sigue cerrada hasta exigir actor autenticado y bitácora verificable.

## Flujos y terceros por verificar

El código contiene una ruta OCR conectable a AWS Textract, servicios AI con integraciones externas (el uso real de cada proveedor debe confirmarse; hay configuración de proveedor que requiere inventario), y servicios de infraestructura con PostgreSQL, Redis, RabbitMQ, Weaviate y Neo4j. Las capacidades AI tienen flags apagados y una barrera adicional `PRIVACY_PROCESSING_APPROVED=false`; el perfil `high-risk` no se inicia por defecto. Que una ruta esté bloqueada en este build no demuestra por sí solo contratos, ubicación, subencargados, eliminación, telemetría o inexistencia de copias en otros entornos.

RabbitMQ puede transportar eventos de empleados, incidentes, capacitaciones o exámenes si sus servicios se habilitan. El esquema `audit.event_log` permite payload JSON, IP y user-agent; el bootstrap local ahora rechaza UPDATE/DELETE/TRUNCATE, pero aún se necesita una migración productiva, separación entre propietario de tabla y rol aplicativo, política de retención de bitácoras y pruebas en PostgreSQL real.

## Mitigaciones técnicas en la copia de hardening

Core bloquea rutas laborales, incidentes, capacitaciones y OCR-RUT después de autenticación, autorización y `CompanyScope`; además, `EmployeeService`, `IncidentService` y `TrainingService` devuelven error centinela antes de tocar repositorios o RabbitMQ. Por eso una solicitud A→B conserva 403 y una solicitud propia a la superficie laboral responde 404. Exámenes y exportación también fallan cerrados a nivel de servicio. Los servicios de consentimiento no crean concesiones implícitas ni aceptan registros históricos como autorización. Documents sólo permite health; AI devuelve errores 5xx genéricos y su fallback OCR no devuelve mensajes del proveedor.

La barrera `PRIVACY_PROCESSING_APPROVED` es sólo un control técnico de activación. No certifica que exista aprobación legal ni debe cambiarse a `true` hasta que las condiciones externas estén documentadas y aprobadas. OCR médico, biometría y toda inspección por zona permanecen cerrados aun si se activan los otros flags. Prompts/plantillas visibles se ajustaron para no citar automáticamente normas antiguas sin verificar; el catálogo de protocolos de zonas permanece desactivado y requiere revisión profesional antes de cualquier reapertura.

## Evidencia pendiente para cerrar Gate 2

El responsable de privacidad/SST y asesoría legal competente deben completar y aprobar el inventario real de campos y archivos, finalidad y necesidad de cada dato, roles de las partes, fundamento/autorización aplicable, avisos y mecanismos para derechos, proveedores y transferencias, contratos, plazos por categoría y evento, legal hold, disposición de copias/backups/logs/colas, respuesta a incidentes y criterios de interrupción. También se requieren pruebas de integración para lecturas, escrituras, archivos, jobs y auditoría, y evidencia de que ningún dato identificable se usa para entrenamiento, analítica global o pruebas de prompts sin autorización independiente.

**Decisión:** las mitigaciones reducen superficies de riesgo, pero no cierran el gate. No habilitar datos reales, exámenes médicos, incidentes reales, biometría, AI con contenido de personas, exportación ni APIs Documents hasta una aprobación formal y una prueba E2E en infraestructura controlada.
