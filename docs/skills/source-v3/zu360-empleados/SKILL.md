---
name: zu360-empleados
description: >
  Skill para gestionar trabajadores en ZU 360. Activa para: empleado, trabajador,
  hoja de vida ocupacional, directorio de personal, cargo, restricción médica,
  perfil de riesgo por cargo, importación masiva de empleados, vencimientos de
  trabajador, historial de capacitaciones de empleado, contratación, vinculación.
---

# ZU360 — Gestión de Empleados

## Hoja de Vida Ocupacional (campos)

```
DATOS PERSONALES
- Nombres y apellidos completos
- Tipo y número de documento
- Fecha de nacimiento / edad
- Sexo / Estado civil
- Nivel educativo
- Contacto de emergencia
- EPS / Fondo pensión / Caja compensación / ARL

INFORMACIÓN LABORAL
- Cargo actual + fecha ingreso
- Tipo de vinculación (indefinido / fijo / obra / aprendiz)
- Área / Proceso / Turno
- Salario base de cotización
- Clase de riesgo ARL del cargo

HISTORIAL SST
- Exámenes médicos: tipo, fecha, resultado, vigencia
- Capacitaciones: curso, fecha, calificación, vigencia
- Certificados vigentes
- Restricciones médicas activas
- Incidentes/AT en que estuvo involucrado
- EPP asignado (tipo, talla, fecha entrega)
```

## Importación Masiva — Formato CSV/Excel

```
Columnas requeridas:
nombre_completo, tipo_doc, numero_doc, fecha_nacimiento, sexo,
cargo, area, tipo_vinculacion, fecha_ingreso, eps, arl,
contacto_emergencia_nombre, contacto_emergencia_tel
```

## Alertas Automáticas por Trabajador

```
- Examen médico próximo a vencer (30/15/7/0 días)
- Certificado de capacitación por vencer (30/15/7/0 días)
- Cumpleaños (cortesía)
- Aniversario laboral
- Restricción médica activa (al asignar tarea incompatible)
```

## Archivos Generados
1. **Directorio_Empleados_[Empresa].xlsx** — Con estado de vencimientos (semáforo)
2. **Hoja_Vida_Ocupacional_[Nombre].pdf** — Individual
3. **Reporte_Vencimientos_[Empresa].pdf** — Pendientes urgentes
