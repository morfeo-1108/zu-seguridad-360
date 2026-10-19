import pytest

from app.core.privacy_gate import documents_path_enabled


def test_health_is_the_only_allowed_documents_path():
    assert documents_path_enabled("/health") is True


@pytest.mark.parametrize(
    "path",
    [
        "/v1/documents/generate",
        "/v1/documents/generate/inline",
        "/v1/certificates/generate",
        "/v1/certificates/verify/demo-folio",
        "/v1/reports/monthly/generate",
        "/v1/reports/res0312/generate",
        "/openapi.json",
        "/docs",
        "/unknown",
    ],
)
def test_personal_data_and_api_discovery_routes_are_closed(path):
    assert documents_path_enabled(path) is False
