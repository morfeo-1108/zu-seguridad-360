"""Fail-closed gates for AI endpoints in the closed pilot."""

from collections.abc import Mapping
from typing import Any


def _enabled(flags: Any, name: str) -> bool:
    if isinstance(flags, Mapping):
        value = flags.get(name, False)
    else:
        value = getattr(flags, name, False)
    return value is True


def _matches(path: str, prefix: str) -> bool:
    return path == prefix or path.startswith(prefix + "/")


def ai_path_enabled(path: str, flags: Any) -> bool:
    """Return whether a request path is enabled by explicit boolean flags.

    Health/metrics endpoints are intentionally not controlled here. Every API
    path under /v1 is closed unless AI_SERVICE_API_ENABLED is explicitly true;
    higher-risk groups additionally require their own feature flag.

    /v1/ocr and /v1/zone-inspection used to be hardcoded closed regardless of
    flags, pending a verified Core consent/company authorization integration.
    zone-inspection's biometric enroll now verifies company membership against
    Core before enrolling (see _employee_belongs_to_company in
    app/api/zone_inspection.py); ocr/medical-exam never persists data itself —
    review/storage happens through Core's own company-scoped endpoint. Both are
    now governed only by their feature flags, same as every other group below.
    consent_verified on biometric enroll remains client-reported, not a real
    consent-capture/verification system.
    """
    if not path.startswith("/v1/"):
        return True
    if not _enabled(flags, "AI_SERVICE_API_ENABLED"):
        return False
    if not _enabled(flags, "PRIVACY_PROCESSING_APPROVED"):
        return False

    gates = (
        (("/v1/agents", "/v1/chat", "/v1/cognition", "/v1/incidents"), "AI_AUTONOMOUS_ENABLED"),
        (("/v1/vision", "/v1/zone-inspection"), "VISION_ENABLED"),
        (("/v1/ocr",), "SENSITIVE_AI_ENABLED"),
    )
    for prefixes, flag_name in gates:
        if any(_matches(path, prefix) for prefix in prefixes):
            return _enabled(flags, flag_name)
    return True
