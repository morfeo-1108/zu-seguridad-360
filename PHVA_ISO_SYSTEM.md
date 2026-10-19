# Dominio PHVA + Cumplimiento ISO/Normativo (v12)

## Qué se construyó
El SG-SST dejó de ser un conjunto de módulos sueltos: ahora vive en el ciclo
**Planear-Hacer-Verificar-Actuar** como sistema de gestión, con cumplimiento
normativo vivo y vigilancia regulatoria con humano en el lazo.

## Principio de diseño (innegociable)
El sistema **PROPONE, ORGANIZA y GENERA**; el profesional SST **APRUEBA y FIRMA**.
- El score de cumplimiento cuenta SOLO estándares validados por humano.
- Los cambios normativos se detectan y encolan, nunca se aplican solos.
- Nada legalmente vinculante se cierra de forma automática.

## Modelo (core/internal/model/)
- **ManagementSystem** — agregado raíz: el SG-SST de una empresa para un año,
  con su fase PHVA actual y el enlace al ciclo anterior (cierre del lazo).
- **ComplianceStandard** — un requisito de una norma ligado a su fase PHVA, al
  módulo que lo satisface y a su evidencia. Lleva la evaluación automática del
  sistema Y el flag de validación humana (lo único que cuenta legalmente).
- **SSTObjective** — objetivo medible (fase Planear): meta vs valor actual.
- **CorrectiveAction** — cierre del lazo (fase Actuar): un hallazgo genera acción,
  su cierre retroalimenta el plan del siguiente ciclo.
- **RegulatorySource / RegulatoryChange** — vigilancia normativa en cola de
  aprobación humana.

## Marcos normativos modelados (versiones vigentes)
- **ISO 45001:2018** — SST (16 cláusulas, estructura HLS 4–10)
- **ISO 9001:2015** — calidad (6 cláusulas relevantes al SG)
- **ISO 14001:2015** — ambiental (5 cláusulas, para quien lo integra)
- **ISO 31000:2018** — gestión de riesgos (5 cláusulas del proceso)
- **Res. 0312/2019** — estándares mínimos Colombia (13 estándares, núcleo legal)

Cada estándar mapea: cláusula → fase PHVA → módulo que la satisface. Esa es la
matriz de trazabilidad que pide la auditoría del Ministerio.

## Servicios (core/internal/service/)
- **ManagementSystemService**: InitializeSystem (siembra estándares),
  AssessCompliance (el sistema propone), ValidateStandard (el humano aprueba),
  ComplianceScore (solo validados), AdvancePhase (avanza PHVA), CloseLoop
  (crea el año siguiente heredando lecciones).
- **RegulatoryWatchService**: SeedSources (6 fuentes oficiales), QueueDetectedChange
  (encola), PendingChanges (cola), ReviewChange (experto aprueba/descarta).

## Endpoints (8)
- POST /v1/management/systems · POST /…/:id/assess · GET /…/:id/score
- POST /…/:id/advance · POST /…/:id/close-loop · POST /…/standards/:id/validate
- GET /v1/regulatory/pending · POST /v1/regulatory/changes/:id/review

## Frontend
Tablero "Ciclo PHVA & ISO" (/dashboard/management): las 4 fases con la actual
resaltada, cumplimiento por cada marco normativo, matriz de trazabilidad de
estándares con su estado y validación, score global (solo validados) y la cola
de vigilancia normativa con botones aprobar/descartar.

## Tests (internal/phva/, ejecutados y pasando)
- ComplianceScore cuenta SOLO validados por humano que cumplen.
- Una propuesta del sistema sin validar NO cuenta (garantía legal).
- El ciclo PHVA avanza y cierra correctamente.

## Lo honesto que esto NO hace
- NO se mantiene legal solo: detecta y propone; un experto decide.
- NO garantiza documentos "sin errores": genera borradores que el profesional
  valida y firma. La exactitud es responsabilidad humana + validación, no magia.
