"""Regenera frontend/web/lib/normative-corpus.ts desde services/ai/corpus/sources.json."""
import json
from pathlib import Path

root = Path(__file__).resolve().parents[1]
src = json.loads((root / "services/ai/corpus/sources.json").read_text())
lines = ["// GENERADO desde services/ai/corpus/sources.json (scripts/gen-normative-list.py). No editar a mano.",
         "export const NORMS_INDEXED: { name: string; topic: string; status: string; derogatedBy?: string }[] = ["]
for n in src["norms"]:
    d = {"name": n["name"], "topic": n["topic"].replace(" (DEROGADA)", ""), "status": n["status"]}
    if n.get("derogated_by"):
        d["derogatedBy"] = n["derogated_by"]
    lines.append("  " + json.dumps(d, ensure_ascii=False) + ",")
lines.append("];")
(root / "frontend/web/lib/normative-corpus.ts").write_text("\n".join(lines) + "\n")
print("ok:", len(src["norms"]), "normas")
