from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from app.core.config import settings
from app.core.privacy_gate import documents_path_enabled
from app.api import generator, certificates, reports


@asynccontextmanager
async def lifespan(app: FastAPI):
    print("✅ ZU Documents Service started — 49 templates ready")
    yield


app = FastAPI(title="ZU Documents Service", version="6.0.0", lifespan=lifespan)

app.add_middleware(CORSMiddleware, allow_origins=settings.ALLOWED_ORIGINS,
                   allow_credentials=True, allow_methods=["*"], allow_headers=["*"])


@app.middleware("http")
async def privacy_gate(request, call_next):
    if not documents_path_enabled(request.url.path):
        return JSONResponse(status_code=404, content={"detail": "resource unavailable"})
    return await call_next(request)

app.include_router(generator.router, prefix="/v1/documents", tags=["Generación"])
app.include_router(certificates.router, prefix="/v1/certificates", tags=["Certificados"])
app.include_router(reports.router, prefix="/v1/reports", tags=["Reportes"])


@app.get("/health")
async def health():
    return {"status": "ok", "service": "zu-documents", "templates": 49}
