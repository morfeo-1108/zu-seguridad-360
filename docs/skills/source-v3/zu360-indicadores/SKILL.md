---
name: zu360-indicadores
description: >
  Skill para calcular, visualizar y reportar indicadores SST en ZU 360.
  Activa cuando el usuario mencione: IF, IS, ILI, IIEL, indicadores de accidentalidad,
  KPIs SST, tasa de incidentalidad, cumplimiento Res. 0312, estándares mínimos,
  dashboard de indicadores, informe estadístico, morbilidad, ausentismo,
  cobertura de capacitaciones, cobertura de exámenes, o cualquier métrica de gestión SST.
---

# ZU360 — Indicadores y KPIs SST

## Indicadores de Accidentalidad (Colombia)

```python
# Constante estándar: 240.000 (basada en 100 trabajadores × 48 horas × 50 semanas)

IF   = (N_AT × 240_000) / HHT
# Índice de Frecuencia: ATs por cada 240.000 HHT

IS   = (Días_perdidos × 240_000) / HHT  
# Índice de Severidad: días perdidos por cada 240.000 HHT

ILI  = (IF × IS) / 1_000
# Índice de Lesión Incapacitante

IIEL = (N_EL × 240_000) / HHT
# Índice de Incidencia de Enfermedad Laboral

Tasa_AT = (N_AT / N_trabajadores) × 100
# Tasa de accidentalidad por 100 trabajadores

Tasa_Ausentismo = (Días_ausencia / (N_trabajadores × Días_hábiles)) × 100
```

## Indicadores de Gestión SG-SST (Res. 0312/2019)

```
Cobertura Capacitaciones   = (Trabajadores capacitados / Total) × 100
Cobertura Exámenes         = (Exámenes realizados / Programados) × 100
Cumplimiento Cronograma    = (Actividades ejecutadas / Programadas) × 100
Cumplimiento Inspecciones  = (Inspecciones realizadas / Programadas) × 100
Cierre Planes de Acción    = (Acciones cerradas / Total) × 100
Cumplimiento Res. 0312     = (Estándares cumplidos / Total aplicables) × 100
```

## Calificación Res. 0312/2019

```
CRÍTICO        < 60%    → Medida de prevención o sanción
MODERADAMENTE
ACEPTABLE    60% - 85%  → Plan de mejoramiento 6 meses
ACEPTABLE      > 85%    → Plan de mejoramiento anual (autoevaluación)
```

### Estándares por Tamaño de Empresa
- **≤ 10 trabajadores, riesgo I-II**: 7 estándares
- **11-50 trabajadores, riesgo I-II**: 21 estándares  
- **> 50 trabajadores o riesgo III-IV-V**: 60 estándares

## Dashboard — Estructura del Informe

### Sección 1: Resumen Ejecutivo
- Período analizado
- Número de empresas / trabajadores
- Semáforo de cumplimiento por empresa

### Sección 2: Accidentalidad
- Tabla de ATs por mes con IF/IS/ILI
- Comparativo período anterior
- Gráfica de tendencia mensual
- Pareto de causas raíz
- Distribución por área / cargo / turno

### Sección 3: Enfermedad Laboral
- Casos activos vs resueltos
- IIEL por período
- Diagnósticos más frecuentes

### Sección 4: Ausentismo
- Días perdidos por causa (AT / EL / Enfermedad común / Permisos)
- Tasa de ausentismo por área
- Costo estimado de ausentismo

### Sección 5: Gestión SST
- Cobertura de capacitaciones (%)
- Cobertura de exámenes médicos (%)
- Cumplimiento de inspecciones (%)
- Estado de planes de acción

### Sección 6: Cumplimiento Res. 0312
- Puntaje por ciclo PHVA
- Estándares críticos sin cumplir
- Plan de mejoramiento sugerido

## Archivos Generados

1. **Indicadores_SST_[Empresa]_[Período].xlsx**
   - Hoja 1: Datos base (ATs, HHT, trabajadores)
   - Hoja 2: Indicadores calculados con fórmulas
   - Hoja 3: Gráficas automáticas
   - Hoja 4: Cumplimiento Res. 0312

2. **Informe_Indicadores_[Empresa]_[Período].docx**
   - Informe ejecutivo completo para presentar a gerencia

3. **Dashboard_SST.pdf**
   - Versión imprimible del dashboard con semáforos visuales
