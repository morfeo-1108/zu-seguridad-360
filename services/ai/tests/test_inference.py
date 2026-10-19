"""Tests del motor de inferencia de visión — severidad y clasificación."""
import asyncio
import sys, os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.vision.inference import (
    VisionInferenceEngine, Detection, DetectionClass, InferenceResult, SEVERITY,
)


def run(coro):
    return asyncio.run(coro)


def test_severity_mapping_critical_events():
    """Caídas, fuego y falta de arnés son críticos (potencialmente fatales)."""
    assert SEVERITY[DetectionClass.FALL] == "critical"
    assert SEVERITY[DetectionClass.FIRE] == "critical"
    assert SEVERITY[DetectionClass.NO_HARNESS] == "critical"


def test_detection_severity_property():
    d = Detection(DetectionClass.FALL, 0.9, (0, 0, 0.1, 0.1))
    assert d.severity == "critical"
    d2 = Detection(DetectionClass.NO_VEST, 0.8, (0, 0, 0.1, 0.1))
    assert d2.severity == "medium"


def test_critical_events_filter():
    result = InferenceResult(
        frame_id="f1", company_id="c1",
        detections=[
            Detection(DetectionClass.HELMET, 0.9, (0, 0, 0.1, 0.1)),
            Detection(DetectionClass.FIRE, 0.85, (0.5, 0.5, 0.1, 0.1)),
            Detection(DetectionClass.NO_HELMET, 0.8, (0.6, 0.2, 0.1, 0.1)),
        ],
        epp_compliant=False, latency_ms=10, model_version="test",
    )
    crits = result.critical_events
    assert len(crits) == 1
    assert crits[0].cls == DetectionClass.FIRE


def test_epp_violations_filter():
    result = InferenceResult(
        frame_id="f1", company_id="c1",
        detections=[
            Detection(DetectionClass.NO_HELMET, 0.8, (0, 0, 0.1, 0.1)),
            Detection(DetectionClass.NO_HARNESS, 0.75, (0, 0, 0.1, 0.1)),
            Detection(DetectionClass.HELMET, 0.9, (0, 0, 0.1, 0.1)),
        ],
        epp_compliant=False, latency_ms=10, model_version="test",
    )
    violations = result.epp_violations
    assert len(violations) == 2
    codes = {v.cls for v in violations}
    assert DetectionClass.NO_HELMET in codes
    assert DetectionClass.NO_HARNESS in codes


def test_infer_returns_result_in_degraded_mode():
    """Sin modelo ONNX, el motor degrada con elegancia y devuelve un resultado."""
    eng = VisionInferenceEngine(model_dir="/nonexistent")
    result = run(eng.infer(b"fake_frame", "frame1", "company1", "ground"))
    assert isinstance(result, InferenceResult)
    assert result.frame_id == "frame1"
    assert result.latency_ms >= 0
