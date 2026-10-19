"""Public error payloads that never include exception or request contents."""


def public_error_payload(request_id: str) -> dict[str, str]:
    return {"detail": "Internal server error", "request_id": request_id}


def ocr_failure_payload() -> dict[str, object]:
    """Return only empty, non-identifying OCR fields after provider failure."""
    return {"text": "", "confidence": 0.0, "fields": {}}
