# Makefile for Goose, Chicken, Guinea Fowl Egg Tracker Project

PROJECT_NAME=egg-tracker

# Directories
BACKEND_DIR=backend
FRONTEND_DIR=frontend
DOCKER_DIR=docker
BACKUP_DIR=backups

# Docker Compose file
COMPOSE_FILE=docker-compose.yml

# Targets

## Build

build-backend:
	cd $(BACKEND_DIR) && go build -o app

build-frontend:
	cd $(FRONTEND_DIR) && npm install && npm run build

## Run Local Dev Servers (without Docker)

run-backend:
	cd $(BACKEND_DIR) && go run main.go

run-frontend:
	cd $(FRONTEND_DIR) && npm run dev

## Docker Compose

docker-up:
	docker-compose -f $(COMPOSE_FILE) up --build

docker-down:
	docker-compose -f $(COMPOSE_FILE) down

## Database Backup

backup:
	@echo "Creating backup..."
	mkdir -p $(BACKUP_DIR)
	cp $(BACKEND_DIR)/data/*.db $(BACKUP_DIR)/ || true
	cp $(BACKEND_DIR)/data/*.duckdb $(BACKUP_DIR)/ || true
	@echo "Backup completed."

## Clean

clean:
	@echo "Cleaning builds..."
	rm -rf $(BACKEND_DIR)/app
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -f $(FRONTEND_DIR)/.env*
	@echo "Clean completed."

## Testing

test-backend:
	cd $(BACKEND_DIR) && go test ./...

# test-frontend will not fail if no test script is present, but will print a warning

test-frontend:
	cd $(FRONTEND_DIR) && if npm run | grep -q " test"; then npm run test; else echo "No test script found in package.json"; fi

## Help

help:
	@echo ""
	@echo "Available commands:"
	@echo "  make build-backend      Build the Go backend"
	@echo "  make build-frontend     Build the React frontend"
	@echo "  make run-backend        Run backend locally"
	@echo "  make run-frontend       Run frontend locally"
	@echo "  make docker-up          Start docker-compose services"
	@echo "  make docker-down        Stop docker-compose services"
	@echo "  make backup             Backup SQLite and DuckDB databases"
	@echo "  make clean              Clean build artifacts"
	@echo "  make test-backend       Run backend Go tests"
	@echo "  make test-frontend      Run frontend tests"
	@echo "  make help               Show this help message"
	@echo ""