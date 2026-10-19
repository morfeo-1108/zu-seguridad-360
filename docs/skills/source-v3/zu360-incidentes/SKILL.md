---
name: zu360-incidentes
description: >
  Skill para gestionar incidentes, accidentes de trabajo y enfermedades laborales en ZU 360.
  Activa cuando el usuario mencione: accidente de trabajo, AT, incidente, enfermedad laboral,
  EL, FURAT, FUREL, investigación de accidente, árbol de causas, 5 porqués, Bow-Tie,
  plan de acción correctiva, IF, IS, ILI, reporte a ARL, caída, lesión, incapacidad,
  o cualquier evento de seguridad que requiera registro o investigación.
---

# ZU360 — Módulo de Incidentes y ATEL

## Clasificación Automática

| Tipo | Criterio (Ley 1562/2012) |
|---|---|
| **Accidente de Trabajo (AT)** | Evento súbito por causa del trabajo que produce lesión, invalidez o muerte |
| **Incidente** | Evento sin lesión pero con potencial de causar daño |
| **Enfermedad Laboral (EL)** | Contraída por exposición a factores inherentes a la actividad laboral |
| **Accidente de Tránsito Laboral** | Ocurrido en cumplimiento de función fuera de sede |

## Flujo de Registro de AT/Incidente

### Datos Obligatorios FURAT
```
SECCIÓN A — DATOS DEL TRABAJADOR
- Nombres y apellidos completos
- Tipo y número de documento
- Fecha de nacimiento / edad
- Sexo
- Ocupación (código CIUO)
- Tipo de vinculación (indefinido, fijo, obra, aprendiz, etc.)
- Salario base de cotización
- ARL a la que está afiliado

SECCIÓN B — DATOS DE LA EMPRESA
- Razón social y NIT
- ARL contratada
- Clase de riesgo
- Actividad económica (CIIU)
- Dirección centro de trabajo

SECCIÓN C — DATOS DEL ACCIDENTE
- Fecha y hora exacta del accidente
- Jornada (ordinaria/extra/otro)
- Tiempo laborado antes del accidente
- Sitio exacto del accidente
- Área de la empresa
- ¿Estaba haciendo su tarea habitual? (S/N)
- Descripción detallada del accidente
- Agente causante
- Mecanismo o forma del accidente
- Tipo de lesión
- Parte del cuerpo afectada
- Objeto o substancia causante

SECCIÓN D — CONSECUENCIAS
- ¿Requirió atención médica? (S/N)
- Días de incapacidad
- Secuelas (S/N)
- ¿Fue accidente mortal? (S/N)
```

### Plazos Legales FURAT
- **AT grave o mortal**: Reportar a ARL dentro de **2 días hábiles** (Dec. 1072/2015 Art. 2.2.4.2.1)
- **AT con incapacidad**: Reportar a ARL dentro de **2 días hábiles**
- **Enfermedad Laboral**: FUREL en los plazos establecidos por la ARL

## Metodología de Investigación

### 5 Porqués (incidentes leves a moderados)
```
Evento: [descripción del accidente]
¿Por qué 1? → Causa inmediata
¿Por qué 2? → Causa básica
¿Por qué 3? → Causa raíz operativa
¿Por qué 4? → Causa raíz sistémica
¿Por qué 5? → Causa raíz organizacional
→ ACCIÓN CORRECTIVA sobre causa raíz
```

### Bow-Tie (AT graves o mortales)
```
AMENAZA → [EVENTO CENTRAL] → CONSECUENCIAS
   ↓                               ↓
Barreras preventivas         Barreras mitigadoras
(pre-evento)                 (post-evento)
```

## Indicadores Estadísticos

```python
# Fórmulas oficiales Colombia
IF  = (N_AT × 240000) / HHT        # Índice de Frecuencia
IS  = (Días_perdidos × 240000) / HHT  # Índice de Severidad
ILI = IF × IS / 1000               # Índice de Lesión Incapacitante
IIEL = (N_EL × 240000) / HHT       # Índice Incidencia EL

# HHT = Horas Hombre Trabajadas en el período
# 240.000 = constante (trabajadores × horas × días estándar)
```

## Plan de Acción Correctiva

Para cada causa raíz identificada:
```
- Acción correctiva específica
- Tipo: Eliminación / Sustitución / Ingeniería / Administrativo / EPP
- Responsable (nombre y cargo)
- Fecha límite de implementación
- Recursos requeridos
- Indicador de verificación
- Estado: Pendiente / En proceso / Cerrada
```

## Archivos Generados

1. **FURAT.pdf** — Formulario oficial con todos los campos
2. **Informe_Investigacion_AT.docx** — Narrativa + árbol de causas + plan de acción
3. **Estadisticas_Accidentalidad.xlsx** — IF, IS, ILI con gráficas por mes
4. **Plan_Accion_Correctiva.xlsx** — Seguimiento con semáforo de cumplimiento
