"""Recuperación normativa SST sobre Weaviate (colección ZUNormativa).

- Embeddings locales en CPU (fastembed / ONNX): ningún texto sale a terceros.
- Búsqueda híbrida: BM25 (términos legales exactos) + vector (paráfrasis).
- Cada resultado es un artículo concreto con su norma, referencia y URL oficial.
- Lo derogado se excluye por defecto; si una norma derogada coincide fuerte con
  la pregunta, se devuelve como advertencia ("está derogada por ...") y nunca
  como fundamento.
"""
from __future__ import annotations

import json
import os
import uuid
from dataclasses import dataclass, field
from pathlib import Path

import weaviate
from weaviate.classes.config import Configure, DataType, Property, Tokenization
from weaviate.classes.init import Auth
from weaviate.classes.query import Filter, HybridFusion, MetadataQuery

COLLECTION = "ZUNormativa"
CORPUS = Path(__file__).resolve().parents[2] / "corpus" / "normativa.jsonl"
MODEL = os.getenv("NORMATIVE_EMBED_MODEL", "intfloat/multilingual-e5-large")
TITLE_BOOST = int(os.getenv("NORMATIVE_TITLE_BOOST", "2"))
ALPHA = float(os.getenv("NORMATIVE_HYBRID_ALPHA", "0.75"))  # mejor resultado medido (corpus/evaluate.py)

TEXT_PROPS = ["norm_id", "norm_name", "issuer", "status", "derogated_by", "topic", "article", "paragraph",
              "reference", "article_title", "section", "source_url", "content", "article_status", "chunk_id"]


class Embedder:
    """e5 models expect "query: " / "passage: " prefixes; fastembed does not add them."""

    def __init__(self, model: str = MODEL, cache_dir: str | None = None):
        from fastembed import TextEmbedding

        self.model = model
        # In the image the model is baked into FASTEMBED_CACHE and HF_HUB_OFFLINE=1:
        # load strictly from disk, never reach the network at runtime.
        self._m = TextEmbedding(model_name=model, cache_dir=cache_dir or os.getenv("FASTEMBED_CACHE", None), threads=os.cpu_count(),
                                local_files_only=os.getenv("HF_HUB_OFFLINE") == "1")
        self._e5 = "e5" in model.lower()

    def passages(self, texts: list[str]) -> list[list[float]]:
        if self._e5:
            texts = ["passage: " + t for t in texts]
        return [v.tolist() for v in self._m.embed(texts, batch_size=16)]

    def query(self, text: str) -> list[float]:
        if self._e5:
            text = "query: " + text
        return next(iter(self._m.embed([text]))).tolist()


def connect(url: str | None = None, api_key: str | None = None) -> weaviate.WeaviateClient:
    url = url or os.getenv("WEAVIATE_URL", "http://weaviate:8080")
    host_port = url.split("://", 1)[-1]
    host, _, port = host_port.partition(":")
    key = api_key if api_key is not None else os.getenv("WEAVIATE_API_KEY", "")
    return weaviate.connect_to_local(host=host, port=int(port or 8080), grpc_port=int(os.getenv("WEAVIATE_GRPC_PORT", "50051")),
                                     auth_credentials=Auth.api_key(key) if key else None)


def ensure_collection(client: weaviate.WeaviateClient) -> None:
    if client.collections.exists(COLLECTION):
        return
    props = [Property(name=p, data_type=DataType.TEXT,
                      tokenization=Tokenization.WORD if p in ("content", "article_title", "topic", "norm_name") else Tokenization.FIELD,
                      skip_vectorization=True)
             for p in TEXT_PROPS]
    props += [Property(name="year", data_type=DataType.INT), Property(name="editorial_notes", data_type=DataType.TEXT_ARRAY)]
    client.collections.create(COLLECTION, properties=props, vectorizer_config=Configure.Vectorizer.none())


def load_corpus(path: Path = CORPUS) -> list[dict]:
    with open(path, encoding="utf-8") as fh:
        return [json.loads(line) for line in fh if line.strip()]


def index_corpus(client: weaviate.WeaviateClient, embedder: Embedder, chunks: list[dict], recreate: bool = False) -> int:
    """Idempotent: object UUIDs derive from chunk_id, so re-running replaces in place."""
    if recreate and client.collections.exists(COLLECTION):
        client.collections.delete(COLLECTION)
    ensure_collection(client)
    col = client.collections.get(COLLECTION)
    # Embed what the article is about, not only its body: title and norm name carry meaning.
    texts = [f"{c['norm_name']} — {c['reference']} {c.get('article_title', '')}\n{c['content']}" for c in chunks]
    vectors = embedder.passages(texts)
    with col.batch.fixed_size(batch_size=64) as batch:
        for c, v in zip(chunks, vectors, strict=True):
            props = {k: c.get(k, "") for k in TEXT_PROPS}
            props["year"] = int(c.get("year", 0))
            props["editorial_notes"] = c.get("editorial_notes", [])
            batch.add_object(properties=props, vector=v, uuid=uuid.uuid5(uuid.NAMESPACE_URL, "zu-normativa/" + c["chunk_id"]))
    failed = col.batch.failed_objects
    if failed:
        raise RuntimeError(f"{len(failed)} fragmentos no se indexaron: {failed[0].message}")
    return len(chunks)


@dataclass
class Hit:
    norm_name: str
    reference: str
    article: str
    article_title: str
    status: str
    derogated_by: str
    source_url: str
    content: str
    score: float
    norm_id: str = ""

    @property
    def citation(self) -> str:
        return f"{self.norm_name}, {self.reference}"


@dataclass
class SearchResult:
    hits: list[Hit]
    warnings: list[str] = field(default_factory=list)


def _hit(o) -> Hit:
    p = o.properties
    return Hit(norm_name=p["norm_name"], reference=p["reference"], article=p["article"], article_title=p.get("article_title", ""),
               status=p.get("article_status", ""), derogated_by=p.get("derogated_by", ""), source_url=p["source_url"],
               content=p["content"], score=float(o.metadata.score or 0), norm_id=p["norm_id"])


def search(client: weaviate.WeaviateClient, embedder: Embedder, question: str, k: int = 5,
           alpha: float = ALPHA, include_derogated: bool = False) -> SearchResult:
    col = client.collections.get(COLLECTION)
    vec = embedder.query(question)
    common = dict(query=question, vector=vec, alpha=alpha, fusion_type=HybridFusion.RELATIVE_SCORE,
                  query_properties=["content", f"article_title^{TITLE_BOOST}", "norm_name", "topic"],
                  return_metadata=MetadataQuery(score=True))
    flt = None if include_derogated else Filter.by_property("article_status").equal("vigente")
    res = col.query.hybrid(limit=k * 3, filters=flt, **common)
    # One entry per article: several fragments of the same article are one source.
    seen, hits = set(), []
    for o in res.objects:
        h = _hit(o)
        key = (h.norm_id, h.article)
        if key not in seen:
            seen.add(key)
            hits.append(h)
        if len(hits) == k:
            break
    warnings = []
    if not include_derogated:
        # Warn only when derogated text competes with the valid answer *in the
        # same ranking*: scores of separate hybrid queries are each normalized
        # to their own best hit and cannot be compared.
        mixed = col.query.hybrid(limit=6, **common)
        for o in mixed.objects[:3]:
            p = o.properties
            if p.get("article_status") != "derogado":
                continue
            if p.get("status") == "derogada":
                msg = f"{p['norm_name']} también trata este tema, pero está derogada" + (f" ({p['derogated_by']})" if p.get("derogated_by") else "") + "; no se usa como fundamento."
            else:
                msg = f"{p['norm_name']}, {p['reference']} también trata este tema, pero ese artículo está derogado; no se usa como fundamento."
            if msg not in warnings:
                warnings.append(msg)
    return SearchResult(hits=hits, warnings=warnings)
