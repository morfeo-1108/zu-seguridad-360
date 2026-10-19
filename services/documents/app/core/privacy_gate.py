"""Route policy for the Documents service pilot.

Per explicit product decision (2026-10-19), document/certificate/report
generation is unblocked for the synthetic pilot now that every generating
route requires a valid Core-issued JWT (see app/core/auth.py). This service
has no direct access to Core's database, so it still cannot itself verify
that request-body fields belong to the caller's active company — that
remains a gap, not something this gate can close. /v1/certificates/verify/*
stays public by design (QR-code verification). API discovery (/docs,
/openapi.json) and anything unlisted stay closed.
"""

ALLOWED_PREFIXES = (
    "/health",
    "/v1/documents/generate",
    "/v1/certificates/generate",
    "/v1/certificates/verify/",
    "/v1/reports/monthly/generate",
    "/v1/reports/res0312/generate",
)


def documents_path_enabled(path: str) -> bool:
    return any(path == p or path.startswith(p) for p in ALLOWED_PREFIXES)
