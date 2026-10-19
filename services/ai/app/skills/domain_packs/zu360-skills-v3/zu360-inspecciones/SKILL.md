---
name: zu360-inspecciones
description: >
  Skill para gestionar inspecciones de seguridad en ZU 360. Activa cuando el usuario
  mencione: inspección, checklist de seguridad, hallazgo, EPP en campo, informe de
  inspección, programa de inspecciones, extintor, andamio, eléctrico, orden y aseo,
  señalización, condiciones inseguras, actos inseguros, planes de acción de inspección,
  o cualquier actividad de verificación de condiciones de trabajo en campo.
---

# ZU360 — Módulo de Inspecciones de Seguridad

## Tipos de Inspección Programadas

| Tipo | Frecuencia Sugerida | Responsable |
|---|---|---|
| General de instalaciones | Mensual | COPASST + SISO |
| EPP por área | Mensual | Supervisor + SISO |
| Extintores y equipos contra incendio | Mensual | Brigadista |
| Botiquines de primeros auxilios | Mensual | Brigadista |
| Herramientas manuales | Trimestral | Supervisor |
| Equipos y maquinaria | Trimestral | Mantenimiento + SISO |
| Instalaciones eléctricas | Semestral | Electricista + SISO |
| Andamios y trabajo en alturas | Antes de cada uso | Coordinador alturas |
| Orden y aseo (5S) | Semanal | Supervisor |
| Vehículos de carga | Mensual | Conductor + SISO |

## Estructura del Informe de Inspección

```
INFORME DE INSPECCIÓN DE SEGURIDAD
Empresa: [Nombre] | NIT: [XXX]
Área/Proceso: [XXX]
Fecha: [DD/MM/AAAA] | Hora: [HH:MM]
Inspector(es): [Nombres y cargos]
Tipo de inspección: [Programada / No programada / Seguimiento]

HALLAZGOS:
| # | Descripción | Foto | Área | Clasificación | Riesgo | Responsable | Fecha cierre |

Clasificación de hallazgos:
  CRÍTICO     → Riesgo inminente — acción inmediata
  MAYOR       → Riesgo alto — corrección en 48 horas
  MENOR       → Riesgo bajo — corrección en 30 días
  OBSERVACIÓN → Sin riesgo inmediato — mejora recomendada

PLAN DE ACCIÓN:
[Tabla con acciones correctivas por hallazgo]

Firma inspector + Firma responsable del área
```

## Checklists por Tipo

### Checklist Extintores
```
□ Ubicación accesible y señalizada
□ Tipo correcto para el riesgo del área (A/B/C/D/K)
□ Presión en zona verde del manómetro
□ Pin de seguridad intacto y con precintos
□ Manguera en buen estado (sin grietas)
□ Etiqueta de carga vigente (≤1 año)
□ Sin corrosión visible en el cuerpo
□ Soporte de pared firme, a 1.20m del suelo
```

### Checklist EPP
```
□ Casco: sin grietas, arnés en buen estado, fecha fabricación vigente
□ Gafas: lentes sin rayaduras, sin deformación
□ Guantes: tipo correcto para la tarea, sin perforaciones
□ Calzado de seguridad: punta de acero, suela antideslizante
□ Protección auditiva: tipo correcto, en buen estado
□ Arnés de seguridad: sin desgaste, costuras íntegras, argollas sin deformación
□ Uniforme: completo, limpio, sin elementos que cuelguen
```

### Checklist Orden y Aseo (5S)
```
□ Clasificación: solo elementos necesarios en el área
□ Orden: cada cosa en su lugar señalizado
□ Limpieza: pisos, máquinas y superficies limpias
□ Estandarización: señalización visible y actualizada
□ Disciplina: personal cumple los estándares establecidos
```

## Seguimiento de Planes de Acción

```
Estado semáforo:
🟢 CERRADA   → Acción implementada y verificada
🟡 EN PROCESO → Dentro del plazo, en ejecución
🔴 VENCIDA   → Plazo cumplido sin implementar — escalar
⚪ PENDIENTE  → Plazo futuro
```

## Archivos Generados

1. **Informe_Inspeccion_[Tipo]_[Empresa]_[Fecha].docx** — Con hallazgos y fotos
2. **Checklist_[Tipo]_[Fecha].pdf** — Para uso en campo (imprimible)
3. **Plan_Accion_Inspeccion_[Empresa].xlsx** — Seguimiento con semáforo
4. **Programa_Inspecciones_[Empresa]_[Año].xlsx** — Cronograma anual
