---
name: zu360-capacitaciones
description: >
  Skill para gestionar capacitaciones y certificados SST en ZU 360.
  Activa cuando el usuario mencione: capacitación, curso, certificado, trabajo en alturas,
  espacios confinados, primeros auxilios, extintores, inducción, re-inducción, COPASST
  formación, e-learning, evaluación de conocimiento, asistencia, vencimiento de certificado,
  programa anual de capacitaciones, o cualquier actividad formativa SST.
---

# ZU360 — Módulo de Capacitaciones y Certificados

## Cursos Normativos Obligatorios Colombia

| Curso | Norma | Frecuencia | Aplica a |
|---|---|---|---|
| Trabajo en Alturas | Res. 1409/2012 | Anual | Expuestos a alturas ≥1.5m |
| Espacios Confinados | Res. 4276/2022 | Según programa | Entrantes, vigías, supervisores |
| Primeros Auxilios | Dec. 1072/2015 | Bienal | Brigadistas |
| Manejo Extintores | Dec. 1072/2015 | Anual | Brigadistas y personal en general |
| Riesgo Eléctrico / LOTO | Res. 90708/2013 | Anual | Expuestos a riesgo eléctrico |
| Sustancias Peligrosas | Dec. 1496/2018 | Anual | Manejadores de químicos |
| Ergonomía | Dec. 1072/2015 | Anual | Todo el personal |
| Inducción SST | Dec. 1072/2015 | Ingreso | Trabajadores nuevos |
| Re-inducción SST | Dec. 1072/2015 | Anual | Todo el personal |
| COPASST | Dec. 1072/2015 | Al elegir | Miembros COPASST |
| Brigadas Emergencia | Dec. 1072/2015 | Semestral | Brigadistas |

## Flujo de Registro de Capacitación

### Datos Obligatorios
```
- Empresa y NIT
- Nombre del curso
- Norma habilitante
- Fecha de realización
- Duración (horas)
- Modalidad: presencial / virtual / mixta
- Instructor: nombre, empresa, licencia (si aplica)
- Lista de asistentes: nombre, cédula, cargo, firma
- Evaluación: nota mínima aprobatoria (≥60%)
- Evidencias: fotos, material de capacitación
```

### Certificado Generado — Elementos Obligatorios
```
✓ Logo empresa capacitadora + empresa beneficiaria
✓ Nombre completo del trabajador
✓ Número de cédula
✓ Nombre del curso
✓ Norma que lo habilita
✓ Duración en horas
✓ Fecha de realización
✓ Fecha de vencimiento
✓ Firma del instructor (con licencia si aplica)
✓ Firma Representante Legal
✓ Folio UUID único
✓ Código QR de verificación pública
```

## Programa Anual de Capacitaciones

Estructura del PAC por empresa:
- Enero–Diciembre con curso por mes
- Responsable, recursos, costo estimado
- Indicador de cumplimiento al cierre del año

## Alertas de Vencimiento

```
30 días antes → Notificación al trabajador y supervisor
15 días antes → Segunda notificación + alerta a profesional SST
7 días antes  → Alerta urgente
0 días        → Vencido — bloqueo de tareas incompatibles
```

## Archivos Generados

1. **Certificado_[Nombre]_[Curso]_[Fecha].pdf** — Con QR verificable
2. **Lista_Asistencia_[Curso]_[Fecha].xlsx** — Con firmas y evaluaciones
3. **Programa_Anual_Capacitaciones_[Empresa].xlsx** — Cronograma completo
4. **Informe_Cobertura_Capacitaciones.pdf** — Por empresa y período
