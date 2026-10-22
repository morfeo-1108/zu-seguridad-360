FROM python:3.12-slim AS builder
WORKDIR /app
ARG EXTENDED=0
RUN apt-get update && apt-get install -y --no-install-recommends gcc && rm -rf /var/lib/apt/lists/*
COPY requirements.txt requirements-extended.txt ./
RUN pip install --no-cache-dir --user -r $( [ "$EXTENDED" = 1 ] && echo requirements-extended.txt || echo requirements.txt )
# Embedding model baked into the image: no download at runtime.
ENV FASTEMBED_CACHE=/models
RUN PATH=/root/.local/bin:$PATH python -c "from fastembed import TextEmbedding; TextEmbedding('sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2', cache_dir='/models')"

FROM python:3.12-slim
WORKDIR /app
RUN useradd --uid 10001 --no-create-home zu
COPY --from=builder /root/.local /usr/local
COPY --from=builder /models /models
COPY . .
ENV PYTHONPATH=/app FASTEMBED_CACHE=/models NORMATIVE_EMBED_MODEL=sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2 HF_HUB_OFFLINE=1
USER zu
EXPOSE 4002
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "4002", "--workers", "1"]
