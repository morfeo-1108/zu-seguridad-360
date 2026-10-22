import logging
import uuid
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from prometheus_fastapi_instrumentator import Instrumentator

from app.api import rag
from app.core.config import settings
from app.core.feature_gates import ai_path_enabled
from app.core.privacy_errors import public_error_payload

logger = logging.getLogger(__name__)


# Feature groups whose code (and heavy dependencies: langchain, neo4j, vision)
# is imported only when their flag is on. With the flags off — the production
# default — they are not loaded at all, not just hidden behind the gate.
AUTONOMOUS = settings.AI_AUTONOMOUS_ENABLED and settings.PRIVACY_PROCESSING_APPROVED
VISION = settings.VISION_ENABLED and settings.PRIVACY_PROCESSING_APPROVED
SENSITIVE = settings.SENSITIVE_AI_ENABLED and settings.PRIVACY_PROCESSING_APPROVED


@asynccontextmanager
async def lifespan(app: FastAPI):
    if AUTONOMOUS or VISION or SENSITIVE:
        from app.core.database import init_db
        await init_db()
    if AUTONOMOUS:
        from app.core.neo4j_client import init_neo4j
        from app.core.weaviate_client import init_weaviate
        await init_weaviate()
        await init_neo4j()
    print(f"ZU AI Service started (rag=on autonomous={AUTONOMOUS} vision={VISION} sensitive={SENSITIVE})")
    yield


app = FastAPI(
    title="ZU AI Service",
    description="Motor de Inteligencia Artificial para ZU Seguridad Integral 360",
    version="6.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=settings.ALLOWED_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.middleware("http")
async def high_risk_feature_gate(request: Request, call_next):
    if not ai_path_enabled(request.url.path, settings):
        return JSONResponse(status_code=404, content={"detail": "AI capability disabled"})
    return await call_next(request)

# Prometheus metrics
Instrumentator().instrument(app).expose(app)

# Routers
app.include_router(rag.router, prefix="/v1/rag", tags=["RAG Normativo"])
if AUTONOMOUS:
    from app.api import agents, chat, cognition, documents_ai, incidents_ai
    app.include_router(documents_ai.router, prefix="/v1/documents", tags=["Generación Documental"])
    app.include_router(incidents_ai.router, prefix="/v1/incidents", tags=["Análisis Incidentes"])
    app.include_router(cognition.router, prefix="/v1/cognition", tags=["Cognición Adaptativa"])
    app.include_router(chat.router, prefix="/v1/chat", tags=["SISO-Bot"])
    app.include_router(agents.router, prefix="/v1/agents", tags=["Agentes"])
if VISION:
    from app.api import vision, zone_inspection
    app.include_router(vision.router, prefix="/v1/vision", tags=["Visión"])
    app.include_router(zone_inspection.router, prefix="/v1/zone-inspection", tags=["Inspección por Zona"])
if SENSITIVE:
    from app.api import ocr
    app.include_router(ocr.router, prefix="/v1/ocr", tags=["OCR Médico"])


@app.get("/health")
async def health():
    return {"status": "ok", "service": "zu-ai", "version": "6.1.0", "rag": True,
            "autonomous": AUTONOMOUS, "vision": VISION, "sensitive": SENSITIVE}


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    request_id = uuid.uuid4().hex
    route = request.scope.get("route")
    route_template = getattr(route, "path", "unmatched")
    # Do not log exception messages, request URLs, query strings, or bodies.
    logger.error(
        "unhandled_request request_id=%s method=%s route=%s exception_type=%s",
        request_id,
        request.method,
        route_template,
        type(exc).__name__,
    )
    return JSONResponse(
        status_code=500,
        content=public_error_payload(request_id),
        headers={"X-Request-ID": request_id},
    )
