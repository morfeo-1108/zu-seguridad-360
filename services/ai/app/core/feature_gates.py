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
    """
    # These data classes have no verified Core consent/company authorization
    # integration. Keep them closed even if an operator toggles feature flags.
    if _matches(path, "/v1/ocr") or _matches(path, "/v1/zone-inspection"):
        return False
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
