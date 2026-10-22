import uuid

from fastapi import Header, HTTPException, status
from jose import JWTError, jwt

from app.core.config import settings


def _uuid(value) -> str:
    try:
        return str(uuid.UUID(str(value)))
    except (ValueError, TypeError):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token") from None


async def get_current_user(authorization: str = Header(None)) -> dict:
    """Validates Core's HS256 access token with Core's own rules
    (services/core/internal/middleware/auth.go): access tokens only, expiry
    required, sub == user_id, UUID claims. Scope (tenant, company) always
    comes from these signed claims, never from the request body."""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing token")
    token = authorization.split(" ", 1)[1].strip()
    try:
        payload = jwt.decode(token, settings.JWT_SECRET, algorithms=["HS256"], options={"require_exp": True})
    except JWTError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token") from None
    if payload.get("token_type") != "access" or not payload.get("role") or payload.get("sub") != payload.get("user_id"):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
    company_id = payload.get("company_id")
    return {
        "tenant_id": _uuid(payload.get("tenant_id")),
        "user_id": _uuid(payload.get("user_id")),
        "role": payload["role"],
        "company_id": _uuid(company_id) if company_id else "",
        "email": payload.get("email", ""),
    }


def require_company(user: dict) -> str:
    if not user["company_id"]:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail="No active company in session")
    return user["company_id"]
