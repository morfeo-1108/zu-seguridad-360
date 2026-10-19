from fastapi import APIRouter
router = APIRouter()

@router.get("/")
async def list(): return {"message": "49 templates available — see /v1/documents/templates"}
