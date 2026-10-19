# Opinión técnica profesional — ZU Seguridad 360

**Fecha de corte: 17 de octubre de 2026**  
**Conclusión ejecutiva: NO-GO para producción, comercialización por licencias o datos reales sensibles en el estado actual.** Esto no significa que el producto no tenga potencial; significa que la amplitud funcional supera el nivel de evidencia, seguridad y operación disponible para venderlo responsablemente hoy.

## Diagnóstico

ZU Seguridad 360 apunta a un problema real: permitir que un profesional SST administre varias empresas, conserve trazabilidad de ejecución y reduzca trabajo administrativo. Esa propuesta —“operación SST de portafolio para profesionales y consultores”— es más defendible que prometer cumplimiento legal automático.

La arquitectura tiene ambición amplia: múltiples servicios, módulos de SST, IA/RAG, integración documental, eventos, billing y conceptos de drones/visión. El riesgo actual es que algunas capacidades están representadas en modelos, rutas, prompts o documentación, pero no todas tienen autorización coherente, pruebas de extremo a extremo, actualización normativa, operación medible o comportamiento real certificado. Hay que distinguir con rigor entre **código presente**, **ruta disponible**, **capacidad integrada**, **capacidad validada** y **garantía comercial**; hoy no son equivalentes.

## Hallazgos que bloquean salida

1. **Aislamiento multiempresa: primera barrera añadida, gate todavía abierto.** Los JWT ahora incluyen company_id derivado de usuario y validado contra el tenant; middleware devuelve 403 si un token de A accede a la ruta de B, y el test comprueba que header/query/body no amplían alcance. Rutas de empresas, empleados, incidentes y capacitaciones tienen scope; queries sensibles filtran empresa+tenant. Aún faltan membresías para portafolios multiempresa, tests con DB y cobertura de cada recurso/archivo/job. Por eso evidencia, PHVA por objeto y eventos de agente quedaron sin registrar. Roles de empleador/trabajador/IPS/auditor/contratista continúan sin grants.
2. **Build y pruebas: avance local, integración pendiente.** Los diez módulos pasan `go test ./... -count=1`, `go vet ./...` y `make build`; los 39 tests AI pasan. `go.sum` fue regenerado/verificado tras limpiar la caché Go y todos los módulos compilaron. Esto no demuestra migraciones, imágenes, CI GitHub, Compose, web/móvil ni integración con servicios externos. El sandbox no tiene Docker y no ejecutó la prueba HTTP A/B.
3. **Autenticación y sesiones requieren cierre.** Se aplicó una primera defensa de tokens/claims y refresh, pero faltan rotación/revocación de refresh, revocación inmediata de access, rotación de llaves y pruebas de sesión; logout por sí solo no invalida un access ya emitido.
4. **Contenido normativo de alto riesgo desactualizado o no verificado.** El código citaba Resolución 1409/2012 para alturas; el texto consultado de la Resolución 4272/2021 la deroga. Skills de exámenes usaban Resolución 2346/2007, derogada por Resolución 1843/2025. Plantillas de espacio confinado citaban “4276/2022” sin respaldo oficial identificado; el marco consultado es Resolución 491/0491 de 2020, con cambios que también deben revisarse. Se deshabilitaron rutas de protocolos hasta validar artículo por artículo.
5. **Datos personales y de salud.** Exámenes, restricciones, incidentes, fotografías y posiblemente biometría requieren autorización, minimización, retención/borrado, cifrado, registro de acceso, consentimiento y acuerdos contractuales revisados. Las rutas de exámenes y supresión/anonimización están deshabilitadas; exportación general y verificación pública de certificados siguen retiradas. El consentimiento es parcial y aún no bloquea todas las operaciones sensibles.
6. **Herramientas IA/integraciones.** El paquete nuevo de skills queda como referencia con `approved_for_runtime=false`; además, el registro ejecutable existente en Python expone herramientas que no tenían de forma uniforme auth servicio-a-servicio y autorización por empresa. Se dejó el registro de herramientas vacío por defecto hasta auditarlo y probarlo.
7. **Licencias por planes no comprobadas.** El sistema no demuestra aún entitlements autoritativos por plan ni un ciclo de cobro/suspensión/reconciliación completo. La existencia de esquemas de billing no significa que la monetización esté implementada.
8. **Operación productiva pendiente.** No se verificaron backups y restauración, RPO/RTO, disponibilidad, escalado, alertas, respuesta a incidentes, pruebas de carga, pentest, CI/CD ni soporte contractual.

## Trabajo de hardening ya aplicado en la copia de trabajo

Se endurecieron JWT y refresh, se retiraron headers de tenant/company como fuente de contexto, se incorporó validación fail-fast de defaults inseguros y Casbin detiene el arranque si falla. El scope empresa cubre las rutas de empresa/empleados/incidentes/capacitaciones; la prueba unitaria negativa A→B pasa. PHVA conserva cálculo conservador, pero sus rutas por objeto quedan sin registrar hasta probar ownership de extremo a extremo. También se retiraron evidencia, eventos de agente, exportación genérica, certificados públicos, rutas médicas, borrado de datos, Billing financiero, zonas/checklists y operación de drones/autopilot. AI API, visión, biometría/OCR permanecen cerrados por flags.

Estas son medidas de defensa y transparencia, **no** una certificación ni una habilitación para datos productivos. El resumen de los cambios, pruebas y pendientes está en [Estado de hardening](docs/production/ESTADO_HARDENING_2026-10-17.md).

La evidencia ejecutada se resume en [Resultados de pruebas locales](docs/testing/RESULTADOS_PRUEBA_LOCAL_2026-10-17.md); no se levantó Docker/Compose, por lo que ni el seed PostgreSQL ni el smoke HTTP A/B se ejecutaron.

## Diferenciación y mercado

La comparación de productos comerciales y opciones de ARL está documentada en [Comparativa SG-SST Colombia](docs/market/COMPARATIVA_COMPETIDORES_SG-SST_COLOMBIA.md). Se basa en material publicado, no en una matriz de demos/pruebas ni en precios negociados. No permite afirmar superioridad funcional frente al mercado.

La diferenciación potencial es combinar **multiempresa para el profesional**, **trazabilidad auditable** y **automatización asistida**. Para convertir esa tesis en ventaja competitiva se deben medir, en piloto, tiempo de incorporación por empresa, tasa de cierre oportuno, trazabilidad recuperable, precisión y límites de IA, incidencias de privacidad y disposición a pagar. Drones/visión deben permanecer como experimento controlado hasta pasar validación de seguridad física, permisos, confiabilidad y aprobación humana.

## Recomendación

No vender todavía “planes productivos” ni cargar historias médicas reales. Ejecutar gates en este orden: (1) build/CI limpia y pruebas; (2) scope empresa–usuario/objeto y pentest; (3) privacidad y revisión normativa; (4) entitlements y pagos; (5) backups, incidentes y observabilidad; (6) piloto cerrado con datos sintéticos y luego datos reales bajo controles; (7) evaluación independiente y lanzamiento gradual.

La postura técnica responsable es invertir primero en **seguridad, evidencia y límites de producto**, no en sumar más módulos. El potencial comercial se conserva; las promesas de cumplimiento, automatización normativa y producción deben esperar a evidencia verificable.
