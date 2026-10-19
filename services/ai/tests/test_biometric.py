"""Tests del motor biométrico — enrolamiento, identificación, umbral de match."""
import asyncio
import numpy as np
import pytest
import sys, os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from app.vision.biometric import BiometricEngine, FaceMatch


def run(coro):
    return asyncio.get_event_loop().run_until_complete(coro)


def test_cosine_identical_is_one():
    eng = BiometricEngine()
    v = np.array([1.0, 2.0, 3.0])
    assert abs(eng._cosine(v, v) - 1.0) < 1e-6


def test_cosine_orthogonal_is_zero():
    eng = BiometricEngine()
    a = np.array([1.0, 0.0])
    b = np.array([0.0, 1.0])
    assert abs(eng._cosine(a, b)) < 1e-6


def test_enroll_requires_consent():
    """Ley 1581: sin consentimiento verificado, NO se enrola el dato biométrico."""
    eng = BiometricEngine()
    result = run(eng.enroll("tenant1", "emp1", b"face_data", consent_verified=False))
    assert result.enrolled is False
    assert "onsentimiento" in (result.error or "")


def test_enroll_with_consent_succeeds():
    eng = BiometricEngine()
    result = run(eng.enroll("tenant1", "emp1", b"face_data_unique", consent_verified=True))
    assert result.enrolled is True
    assert result.liveness_passed is True


def test_identify_matches_enrolled_worker():
    """Un trabajador enrolado se identifica a sí mismo con su propio rostro."""
    eng = BiometricEngine()
    face = b"carlos_face_bytes"
    run(eng.enroll("tenantX", "carlos", face, consent_verified=True))
    match = run(eng.identify("tenantX", face))
    assert match is not None
    assert match.employee_id == "carlos"
    assert match.is_match is True
    assert match.similarity >= eng.MATCH_THRESHOLD


def test_identify_empty_index_returns_none():
    eng = BiometricEngine()
    match = run(eng.identify("tenant_vacio", b"cualquier_rostro"))
    assert match is None


def test_identify_isolates_by_tenant():
    """Aislamiento multi-tenant: un rostro de tenant A no se busca en tenant B."""
    eng = BiometricEngine()
    run(eng.enroll("tenantA", "empA", b"face_A", consent_verified=True))
    # Buscar en tenant B (vacío) no debe encontrar nada
    match = run(eng.identify("tenantB", b"face_A"))
    assert match is None
