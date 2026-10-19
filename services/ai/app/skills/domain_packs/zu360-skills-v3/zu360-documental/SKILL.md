---
name: zu360-documental
description: >
  Skill generador de los 49 documentos normativos SST de ZU 360. Activa cuando el usuario
  pida crear, generar, redactar, elaborar o exportar cualquier documento SST:
  políticas, procedimientos, programas, actas, informes, formatos, matrices, certificados,
  planes de emergencia, reglamentos, permisos de trabajo, cronogramas, inventarios,
  o cualquier documento en formato Word, PDF, Excel o PowerPoint relacionado con SST Colombia.
  También activa para "llena este formato", "genera la plantilla de...", "crea el documento de..."
---

# ZU360 — Generador Documental SST (49 Plantillas)

## Catálogo Completo de Documentos

### POLÍTICAS (6)
| Código | Documento | Norma Base |
|---|---|---|
| POL-001 | Política de Seguridad y Salud en el Trabajo | Dec. 1072/2015 Art. 2.2.4.6.5 |
| POL-002 | Programa Anual de Trabajo SST | Res. 0312/2019 |
| POL-003 | Plan de Emergencias y Contingencias | Dec. 1072/2015 |
| POL-004 | Plan de Evacuación | NSR-10 / Dec. 1072 |
| POL-005 | Procedimientos de Trabajo Seguro | Dec. 1072/2015 |
| POL-006 | Reglamento de Higiene y Seguridad Industrial | Res. 2400/1979 Art. 106 |

### ACTAS (5)
| Código | Documento |
|---|---|
| ACT-001 | Acta de Reunión COPASST |
| ACT-002 | Acta de Inspección de Seguridad |
| ACT-003 | Acta de Entrega de EPP |
| ACT-004 | Acta de Inducción SST |
| ACT-005 | Acta de Elección COPASST |

### INFORMES (7)
| Código | Documento |
|---|---|
| INF-001 | Informe de Investigación de Accidente de Trabajo |
| INF-002 | Informe de Auditoría Interna SG-SST |
| INF-003 | Informe de Inspección de Seguridad |
| INF-004 | Informe de Indicadores SST (mensual/trimestral/anual) |
| INF-005 | Informe de Exámenes Médicos Ocupacionales |
| INF-006 | Informe de Capacitaciones Realizadas |
| INF-007 | Informe de Riesgos Psicosociales |

### FORMATOS (6)
| Código | Documento |
|---|---|
| FOR-001 | FURAT — Formulario Único de Reporte de AT |
| FOR-002 | FUREL — Formulario de Reporte Enfermedad Laboral |
| FOR-003 | Formato de Investigación de AT/Incidente |
| FOR-004 | Permiso de Trabajo de Alto Riesgo |
| FOR-005 | Checklist de Inspección (adaptable por área) |
| FOR-006 | Formato de Entrega de Área |

### MATRICES (4)
| Código | Documento |
|---|---|
| MAT-001 | Matriz de Peligros y Valoración de Riesgos GTC-45 (Excel) |
| MAT-002 | Matriz de Peligros GTC-45 (Informe HTML/Word) |
| MAT-003 | Matriz de EPP por Cargo |
| MAT-004 | Matriz de Cumplimiento Legal SST |

### CERTIFICADOS (6)
| Código | Documento |
|---|---|
| CER-001 | Certificado de Capacitación General |
| CER-002 | Certificado Trabajo en Alturas (Res. 1409/2012) |
| CER-003 | Certificado Espacios Confinados (Res. 4276/2022) |
| CER-004 | Certificado Primeros Auxilios |
| CER-005 | Certificado Manejo de Extintores |
| CER-006 | Certificado de Investigación de AT |

### EXCEL (7)
| Código | Documento |
|---|---|
| XLS-001 | Inventario y Control de EPP |
| XLS-002 | Programa Anual de Capacitaciones |
| XLS-003 | Control de Asistencia a Capacitación |
| XLS-004 | Registro de Exámenes Médicos Ocupacionales |
| XLS-005 | Indicadores SST con Fórmulas Automáticas |
| XLS-006 | Programa de Inspecciones |
| XLS-007 | Control de Dotación por Trabajador |

### PROGRAMAS DE VIGILANCIA EPIDEMIOLÓGICA (8)
| Código | Documento |
|---|---|
| PVE-001 | PVE Riesgo Cardiovascular |
| PVE-002 | PVE Riesgo Ergonómico |
| PVE-003 | PVE Riesgo Psicosocial |
| PVE-004 | PVE Riesgo Auditivo (HNIR) |
| PVE-005 | PVE Riesgo Químico |
| PVE-006 | PVE Riesgo Visual |
| PVE-007 | PVE Trabajo en Alturas |
| PVE-008 | PVE Riesgo Biológico |

## Flujo de Generación

### 1. Identificar documento solicitado
Mapear solicitud del usuario al código de documento arriba.

### 2. Recopilar datos de empresa (si no están en contexto)
```
- Razón social y NIT
- Representante Legal
- ARL vinculada
- Clase de riesgo (I al V)
- Número de trabajadores
- Actividad económica (CIIU)
- Ciudad y dirección
- Fecha del documento
- Profesional SST responsable (nombre + licencia)
```

### 3. Generar el archivo
- **Word (.docx)**: Documentos narrativos, informes, políticas, procedimientos
- **Excel (.xlsx)**: Matrices, cronogramas, inventarios, estadísticas
- **PDF**: Certificados, formatos impresos, actas firmadas
- **PowerPoint (.pptx)**: Capacitaciones, presentaciones a gerencia

### 4. Elementos obligatorios en todo documento SST
```
✓ Encabezado: Logo empresa + nombre + NIT
✓ Código del documento (ej: FOR-SST-001)
✓ Versión y fecha de vigencia
✓ Nombre del documento
✓ Firma y licencia de la profesional SST
✓ Firma del Representante Legal (cuando aplique)
✓ Control de versiones al final
```

## Instrucciones de Uso

Cuando el usuario pida un documento:
1. Lee los datos de empresa del contexto activo
2. Selecciona la plantilla correcta del catálogo
3. Pre-llena todos los campos con los datos disponibles
4. Genera el archivo real (.docx/.xlsx/.pdf/.pptx)
5. Entrega el archivo descargable
6. Indica qué campos quedaron pendientes de completar manualmente
