from types import SimpleNamespace

from app.core.config import Settings
from app.core.feature_gates import ai_path_enabled


def test_all_ai_api_paths_are_closed_by_default():
    flags = SimpleNamespace(
        AI_SERVICE_API_ENABLED=False,
        AI_AUTONOMOUS_ENABLED=False,
        VISION_ENABLED=False,
        BIOMETRIC_ENABLED=False,
        SENSITIVE_AI_ENABLED=False,
        PRIVACY_PROCESSING_APPROVED=False,
    )
    for path in (
        "/v1/rag/query",
        "/v1/agents/run",
        "/v1/vision/infer",
        "/v1/ocr/process",
        "/v1/zone-inspection/analyze",
        "/v1/zone-inspection/enroll-biometric",
    ):
        assert ai_path_enabled(path, flags) is False
    assert ai_path_enabled("/health", flags) is True


def test_high_risk_features_require_independent_opt_in():
    flags = {
        "AI_SERVICE_API_ENABLED": True,
        "AI_AUTONOMOUS_ENABLED": False,
        "VISION_ENABLED": False,
        "BIOMETRIC_ENABLED": False,
        "SENSITIVE_AI_ENABLED": False,
        "PRIVACY_PROCESSING_APPROVED": True,
    }
    assert ai_path_enabled("/v1/rag/query", flags) is True
    assert ai_path_enabled("/v1/agents/run", flags) is False
    assert ai_path_enabled("/v1/vision/infer", flags) is False
    assert ai_path_enabled("/v1/ocr/process", flags) is False


def test_enabled_feature_does_not_match_similar_prefix():
    flags = {
        "AI_SERVICE_API_ENABLED": True,
        "AI_AUTONOMOUS_ENABLED": False,
        "VISION_ENABLED": True,
        "BIOMETRIC_ENABLED": False,
        "SENSITIVE_AI_ENABLED": False,
        "PRIVACY_PROCESSING_APPROVED": True,
    }
    assert ai_path_enabled("/v1/vision/infer", flags) is True
    assert ai_path_enabled("/v1/visionary/report", flags) is True


def test_settings_default_all_high_risk_flags_to_false(monkeypatch):
    for name in (
        "AI_SERVICE_API_ENABLED",
        "AI_AUTONOMOUS_ENABLED",
        "VISION_ENABLED",
        "BIOMETRIC_ENABLED",
        "SENSITIVE_AI_ENABLED",
        "PRIVACY_PROCESSING_APPROVED",
    ):
        monkeypatch.delenv(name, raising=False)
    settings = Settings(_env_file=None)
    assert settings.AI_SERVICE_API_ENABLED is False
    assert settings.AI_AUTONOMOUS_ENABLED is False
    assert settings.VISION_ENABLED is False
    assert settings.BIOMETRIC_ENABLED is False
    assert settings.SENSITIVE_AI_ENABLED is False
    assert settings.PRIVACY_PROCESSING_APPROVED is False


def test_medical_ocr_and_biometric_enrollment_follow_their_own_flags():
    """Until 2026-10-19 these paths stayed hardcoded closed regardless of flags,
    pending a verified Core consent/company authorization integration. That
    integration now exists for biometric enroll (company membership is checked
    against Core before enrolling — see _employee_belongs_to_company in
    app/api/zone_inspection.py) and was never needed for OCR (it never persists
    data itself). Per explicit product decision they now follow the same
    per-group flag as every other route; consent_verified on biometric enroll
    is still client-reported, not a verified consent-capture system."""
    flags = {
        "AI_SERVICE_API_ENABLED": True,
        "AI_AUTONOMOUS_ENABLED": True,
        "VISION_ENABLED": True,
        "BIOMETRIC_ENABLED": True,
        "SENSITIVE_AI_ENABLED": True,
        "PRIVACY_PROCESSING_APPROVED": True,
    }
    assert ai_path_enabled("/v1/ocr/medical-exam", flags) is True
    assert ai_path_enabled("/v1/zone-inspection/enroll-biometric", flags) is True
    assert ai_path_enabled("/v1/zone-inspection/inspect", flags) is True
    assert ai_path_enabled("/v1/vision/infer", flags) is True

    flags_no_sensitive = {**flags, "SENSITIVE_AI_ENABLED": False}
    assert ai_path_enabled("/v1/ocr/medical-exam", flags_no_sensitive) is False

    flags_no_vision = {**flags, "VISION_ENABLED": False}
    assert ai_path_enabled("/v1/zone-inspection/enroll-biometric", flags_no_vision) is False
