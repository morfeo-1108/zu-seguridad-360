# Verificación normativa inicial — Colombia

**Consulta:** 17 de octubre de 2026. Este documento registra hallazgos de referencias, no es concepto jurídico ni sustituye revisión normativa completa por profesional competente.

## Hallazgos confirmados en textos normativos consultados

1. **Trabajo en alturas:** la Resolución 4272 de 2021 fija requisitos mínimos de seguridad para trabajo en alturas; el texto consultado en el Sistema de Información Jurídica de Bogotá indica expedición 27-dic-2021 y entrada en vigencia 08-ago-2022. Su artículo 68 dice expresamente que deroga la Resolución 1409 de 2012, entre otras. Por tanto, referencias operativas del código/skills a Resolución 1409/2012 no deben presentarse como la regla vigente; requieren migración y revisión de cada detalle técnico. El propio portal advierte que su información es informativa y que la vigencia depende del análisis/competencia legal. [1]
2. **Evaluaciones médicas ocupacionales:** la Resolución 1843 de 2025, expedida 29-abr-2025, regula la práctica y el manejo/contenido de evaluaciones médicas ocupacionales. El artículo 36 del texto consultado deroga expresamente las Resoluciones 2346/2007, 1918/2009, 1075/1992 y 4050/1994; el artículo 35 fija seis meses de transición. Por tanto, el skill `zu360-examenes` que usa Res. 2346/2007 como “marco normativo” vigente está obsoleto y se debe retirar de uso productivo hasta actualizar reglas y revisar su tratamiento de conceptos de aptitud, acceso y conservación. [2]
3. El Ministerio de Trabajo publica una página de centros de entrenamiento que hace referencia a la implementación de la Resolución 4272/2021 y a formación/actualización. La página también conserva instrucciones históricas de 2020; no reutilizarlas como reglas actuales sin revisar su fecha/contexto. [3]
4. El Ministerio indexa la Resolución 1843 de 2025 en su página de resoluciones; el extractor devuelve un banner de cookies, por lo que se usó el texto normativo del sistema jurídico de Bogotá para leer artículos y no se atribuyen al índice detalles que éste no mostró. [4]

## Consecuencias técnicas inmediatas

- Mantener deshabilitado o en borrador cualquier checklist que cite Res. 1409/2012 o 2346/2007 como regla vigente.
- No ejecutar bloqueo de tareas, clasificar “apto/no apto”, calcular periodicidades, definir exámenes, validar licencias, ni atribuir 20 años de retención por una instrucción del modelo. Cada regla requiere fuente, artículo, fecha, población aplicable, estado y aprobación competente.
- Añadir `source_id`, versión/fecha de vigencia, jurisdicción, responsable de revisión y fecha de revisión a reglas del motor; usar versionado y un proceso de actualización controlado.
- En el código, referencias a la Resolución 1409/2012 están en plantillas de protocolos de alturas/zona (`services/core/internal/service/zone_service.go` y comentarios de modelos). No sustituir automáticamente texto por número nuevo: auditar equipos, alturas, training, permisos y protocolos completos antes de reabrir el módulo.
- El skill médico incluye frecuencias por clase de riesgo, plazos y formatos de aptitud/restricciones; no se han validado todos contra el texto 1843/2025. Mantenerlo inactivo hasta una revisión por SST/salud ocupacional y privacidad.

## Fuentes

[1]: https://www.alcaldiabogota.gov.co/sisjur/normas/Norma1.jsp?i=120880 — Resolución 4272 de 2021, texto con fecha de vigencia y Art. 68 (derogatoria).
[2]: https://www.alcaldiabogota.gov.co/sisjur//normas/Norma1.jsp?i=178947 — Resolución 1843 de 2025, texto con artículos 35 y 36.
[3]: https://www.mintrabajo.gov.co/empleo-y-pensiones/movilidad-y-formacion/centros-de-entrenamiento — MinTrabajo, centros de entrenamiento; referencia a Resolución 4272.
[4]: https://www.mintrabajo.gov.co/atencion-al-ciudadano/transparencia/resoluciones — Índice de resoluciones del Ministerio de Trabajo.

## Hallazgo adicional — espacios confinados

5. La Resolución 491/0491 de 2020 establece requisitos mínimos de seguridad para trabajos en espacios confinados; su texto consultado indica entrada en vigor el 24-ago-2020 y contiene una modificación expresa del artículo 42 por la Resolución 2605 de 2020. La página oficial de MinTrabajo también remite a la Resolución 0491/2020 y habla de correcciones/modificaciones, por lo que deben revisarse las versiones consolidadas antes de usar una regla operativa. [5][6]
6. El código actual cita “Res. 4276/2022” en plantillas de zonas para espacios confinados. No se encontró soporte oficial de esa referencia en las consultas realizadas; se clasifica como **no verificada**, no como norma vigente. Las rutas de zonas/checklists se retiraron del arranque en este hardening.

[5]: https://www.alcaldiabogota.gov.co/sisjur/normas/Norma1.jsp?i=90906 — Resolución 491 de 2020, texto con vigencia y modificación del art. 42 por Res. 2605/2020.
[6]: https://www.mintrabajo.gov.co/marco-legal — Índice oficial consultado; el extractor devolvió banner de cookies, no texto suficiente para identificar el acto de corrección. La referencia requiere confirmar número y texto en una consulta normativa completa.
