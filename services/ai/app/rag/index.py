"""Indexa corpus/normativa.jsonl en Weaviate (idempotente).

    python -m app.rag.index            reemplaza la colección completa
"""
import time

import app.core.config  # noqa: F401 — loads secrets from Vault into the environment
from app.rag import normative as nv


def main() -> None:
    client = nv.connect()
    try:
        emb = nv.Embedder()
        t = time.time()
        n = nv.index_corpus(client, emb, nv.load_corpus(), recreate=True)
        print(f"indexados {n} fragmentos con {emb.model} en {time.time() - t:.0f}s")
    finally:
        client.close()


if __name__ == "__main__":
    main()
