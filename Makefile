.PHONY: help build build-admin build-voice dev dev-admin dev-voice clean download-model swagger docker-build docker-buildx docker-push postgres redis ollama infra infra-stop infra-clean

CONFIG ?= config.yaml
CONTAINER_ENGINE ?= $(shell command -v podman 2>/dev/null || echo docker)

help:
	@echo "Magec - Multi-Agent AI Platform"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development:"
	@echo "  build                Build the server binary"
	@echo "  build-admin          Build the admin UI (Vue)"
	@echo "  build-voice          Build the voice UI (Vue)"
	@echo "  dev                  Build all and start server (CONFIG=config.yaml)"
	@echo "  dev-admin            Start admin UI dev server (Vite, port 5173)"
	@echo "  dev-voice            Start voice UI dev server (Vite, port 5174)"
	@echo "  swagger              Regenerate Swagger docs from annotations"
	@echo "  clean                Remove generated files"
	@echo ""
	@echo "Models:"
	@echo "  download-model       Download wake word model (interactive)"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build         Build Docker image for current arch (IMAGE_TAG=latest)"
	@echo "  docker-buildx        Build multi-arch image: linux/amd64,linux/arm64 (IMAGE_TAG=latest)"
	@echo "  docker-push          Build multi-arch and push to GHCR (IMAGE_TAG=latest)"
	@echo ""
	@echo "Infrastructure (Docker):"
	@echo "  postgres             Start PostgreSQL container"
	@echo "  redis                Start Redis container"
	@echo "  ollama               Start Ollama with qwen3:8b + nomic-embed-text"
	@echo "  infra                Start postgres + redis"
	@echo "  infra-stop           Stop and remove postgres + redis"
	@echo "  infra-clean          Stop all containers and remove volumes"

build-admin:
	@cd frontend/admin-ui && npm install --silent && npx vite build
	@rm -rf server/frontend/admin-ui && cp -r frontend/admin-ui/dist server/frontend/admin-ui
	@echo "Admin UI built and copied to server/frontend/admin-ui/"

build-voice:
	@cd frontend/voice-ui && npm install --silent && npx vite build
	@rm -rf server/frontend/voice-ui && cp -r frontend/voice-ui/dist server/frontend/voice-ui
	@echo "Voice UI built and copied to server/frontend/voice-ui/"

build: build-admin build-voice embed-models
	@mkdir -p bin
	@cd server && go build -o ../bin/magec-server .

embed-models: download-model
	@rm -rf server/models/wakeword server/models/auxiliary
	@cp -r models/wakeword server/models/wakeword
	@cp -r models/auxiliary server/models/auxiliary
	@echo "Models copied to server/models/"

swagger:
	@cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./api/admin --generalInfo doc.go --output ./api/admin/docs --parseDependency --parseInternal --templateDelims "[[,]]"
	@echo "Admin API swagger generated in server/api/admin/docs/"
	@cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./api/user --generalInfo doc.go --output ./api/user/docs --parseDependency --parseInternal --instanceName userapi --templateDelims "[[,]]"
	@echo "User API swagger generated in server/api/user/docs/"

dev: build
	@./bin/magec-server -config=$(CONFIG)

dev-admin:
	@cd frontend/admin-ui && npx vite

dev-voice:
	@cd frontend/voice-ui && npx vite

download-model:
	@go run scripts/download-model.go

clean:
	@rm -rf bin
	@rm -rf frontend/admin-ui/dist
	@rm -rf frontend/voice-ui/dist
	@rm -rf server/frontend/admin-ui
	@rm -rf server/frontend/voice-ui
	@rm -rf server/models/wakeword
	@rm -rf server/models/auxiliary
	@rm -rf models/auxiliary
	@find . -name ".DS_Store" -delete
	@echo "Cleaned"

# Docker image

IMAGE_NAME ?= ghcr.io/achetronic/magec
IMAGE_TAG ?= latest

DOCKER_PLATFORMS ?= linux/amd64,linux/arm64

docker-build:
	@$(CONTAINER_ENGINE) build -f docker/build/Dockerfile -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Image built: $(IMAGE_NAME):$(IMAGE_TAG)"

docker-buildx:
	@$(CONTAINER_ENGINE) buildx build -f docker/build/Dockerfile --platform $(DOCKER_PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@echo "Multi-arch image built: $(IMAGE_NAME):$(IMAGE_TAG) [$(DOCKER_PLATFORMS)]"

docker-push:
	@$(CONTAINER_ENGINE) buildx build -f docker/build/Dockerfile --platform $(DOCKER_PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) --push .
	@echo "Image pushed: $(IMAGE_NAME):$(IMAGE_TAG) [$(DOCKER_PLATFORMS)]"

# Infrastructure (Docker)

postgres:
	@$(CONTAINER_ENGINE) run -d --name magec-postgres \
		-p 5432:5432 \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=magec \
		docker.io/pgvector/pgvector:pg17
	@echo "PostgreSQL (pgvector) started on localhost:5432"

redis:
	@$(CONTAINER_ENGINE) run -d --name magec-redis \
		-p 6379:6379 \
		docker.io/library/redis:alpine
	@echo "Redis started on localhost:6379"

ollama:
	@$(CONTAINER_ENGINE) run -d --name magec-ollama \
		-p 11434:11434 \
		-v ollama:/root/.ollama \
		docker.io/ollama/ollama
	@echo "Waiting for Ollama to start..."
	@sleep 3
	@$(CONTAINER_ENGINE) exec magec-ollama ollama pull qwen3:8b
	@$(CONTAINER_ENGINE) exec magec-ollama ollama pull nomic-embed-text
	@echo "Ollama started on localhost:11434 with qwen3:8b and nomic-embed-text"

infra: postgres redis
	@echo "Infrastructure ready"

infra-stop:
	@$(CONTAINER_ENGINE) stop magec-postgres magec-redis 2>/dev/null || true
	@$(CONTAINER_ENGINE) rm magec-postgres magec-redis 2>/dev/null || true
	@echo "Infrastructure stopped"

infra-clean: infra-stop
	@$(CONTAINER_ENGINE) stop magec-ollama 2>/dev/null || true
	@$(CONTAINER_ENGINE) rm magec-ollama 2>/dev/null || true
	@$(CONTAINER_ENGINE) volume rm ollama 2>/dev/null || true
	@echo "All containers and volumes removed"
