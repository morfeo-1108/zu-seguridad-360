from app.core.privacy_errors import ocr_failure_payload, public_error_payload


def test_public_error_payload_is_generic_and_does_not_accept_exception_text():
    payload = public_error_payload("request-123")
    assert payload == {"detail": "Internal server error", "request_id": "request-123"}
    assert "traceback" not in " ".join(payload.values()).lower()
    assert "query" not in payload


def test_ocr_failure_payload_contains_no_provider_or_document_details():
    payload = ocr_failure_payload()
    assert payload == {"text": "", "confidence": 0.0, "fields": {}}
