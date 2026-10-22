"""Evalúa la recuperación normativa contra eval_questions.json.

    WEAVIATE_URL=... WEAVIATE_API_KEY=... python3 evaluate.py [--index] [--alpha 0.5]
Acierto = el artículo esperado aparece entre los k primeros resultados vigentes.
"""
import argparse
import json
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from app.rag import normative as nv  # noqa: E402

ap = argparse.ArgumentParser()
ap.add_argument("--index", action="store_true")
ap.add_argument("--alpha", type=float, nargs="*", default=[nv.ALPHA])
ap.add_argument("--k", type=int, default=5)
args = ap.parse_args()

client = nv.connect()
emb = nv.Embedder()
try:
    if args.index:
        t = time.time()
        n = nv.index_corpus(client, emb, nv.load_corpus(), recreate=True)
        print(f"indexados {n} fragmentos con {emb.model} en {time.time() - t:.0f}s")
    qs = json.loads((Path(__file__).parent / "eval_questions.json").read_text())
    for alpha in args.alpha:
        top1 = top3 = topk = 0
        for item in qs:
            res = nv.search(client, emb, item["q"], k=args.k, alpha=alpha)
            got = [(h.norm_id, h.article) for h in res.hits]
            exp = {tuple(e) for e in item["expect"]}
            rank = next((i + 1 for i, g in enumerate(got) if g in exp), None)
            top1 += rank == 1
            top3 += bool(rank and rank <= 3)
            topk += bool(rank)
            if len(args.alpha) == 1:
                mark = f"#{rank}" if rank else "FALLA"
                print(f"[{mark:>5}] {item['q']}\n        → {res.hits[0].citation if res.hits else '-'}" + (f"   ⚠ {res.warnings[0]}" if res.warnings else ""))
        n = len(qs)
        print(f"alpha={alpha}: top-1 {top1}/{n} · top-3 {top3}/{n} · top-{args.k} {topk}/{n}")
finally:
    client.close()
