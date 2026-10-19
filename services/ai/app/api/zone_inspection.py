"""API de inspección por zona: el dron envía un frame y recibe validación por trabajador."""
from fastapi import APIRouter, UploadFile, File, Form, Depends, HTTPException
from typing import Any
import json
import httpx

from app.vision.compliance_validator import validator
from app.vision.biometric import biometric_engine
from app.core.config import settings
from app.core.auth import get_current_user

router = APIRouter()


@router.post("/enroll-biometric")
async def enroll_worker(
    employee_id: str = Form(...),
    face: UploadFile = File(...),
    consent_verified: bool = Form(False),
    user=Depends(get_current_user),
):
    """Enrola el rostro de un trabajador. Requiere consentimiento Ley 1581."""
    face_bytes = await face.read()
    result = await biometric_engine.enroll(user["tenant_id"], employee_id, face_bytes, consent_verified)
    if not result.enrolled:
        raise HTTPException(status_code=422, detail=result.error)
    return {"employee_id": result.employee_id, "enrolled": True,
            "liveness_passed": result.liveness_passed, "quality_score": result.quality_score}


@router.post("/inspect")
async def inspect_zone(
    project_id: str = Form(...),
    frame: UploadFile = File(...),
    face_crops: str = Form("[]"),     # lista de rostros base64, o se extraen del frame
    frame_id: str = Form(...),
    lat: float = Form(None),
    lng: float = Form(None),
    marker_id: str = Form(None),
    user=Depends(get_current_user),
):
    """
    Flujo completo de inspección del dron:
      1. Resuelve la ZONA (por marcador visual o GPS) consultando a core.
      2. Obtiene el checklist de esa zona.
      3. Valida visión + biometría por trabajador.
      4. Escala lo crítico.
    """
    frame_bytes = await frame.read()

    # 1. Reconocimiento de zona vía core
    zone = await _resolve_zone(user, project_id, marker_id, lat, lng)
    if not zone:
        raise HTTPException(status_code=404, detail="No se pudo reconocer la zona (GPS/marcador)")

    # 2. Nombres de empleados para el reporte
    employee_names = await _fetch_employee_names(user, zone.get("company_id"))

    # 3. Decodificar rostros (en producción se extraen del frame con el detector facial)
    try:
        crops = json.loads(face_crops)
        face_bytes_list = [c.encode() if isinstance(c, str) else c for c in crops]
    except Exception:
        face_bytes_list = []

    # 4. Validar
    result = await validator.validate_zone_frame(
        tenant_id=user["tenant_id"], project_id=project_id, zone=zone,
        frame_bytes=frame_bytes, face_crops=face_bytes_list,
        frame_id=frame_id, employee_names=employee_names,
    )

    # 5. Registrar evidencia fotográfica + eventos de agente (registro + feed en vivo)
    await _emit_evidence_and_events(user, project_id, result)

    # 6. Escalar críticos a notificaciones + SLA
    if result.needs_escalation:
        await _escalate_critical(user, result)

    return {
        "zone": {"id": result.zone_id, "name": result.zone_name, "type": result.zone_type},
        "zone_compliance_pct": result.zone_compliance_pct,
        "needs_escalation": result.needs_escalation,
        "unidentified_count": result.unidentified_count,
        "workers": [
            {
                "employee_id": w.employee_id,
                "employee_name": w.employee_name,
                "identification_confidence": w.identification_confidence,
                "compliant": w.compliant,
                "compliance_pct": w.compliance_pct,
                "missing_items": [{"code": i.code, "description": i.description, "severity": i.severity}
                                  for i in w.missing_items],
                "missing_critical": [i.code for i in w.missing_critical],
            }
            for w in result.workers
        ],
        "latency_ms": result.latency_ms,
    }


async def _resolve_zone(user, project_id, marker_id, lat, lng) -> dict[str, Any] | None:
    async with httpx.AsyncClient(timeout=10.0) as http:
        headers = {"Authorization": f"Bearer {user.get('token','')}", "X-Tenant-ID": user["tenant_id"]}
        if marker_id:
            url = f"{settings.CORE_URL}/v1/projects/{project_id}/zones/resolve-marker?marker_id={marker_id}"
        elif lat is not None and lng is not None:
            url = f"{settings.CORE_URL}/v1/projects/{project_id}/zones/resolve-location?lat={lat}&lng={lng}"
        else:
            return None
        try:
            r = await http.get(url, headers=headers)
            return r.json() if r.status_code < 300 else None
        except Exception:
            return None


async def _fetch_employee_names(user, company_id) -> dict[str, str]:
    return {}  # en producción: consulta a core el directorio


async def _escalate_critical(user, result):
    """Notifica al supervisor y abre SLA por cada trabajador con falta crítica."""
    async with httpx.AsyncClient(timeout=10.0) as http:
        for w in result.workers:
            if w.missing_critical:
                msg = (f"{w.employee_name or w.employee_id} en zona {result.zone_name}: "
                       f"falta crítica de {', '.join(i.code for i in w.missing_critical)}")
                try:
                    await http.post(f"{settings.NOTIFICATIONS_URL}/v1/dispatch",
                                    json={"channel": "whatsapp", "message": msg,
                                          "recipients": [user.get("supervisor", "")]})
                except Exception:
                    pass


async def _emit_evidence_and_events(user, project_id, result, frame_url=""):
    """Por cada trabajador validado, registra prueba fotográfica + evento de agente."""
    async with httpx.AsyncClient(timeout=10.0) as http:
        headers = {"Authorization": f"Bearer {user.get('token','')}", "X-Tenant-ID": user["tenant_id"]}
        for w in result.workers:
            compliant = w.compliant
            # 1. Prueba fotográfica ligada al trabajador
            for item in w.items:
                try:
                    await http.post(f"{settings.CORE_URL}/v1/evidence", headers=headers, json={
                        "employee_id": w.employee_id,
                        "project_id": project_id,
                        "zone_id": result.zone_id,
                        "type": "epp_validation",
                        "photo_url": frame_url,
                        "source": "drone",
                        "checklist_code": item.code,
                        "compliant": item.detected,
                        "severity": item.severity,
                        "detection_data": {"confidence": item.confidence},
                        "captured_at": __import__("datetime").datetime.utcnow().isoformat(),
                    })
                except Exception:
                    pass
            # 2. Evento de agente (alimenta el feed en vivo + registro)
            action = (f"Validó a {w.employee_name or w.employee_id} en zona {result.zone_name}: "
                      f"{w.compliance_pct}% cumplimiento" +
                      (f" — FALTA CRÍTICA: {', '.join(i.code for i in w.missing_critical)}" if w.missing_critical else ""))
            try:
                await http.post(f"{settings.CORE_URL}/v1/agent-events", headers=headers, json={
                    "agent_key": "vision", "agent_class": "reasoning",
                    "action": action,
                    "severity": "critical" if w.missing_critical else ("warning" if not compliant else "info"),
                    "resource_type": "zone", "employee_id": w.employee_id,
                    "project_id": project_id,
                    "metadata": {"zone": result.zone_name, "compliance_pct": w.compliance_pct},
                })
            except Exception:
                pass
