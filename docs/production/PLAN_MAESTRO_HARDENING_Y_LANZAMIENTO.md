# Plan maestro de hardening y lanzamiento de ZU Seguridad 360

## Decisión de salida hoy

**NO-GO para producción o facturación con clientes reales.** El estado inicial auditado no supera los gates de build, aislamiento, autorización, billing y operación. El checklist ampliado recibido se conserva en `PRODUCTION_CHECKLIST_AMPLIADO_INPUT.md` y sigue siendo la especificación objetivo; en este plan no se marca como completado nada que no tenga cambio y evidencia de prueba.

## Secuencia de trabajo

### Fase 0 — Congelar base, alcance y recuperación

- [ ] Mantener intactos los ZIP de entrada y trabajar sólo sobre la copia versionable `zu-seguridad-360/`.
- [ ] Establecer rama/commit inicial y registrar hashes de los archivos de entrada.
- [ ] Definir release MVP: core SST multiempresa + web/portal; posponer cobros reales, IA de decisión, biometría y vuelos.
- [ ] Retirar credenciales de desarrollo del artefacto y asegurar que ninguna sea reutilizada en sistemas reales.
- [ ] Levantar inventario de endpoints, roles, datos, dependencias y features por plan.

**Criterio de cierre:** alcance firmado, secretos rotados cuando corresponda, entorno de prueba aislado y fuente original recuperable.

### Fase 1 — Compilación e identidad

- [ ] Corregir errores de build en todos los módulos; generar/fijar lockfiles (`go.sum`, Python y frontend).
- [ ] CI limpio: format, lint, build, pruebas unitarias, import-smoke Python, análisis de dependencias, build de contenedores.
- [ ] JWT de tipo de uso correcto, claims obligatorios, algoritmo fijado, expiración y membership validadas; refresh con rotación/revocación.
- [ ] El tenant sólo se deriva de identidad validada y membership en servidor; un header/body/path no puede ampliar alcance.
- [ ] RBAC deny-by-default por endpoint, recurso, acción y empresa.

**Criterio de cierre:** todos los servicios del alcance compilan en CI; toda ruta tiene contrato de autenticación y autorización documentado y probado.

### Fase 2 — Aislamiento y datos SST

- [ ] Corregir todas las consultas/mutaciones cross-tenant, incluidas exportación, evidencias, PHVA, zonas, proyectos y datos médicos.
- [ ] Verificar cada referencia padre/hijo en transacción.
- [ ] Añadir claves, índices y constraints de pertenencia; adoptar RLS si encaja con el modelo y se prueba con el rol de DB de producción.
- [ ] Pruebas negativas con tenants A/B y UUIDs controlados en todos los recursos y canales HTTP, SSE y mensajería.
- [ ] Definir permisos médicos mínimos, exportación segura, retención, borrado, auditoría y acceso interno.

**Criterio de cierre:** cero accesos cruzados en pruebas automatizadas y demostración independiente desde rutas HTTP y consultas de repositorio.

### Fase 3 — Producto real y operación móvil

- [ ] Sustituir login fake, datos fijos y `setTimeout` por autenticación y mutaciones reales.
- [ ] Completar portal profesional multiempresa y portal de cliente separado por membership.
- [ ] Implementar offline-first: base local cifrada, outbox local, UUID/idempotency key, estados de sincronización, resolución de conflictos y prueba de pérdida de conectividad.
- [ ] Compilar Android/iOS, probar permisos, privacidad, accesibilidad y política de publicación en Google Play/App Store.
- [ ] Configurar media con auth, tenant scope, escaneo real, límites, storage privado y URL firmada temporal.

**Criterio de cierre:** flujos end-to-end desde navegador y dispositivo; no hay confirmación hasta guardar en backend.

### Fase 4 — Planes, licencia profesional y billing

- [ ] Modelar productos/planes y entitlements versionados; aplicar límites de empresas gestionadas por profesional, seats, almacenamiento, automatizaciones y features en server-side.
- [ ] Probar transiciones active/trial/past_due/suspended/cancelled/grace/renewal/up/downgrade.
- [ ] Implementar webhook firmado, deduplicación e idempotencia por evento y período, reconciliación, recibos y outbox/retry/DLQ.
- [ ] Evitar que el cliente suministre tenant/company para seleccionar la cuenta cobrada.
- [ ] Verificar proceso, alcance y vigencia de la licencia profesional si se promociona como requisito de elegibilidad.
- [ ] Validar precios y disposición a pagar en Colombia. Los valores US$70/130/300 del checklist son hipótesis y no plan publicado; evaluar cobro COP/USD, IVA, costos, soporte, IA y margen.

**Criterio de cierre:** pagos de sandbox reconciliados, cambios de plan probados y límites aplicados antes de habilitar cobro real.

### Fase 5 — Integración y gobierno de skills/IA

- [ ] Mantener el paquete v3 como fuente versionada, no como código ejecutable ni prompt de confianza automática.
- [ ] Revisar normativamente y técnicamente cada skill; documentar versión, fuente, fecha y responsable.
- [ ] Separar RAG público global (norma oficial) de datos/feedback privados por tenant.
- [ ] Crear herramientas por acción con identity propagation, ACL y validación objeto; desactivar tools que no tengan endpoint verificado.
- [ ] Borrador IA con fuentes/abstención; aprobación humana obligatoria para salidas médicas, documentos oficiales, firma, notificaciones y acciones sensibles.
- [ ] Feedback guardado bajo tenant y sin entrenamiento cruzado automático.
- [ ] Feature flags y kill switch por modelo/skill/tenant; evaluación offline, límites, presupuesto y auditoría.

**Criterio de cierre:** no existe ruta de ejecución sólo por prompt; cada acción requiere permiso server-side y los datos no cruzan tenants.

### Fase 6 — Infraestructura y go-live

- [ ] Completar Terraform/Kubernetes, nombres de imagen, overlays y health/readiness; validar ruta limpia en staging.
- [ ] TLS externo, servicios internos privados, secretos gestionados y rotados, controles de acceso mínimos.
- [ ] Migraciones versionadas con lock, historial y operación desde DB vacía; no ocultar fallos.
- [ ] Logs, métricas, alertas, trazas, SLO, runbooks y procedimientos de seguridad.
- [ ] Backups cifrados y restauración periódica demostrada; aprobar RPO/RTO.
- [ ] SBOM, escaneo con gate, artefactos inmutables y release por digest probado; rollback ensayado.
- [ ] Pruebas de carga y aislamiento de ruido entre tenants.

**Criterio final GO:** gates P0/P1 aplicables con pruebas adjuntas, sin riesgo crítico abierto, responsables y rollback disponible. IA/biometría/drones sólo se habilitan tras sus gates separados.

## Comercialización y diferenciación

La comparativa pública documenta capacidades ya anunciadas por seis proveedores colombianos. No se debe basar el posicionamiento sólo en “tener IA”, “tener 20 módulos” o “ser multiempresa”; otras ofertas públicas usan claims similares. Hipótesis de diferenciación a validar: experiencia superior para profesional SST con portafolio, onboarding masivo, operación móvil offline confiable, trazabilidad explicable de IA con control humano y pricing transparente con límites medibles.

- [ ] Entrevistar al menos 10 profesionales SST independientes/consultoras y 10 PYMES/empresas objetivo.
- [ ] Medir tareas que más tiempo consumen y qué datos migran desde Excel.
- [ ] Validar willingness-to-pay y moneda con prototipos de planes; separar disposición a pagar por empresas, trabajadores, módulos y almacenamiento.
- [ ] Demo comparativa estandarizada con competidores usando el mismo guion; documentar evidencia observada versus declaración comercial.
- [ ] Seleccionar un piloto cerrado sin datos médicos reales ni funciones de riesgo hasta cerrar aislamiento y consentimiento.

## Estado de trabajo en esta copia

- [x] ZIP base extraído y preservado en una copia de hardening separada.
- [x] Paquete v3 recibido, inventariado y copiado sin ejecutarlo.
- [x] Checklist ampliado recibido guardado como referencia de requisitos.
- [x] Comparativa inicial de seis ofertas públicas colombianas preparada con fuentes oficiales de proveedores y una guía independiente.
- [x] Política inicial de integración de skills documentada; el paquete queda inactivo hasta revisión.
- [x] Primera pasada de hardening; los 10 módulos Go pasan `go test`, `go vet` y `make build`; los 39 tests AI pasan; `go.sum` se regeneró y verificó tras limpiar caché en todos los módulos.
- [x] CompanyScope con company_id derivado/verificado en JWT, rutas piloto acotadas y test unitario de denegación A→B (403); seed preparado para dos empresas/tenants sintéticos y script smoke HTTP creado.
- [ ] CI en GitHub, build de imágenes Docker/Compose, lint y pruebas de integración.
- [ ] Ejecutar setup con Postgres + Core y `make test-company-isolation`; completar pruebas A/B en todos los endpoints/objetos (incluyendo dos empresas del mismo tenant) y membership multiempresa.
- [ ] Producción habilitada.

Resultados detallados: `docs/testing/RESULTADOS_PRUEBA_LOCAL_2026-10-17.md`. Billing y módulos sensibles siguen deshabilitados; no se considera un piloto con datos reales.

**No se declara el sistema listo para producción.** Esta copia de trabajo no se ha desplegado en un servidor/VPS, las credenciales de producción no se suministraron y aún deben cerrarse los gates pendientes.
