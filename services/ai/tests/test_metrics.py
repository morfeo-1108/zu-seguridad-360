"""Tests del colector de métricas de IA (instrumentación para el piloto)."""
import sys, os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))
from app.metrics.ai_metrics import (
    MetricsCollector, RAGMetric, OCRMetric, BiometricMetric,
)


def test_biometric_far_frr_computation():
    c = MetricsCollector()
    c.record_biometric(BiometricMetric(0.9, 0.62, True, 40, ground_truth_match=True))   # TP
    c.record_biometric(BiometricMetric(0.7, 0.62, True, 40, ground_truth_match=False))  # FA
    c.record_biometric(BiometricMetric(0.4, 0.62, False, 40, ground_truth_match=True))  # FR
    c.record_biometric(BiometricMetric(0.3, 0.62, False, 40, ground_truth_match=False)) # TN
    s = c.biometric_summary()
    assert s["FAR"] == 0.5   # 1 falsa aceptación / 2 negativos reales (FA+TN)
    assert s["FRR"] == 0.5   # 1 falso rechazo / 2 positivos reales


def test_biometric_outcome_classification():
    assert BiometricMetric(0.9, 0.62, True, 40, ground_truth_match=True).outcome == "true_positive"
    assert BiometricMetric(0.7, 0.62, True, 40, ground_truth_match=False).outcome == "false_accept"
    assert BiometricMetric(0.4, 0.62, False, 40, ground_truth_match=True).outcome == "false_reject"


def test_ocr_field_accuracy():
    m = OCRMetric(ips_format="colsanitas", fields_extracted=10, fields_expected=10,
                  latency_ms=200, confidence_avg=0.9, human_corrected_fields=2)
    assert m.field_accuracy == 0.8  # 8 de 10 correctos


def test_empty_collector_no_crash():
    c = MetricsCollector()
    report = c.full_report()
    assert report["rag"]["n"] == 0
    assert "note" in report["biometric"]


def test_rag_latency_p95():
    c = MetricsCollector()
    for i in range(100):
        c.record_rag(RAGMetric(query=f"q{i}", retrieved_chunks=5, latency_ms=float(i)))
    s = c.rag_summary()
    assert s["n"] == 100
    assert 93 <= s["latency_p95_ms"] <= 96  # p95 ~ 94-95
