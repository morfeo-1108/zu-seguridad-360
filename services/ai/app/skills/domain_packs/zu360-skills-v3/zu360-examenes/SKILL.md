---
name: zu360-examenes
description: >
  Skill para gestionar exámenes médicos ocupacionales en ZU 360.
  Activa cuando el usuario mencione: examen médico, examen de ingreso, examen periódico,
  examen de retiro, aptitud médica, restricción médica, IPS, médico ocupacional,
  concepto médico, OCR de examen, Resolución 2346, historia clínica ocupacional,
  vencimiento de examen, perfil del cargo, o cualquier gestión del ciclo médico laboral.
---

# ZU360 — Módulo de Exámenes Médicos Ocupacionales

## Marco Normativo
**Resolución 2346 de 2007** — Regula la práctica de evaluaciones médicas ocupacionales

## Tipos de Examen

| Tipo | Momento | Obligatorio |
|---|---|---|
| **Ingreso** | Antes de iniciar labores | Sí — sin excepción |
| **Periódico** | Según perfil de riesgo | Sí — frecuencia variable |
| **Retiro** | Dentro de 5 días hábiles post-retiro | Sí |
| **Post-incapacidad** | Retorno tras ≥30 días de incapacidad | Sí |
| **Cambio de ocupación** | Al cambiar de cargo con diferente riesgo | Sí |

## Frecuencia Exámenes Periódicos por Riesgo

```
Riesgo I   (bajo)    → Cada 2 años
Riesgo II  (medio)   → Cada año
Riesgo III (alto)    → Cada año
Riesgo IV  (muy alto)→ Cada 6 meses
Riesgo V   (máximo)  → Cada 6 meses
```

## Datos del Perfil Sociodemográfico del Cargo

Para cada cargo se requiere definir:
```
- Denominación del cargo
- Funciones principales
- Exposición a: ruido / vibración / polvos / químicos / alturas /
  posturas / esfuerzo físico / jornada nocturna / radiaciones / otros
- Clase de riesgo ARL del cargo
- Batería de exámenes requeridos:
  □ Optometría
  □ Audiometría (si ruido ≥ 85 dB)
  □ Espirometría (si exposición a polvos/gases)
  □ Visiometría
  □ Electrocardiograma (si cardiovascular)
  □ Hemoleucograma / Química sanguínea
  □ Examen físico general obligatorio
  □ Otros según criterio médico
```

## Pipeline OCR de Resultados

Cuando se sube PDF de resultado de IPS:
```
1. Identificar IPS → cargar template aprendido
2. Extraer: nombre trabajador, cédula, cargo, empresa
3. Extraer: fecha examen, médico, registro médico
4. Extraer: APTITUD (Apto / Apto con restricciones / No apto)
5. Extraer: restricciones específicas si las hay
6. Extraer: vigencia del concepto
7. Human-in-the-loop → profesional SST VALIDA antes de guardar
8. Si restricción → alertar supervisor y bloquear tareas incompatibles
```

## Conceptos de Aptitud

| Concepto | Significado | Acción |
|---|---|---|
| **Apto** | Sin restricciones | Puede ejercer el cargo |
| **Apto con restricciones** | Condicionado | Ajustar tareas según restricción |
| **No apto temporal** | Incapacidad temporal | No puede ejercer el cargo temporalmente |
| **No apto permanente** | Condición definitiva | Reubicación laboral obligatoria |
| **Pendiente de calificación** | En proceso | No puede asumir el cargo aún |

## Restricciones Médicas Frecuentes SST

```
□ No trabajar en alturas (>1.5m)
□ No levantar cargas > [X] kg
□ No trabajo en turnos nocturnos
□ No exposición a ruido > [X] dB
□ No exposición a químicos [tipo]
□ No bipedestación prolongada (>2h continuas)
□ No posturas de flexión/extensión forzada
□ Uso obligatorio de lentes de seguridad graduados
□ Control médico periódico cada [X] meses
```

## Alertas y Vencimientos

```
30 días antes del vencimiento → Notificar trabajador + empleador
15 días                       → Segunda notificación + profesional SST
0 días (vencido)              → Alerta crítica: examen vencido activo
```

## Retención Legal (Res. 2346/2007 Art. 17)
**20 años** en almacenamiento WORM cifrado AES-256

## Archivos Generados

1. **Hoja_Vida_Ocupacional_[Nombre]_[Empresa].pdf** — Historial completo
2. **Registro_Examenes_[Empresa].xlsx** — Control de vigencias y pendientes
3. **Informe_Cobertura_Medica_[Empresa].pdf** — Para auditorías
4. **Alerta_Vencimientos_Examenes.xlsx** — Con semáforo rojo/amarillo/verde
