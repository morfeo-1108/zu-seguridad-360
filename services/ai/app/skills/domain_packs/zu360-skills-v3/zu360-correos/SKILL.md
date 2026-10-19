---
name: zu360-correos
description: >
  Skill para redactar y gestionar correos SST en ZU 360. Activa para: redactar correo,
  enviar notificación, comunicado, carta a empresa cliente, alerta de vencimiento por
  correo, invitación a capacitación, resultados médicos, reporte mensual por email,
  plantilla de correo SST, o cualquier comunicación escrita de la profesional SISO.
---

# ZU360 — Módulo de Correos IA

## Plantillas de Correo SST

### Alerta de Vencimiento de Certificado
```
ASUNTO: ⚠️ Certificado por vencer — [Nombre trabajador] — [Empresa]
Estimado/a [Supervisor/RL]:
Le informamos que el certificado de [curso] del trabajador [nombre],
identificado con [documento], vence el [fecha] (en [X] días).
Por favor coordinar la renovación a la brevedad.
[Firma profesional SST]
```

### Convocatoria a Capacitación
```
ASUNTO: Convocatoria — Capacitación [Nombre Curso] — [Fecha]
Estimado/a [nombre]:
Lo invitamos a participar en la capacitación de [curso], de carácter
[obligatorio/voluntario] según [norma].
Fecha: [DD/MM/AAAA] | Hora: [HH:MM] | Lugar: [X]
Su asistencia es indispensable. Por favor confirmar a [contacto].
[Firma profesional SST]
```

### Reporte Mensual de Gestión SST
```
ASUNTO: Reporte SST [Mes Año] — [Empresa]
Estimado/a [RL/Gerencia]:
Adjunto el informe de gestión SST del mes de [mes], que incluye:
indicadores de accidentalidad, cumplimiento de capacitaciones,
estado de exámenes médicos y avance del programa anual.
[Resumen ejecutivo de 3 líneas con semáforo]
[Firma profesional SST]
```

### Notificación de Accidente de Trabajo
```
ASUNTO: 🚨 URGENTE — Reporte AT — [Nombre trabajador] — [Empresa] — [Fecha]
Se informa a la gerencia el ocurrimiento de un accidente de trabajo:
Trabajador: [nombre] | Cargo: [cargo] | Área: [área]
Fecha/Hora: [DD/MM/AAAA HH:MM]
Descripción breve: [descripción]
Estado: [atendido en IPS / hospitalizado / en campo]
El FURAT fue reportado a [ARL] en la fecha [fecha]. Se iniciará
investigación formal en las próximas 24 horas.
[Firma profesional SST]
```

## Comportamiento del Módulo

- Detecta urgencia del contexto (AT → tono formal urgente; convocatoria → amigable)
- Pre-llena destinatarios según el tipo de correo
- Sugiere horario de envío óptimo
- Registra correos enviados con confirmación de lectura
