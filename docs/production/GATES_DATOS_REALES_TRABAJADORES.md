# Gates críticos para habilitar datos reales de trabajadores

**Decisión actual: NO-GO.** No cargar datos reales, ni siquiera en un piloto informal, hasta cerrar y aprobar los gates P0 que correspondan al alcance habilitado. Este documento operacionaliza el checklist general con criterios de aceptación y evidencia exigible. No sustituye asesoría legal, de privacidad ni de un profesional SST competente.

## Estado conocido del sistema

- Los diez módulos Go pasan `go test`; la suite de AI (42) y Documents (10), más el guard de Docker, pasan desde el virtualenv aislado. `go.sum` se regeneró/verificó en diez módulos después de limpiar la caché Go.
- No se ejecutó Docker/Compose, build de imágenes, stack de servicios, migraciones integradas, pruebas HTTP E2E ni restore de backups.
- Hay tests unitarios que exigen 403 para token A→empresa B, verifican que header/query/body no sustituyen el scope firmado, y prueban 404 para rutas laborales propias mientras Gate 2 esté abierto. El smoke HTTP con Postgres sigue sin ejecutarse; no existe cobertura de integración de todos los recursos.
- Se añadieron barreras fail-closed a empleados/incidentes/capacitaciones/OCR-RUT tanto en HTTP como en sus servicios Core, además de APIs Documents, exámenes médicos, consentimiento no verificado y exportación. Las rutas AI médicas/biométricas/de inspección por zona quedan bloqueadas incluso con sus flags activados. No se registra consentimiento ni se aprueba procesamiento sensible hasta completar controles externos.
- El bootstrap local hace `audit.event_log` resistente a UPDATE/DELETE/TRUNCATE; eso no reemplaza una migración productiva ni una base/rol aplicativo no propietario.
- Billing se mantiene en health-only; no se deben ejecutar pruebas de cobro o suscripciones con este build.
- El seed crea dos profesionales/empresas sintéticos en dos tenants distintos; no crea trabajadores. Los checks unitarios no acreditan autorización, confidencialidad ni integridad de datos reales.
- El [inventario técnico preliminar de privacidad](INVENTARIO_TECNICO_PRIVACIDAD_GATE2_2026-10-18.md) enumera datos y flujos inferidos del código; requiere completar y aprobar por responsables competentes.

## P0 — Bloqueos absolutos antes de cualquier dato personal real

### 1. Inventario, finalidad y minimización

- [ ] Enumerar campos y archivos que pueden identificar o describir a una persona: identidad/contacto, relación laboral, incidentes, restricciones, salud/exámenes, biometría, fotos, ubicación, firmas, evidencias y metadatos.
- [ ] Para cada categoría, documentar finalidad concreta, necesidad, fuente, responsable del tratamiento, quién decide el uso, usuarios autorizados, proveedores que la reciben, ubicación/transferencia, retención y mecanismo de eliminación.
- [ ] Reducir el alcance inicial al mínimo necesario. Excluir diagnósticos, exámenes, biometría, ubicación continua, fotos identificables e incidentes reales si no son esenciales para la primera prueba.
- [ ] Prohibir que datos identificables se usen para entrenar modelos, mejorar servicios globales, analytics compartidos o probar prompts sin aprobación independiente y controles de anonimización validados.

**Criterio de cierre:** inventario y mapa de flujo aprobados por responsable de privacidad/SST; cada campo tiene finalidad, base/autorización documentada, acceso y retención; categorías no aprobadas están deshabilitadas técnicamente.

### 2. Revisión legal, privacidad y relación con clientes

- [ ] Obtener revisión de asesoría legal/privacidad competente para Colombia y para el rol real de cada parte: cliente, profesional SST, empleador, trabajador, proveedor y subencargado.
- [ ] Determinar y documentar la base jurídica/autorización aplicable a cada tratamiento. No asumir que una casilla de consentimiento basta ni usar un consentimiento genérico para salud, biometría, transferencias o usos secundarios.
- [ ] Aprobar avisos de privacidad, autorizaciones, tratamiento de datos sensibles, canales de consulta/reclamo, derechos del titular, revocación cuando proceda y mecanismos para acreditar versión/fecha/finalidad.
- [ ] Firmar contratos de tratamiento/encargo y confidencialidad con clientes y proveedores; definir instrucciones, subencargados, incidentes, devolución/eliminación y transferencias internacionales.
- [ ] Definir responsable interno, registro de tratamientos, evaluación de impacto/riesgo para categorías sensibles y proceso para atender solicitudes de titulares.

**Criterio de cierre:** conceptos y documentos aprobados por asesor competente y responsable de privacidad; contratos firmados para el piloto; flujo que bloquea la captura si falta la autorización exigida para esa finalidad.

### 3. Membresía y autorización por empresa, objeto y tenant

- [ ] Implementar una relación explícita y vigente `usuario ↔ tenant ↔ empresa ↔ rol/alcance`; no inferir acceso sólo del rol global ni aceptar `tenant_id`/`company_id` confiado desde URL, body o header.
- [ ] Aplicar denegación por defecto en cada API, worker, job, exportación, almacenamiento de archivos y llamada servicio-a-servicio; autorizar la acción sobre el objeto concreto en cada request.
- [ ] Validar pertenencia de todos los vínculos: trabajador–empresa, evidencia–trabajador/proyecto, incidente–empresa, archivo–registro, examen–persona/empresa, suscripción–tenant y notificación–destinatario.
- [ ] Separar privilegios de profesional SST, empleador, trabajador, IPS, auditor, contratista, soporte y administrador; restringir datos de salud al rol estrictamente necesario y según finalidad.
- [ ] Evitar acceso por UUID conocido, búsqueda/listado, filtros manipulados, cargas, descargas, endpoints legacy, exportaciones y tareas asíncronas.
- [ ] Agregar defensa de base de datos (RLS o equivalente revisado, claves foráneas/restricciones compuestas) y documentar el comportamiento de conexiones/pools al cambiar tenant.

**Criterio de cierre:** suite de integración automatizada con mínimo dos tenants y dos empresas por tenant, usuarios de todos los roles y objetos con IDs conocidos. Cada caso cruzado de lectura, escritura, listado, búsqueda, archivo, exportación, job y endpoint heredado debe devolver rechazo/no revelar existencia. Evidencia: reportes CI repetibles y revisión independiente del modelo de autorización. El test unitario CompanyScope pasó, pero hoy este gate sigue abierto: falta el smoke HTTP/Postgres, membership multiempresa y cobertura completa por endpoint/objeto.

### 4. Ciclo de vida de datos sensibles: acceso, auditoría, conservación y borrado

- [ ] No habilitar exámenes, restricciones, diagnósticos, biometría ni documentos médicos hasta que exista un diseño específico de acceso, segregación, finalidad, retención y eliminación.
- [ ] Registrar de forma inmutable y consultable accesos, lecturas/descargas, cambios, correcciones, consentimientos, revocaciones, exportaciones y acciones administrativas sobre datos sensibles; no guardar valores sensibles en logs.
- [ ] Definir retención por categoría y evento desencadenante, bloqueos por conservación legal cuando aplique, borrado seguro y tratamiento de copias, índices, cachés, colas, logs y backups.
- [ ] Diseñar portabilidad y eliminación con aprobación y trazabilidad; probar la supresión sin dejar copias huérfanas ni borrar registros que deban conservarse.
- [ ] Separar la función de soporte técnico de la autorización para consultar contenido; usar acceso excepcional temporal, aprobado, registrado y revisado.

**Criterio de cierre:** matriz de retención y permisos aprobada; pruebas que demuestren autorización, auditoría y borrado/retención en base, objetos, cachés, logs y backups; funcionalidades destructivas permanecen apagadas hasta aprobarlas.

### 5. Cifrado, secretos y separación de entornos

- [ ] TLS extremo a extremo en tráfico externo e interno pertinente; cifrado de bases, objetos y backups en reposo con llaves gestionadas, rotación y separación de funciones.
- [ ] Secretos fuera de código, imágenes, repositorio, artefactos, logs y `.env` distribuibles; gestor de secretos, permisos mínimos, rotación/revocación y procedimiento de exposición.
- [ ] Separar desarrollo, QA, staging y producción: cuentas, redes, llaves, buckets, bases, colas, logging y proyectos de proveedores distintos. Impedir que staging importe datos productivos sin proceso aprobado y desidentificación.
- [ ] Limitar e inventariar subprocesadores externos (IA, email, storage, analítica); enviar sólo datos mínimos y bloquear proveedores en módulos apagados.

**Criterio de cierre:** configuración productiva revisada; escaneo de secretos/SBOM sin hallazgos críticos abiertos; prueba de TLS/cifrado/rotación; matriz de accesos y evidencia de separación de entornos.

### 6. Archivos, evidencias y documentos

- [ ] Autorización de subida/descarga por tenant, empresa, persona y objeto en servidor; no confiar en nombre, ruta, URL o metadatos del cliente.
- [ ] Validar tamaño, MIME y firma real del archivo; renombrar y almacenar fuera de webroot; bloquear contenido activo; análisis antimalware efectivo y manejo de cuarentena/fallo.
- [ ] Enlaces de descarga temporales, no enumerables, revocables y con auditoría; impedir caché pública y filtración por logs/referrer.
- [ ] Cifrar y retener archivos conforme a la política del registro que respaldan; revisar OCR y extracción de texto para no generar una copia sin control.

**Criterio de cierre:** tests de autorización negativa y de archivos maliciosos/tamaño/tipo; verificación de expiración de enlaces, aislamiento de objetos y eliminación/copia de backups.

### 7. Seguridad de sesión y API

- [ ] Completar rotación/revocación de refresh, cierre de sesión y revocación de llaves; establecer emisor/audiencia, expiraciones y recuperación de cuenta segura.
- [ ] Rate limits y protección contra fuerza bruta, abuso, enumeración y denegación de servicio; CORS y CSRF según cliente; errores que no revelen datos ni estructura interna.
- [ ] Identidad de servicio a servicio, rotación de credenciales y autorización entre servicios; proteger colas, consumidores, admin y rutas de health/metrics que revelen información.
- [ ] Revisar secretos por cuenta de usuario, logs, traces, errores y analytics; redactar identificadores y contenido sensible.

**Criterio de cierre:** pruebas de authz/authn automatizadas, pruebas de abuso y revisión de configuración; tokens revocados no permiten continuar sesiones fuera de la ventana aceptada.

### 8. Recuperación, incidentes y continuidad

- [ ] Backups cifrados, con acceso separado, periodicidad y retención definidas; objetivo de RPO/RTO aprobado para los datos y el servicio.
- [ ] Restauración completa ensayada en entorno aislado; verificar consistencia, control de acceso post-restore y que las eliminaciones/revocaciones respeten la política.
- [ ] Procedimiento de incidente con responsables, contención, preservación de evidencia, evaluación de impacto, comunicación al cliente/titulares y notificaciones que exija el marco aplicable.
- [ ] Alertas operativas para accesos anómalos, fallos de autorización, exportaciones, errores de backup, colas, exposición de archivos y errores de autenticación.

**Criterio de cierre:** acta reciente de restore con tiempos medidos y runbook de incidente probado en simulacro, con responsables y contactos confirmados.

### 9. Integración, despliegue y revisión independiente

- [ ] Ejecutar `docker compose config`, build de imágenes y stack end-to-end desde checkout limpio; probar migraciones nuevas y actualización desde versión anterior.
- [ ] Tests de contrato/integración para Core, AI, Documents, media/archivos, colas y proveedores; revisar workers, reintentos, idempotencia y aislamiento del tenant en procesos async.
- [ ] CI protegida: build/test/lint/scans no pueden fallar silenciosamente; artefactos inmutables, aprobaciones de release, rollback probado y migraciones controladas.
- [ ] SBOM y escaneo de dependencias/imagen/secretos; SAST/DAST y pentest independiente enfocado en IDOR/BOLA, multi-tenant, archivos, sesión y datos sensibles; corregir hallazgos críticos/altos antes del piloto.
- [ ] Verificar métricas, alertas, logs redactados, guardia, soporte y acceso de operadores con privilegio mínimo.

**Criterio de cierre:** pipeline verde desde entorno limpio, build desplegado y restaurable; informe de seguridad firmado con cero hallazgos críticos/altos sin mitigación aceptada; deploy/rollback ensayados.

### 10. Aprobación del piloto controlado con datos reales

Sólo después de cerrar los gates anteriores:

- [ ] Definir una cohorte pequeña, una duración y las categorías mínimas de datos; excluir módulos apagados y datos de alto riesgo no esenciales.
- [ ] Aprobar por escrito responsables, finalidad, usuarios nominados, soporte, retención, procedimiento de revocación/fin del piloto y criterios de interrupción.
- [ ] Capacitar a usuarios; confirmar avisos/autorizaciones individuales antes de recolectar los datos que correspondan.
- [ ] Empezar con un canario limitado, monitorear accesos y errores, revisar bitácora diariamente al inicio y tener mecanismo inmediato para detener ingesta y aislar el tenant.
- [ ] Finalizado el piloto, revisar incidentes, solicitudes, exportación/devolución y borrado conforme a retención aprobada; emitir acta GO/NO-GO para ampliar.

**Criterio de cierre:** acta firmada por dueño de producto, responsable técnico/seguridad, responsable de privacidad/legal y profesional SST; alcance, riesgos residuales, rollback y fecha de revisión documentados.

## Orden práctico recomendado

1. Mantener las pruebas actuales exclusivamente con datos sintéticos.
2. Cerrar CI/build de imágenes/Compose y pruebas de integración.
3. Resolver membership y authorization por objeto; crear y pasar pruebas multiempresa A/B.
4. Completar inventario, base jurídica/avisos/contratos y política de retención con responsables competentes.
5. Implementar y verificar controles técnicos de datos, archivos, auditoría, cifrado, backup/restore e incidentes.
6. Ejecutar revisión externa y remediar hallazgos.
7. Aprobar piloto pequeño y limitado por escrito. No habilitar módulos médicos, IA, biometría o drones por el solo hecho de que el piloto general haya sido aprobado.

## Regla de parada

Un solo fallo de autorización cruzada, exposición de archivo, acceso sensible no auditado, ausencia de base/autorización documentada, restore fallido, secreto productivo expuesto o vulnerabilidad crítica/alta sin decisión formal implica **NO-GO** y pausa de ingesta hasta remediación. La aprobación para datos personales básicos no autoriza automáticamente datos de salud, biometría, IA ni usos secundarios.
