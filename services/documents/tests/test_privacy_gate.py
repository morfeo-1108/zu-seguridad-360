import pytest

from app.core.privacy_gate import documents_path_enabled


def test_health_is_allowed():
    assert documents_path_enabled("/health") is True


@pytest.mark.parametrize(
    "path",
    [
        "/v1/documents/generate",
        "/v1/documents/generate/inline",
        "/v1/certificates/generate",
        "/v1/documents",
        "/v1/documents/templates",
        "/v1/documents/0a0a0a0a-0000-4000-8000-000000000001/download",
        "/v1/reports/monthly/generate",
        "/v1/reports/res0312/generate",
    ],
)
def test_generation_and_public_verification_routes_are_open(path):
    """Unblocked 2026-10-19: every generating route now requires a valid
    Core-issued JWT (see app/core/auth.py); /verify/* stays public by design
    (QR-code verification), not because of a gap."""
    assert documents_path_enabled(path) is True


@pytest.mark.parametrize("path", ["/openapi.json", "/docs", "/unknown", "/v1/certificates/verify/demo-folio"])
def test_api_discovery_and_unknown_routes_stay_closed(path):
    assert documents_path_enabled(path) is False
