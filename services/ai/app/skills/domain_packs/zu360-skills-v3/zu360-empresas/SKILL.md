---
name: zu360-empresas
description: >
  Skill para gestionar empresas cliente en ZU 360. Activa cuando el usuario mencione:
  empresa nueva, onboarding empresa, RUT, NIT, CIIU, ARL, representante legal,
  scoring de riesgo, portafolio de empresas, dashboard multi-empresa, clase de riesgo,
  perfil de empresa, o cualquier gestión de alta y administración de empresas cliente.
---

# ZU360 — Gestión de Empresas Cliente

## Onboarding en 9 Pasos

```
Paso 1: Carga del RUT → OCR extrae NIT, razón social, CIIU, dirección, RL
Paso 2: Clasificación automática por CIIU → clase de riesgo ARL (I a V)
Paso 3: Configuración de ARL vinculada (Sura, Positiva, Colmena, Colpatria)
Paso 4: Configuración de IPS contratada para exámenes médicos
Paso 5: Carga de trabajadores (CSV/Excel o manual)
Paso 6: Configuración de estructura organizacional (áreas/procesos)
Paso 7: Asignación de responsable SST interno
Paso 8: Carga de documentos existentes del SG-SST
Paso 9: Primera ronda de diagnóstico → scoring inicial
```

## Clases de Riesgo ARL por CIIU

```
Clase I   (0.348%) → Financiero, comercio, administración
Clase II  (0.435%) → Manufactura ligera, educación
Clase III (0.783%) → Manufactura pesada, transporte
Clase IV  (1.044%) → Construcción, minería superficial
Clase V   (1.044%) → Minería subterránea, explosivos
```

## Scoring de Riesgo por Empresa

```
Variables del score:
- Clase de riesgo ARL (peso 30%)
- N° trabajadores (peso 15%)
- IF últimos 12 meses (peso 25%)
- Cumplimiento Res. 0312 (peso 20%)
- Días de vencimientos activos (peso 10%)

Score:
  0–40  → 🔴 CRÍTICO — Intervención urgente
  41–70 → 🟡 MODERADO — Plan de mejora activo
  71–100→ 🟢 ACEPTABLE — Mantenimiento
```

## Datos del Perfil de Empresa
```
- Razón social y NIT
- Código CIIU y actividad económica
- Clase de riesgo ARL
- ARL contratada + ejecutivo de cuenta ARL
- IPS contratada para exámenes
- Dirección sede(s) de trabajo
- Representante Legal: nombre, cédula, cargo, contacto
- Responsable SST interno: nombre, cargo, contacto
- N° de trabajadores directos + contratistas
- Turnos de trabajo
```

## Archivos Generados

1. **Perfil_Empresa_[NIT].pdf** — Ficha completa de la empresa
2. **Diagnostico_Inicial_SG-SST_[Empresa].xlsx** — Estándares Res. 0312
3. **Plan_Trabajo_Anual_[Empresa]_[Año].xlsx** — Con actividades y responsables
