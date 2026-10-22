"""Construye el corpus normativo (normativa.jsonl) a partir de los textos oficiales.

    python3 -I build_corpus.py --fetch <dir_html>   # descarga las fuentes de sources.json
    python3 -I build_corpus.py <dir_html>           # parsea y escribe normativa.jsonl

Chunking legal Artículo → Parágrafo: la unidad es el artículo; si es largo se
parte por parágrafos y, si aún excede, por bloques de líneas (con "parte N").
Se eliminan las notas editoriales del compilador (vigencia, jurisprudencia,
legislación anterior): solo queda el texto de la norma.
"""
import hashlib
import json
import re
import sys
import time
import urllib.request
from pathlib import Path

from bs4 import BeautifulSoup

HERE = Path(__file__).resolve().parent
MAX_CHARS = 2800

# Heading form only ("ARTÍCULO 3o.", "Artículo 1°.", "ARTÍCULO 2.2.4.6.1."): a
# reference in running text ("artículo 25 del Decreto ...") has no ordinal mark.
# Some headings omit the mark but carry an UPPERCASE title ("ARTÍCULO 2.2.4.6.28 CONTRATACIÓN.").
ART_RE = re.compile(r"^\s*ART[ÍI]CULO\s+((?:\d+\.)*\d+)(?:\s*(?:[oº°]\.?|\.)(?:\s+|$)|\s+(?=(?-i:[A-ZÁÉÍÓÚÑ]{3,})))(.*)$", re.I)
PAR_RE = re.compile(r"^\s*PAR[ÁA]GRAFO\s*((?:\d+[oº°]?)|TRANSITORIO|ÚNICO|UNICO)?\.?", re.I)
SECTION_RE = re.compile(r"^\s*(LIBRO|PARTE|T[ÍI]TULO|CAP[ÍI]TULO|SECCI[ÓO]N)\s+[\dIVXLC\.]+\b.*$", re.I)


def fname(norm):
    return norm["id"] + ".html"


def fetch(out_dir: Path, norms):
    out_dir.mkdir(parents=True, exist_ok=True)
    for n in norms:
        req = urllib.request.Request(n["url"], headers={"User-Agent": "Mozilla/5.0 (ZU360 corpus builder)"})
        data = urllib.request.urlopen(req, timeout=120).read()
        (out_dir / fname(n)).write_bytes(data)
        print(f"{n['id']}: {len(data)} bytes")
        time.sleep(1)


def html_to_lines(raw: bytes) -> list[str]:
    soup = BeautifulSoup(raw.decode("latin-1", errors="replace"), "html.parser")
    for tag in soup(["script", "style", "noscript", "header", "footer", "nav"]):
        tag.decompose()
    # Avance Jurídico: cajas de notas (vigencia, jurisprudencia, legislación anterior).
    for tag in soup.find_all(class_=re.compile(r"^(caja_vja|ir-arriba)")):
        tag.decompose()
    body = soup.find(class_="panel-documento") or soup.body or soup
    # Source line breaks inside a paragraph are just wrapping: turn them into
    # spaces, then mark real block boundaries (p, div, li, headings, rows, br)
    # with explicit newlines, so "Artículo" and "1°." in separate inline tags
    # stay on one logical line.
    from bs4 import NavigableString
    for node in list(body.find_all(string=True)):
        if "\n" in node or "\r" in node:
            node.replace_with(NavigableString(re.sub(r"[\r\n]+", " ", str(node))))
    for row in body.find_all("tr"):
        if row.find(["table", "p", "div"]) is not None:
            continue  # layout table (SISJUR wraps the whole document in one), not a data table
        cells = [c.get_text(" ", strip=True) for c in row.find_all(["td", "th"])]
        row.replace_with(NavigableString("\n" + " | ".join(c for c in cells if c) + "\n"))
    for br in body.find_all("br"):
        br.replace_with(NavigableString("\n"))
    for tag in body.find_all(["p", "div", "li", "h1", "h2", "h3", "h4", "h5", "h6", "table", "blockquote"]):
        tag.insert_before(NavigableString("\n"))
        tag.insert_after(NavigableString("\n"))
    text = body.get_text("")
    lines = []
    for ln in text.split("\n"):
        ln = re.sub(r"[ \t\xa0]+", " ", ln).strip()
        if ln:
            lines.append(ln)
    return lines


def split_long(text: str):
    if len(text) <= MAX_CHARS:
        return [text]
    parts, cur = [], []
    for ln in text.split("\n"):
        if cur and sum(len(x) + 1 for x in cur) + len(ln) > MAX_CHARS:
            parts.append("\n".join(cur))
            cur = []
        cur.append(ln)
    if cur:
        parts.append("\n".join(cur))
    return parts


NOTE_RE = re.compile(r"<([^<>]{0,60}(?:derogad|modificad|subrogad|adicionad|inexequible|exequible)[^<>]{0,300})>", re.I)
NORM_DEROGATED_RE = re.compile(r"^(?:Resoluci[óo]n|Ley|Decreto)\s+derogad[ao]\s+por\s+(.+)$", re.I)
ART_DEROGATED_RE = re.compile(r"^(?:Art[íi]culo|Par[áa]grafo|Numeral|Literal)\s+derogad[ao]\s+por\s+(.+)$", re.I)


def chunk_article(norm, article, title, section, body_lines):
    # Split by parágrafos only when the article is too long as a whole.
    full = "\n".join(body_lines)
    groups = [("", body_lines)]
    if len(full) > MAX_CHARS:
        groups, cur_label, cur = [], "", []
        for ln in body_lines:
            m = PAR_RE.match(ln)
            if m and cur:
                groups.append((cur_label, cur))
                cur_label, cur = ("Parágrafo " + (m.group(1) or "")).strip(), []
            cur.append(ln)
        groups.append((cur_label, cur))
    # Editorial notes inserted by the compiler ("<Artículo modificado por ...>")
    # leave the text but become metadata: they drive the validity warnings.
    notes = sorted({m.group(1).strip() for m in NOTE_RE.finditer(full)})
    art_status = "vigente"
    for n in notes:
        if ART_DEROGATED_RE.match(n):
            art_status = "derogado"
        elif NORM_DEROGATED_RE.match(n):
            art_status = "derogado"
    out = []
    for label, lines in groups:
        pieces = split_long("\n".join(lines))
        for i, piece in enumerate(pieces):
            ref = f"Art. {article}" + (f", {label}" if label else "") + (f" (parte {i + 1}/{len(pieces)})" if len(pieces) > 1 else "")
            out.append({
                "norm_id": norm["id"], "norm_name": norm["name"], "issuer": norm["issuer"], "year": norm["year"],
                "status": norm["status"], "derogated_by": norm.get("derogated_by", ""), "topic": norm["topic"],
                "article": article, "paragraph": label, "reference": ref, "article_title": title,
                "section": section, "source_url": norm["url"], "content": NOTE_RE.sub("", piece).strip(),
                "article_status": art_status, "editorial_notes": notes,
            })
    return out


def parse(norm, raw: bytes):
    lines = html_to_lines(raw)
    chunks, article, title, section, body = [], None, "", "", []
    prefix = norm.get("article_prefix")
    started = False
    seen = set()
    last_int = None

    def flush():
        if article and body and (not prefix or article.startswith(prefix)):
            chunks.extend(chunk_article(norm, article, title, section, body))

    for ln in lines:
        m = ART_RE.match(ln)
        if m:
            flush()
            started = True
            if m.group(1) in seen:
                if m.group(1) == "1":
                    # Numbering restarts: the compiler appended another text (e.g. a
                    # Constitutional Court ruling transcribing the bill). Stop here.
                    article, body = None, []
                    break
                # A repeated number inside the text (quoted by a later article): keep the first.
                article, body = None, []
                continue
            num = m.group(1)
            if "." not in num and last_int is not None and not (last_int < int(num) <= last_int + 3):
                # Out of sequence: an article of another norm quoted as the new
                # text it gets ("el artículo 66 quedará así: Artículo 66 ...").
                # It belongs to the current article's body.
                body.append(ln)
                continue
            if "." not in num:
                last_int = int(num)
            seen.add(num)
            article, rest = m.group(1), m.group(2).strip()
            # "ARTÍCULO 3o. TÍTULO EN MAYÚSCULAS. Texto..." → título = primer segmento en mayúsculas
            tm = re.match(r"^([A-ZÁÉÍÓÚÑÜ0-9 ,;()\-–]+)\.\s*", rest)
            title = tm.group(1).strip() if tm and len(tm.group(1)) > 3 else ""
            body = [ln]
            continue
        if SECTION_RE.match(ln) and len(ln) < 200:
            if started and not article:
                section = ln
            section = ln
            continue
        if started and article:
            if ln.lower().startswith(("publíquese", "comuníquese y cúmplase", "dada en bogotá", "dado en bogotá")):
                flush()
                article, body = None, []
                continue
            body.append(ln)
    flush()
    return chunks


def main():
    sources = json.loads((HERE / "sources.json").read_text())["norms"]
    args = sys.argv[1:]
    if args and args[0] == "--fetch":
        fetch(Path(args[1]), sources)
        return
    raw_dir = Path(args[0])
    out, report = [], []
    for n in sources:
        raw = (raw_dir / fname(n)).read_bytes()
        chunks = parse(n, raw)
        # The official text wins over the manifest: a norm the compiler marks as
        # derogated cannot be shipped as "vigente" without a human fixing sources.json.
        derog = [c for c in chunks if any(NORM_DEROGATED_RE.match(x) for x in c["editorial_notes"])]
        if derog and n["status"] != "derogada":
            sys.exit(f"ERROR: {n['name']} figura como '{n['status']}' pero el texto oficial dice: "
                     f"{next(x for x in derog[0]['editorial_notes'] if NORM_DEROGATED_RE.match(x))}. Revise sources.json.")
        if n["status"] == "derogada":
            for c in chunks:
                c["article_status"] = "derogado"
        arts = sorted({c["article"] for c in chunks}, key=lambda a: [int(x) for x in a.split(".")])
        nder = len({c['article'] for c in chunks if c['article_status'] == 'derogado'})
        report.append(f"{n['name']:<28} {n['status']:<15} {len(arts):>4} artículos ({nder:>3} derogados)  {len(chunks):>4} fragmentos  sha256 fuente {hashlib.sha256(raw).hexdigest()[:12]}")
        out.extend(chunks)
    with open(HERE / "normativa.jsonl", "w", encoding="utf-8") as fh:
        for c in out:
            c["chunk_id"] = hashlib.sha256(f"{c['norm_id']}|{c['reference']}|{c['content']}".encode()).hexdigest()[:24]
            fh.write(json.dumps(c, ensure_ascii=False) + "\n")
    print("\n".join(report))
    print(f"TOTAL: {len(out)} fragmentos de {len(sources)} normas → normativa.jsonl")


if __name__ == "__main__":
    main()
