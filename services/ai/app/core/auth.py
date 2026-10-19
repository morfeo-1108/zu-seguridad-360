from fastapi import Header, HTTPException, status
from jose import jwt, JWTError
from app.core.config import settings

async def get_current_tenant(authorization: str = Header(None)) -> str:
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing token")
    token = authorization.split(" ")[1]
    try:
        payload = jwt.decode(token, settings.JWT_SECRET, algorithms=["HS256"])
        tenant_id = payload.get("tenant_id")
        if not tenant_id:
            raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
        return tenant_id
    except JWTError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")


async def get_current_user(authorization: str = Header(None)) -> dict:
    """Decodes Core's JWT (same HS256 secret/claims) into the shape AI routes expect:
    tenant_id, user_id, role, company_id, email, and the raw token for service-to-service calls."""
    if not authorization or not authorization.startswith("Bearer "):
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Missing token")
    token = authorization.split(" ")[1]
    try:
        payload = jwt.decode(token, settings.JWT_SECRET, algorithms=["HS256"])
    except JWTError:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
    tenant_id = payload.get("tenant_id")
    user_id = payload.get("user_id")
    role = payload.get("role")
    if not tenant_id or not user_id or not role:
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED, detail="Invalid token")
    return {
        "tenant_id": tenant_id,
        "user_id": user_id,
        "role": role,
        "company_id": payload.get("company_id", ""),
        "email": payload.get("email", ""),
        "token": token,
    }
