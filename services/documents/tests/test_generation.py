import time

import pytest
from fastapi import HTTPException
from jose import jwt

from app.core import auth
from app.core.config import settings
from app.services.pdf_generator import TEMPLATES, DocumentGenerator

UID = "11111111-1111-4111-8111-111111111111"
TID = "22222222-2222-4222-8222-222222222222"


def token(**over):
    claims = {"user_id": UID, "sub": UID, "tenant_id": TID, "role": "professional_sst",
              "token_type": "access", "exp": int(time.time()) + 60}
    claims.update(over)
    claims = {k: v for k, v in claims.items() if v is not None}
    return "Bearer " + jwt.encode(claims, settings.JWT_SECRET, algorithm="HS256")


@pytest.mark.asyncio
async def test_accepts_core_access_token():
    user = await auth.get_current_user(token(company_id="33333333-3333-4333-8333-333333333333"))
    assert user["tenant_id"] == TID and user["company_id"].startswith("3333")


@pytest.mark.asyncio
@pytest.mark.parametrize("over", [
    {"token_type": "refresh"}, {"sub": "other"}, {"exp": None}, {"exp": int(time.time()) - 10},
    {"tenant_id": "not-a-uuid"}, {"role": None},
])
async def test_rejects_invalid_tokens(over):
    with pytest.raises(HTTPException) as e:
        await auth.get_current_user(token(**over))
    assert e.value.status_code == 401


def test_every_catalog_template_renders_a_pdf():
    gen = DocumentGenerator()
    for key in TEMPLATES:
        pdf = gen.generate(key, {"company_name": "ACME & Cía", "content": "Texto <b>no</b> es marcado", "folio": "F-1"}, "doc-1")
        assert pdf.startswith(b"%PDF"), key


def test_user_text_is_not_interpreted_as_markup():
    # An unbalanced tag would make ReportLab raise if it reached the parser unescaped.
    pdf = DocumentGenerator().generate("generico", {"content": "<link href='x'>a < b & c", "company_name": "X"}, "d")
    assert pdf.startswith(b"%PDF")
