import json
from typing import Any, Dict, List, Optional

from fastapi import APIRouter, Depends, WebSocket, WebSocketDisconnect
from langchain.prompts import ChatPromptTemplate
from langchain_openai import ChatOpenAI
from pydantic import BaseModel

from app.core.auth import get_current_tenant
from app.core.config import settings

router = APIRouter()

SISO_SYSTEM = """Eres SISO-Bot, asistente de IA especializado en Seguridad y Salud en el Trabajo (SST) 
para Colombia. Trabajas dentro de ZU Seguridad Integral 360.

Puedes ejecutar acciones REALES en el sistema (siempre pidiendo confirmación previa):
- Crear incidentes
- Generar documentos SST
- Programar rondas de dron
- Enviar correos
- Crear capacitaciones
- Actualizar matrices de riesgo

Tienes acceso a toda la normativa SST colombiana vigente.
Responde siempre en español colombiano formal y conciso.
Cuando vayas a ejecutar una acción, explica qué vas a hacer y pide confirmación explícita."""


class ChatMessage(BaseModel):
    role: str  # user | assistant | system
    content: str
    metadata: Optional[Dict[str, Any]] = None


class ChatRequest(BaseModel):
    message: str
    conversation_id: Optional[str] = None
    history: Optional[List[ChatMessage]] = []
    attachments: Optional[List[str]] = []  # URLs de archivos adjuntos
    company_id: Optional[str] = None
    confirm_action: Optional[bool] = None


class ChatResponse(BaseModel):
    message: str
    conversation_id: str
    action_required: Optional[Dict] = None
    action_executed: Optional[Dict] = None
    sources: Optional[List[str]] = []


@router.post("/message", response_model=ChatResponse)
async def chat_message(
    req: ChatRequest,
    tenant_id: str = Depends(get_current_tenant),
):
    """Endpoint principal del SISO-Bot conversacional."""
    import uuid

    conversation_id = req.conversation_id or str(uuid.uuid4())

    llm = ChatOpenAI(
        model=settings.LLM_MODEL,
        temperature=0.2,
        api_key=settings.OPENAI_API_KEY,
        streaming=False,
    )

    # Construir historial de mensajes
    messages = [("system", SISO_SYSTEM)]
    for msg in req.history[-10:]:  # últimos 10 mensajes
        messages.append((msg.role, msg.content))
    messages.append(("human", req.message))

    prompt = ChatPromptTemplate.from_messages(messages)

    # Detectar intención de acción
    action_keywords = {
        "crear incidente": "create_incident",
        "generar documento": "generate_document",
        "programar ronda": "schedule_drone",
        "enviar correo": "send_email",
        "programar capacitación": "create_training",
    }

    detected_action = None
    msg_lower = req.message.lower()
    for keyword, action in action_keywords.items():
        if keyword in msg_lower:
            detected_action = action
            break

    chain = prompt | llm
    response = chain.invoke({})
    answer = response.content

    # Si hay acción detectada y no hay confirmación aún
    if detected_action and req.confirm_action is None:
        return ChatResponse(
            message=answer,
            conversation_id=conversation_id,
            action_required={
                "type": detected_action,
                "description": f"¿Confirmas que deseas ejecutar: {detected_action}?",
                "confirm_field": "confirm_action",
            },
        )

    # Si hay confirmación positiva, ejecutar acción
    action_result = None
    if detected_action and req.confirm_action is True:
        action_result = await execute_action(detected_action, req, tenant_id)

    return ChatResponse(
        message=answer,
        conversation_id=conversation_id,
        action_executed=action_result,
    )


async def execute_action(action_type: str, req: ChatRequest, tenant_id: str) -> Dict:
    """Ejecuta acciones reales en el sistema ZU."""

    actions = {
        "create_incident": lambda: {"status": "pending", "message": "Creando incidente..."},
        "generate_document": lambda: {"status": "pending", "message": "Generando documento..."},
        "schedule_drone": lambda: {"status": "pending", "message": "Programando ronda de dron..."},
        "send_email": lambda: {"status": "pending", "message": "Preparando correo..."},
    }

    handler = actions.get(action_type)
    return handler() if handler else {"status": "error", "message": "Acción no reconocida"}


@router.websocket("/ws/{conversation_id}")
async def chat_websocket(websocket: WebSocket, conversation_id: str):
    """WebSocket para streaming del SISO-Bot."""
    await websocket.accept()
    try:
        while True:
            data = await websocket.receive_text()
            msg = json.loads(data)

            llm = ChatOpenAI(
                model=settings.LLM_MODEL,
                temperature=0.2,
                api_key=settings.OPENAI_API_KEY,
                streaming=True,
            )

            prompt = ChatPromptTemplate.from_messages([
                ("system", SISO_SYSTEM),
                ("human", msg.get("message", "")),
            ])

            chain = prompt | llm

            async for chunk in chain.astream({}):
                await websocket.send_text(json.dumps({
                    "type": "stream",
                    "content": chunk.content,
                    "conversation_id": conversation_id,
                }))

            await websocket.send_text(json.dumps({"type": "end"}))

    except WebSocketDisconnect:
        pass
