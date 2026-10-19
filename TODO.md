# TODO — piloto sintético funcional ZU Seguridad 360

Actualizado: 19 de octubre de 2026. Los estados distinguen implementación local de verificación E2E y no autorizan uso de datos reales.

## 1. Perfil sintético local, fail-closed en producción
- El modo sintético sólo puede habilitarse mediante configuración del servidor; valores del browser, headers o query no pueden activarlo.
- Al arrancar, el servidor debe rechazar modo sintético con `APP_ENV=production`, endpoint DB fuera de la red local/Compose autorizada, configuración de credenciales productivas o egress sensible habilitado.
- Toda escritura de prueba debe quedar marcada inequívocamente como sintética y los identificadores/emails de test no deben coincidir con personas/empresas reales; los seeds deben crear únicamente registros sintéticos.
- Las funciones de AI externa/autónoma, OCR médico, biometría/visión, drones/autopilot, verificación pública, cobros/pagos y exportación divulgativa deben seguir inaccesibles.
- **Estado: PARCIAL.** Validaciones de ambiente, DB/Redis/Rabbit locales, marcadores sintéticos y bloqueos de rutas están implementados; seed y pruebas locales pasan. No hay política Docker efectiva de bloqueo de egress demostrada y faltan pruebas de despliegue independiente.

## 2. Membership profesional↔múltiples compañías con aislamiento
- Persistir la relación usuario↔compañía con el tenant como límite obligatorio.
- Listar y seleccionar sólo compañías asignadas al usuario; el servidor valida pertenencia en cada solicitud sensible.
- Rechazar con 403 cualquier lectura/escritura A→B, miembro de otro tenant, compañía no asignada o intento de ampliar acceso por `X-Tenant-ID`, `X-Company-ID`, query o body.
- Probar usuario con dos compañías del mismo tenant: sólo las dos memberships autorizadas resultan accesibles. Otro usuario del mismo tenant sin membership y todo usuario de otro tenant deben recibir denegación.
- **Estado: PARCIAL.** Tabla, token de compañía activa, middleware de JWT + membership y prueba unitaria negativa están implementados. `seed.sh` prepara dos usuarios, dos empresas por usuario en tenants aislados y cuatro perfiles sintéticos. El nuevo smoke HTTP cubre membership listing, switch-company y 403 A↔B; sintaxis validada, pero aún no ejecutado porque Docker no está disponible.

## 3. Autenticación y sesión web reales
- `POST /api/core/auth/login` debe validar correo/contraseña y sólo navegar al dashboard después de recibir tokens válidos de Core.
- Mostrar errores genéricos de credenciales/cuenta inactiva; impedir redirect por campos vacíos y estados simulados.
- Renovar token usando refresh token, reintentar una solicitud como máximo una vez y cerrar sesión correctamente sin borrar claves no relacionadas del navegador.
- Mostrar un usuario derivado de datos firmados/servidor sólo con fines de UI; Core sigue siendo autoridad de identidad/roles/scope.
- **Estado: PARCIAL.** Login Core, refresh de una sola repetición, logout y cambio de empresa están cableados; `next build`/typecheck pasan. Faltan pruebas E2E de expiración/revocación/rotación con Redis real y navegador.

## 4. CRUD y contexto multiempresa en dashboard
- El selector del header consulta memberships reales, muestra únicamente empresas autorizadas y cambia el contexto mediante flujo autenticado del servidor.
- Listado, filtros y detalle de empresa leen API, tienen estados loading/error/empty y no usan datos de proyecto hardcoded.
- Formularios de empresa validan el contrato, guardan en Core, previenen envío doble, muestran confirmación/error y refrescan la caché. La eliminación exige confirmación y sólo opera en synthetic mode.
- **Estado: PARCIAL.** Listado, búsqueda, detalle, métricas Core, creación, edición limitada y archivo lógico de compañías están implementados. El servidor valida marcas sintéticas, membership, campos reservados y bloquea el score de riesgo ilustrativo. El switch usa token nuevo. Sin E2E HTTP/UI no se marca completo.

## 5. Trabajadores sintéticos
- Lista, búsqueda, alta, edición, consulta y baja lógica se conectan al Core y filtran por company membership.
- CSV sólo se admite en synthetic mode, valida contenido/columnas/IDs no reales, limita tamaño, reporta errores de fila sin registrar PII e importa transaccionalmente.
- El expediente elimina fotos/biometría/evidencias reales; avatar es generado o placeholder rotulado, y toda fila se marca “SINTÉTICO”.
- **Estado: PARCIAL.** Listado, búsqueda, alta, edición restringida a cargo/área/correo reservado, detalle y baja sintética están implementados; CSV permanece cerrado intencionalmente porque no es posible clasificar sus filas de forma segura. Falta probar contra Core/Postgres y validar baja lógica E2E.

## 6. Incidentes y flujo FURAT de prueba
- Crear, listar, filtrar, ver, investigar y cerrar incidentes sintéticos contra la empresa seleccionada; estados/validaciones y auditoría son visibles.
- FURAT sólo es una previsualización de prueba marcada “NO VÁLIDO — SINTÉTICO”; no se envía a ARL ni se presenta como radicado real.
- El backend valida membership y deniega IDs de incidentes pertenecientes a otra compañía antes de consultar/serializar su contenido.
- **Estado: PARCIAL.** Flujos sintéticos Core/UI implementados; el cierre expone fecha y actor, FURAT es sólo preview y las validaciones rechazan lesionados/datos sin marcador. Falta smoke HTTP de pertenencia de objeto y confirmación E2E de persistencia.

## 7. Capacitaciones y certificados simulados
- CRUD/consulta de cursos, agenda y asistencia funcionan sólo en synthetic mode; identificadores de trabajador/empresa se validan contra membership.
- Certificados generados llevan marca de agua de prueba y no tienen código/URL de validación pública ni afirmación de cumplimiento.
- Las vigencias se muestran como datos de prueba configurables, sin atribuirles una regla normativa no aprobada.
- **Estado: PARCIAL.** Crear/listar cursos y registrar asistencia con trabajadores sintéticos produce constancias explícitamente inválidas y sin endpoint público. No están completos la edición/baja/consulta persistente de constancias y el flujo no se ha probado con Postgres.

## 8. Expedientes médicos sintéticos, sin tratamiento externo
- Vista y edición de exámenes simulados sólo en modo local sintético; toda respuesta lleva advertencia de dato ficticio y no se habilitan diagnósticos OCR, biometría ni procesamiento externo.
- La UI no afirma retención normativa ni diagnóstico/aptitud laboral legal; exámenes reales y consentimientos legalmente válidos siguen fuera de alcance.
- **Estado: NO IMPLEMENTADO / BLOQUEADO.** No se abrieron rutas de exámenes, edición médica ni OCR; se conserva fail-closed hasta contar con diseño de privacidad, autorización, auditoría y aprobación competente.

## 9. PHVA y documentos borrador
- PHVA se habilita sólo cuando cada entidad use tenant/company ownership comprobado y operaciones cruzadas de prueba devuelvan 403.
- Documents implementa autenticación y autorización company-scoped antes de generar/descargar; sólo produce borradores sintéticos con marca de agua y referencias pendientes de revisión.
- OpenAPI/docs públicas, verificación pública y certificados reales permanecen cerrados; documentos sensibles no se almacenan ni se envían a terceros.
- **Estado: BLOQUEADO / NO IMPLEMENTADO.** PHVA y Documents no se habilitaron en el piloto; sus APIs públicas/sensibles siguen cerradas. Se requieren pruebas de ownership y revisión normativa independiente.

## 10. Interfaz completa sin mocks ni botones muertos
- Reemplazar todos los KPIs, proyectos, agentes, trabajadores, exámenes, incidentes, certificados y facturas fijos por información autorizada o eliminarla.
- Implementar acciones visibles permitidas con API y errores recuperables; para cualquier acción fuera de alcance, mostrar desactivación y motivo, sin simular éxito.
- Búsqueda, filtros, paginación, navegación, notificaciones, ajustes y selector de compañía deben funcionar o retirarse del menú.
- Banner persistente en el dashboard: “ENTORNO SINTÉTICO — NO USAR DATOS REALES”.
- **Estado: PARCIAL.** Inicio, compañías, trabajadores, incidentes y capacitación consumen Core y tienen estados de carga/error/vacío; módulos de alto riesgo muestran bloqueo explícito y no simulan éxito. Falta auditar cada ruta/acción del menú y realizar pruebas de interacción E2E.

## 11. Dependencias, builds y evidencia
- Definir package manager/version, crear lockfile y usar instalación congelada en `frontend/web`.
- Corregir scripts Next/TS/lint y añadir pruebas de formularios/login/errores/denegación; compilar la app con `next build`.
- Ampliar seeds y smoke/integration tests con dos profesionales sintéticos, varias compañías por tenant y accesos cruzados; ejecutar contra PostgreSQL en entorno local seguro.
- Actualizar runbook y estado Gate 1/2. No declarar producción GO ni datos reales habilitados.
- **Estado: PARCIAL.** Lockfile versionado; `make test` pasó (10 módulos Go, 42 pruebas AI, 10 pruebas Documents y guard Docker simulado), `make build`, `npm run typecheck` y `npm run build` pasaron; smoke y setup tienen sintaxis válida. Runbook actualizado. Docker no existe en el sandbox, por lo que integración/Postgres y validación de browser quedan pendientes.

## 12. Autorregistro real de profesional SST con Google
- Mostrar “Continuar con Google” sólo cuando `GOOGLE_SIGNUP_ENABLED=true` y exista un Google Web Client ID configurado; la opción permanece deshabilitada en `PILOT_SYNTHETIC_ENABLED` y por defecto apagada.
- Verificar en Core la firma, audiencia, emisor, expiración, `email_verified` y nonce de un solo uso del ID token. Usar `sub` de Google como identificador federado único; no vincular automáticamente una cuenta por correo coincidente.
- En primer registro, solicitar datos requeridos de la empresa inicial y aceptación versionada de términos/aviso de privacidad; crear atómicamente un tenant aislado, usuario con rol `professional_sst`, empresa no sintética y membership inicial. No activar plan/cobro automáticamente.
- Dejar clase de riesgo pendiente de verificación; no asignar valores ARL mediante heurística. Mostrar errores genéricos ante token inválido, correo ya registrado o conflictos de NIT.
- Configurar orígenes HTTPS, Client ID, URLs/versiones vigentes de términos y privacidad, migración SQL revisada y egress de Google antes de habilitar registros. El cambio no elimina los gates abiertos de producción ni habilita módulos médicos, IA, biometría o drones.
- **Estado: IMPLEMENTADO EN CÓDIGO; DESHABILITADO PARA USO.** Core verifica ID token/nonce y crea el tenant inicial transaccional; UI GIS y tests de validación/local compilan. `GOOGLE_SIGNUP_ENABLED` queda `false` por defecto y el synthetic guard lo rechaza. Faltan migración aplicada a una BD limpia, Client ID/orígenes de Google Cloud, documentos legales publicados/aprobados, staging, Redis/Postgres reales y E2E/browser independientes. No es autorización para producción ni carga de datos de trabajadores.
