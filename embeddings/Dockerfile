FROM python:3.13-slim

RUN pip install --no-cache-dir fastapi uvicorn[standard] sentence-transformers transformers torch --extra-index-url https://download.pytorch.org/whl/cpu

ENV EMBEDDING_MODEL=intfloat/multilingual-e5-large
ENV RERANK_MODEL=BAAI/bge-reranker-v2-m3

# Warm cache for both models
RUN echo "Precaching models" && python - <<PY
from sentence_transformers import SentenceTransformer
from transformers import AutoTokenizer, AutoModelForSequenceClassification
print("Precaching ${EMBEDDING_MODEL}")
SentenceTransformer("${EMBEDDING_MODEL}")
print("Precaching ${RERANK_MODEL}")
AutoTokenizer.from_pretrained("${RERANK_MODEL}")
AutoModelForSequenceClassification.from_pretrained("${RERANK_MODEL}")
print("Models cached.")
PY

ENV TORCH_NUM_THREADS=4
ENV HF_HUB_DISABLE_TELEMETRY=1
ENV RERANK_BATCH=32
ENV EMBEDDING_BATCH=32

WORKDIR /app
COPY server.py /app/server.py

RUN apt-get update && apt-get -y install curl && rm -rf /var/cache/apt/archives /var/lib/apt/lists/*
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=5 \
  CMD curl -fsS http://127.0.0.1:8000/health || exit 1

EXPOSE 8000
CMD ["uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000"]
