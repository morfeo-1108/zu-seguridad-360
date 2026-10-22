"""Conecta un servicio de docker-compose.dev.yml a Vault (env común + Docker secret).
Uso: python3 scripts/compose-vault-wire.py <svc> [KEY=VAL ...]  — idempotente."""
import re, sys
svc, extra = sys.argv[1], sys.argv[2:]
p = "docker-compose.dev.yml"
s = open(p).read()
m = re.search(rf"\n  {svc}:\n(?:    .*\n|\n)*?(?=\n  [a-z]|\nsecrets:|\Z)", s)
block = m.group(0)
if "vault_token_" + svc not in block:
    lines = block.rstrip("\n").split("\n")
    env = ["    environment:", "      <<: *vault-env", f"      VAULT_SECRET_PATH: secret/data/zu360/{svc}"]
    env += [f"      {k}: \"{v}\"" for k, v in (e.split("=", 1) for e in extra)]
    # drop an existing environment: block (its keys are re-added through extra)
    out, skip = [], False
    for ln in lines:
        if ln.startswith("    environment:"):
            skip = True; continue
        if skip and ln.startswith("      "):
            continue
        skip = False
        out.append(ln)
    i = next(k for k, ln in enumerate(out) if ln.startswith("    env_file")) 
    j = i + 1
    while j < len(out) and out[j].startswith("      "):
        j += 1
    out[j:j] = env + ["    secrets:", f"      - source: vault_token_{svc}", "        target: vault_token"]
    if not any(ln.startswith("    restart:") for ln in out):
        out.append("    restart: unless-stopped")
    out = [re.sub(r"depends_on: \[(.*)\]", lambda mm: "depends_on: [" + ", ".join(dict.fromkeys([x.strip() for x in mm.group(1).split(",")] + ["vault"])) + "]", ln) for ln in out]
    s = s.replace(block, "\n".join(out) + "\n")
if f"  vault_token_{svc}:" not in s:
    s = s.rstrip("\n") + f"\n  vault_token_{svc}:\n    file: ./secrets/vault/{svc}.token\n"
open(p, "w").write(s)
print(f"compose: {svc} conectado a Vault")
