.PHONY: help build build-admin dev dev-admin clean download-model swagger postgres redis ollama infra infra-stop infra-clean

CONFIG ?= config.yaml

help:
	@echo "Magec - Voice Assistant"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Development:"
	@echo "  build                Build the server binary"
	@echo "  build-admin          Build the admin UI (Vue)"
	@echo "  dev                  Build all and start server (CONFIG=config.yaml)"
	@echo "  dev-admin            Start admin UI dev server (Vite, port 5173)"
	@echo "  swagger              Regenerate Swagger docs from annotations"
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

build-admin:
	@cd admin-ui && npm install --silent && npx vite build
	@echo "Admin UI built to admin-ui/dist/"

build: build-admin
	@mkdir -p bin
	@cd server && go build -o ../bin/magec-server .

swagger:
	@cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./admin --generalInfo doc.go --output ./admin/docs --parseDependency --parseInternal
	@echo "Admin API swagger generated in server/admin/docs/"
	@cd server && go run github.com/swaggo/swag/cmd/swag init --dir ./userapi --generalInfo doc.go --output ./userapi/docs --parseDependency --parseInternal --instanceName userapi
	@echo "User API swagger generated in server/userapi/docs/"

dev: build
	@./bin/magec-server -config=$(CONFIG)

dev-admin:
	@cd admin-ui && npx vite

download-model:
	@go run scripts/download-model.go

clean:
	@rm -rf bin
	@rm -rf admin-ui/dist
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
