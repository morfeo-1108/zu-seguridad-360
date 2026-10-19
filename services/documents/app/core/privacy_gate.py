"""Fail-closed route policy for the Documents service pilot."""


ALLOWED_PATHS = frozenset({"/health"})


def documents_path_enabled(path: str) -> bool:
    """Only health is exposed until authz, privacy, and retention are approved."""
    return path in ALLOWED_PATHS
