.PHONY: help build dev clean download-model postgres redis ollama infra infra-stop infra-clean

CONFIG ?= config.yaml

help:
	@echo "Magec - Voice Assistant"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development:"
	@echo "  build                Build the server binary"
	@echo "  dev                  Start development server (CONFIG=config.yaml)"
	@echo "  clean                Remove generated files"
	@echo ""
	@echo "Models:"
	@echo "  download-model       Download wake word model (interactive)"
	@echo ""
	@echo "Infrastructure (Docker):"
	@echo "  postgres             Start PostgreSQL container"
	@echo "  redis                Start Redis container"
	@echo "  ollama               Start Ollama with qwen3:8b + nomic-embed-text"
	@echo "  infra                Start postgres + redis"
	@echo "  infra-stop           Stop and remove postgres + redis"
	@echo "  infra-clean          Stop all containers and remove volumes"

build:
	@mkdir -p bin
	@cd server && go build -o ../bin/magec-server .

dev: build
	@./bin/magec-server -config=$(CONFIG)

download-model:
	@go run scripts/download-model.go

clean:
	@rm -rf bin
	@rm -rf gui/pretrained
	@find . -name ".DS_Store" -delete
	@echo "Cleaned"

# Infrastructure (Docker)

postgres:
	@docker run -d --name magec-postgres \
		-p 5432:5432 \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=magec \
		pgvector/pgvector:pg17
	@echo "PostgreSQL (pgvector) started on localhost:5432"

redis:
	@docker run -d --name magec-redis \
		-p 6379:6379 \
		redis:alpine
	@echo "Redis started on localhost:6379"

ollama:
	@docker run -d --name magec-ollama \
		-p 11434:11434 \
		-v ollama:/root/.ollama \
		ollama/ollama
	@echo "Waiting for Ollama to start..."
	@sleep 3
	@docker exec magec-ollama ollama pull qwen3:8b
	@docker exec magec-ollama ollama pull nomic-embed-text
	@echo "Ollama started on localhost:11434 with qwen3:8b and nomic-embed-text"

infra: postgres redis
	@echo "Infrastructure ready"

infra-stop:
	@docker stop magec-postgres magec-redis 2>/dev/null || true
	@docker rm magec-postgres magec-redis 2>/dev/null || true
	@echo "Infrastructure stopped"

infra-clean: infra-stop
	@docker stop magec-ollama 2>/dev/null || true
	@docker rm magec-ollama 2>/dev/null || true
	@docker volume rm ollama 2>/dev/null || true
	@echo "All containers and volumes removed"
