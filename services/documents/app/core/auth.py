from fastapi import Header, HTTPException, status
from jose import jwt, JWTError
from app.core.config import settings


async def get_current_user(authorization: str = Header(None)) -> dict:
    """Decodes Core's JWT (same HS256 secret/claims) so document/report/
    certificate generation requires an authenticated session. This service has
    no direct access to Core's database, so it cannot itself verify that
    request-body fields (employee name, company, etc.) belong to the caller's
    active company — only Core's own endpoints enforce that today."""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing token")
    token = authorization.split(" ")[1]
    try:
        payload = jwt.decode(token, settings.JWT_SECRET, algorithms=["HS256"])
    except JWTError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
    tenant_id = payload.get("tenant_id")
    user_id = payload.get("user_id")
    if not tenant_id or not user_id:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
    return {
        "tenant_id": tenant_id,
        "user_id": user_id,
        "role": payload.get("role", ""),
        "company_id": payload.get("company_id", ""),
        "token": token,
    }
