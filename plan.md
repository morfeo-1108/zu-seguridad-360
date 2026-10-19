# Plan aprobado: piloto sintético funcional de ZU Seguridad 360

**Aprobado por el usuario:** 19 de octubre de 2026.  
**Objetivo:** habilitar un piloto local y aislado que ejercite flujos SST reales de la aplicación con datos exclusivamente sintéticos, sin abrir producción ni afirmar cierre de Gates 1/2.

## Requisitos y límites aprobados

- El login web autentica contra Core; no entra por un submit vacío.
- Cada acción permitida del dashboard ejecuta un flujo real contra servicios y conserva estados de carga, validación, éxito y error. Las acciones aún fuera de alcance quedan claramente bloqueadas, sin datos ni respuestas ficticias que aparenten ejecución.
- El perfil de datos sintéticos puede habilitar compañías y membresías profesional↔varias empresas, trabajadores de prueba, incidentes, capacitaciones/certificados simulados, expedientes médicos simulados, PHVA y documentos de borrador dentro de una base local dedicada y sin egress a proveedores.
- Todos los registros de ese perfil llevan clasificación inequívoca de datos sintéticos; IDs/documentos de prueba son no válidos y siguen un marcador de test. No se usan nombres, cédulas, NIT, teléfonos, correos o historias de personas/empresas reales.
- El servidor rechaza el perfil sintético si la aplicación no está en un entorno permitido de desarrollo/pruebas, si el host de base de datos no es local/interno a Compose o si hay configuración de producción. No basta un flag enviado desde el browser.
- Multiempresa se autoriza por pertenencia persistida y comprobada por `tenant_id`, `user_id` y `company_id`; no se confía en headers/query/body para ampliar permisos. Pruebas negativas deben demostrar 403 entre empresas y tenants.
- Mantener apagados durante este piloto: proveedores IA externos/autónomos, OCR médico, visión/biometría, drones/autopilot, verificación pública de certificados, pagos/cobros, y cualquier exportación que pueda divulgar información. Documentos normativos sólo como borradores no certificados, a menos que el contenido haya sido revisado por un profesional competente.
- No introducir datos reales ni publicar/desplegar. Gate 1 y Gate 2 continúan NO-GO hasta contar con pruebas E2E independientes y aprobaciones externas.

## Diseño y experiencia

Se conserva el lenguaje visual ZU ya presente: navegación compacta, superficies claras, acento verde, estados de riesgo legibles y tablas densas. Se retiran mocks de personas, agentes, proyectos, facturas y resultados médicos; se agrega una franja persistente de “ENTORNO SINTÉTICO — NO USAR DATOS REALES”. Los flujos de prueba deben ser comprensibles sin ocultar sus límites normativos.

## Enfoque técnico

1. **Línea base:** comprobar fuente original, cambios Gate 2 y scripts; recuperar/implementar métodos de servicios cerrados desde contratos existentes, sin confiar en archivos de entorno.
2. **Modo sintético de backend:** configuración segura validada al arranque, clasificación/validación de datos de prueba y escritura sólo a DB local dedicada; mantener barrera absoluta en producción y bloquear llamadas externas.
3. **Membership multiempresa:** tabla/relación explícita usuario↔compañía, tenant-scoped; middleware y endpoints de selección de compañía verifican pertenencia antes de leer/escribir. Conservar claims JWT verificadas y no permitir que el cliente los sustituya.
4. **Flujos Core sintéticos:** reabrir, sólo bajo modo sintético, CRUD de compañías/miembros, trabajadores, incidentes, capacitación/certificados simulados, exámenes sintéticos y PHVA. Las acciones no disponibles por depender de proveedores, revisión normativa o infraestructura externa quedan deshabilitadas.
5. **Documentos:** añadir autenticación/scope de compañía al servicio; habilitar únicamente borradores sintéticos y retirar normativa/afirmaciones no validadas. No exponer descarga/verificación pública.
6. **Dashboard:** corregir autenticación, selector multiempresa basado en membership real, KPIs desde API, búsqueda/filtros/paginación y formularios conectados; retirar datos fijos y conectar o desactivar cada botón.
7. **Dependencias frontend:** fijar package manager y lockfile, corregir scripts Next.js, ejecutar build/lint/typecheck y añadir pruebas de interacción. Los cierres de alto riesgo deben tener tests positivos en synthetic mode y negativos en producción.
8. **Pruebas de aislamiento:** ampliar seed sintético para dos profesionales y varias compañías de prueba por tenant; cubrir list/get/create/update/delete y búsquedas por compañía, más escenarios A→B denegados; ejecutar contra Postgres real en entorno local controlado o documentar cada caso no ejecutable.

## Estructura del proyecto existente

- `frontend/web/app/`: login y 20 rutas de dashboard; se conectarán a los endpoints reales o mostrarán cierre explícito.
- `frontend/web/components/`: shell, navegación, formularios/estados y banner sintético compartidos.
- `frontend/web/lib/`: cliente API, hooks, tipos y stores; el contexto de empresa se basará en memberships del servidor.
- `services/core/internal/`: handlers, servicios, modelos, repositorios y middleware para auth, scope, membership, datos sintéticos y flujos Core.
- `services/core/migrations/` y `scripts/`: cambios de schema y seed sólo sintético/local.
- `services/documents/`: API autenticada/scoped y generación de borradores sintéticos.
- `docs/production/` y `docs/testing/`: límites de gates, procedimientos y resultados verificables.

## Límite de la afirmación final

“Funcional” significará que los flujos autorizados se ejecutan end-to-end con registros sintéticos en el entorno local aislado, con aislamiento negativo probado y una experiencia web coherente. No significará que se puedan usar datos reales, que los documentos tengan validez normativa/legal, que AI/biometría/drones estén disponibles, ni que el sistema esté listo para producción o licenciamiento.

## Enmienda aprobada — registro real de profesionales con Google

**Aprobada por el usuario:** 19 de octubre de 2026, en una configuración separada del piloto sintético. El usuario seleccionó el autoregistro con tenant y empresa inicial.

- El login ofrecerá Google Identity Services cuando `GOOGLE_SIGNUP_ENABLED=true`, exista un Client ID Web válido y el entorno no esté en modo sintético. El flag será `false` por defecto y la configuración sintética rechazará explícitamente su activación.
- Core verificará criptográficamente el ID token, `aud`, `iss`, `exp`, correo verificado y un nonce de un solo uso; identificará la cuenta únicamente por el `sub` de Google. No se vinculará una cuenta existente sólo porque coincida el correo.
- En el primer registro, la persona completará los datos de su empresa inicial y aceptará las versiones configuradas de términos y aviso de privacidad. Una transacción creará usuario `professional_sst`, tenant nuevo, empresa no sintética y membership inicial. No se asigna un plan ni se inicia un cobro automáticamente.
- Empresas ingresadas por autorregistro no recibirán clase ARL heurística: la clasificación queda pendiente de validación. No se abren por este cambio módulos médicos, biométricos, IA, drones ni documentos de cumplimiento.
- La activación requiere Client ID y orígenes HTTPS autorizados, URLs públicas vigentes de términos/privacidad, versiones de documentos, política de egress para la verificación de Google y migración SQL revisada. Esta enmienda no cambia el NO-GO existente para datos laborales sensibles o producción.
