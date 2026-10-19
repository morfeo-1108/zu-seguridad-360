---
name: zu360-organizador
description: >
  Skill para organizar el computador de la profesional SST en ZU 360. Activa cuando
  el usuario mencione: organizar el computador, organizar archivos, ordenar carpetas,
  clasificar documentos, separar trabajo de personal, limpiar el escritorio,
  encontrar archivos de trabajo, estructura de carpetas SST, mover documentos,
  inventario de archivos, qué tengo en el computador, organizar mis documentos,
  archivos duplicados, carpeta desordenada, o cualquier tarea de gestión y
  organización del sistema de archivos. También activa para "escanea mi computador",
  "dime qué archivos tengo", "organiza mis carpetas de SST".
---

# ZU360 — Organizador Inteligente del Computador

## Propósito
Escanear el computador de la profesional SST, clasificar automáticamente cada
archivo como TRABAJO o PERSONAL, construir una estructura de carpetas ordenada
para ZU 360, y mover todo a su lugar correcto — sin borrar nada.

---

## Fase 1 — Escaneo del Sistema

### Ubicaciones a escanear (Windows)
```
C:\Users\[NombreUsuario]\Desktop\
C:\Users\[NombreUsuario]\Documents\
C:\Users\[NombreUsuario]\Downloads\
C:\Users\[NombreUsuario]\OneDrive\        (si existe)
C:\Users\[NombreUsuario]\Google Drive\    (si existe)
Cualquier carpeta raíz adicional indicada por la usuaria
```

### Comandos de escaneo (PowerShell)
```powershell
# Inventario completo de archivos con metadatos
Get-ChildItem -Path "C:\Users\$env:USERNAME" -Recurse -File |
  Select-Object Name, Extension, Length, LastWriteTime, DirectoryName |
  Export-Csv "C:\ZU360\inventario_archivos.csv" -NoTypeInformation

# Contar por extensión para tener panorama general
Get-ChildItem -Path "C:\Users\$env:USERNAME" -Recurse -File |
  Group-Object Extension | Sort-Object Count -Descending |
  Select-Object Name, Count | Format-Table
```

---

## Fase 2 — Clasificación Automática

### Reglas de clasificación TRABAJO (SST / Profesional)

```
EXTENSIONES DE TRABAJO:
  .docx .doc  → Documentos Word (informes, formatos, procedimientos)
  .xlsx .xls  → Excel (matrices, indicadores, cronogramas)
  .pdf        → Documentos oficiales, normas, certificados
  .pptx .ppt  → Presentaciones de capacitación
  .csv        → Datos de empleados, estadísticas

PALABRAS CLAVE EN EL NOMBRE (detectar como trabajo):
  sst, sgsst, siso, sst, ohc, ohse, hse
  riesgo, matriz, gtc, peligro, nrc, nr,
  accidente, incidente, furat, furel, at, el,
  capacitacion, certificado, alturas, epp, arnés,
  copasst, acta, inspeccion, auditoria,
  empresa, empleado, trabajador, cargo,
  indicador, kpi, informe, reporte,
  emergencia, brigada, simulacro,
  examen, medico, aptitud, restriccion,
  norma, decreto, resolucion, ley, gtc45,
  politica, procedimiento, programa,
  contratista, permiso, loto

CARPETAS QUE SON TRABAJO (mover contenido completo):
  SST, SGSST, Seguridad, Trabajo en Alturas, Empresas,
  Clientes, Informes, Matrices, Capacitaciones,
  Certificados, Incidentes, Auditorias, COPASST
```

### Reglas de clasificación PERSONAL

```
PALABRAS CLAVE EN EL NOMBRE (detectar como personal):
  foto, imagen, vacaciones, viaje, familia,
  receta, musica, pelicula, serie, juego,
  banco, factura personal, seguro personal,
  cumpleaños, fiesta, mascota, hogar, casa

EXTENSIONES TÍPICAMENTE PERSONALES:
  .jpg .jpeg .png .gif .mp4 .mp3 .avi .mkv
  → NOTA: analizar nombre + carpeta origen antes de mover
  → Un .jpg puede ser evidencia fotográfica de SST

AMBIGUO → PREGUNTAR A LA USUARIA:
  Archivos sin nombre descriptivo (doc1.docx, sin título.xlsx)
  Archivos mezclados en carpetas sin clasificar
  Fotos sin contexto claro
```

---

## Fase 3 — Estructura de Carpetas ZU360 (crear en C:\ZU360\)

```
C:\ZU360\
│
├── 📁 TRABAJO\
│   ├── 📁 EMPRESAS\
│   │   ├── [NombreEmpresa1]\
│   │   │   ├── Documentos\
│   │   │   ├── Matrices\
│   │   │   ├── Incidentes\
│   │   │   ├── Capacitaciones\
│   │   │   ├── Examenes_Medicos\
│   │   │   ├── Inspecciones\
│   │   │   ├── COPASST\
│   │   │   └── Informes\
│   │   └── [NombreEmpresa2]\
│   │       └── [misma estructura]
│   │
│   ├── 📁 PLANTILLAS\
│   │   ├── Word\
│   │   ├── Excel\
│   │   ├── PDF\
│   │   └── PowerPoint\
│   │
│   ├── 📁 NORMATIVA\
│   │   ├── Decretos\
│   │   ├── Resoluciones\
│   │   ├── Leyes\
│   │   └── Normas_Tecnicas\
│   │
│   ├── 📁 FORMACION_PROPIA\
│   │   ├── Cursos\
│   │   ├── Certificados_Propios\
│   │   └── Material_Estudio\
│   │
│   └── 📁 ADMINISTRATIVO\
│       ├── Contratos\
│       ├── Facturacion\
│       └── Propuestas\
│
├── 📁 PERSONAL\
│   ├── Fotos\
│   ├── Documentos_Personales\
│   ├── Finanzas_Personales\
│   └── Entretenimiento\
│
├── 📁 PENDIENTE_CLASIFICAR\
│   └── [archivos ambiguos que requieren decisión]
│
└── 📁 RESPALDO_ORIGINAL\
    └── [copia de seguridad antes de reorganizar — NO borrar]
```

---

## Fase 4 — Ejecución Segura

### Protocolo de seguridad ANTES de mover cualquier archivo
```
PASO 1: Crear respaldo completo en C:\ZU360\RESPALDO_ORIGINAL\
PASO 2: Mostrar a la usuaria el plan de reorganización (lista de movimientos)
PASO 3: Pedir CONFIRMACIÓN EXPLÍCITA antes de ejecutar
PASO 4: Mover (nunca borrar) — los originales quedan en el respaldo
PASO 5: Generar reporte de lo que se movió y a dónde
PASO 6: Mantener RESPALDO_ORIGINAL durante 30 días antes de sugerir eliminar
```

### PowerShell — Crear estructura y mover archivos
```powershell
# Crear estructura de carpetas ZU360
$base = "C:\ZU360"
$carpetas = @(
  "TRABAJO\EMPRESAS",
  "TRABAJO\PLANTILLAS\Word",
  "TRABAJO\PLANTILLAS\Excel",
  "TRABAJO\PLANTILLAS\PDF",
  "TRABAJO\PLANTILLAS\PowerPoint",
  "TRABAJO\NORMATIVA\Decretos",
  "TRABAJO\NORMATIVA\Resoluciones",
  "TRABAJO\NORMATIVA\Leyes",
  "TRABAJO\FORMACION_PROPIA",
  "TRABAJO\ADMINISTRATIVO",
  "PERSONAL\Fotos",
  "PERSONAL\Documentos_Personales",
  "PERSONAL\Finanzas_Personales",
  "PENDIENTE_CLASIFICAR",
  "RESPALDO_ORIGINAL"
)
foreach ($c in $carpetas) {
  New-Item -ItemType Directory -Path "$base\$c" -Force
}
Write-Host "Estructura ZU360 creada exitosamente"
```

---

## Fase 5 — Reporte Final de Organización

Al terminar, entregar en pantalla y como archivo:

```
REPORTE DE ORGANIZACIÓN — ZU360
Fecha: [DD/MM/AAAA] | Computador: [NombrePC]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
RESUMEN:
  Total archivos escaneados:        [X]
  Clasificados como TRABAJO:        [X] ([%])
  Clasificados como PERSONAL:       [X] ([%])
  Pendientes de clasificar:         [X]
  Duplicados encontrados:           [X]
  Espacio liberado (duplicados):    [X] MB

MOVIMIENTOS REALIZADOS:
  [Archivo] → [Carpeta destino]
  ...

ARCHIVOS PENDIENTES DE CLASIFICAR:
  [Lista de archivos ambiguos]
  → Revisar en C:\ZU360\PENDIENTE_CLASIFICAR\

RESPALDO DISPONIBLE EN:
  C:\ZU360\RESPALDO_ORIGINAL\
  (mantener hasta [fecha 30 días])
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Mantenimiento Periódico

Cada semana al iniciar Claude Code, verificar:
```
□ ¿Hay archivos nuevos en Descargas sin clasificar?
□ ¿Hay archivos en el Escritorio que deberían estar en ZU360?
□ ¿Hay documentos de empresas nuevas sin carpeta propia?
□ ¿Se generaron documentos SST que no se guardaron en la empresa correcta?
```
