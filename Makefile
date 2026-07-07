.PHONY: help dev dev-down master-run master-test agent-test frontend-dev db-up db-down lint

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Local Development ───────────────────────────────────────────────

dev: ## Start all services (docker-compose)
	docker compose up --build -d

dev-down: ## Stop all services
	docker compose down

dev-logs: ## Tail logs for all services
	docker compose logs -f

db-up: ## Start only PostgreSQL
	docker compose up -d postgres

db-down: ## Stop PostgreSQL
	docker compose down postgres

# ─── Master Server ───────────────────────────────────────────────────

master-run: ## Run the master server locally
	cd master && go run cmd/server/main.go

master-test: ## Run master tests
	cd master && go test ./... -v -cover

master-lint: ## Lint master code
	cd master && golangci-lint run ./...

master-build: ## Build master binary
	cd master && go build -o bin/server cmd/server/main.go

# ─── Worker Agent ────────────────────────────────────────────────────

agent-run: ## Run the agent locally
	cd agent && go run cmd/agent/main.go

agent-test: ## Run agent tests
	cd agent && go test ./... -v -cover

agent-lint: ## Lint agent code
	cd agent && golangci-lint run ./...

agent-build: ## Build agent binary
	cd agent && go build -o bin/agent cmd/agent/main.go

# ─── Frontend ────────────────────────────────────────────────────────

frontend-dev: ## Run frontend dev server
	cd frontend && npm run dev

frontend-build: ## Build frontend for production
	cd frontend && npm run build

frontend-test: ## Run frontend tests
	cd frontend && npm test

# ─── Docker ──────────────────────────────────────────────────────────

docker-master: ## Build master Docker image
	docker build -t uptimehub-master:latest ./master

docker-agent: ## Build agent Docker image
	docker build -t uptimehub-agent:latest ./agent

docker-frontend: ## Build frontend Docker image
	docker build -t uptimehub-frontend:latest ./frontend
