.PHONY: venv embeddings install-deps warm-models

export EMBEDDING_MODEL ?= intfloat/multilingual-e5-large
export RERANK_MODEL ?= BAAI/bge-reranker-v2-m3
export TORCH_NUM_THREADS ?= 4
export EMBEDDING_BATCH ?= 32
export RERANK_BATCH ?= 32

venv:
	test -d embeddings/.venv || python -m venv embeddings/.venv

install-deps: venv
	embeddings/.venv/bin/python -m pip install --upgrade pip
	embeddings/.venv/bin/python -m pip install fastapi pydantic sentence-transformers numpy torch transformers uvicorn

warm-models: install-deps
	@port=$$(python -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()'); \
	embeddings/.venv/bin/python -m uvicorn embeddings.server:app --host 127.0.0.1 --port $$port --log-level warning & \
	pid=$$!; \
	while :; do \
		if embeddings/.venv/bin/python -c 'import urllib.request; urllib.request.urlopen("http://127.0.0.1:'"$$port"'/health", timeout=5).read()' >/dev/null 2>&1; then \
			break; \
		fi; \
		echo "Waiting for model warm-up..."; \
		sleep 10; \
	done; \
	kill $$pid; \
	wait $$pid 2>/dev/null || true

embeddings: warm-models
	go run internal/cmd/embeddings.go
