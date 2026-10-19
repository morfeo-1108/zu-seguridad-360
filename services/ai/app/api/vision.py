"""API de inferencia de visión para el sistema de drones."""
from fastapi import APIRouter, UploadFile, File, Form
from app.vision.inference import engine

router = APIRouter()


@router.post("/infer")
async def infer_frame(
    file: UploadFile = File(...),
    company_id: str = Form(...),
    frame_id: str = Form(...),
    altitude_context: str = Form("ground"),
):
    """Ejecuta inferencia de visión sobre un frame del dron."""
    frame_bytes = await file.read()
    result = await engine.infer(frame_bytes, frame_id, company_id, altitude_context)
    return {
        "frame_id": result.frame_id,
        "company_id": result.company_id,
        "epp_compliant": result.epp_compliant,
        "detections": [
            {"class": d.cls.value, "confidence": d.confidence, "severity": d.severity, "bbox": d.bbox}
            for d in result.detections
        ],
        "critical_events": [d.cls.value for d in result.critical_events],
        "epp_violations": [d.cls.value for d in result.epp_violations],
        "latency_ms": result.latency_ms,
        "model_version": result.model_version,
    }
