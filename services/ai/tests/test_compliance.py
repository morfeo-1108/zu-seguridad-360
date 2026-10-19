"""Tests del validador de cumplimiento — checklist por trabajador y zona."""
import sys, os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.vision.compliance_validator import (
    ComplianceValidator, WorkerCompliance, WorkerComplianceItem,
)


def make_worker(items):
    return WorkerCompliance(
        employee_id="emp1", employee_name="Carlos", identified=True,
        identification_confidence=0.9, items=items,
    )


def test_fully_compliant_worker():
    """Trabajador con todo el EPP obligatorio detectado = cumple."""
    items = [
        WorkerComplianceItem("casco", "Casco", True, "high", detected=True, confidence=0.9),
        WorkerComplianceItem("arnes", "Arnés", True, "critical", detected=True, confidence=0.9),
    ]
    w = make_worker(items)
    assert w.compliant is True
    assert w.compliance_pct == 100.0
    assert len(w.missing_critical) == 0


def test_missing_critical_epp():
    """Falta de arnés (crítico) en alturas = no cumple, falta crítica detectada."""
    items = [
        WorkerComplianceItem("casco", "Casco", True, "high", detected=True, confidence=0.9),
        WorkerComplianceItem("arnes", "Arnés", True, "critical", detected=False, confidence=0.0),
        WorkerComplianceItem("linea_vida", "Línea de vida", True, "critical", detected=False, confidence=0.0),
    ]
    w = make_worker(items)
    assert w.compliant is False
    assert len(w.missing_critical) == 2
    assert {i.code for i in w.missing_critical} == {"arnes", "linea_vida"}


def test_compliance_pct_partial():
    """2 de 4 obligatorios cumplidos = 50%."""
    items = [
        WorkerComplianceItem("a", "A", True, "high", detected=True, confidence=0.9),
        WorkerComplianceItem("b", "B", True, "high", detected=True, confidence=0.9),
        WorkerComplianceItem("c", "C", True, "high", detected=False, confidence=0.0),
        WorkerComplianceItem("d", "D", True, "high", detected=False, confidence=0.0),
    ]
    w = make_worker(items)
    assert w.compliance_pct == 50.0


def test_optional_items_dont_affect_compliance():
    """Un ítem no obligatorio faltante no rompe el cumplimiento."""
    items = [
        WorkerComplianceItem("casco", "Casco", True, "high", detected=True, confidence=0.9),
        WorkerComplianceItem("gafas", "Gafas", False, "low", detected=False, confidence=0.0),
    ]
    w = make_worker(items)
    assert w.compliant is True  # gafas es opcional


def test_evaluate_checklist_detects_present_epp():
    """El validador marca cumplido lo que está en las clases detectadas."""
    validator = ComplianceValidator()
    checklist = [
        {"code": "casco", "description": "Casco", "mandatory": True, "severity": "high", "detection_class": "casco"},
        {"code": "arnes", "description": "Arnés", "mandatory": True, "severity": "critical", "detection_class": "arnes"},
    ]
    present = {"casco"}  # solo casco detectado
    items = validator._evaluate_checklist(checklist, present)
    by_code = {i.code: i for i in items}
    assert by_code["casco"].detected is True
    assert by_code["arnes"].detected is False
