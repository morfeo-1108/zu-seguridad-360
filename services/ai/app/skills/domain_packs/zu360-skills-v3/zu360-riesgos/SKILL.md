---
name: zu360-riesgos
description: >
  Skill para construir, calcular y exportar matrices de riesgos GTC-45 en ZU 360.
  Activa cuando el usuario mencione: matriz de riesgos, GTC 45, peligro, riesgo,
  NR, ND, NE, NP, NC, mapa de calor, nivel de aceptabilidad, jerarquía de controles,
  EPP, controles de ingeniería, evaluación de riesgos, valoración de peligros,
  identificación de peligros, o cualquier tarea de análisis de riesgo ocupacional.
  Genera matrices completas en Excel con mapa de calor y en Word/PDF como informe.
---

# ZU360 — Módulo de Riesgos y Seguridad GTC-45

## Metodología GTC-45 Implementada

### Fórmulas de Cálculo

```
ND  = Nivel de Deficiencia      (10=MA, 6=A, 2=M, 0=B)
NE  = Nivel de Exposición       (4=EC, 3=EF, 2=EO, 1=ESPO)
NP  = ND × NE                   = Nivel de Probabilidad
NC  = Nivel de Consecuencia     (100=ME, 60=MUY GRAVE, 25=GRAVE, 10=LEVE)
NR  = NP × NC                   = Nivel de Riesgo

Aceptabilidad:
  NR > 4000     → I   No Aceptable (Situación Crítica)
  500–4000      → II  No Aceptable o Aceptable con Control Específico
  120–500       → III Aceptable con Control
  < 120         → IV  Aceptable
```

### Jerarquía de Controles (orden obligatorio)

1. **Eliminación** — Remover el peligro completamente
2. **Sustitución** — Reemplazar por algo menos peligroso
3. **Controles de Ingeniería** — Barreras físicas, ventilación, aislamiento
4. **Controles Administrativos** — Procedimientos, rotación, señalización
5. **EPP** — Última línea de defensa, nunca la primera

## Clasificación de Peligros (GTC-45)

- **Biológico**: Virus, bacterias, hongos, parásitos, vectores
- **Físico**: Ruido, vibración, temperatura, iluminación, radiaciones
- **Químico**: Gases, vapores, polvos, líquidos, aerosoles
- **Psicosocial**: Carga mental, jornada, relaciones, violencia
- **Biomecánico**: Posturas, esfuerzo, movimiento repetitivo, manipulación cargas
- **Condiciones de Seguridad**: Locativo, eléctrico, mecánico, tecnológico, públicos, trabajo en alturas
- **Fenómenos Naturales**: Sismos, vendavales, inundaciones

## Flujo de Construcción de Matriz

### Datos Requeridos
```
1. Empresa: nombre, NIT, sector CIIU, clase riesgo ARL
2. Proceso / Área / Actividad
3. Cargo(s) expuesto(s)
4. Número de trabajadores expuestos
5. Tiempo de exposición (horas/día, días/semana)
6. Controles existentes
```

### Pasos de Generación

1. **Inventario de peligros** por proceso/área
2. **Evaluación**: calcular ND, NE, NP, NC, NR para cada peligro
3. **Clasificación** por nivel de aceptabilidad
4. **Sugerencia de controles** según jerarquía GTC-45
5. **Mapa de calor**: rojo (I), naranja (II), amarillo (III), verde (IV)
6. **Plan de acción** para riesgos No Aceptables
7. **Exportar** a Excel (con colores) + informe Word/PDF

## Generación de Archivos

### Excel — Matriz Completa
```python
# Columnas obligatorias GTC-45:
columnas = [
    "Proceso", "Zona/Lugar", "Actividades", "Tarea",
    "Rutinaria (S/N)", "Peligro (Clase)", "Peligro (Descripción)",
    "Efectos Posibles", "Controles Existentes",
    "ND", "NE", "NP", "Nivel Probabilidad",
    "NC", "NR", "Nivel Riesgo", "Aceptabilidad",
    "Trabajadores Expuestos", "Peor Consecuencia",
    "Control Fuente", "Control Medio", "Control Individuo",
    "Responsable", "Fecha Implementación", "Estado"
]
# Aplicar colores: rojo=#C00000, naranja=#FF6600, amarillo=#FFFF00, verde=#00B050
```

### Informe Word — Formato Res. 0312
Incluye: portada, resumen ejecutivo, metodología, tabla de peligros priorizados,
mapa de calor, plan de acción, firma de la profesional SST.

## Ejemplo de Prompt de Uso

```
"Crea la matriz GTC-45 para el área de producción de una empresa 
metalmecánica, 25 trabajadores, riesgo IV. Peligros principales: 
ruido (90dB), trabajo en alturas (3m), polvos metálicos, 
maquinaria sin guarda."
```
