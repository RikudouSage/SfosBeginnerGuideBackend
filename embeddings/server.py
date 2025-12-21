import os
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import numpy as np
import torch
from transformers import AutoTokenizer, AutoModelForSequenceClassification

torch.set_num_threads(int(os.getenv("TORCH_NUM_THREADS", "invalid-int-cannot-continue")))

EMBEDDING_MODEL_NAME = os.getenv("EMBEDDING_MODEL", "nonexistent-model-cannot-fetch")
EMBEDDING_MODEL_BATCH = int(os.getenv("EMBEDDING_BATCH", "invalid-int-cannot-continue"))
RERANK_MODEL_NAME = os.getenv("RERANK_MODEL", "nonexistent-model-cannot-fetch")
RERANK_MODEL_BATCH = int(os.getenv("RERANK_BATCH", "invalid-int-cannot-continue"))

embedding_model = SentenceTransformer(EMBEDDING_MODEL_NAME, device="cpu")
print(f"Loading embedding model: {EMBEDDING_MODEL_NAME} with batch {EMBEDDING_MODEL_BATCH}", flush=True)

print(f"Loading reranking model: {RERANK_MODEL_NAME} with batch {RERANK_MODEL_BATCH}", flush=True)
rerank_tokenizer = AutoTokenizer.from_pretrained(RERANK_MODEL_NAME)
rerank_model = AutoModelForSequenceClassification.from_pretrained(RERANK_MODEL_NAME)
rerank_model.eval()

app = FastAPI()

@app.get("/health")
def health():
    return {"status": "ok"}

class EmbedRequest(BaseModel):
    texts: list[str]
    mode: str = "passage"  # "passage" for docs; "query" for searches

class EmbedResponse(BaseModel):
    vectors: list[list[float]]
    dim: int

def _prefix(texts, mode):
    if mode == "query":
        return [f"query: {t}" for t in texts]
    return [f"passage: {t}" for t in texts]

@app.post("/embed", response_model=EmbedResponse)
def embed(req: EmbedRequest):
    inputs = _prefix(req.texts, req.mode)
    vecs = embedding_model.encode(inputs, batch_size=EMBEDDING_MODEL_BATCH, normalize_embeddings=False).astype(np.float32)
    return {"vectors": vecs.tolist(), "dim": vecs.shape[1]}

class RerankRequest(BaseModel):
    query: str
    candidates: list[str]
    top_n: int | None = None
    normalize: bool = True

class RerankResponse(BaseModel):
    scores: list[float]
    order: list[int]
    top_texts: list[str] | None = None
    top_scores: list[float] | None = None
    model: str

@app.post("/rerank", response_model=RerankResponse)
def rerank(req: RerankRequest):
    if not req.candidates:
        return {"scores": [], "order": [], "top_texts": [], "top_scores": [], "model": RERANK_MODEL_NAME}

    query_candidate_pairs = [(req.query, candidate_text) for candidate_text in req.candidates]
    all_candidate_scores: list[float] = []

    for start_index in range(0, len(query_candidate_pairs), RERANK_MODEL_BATCH):
        batch_pairs = query_candidate_pairs[start_index:start_index + RERANK_MODEL_BATCH]

        batch_queries = [query_text for query_text, _ in batch_pairs]
        batch_candidates = [candidate_text for _, candidate_text in batch_pairs]

        with torch.no_grad():
            inputs = rerank_tokenizer(
                list(zip(batch_queries, batch_candidates)),
                padding=True,
                truncation=True,
                return_tensors="pt",
                max_length=512
            )
            logits = rerank_model(**inputs).logits.view(-1).float()
            if req.normalize:
                candidate_scores = torch.sigmoid(logits)
            else:
                candidate_scores = logits

            all_candidate_scores.extend(candidate_scores.tolist())

    sorted_candidate_indices = list(np.argsort(-np.array(all_candidate_scores)))
    response_data = {
        "scores": all_candidate_scores,
        "order": sorted_candidate_indices,
        "model": RERANK_MODEL_NAME
    }
    if req.top_n:
        top_indices = sorted_candidate_indices[:req.top_n]
        response_data["top_texts"] = [req.candidates[i] for i in top_indices]
        response_data["top_scores"] = [all_candidate_scores[i] for i in top_indices]

    return response_data
