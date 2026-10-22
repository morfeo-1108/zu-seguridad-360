"""Consulta normativa SST (RAG de recuperación, sin generación libre).

La respuesta son los artículos vigentes que responden la pregunta, con norma,
referencia, texto literal y URL oficial. No hay un modelo de lenguaje
redactando: lo que el profesional lee es el texto de la norma. Las normas o
artículos derogados nunca se devuelven como fundamento; si compiten con la
respuesta se informan como advertencia.
"""
import asyncio
import threading

from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel, Field

from app.core.auth import get_current_user
from app.rag import normative as nv

router = APIRouter()

_lock = threading.Lock()
_state: dict = {}


def _resources():
    # Lazy: the ONNX model (~200 MB in RAM) loads on first use, once.
    with _lock:
        if "client" not in _state:
            _state["client"] = nv.connect()
            _state["embedder"] = nv.Embedder()
        return _state["client"], _state["embedder"]


class Query(BaseModel):
    question: str = Field(min_length=5, max_length=500)
    top_k: int = Field(default=3, ge=1, le=8)


class Source(BaseModel):
    norm: str
    reference: str
    article_title: str
    status: str
    source_url: str
    text: str
    citation: str


class Answer(BaseModel):
    question: str
    sources: list[Source]
    warnings: list[str]
    note: str = "Texto literal de la norma vigente. Verifique la aplicabilidad a su caso; no sustituye un concepto jurídico."


def _run(q: Query) -> Answer:
    client, emb = _resources()
    res = nv.search(client, emb, q.question, k=q.top_k)
    return Answer(
        question=q.question,
        sources=[Source(norm=h.norm_name, reference=h.reference, article_title=h.article_title, status=h.status,
                        source_url=h.source_url, text=h.content, citation=h.citation) for h in res.hits],
        warnings=res.warnings,
    )


@router.post("/query", response_model=Answer)
async def query_normative(q: Query, user: dict = Depends(get_current_user)):
    try:
        return await asyncio.to_thread(_run, q)
    except Exception:
        raise HTTPException(status_code=503, detail="Motor normativo no disponible")


@router.get("/norms")
async def list_norms(user: dict = Depends(get_current_user)):
    """Normas cargadas en la base de consulta, con su estado de vigencia."""
    import json
    src = json.loads((nv.CORPUS.parent / "sources.json").read_text(encoding="utf-8"))
    return {"norms": [{k: n.get(k) for k in ("name", "issuer", "year", "status", "derogated_by", "topic", "url")} for n in src["norms"]]}
