"""
Métricas de calidad de los modelos de IA.

Responde a la crítica válida de auditoría: 'ninguna afirmación de IA incluye
precision, recall, latencia'. Este módulo instrumenta la MEDICIÓN. Los valores
reales se poblarán durante el piloto de campo con datos etiquetados; aquí queda
el andamiaje para capturarlos de forma sistemática, no inventada.
"""
from __future__ import annotations
from dataclasses import dataclass, field, asdict
from datetime import datetime
from typing import Optional
import time


@dataclass
class RAGMetric:
    """Métrica de una consulta al RAG normativo."""
    query: str
    retrieved_chunks: int
    latency_ms: float
    cited_source: Optional[str] = None
    cited_article: Optional[str] = None
    # Se completa en evaluación con ground-truth etiquetado:
    is_correct: Optional[bool] = None       # ¿la respuesta fue correcta?
    citation_correct: Optional[bool] = None  # ¿la cita de artículo fue correcta?
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat())


@dataclass
class OCRMetric:
    """Métrica de una extracción OCR médica."""
    ips_format: str
    fields_extracted: int
    fields_expected: int
    latency_ms: float
    confidence_avg: float
    human_corrected_fields: Optional[int] = None  # cuántos campos corrigió el humano
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat())

    @property
    def field_accuracy(self) -> Optional[float]:
        """Exactitud = campos correctos / esperados (requiere corrección humana)."""
        if self.human_corrected_fields is None or self.fields_expected == 0:
            return None
        correct = self.fields_expected - self.human_corrected_fields
        return correct / self.fields_expected


@dataclass
class BiometricMetric:
    """Métrica de una identificación biométrica (para FAR/FRR)."""
    similarity: float
    threshold: float
    matched: bool
    latency_ms: float
    # Se completa en evaluación con ground-truth:
    ground_truth_match: Optional[bool] = None  # ¿era realmente la persona?
    timestamp: str = field(default_factory=lambda: datetime.utcnow().isoformat())

    @property
    def outcome(self) -> Optional[str]:
        """Clasifica en TP/TN/FP/FN para calcular FAR y FRR."""
        if self.ground_truth_match is None:
            return None
        if self.matched and self.ground_truth_match:
            return "true_positive"
        if not self.matched and not self.ground_truth_match:
            return "true_negative"
        if self.matched and not self.ground_truth_match:
            return "false_accept"   # cuenta para FAR
        return "false_reject"       # cuenta para FRR


class MetricsCollector:
    """Acumula métricas y calcula agregados para el reporte del piloto."""

    def __init__(self):
        self.rag: list[RAGMetric] = []
        self.ocr: list[OCRMetric] = []
        self.biometric: list[BiometricMetric] = []

    def record_rag(self, m: RAGMetric): self.rag.append(m)
    def record_ocr(self, m: OCRMetric): self.ocr.append(m)
    def record_biometric(self, m: BiometricMetric): self.biometric.append(m)

    def rag_summary(self) -> dict:
        if not self.rag:
            return {"n": 0, "note": "sin datos aún — se puebla en el piloto"}
        evaluated = [m for m in self.rag if m.is_correct is not None]
        latencies = sorted(m.latency_ms for m in self.rag)
        p95 = latencies[int(len(latencies) * 0.95) - 1] if latencies else 0
        acc = sum(1 for m in evaluated if m.is_correct) / len(evaluated) if evaluated else None
        cit = sum(1 for m in evaluated if m.citation_correct) / len(evaluated) if evaluated else None
        return {"n": len(self.rag), "evaluated": len(evaluated),
                "accuracy": acc, "citation_accuracy": cit, "latency_p95_ms": p95}

    def biometric_summary(self) -> dict:
        evaluated = [m for m in self.biometric if m.ground_truth_match is not None]
        if not evaluated:
            return {"n": len(self.biometric), "note": "FAR/FRR requieren etiquetado del piloto"}
        fa = sum(1 for m in evaluated if m.outcome == "false_accept")
        fr = sum(1 for m in evaluated if m.outcome == "false_reject")
        neg = sum(1 for m in evaluated if not m.ground_truth_match)
        pos = sum(1 for m in evaluated if m.ground_truth_match)
        return {"n": len(evaluated),
                "FAR": fa / neg if neg else None,   # tasa de falsa aceptación
                "FRR": fr / pos if pos else None}   # tasa de falso rechazo

    def ocr_summary(self) -> dict:
        evaluated = [m for m in self.ocr if m.field_accuracy is not None]
        if not evaluated:
            return {"n": len(self.ocr), "note": "exactitud requiere corrección humana registrada"}
        avg_acc = sum(m.field_accuracy for m in evaluated) / len(evaluated)
        return {"n": len(evaluated), "field_accuracy_avg": avg_acc}

    def full_report(self) -> dict:
        return {"rag": self.rag_summary(), "ocr": self.ocr_summary(),
                "biometric": self.biometric_summary(),
                "generated_at": datetime.utcnow().isoformat()}
