# Load environment variables from .env file if it exists
-include .env

# Project Variables with defaults in case .env is missing
APP_NAME ?= ehass
GO_VERSION ?= 1.21

# Docker image names
DOCKER_IMAGE_DEV := $(APP_NAME):dev
DOCKER_IMAGE_PROD := $(APP_NAME):prod
DOCKER_IMAGE_STAG := $(APP_NAME):stag

# Default target
.PHONY: all
all: build

# Build target depends on build-dev by default
.PHONY: build
build: build-dev

# Build the Docker image (dev environment)
.PHONY: build-dev
build-dev:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE_DEV) -f deployments/docker/Dockerfile.dev .

# Build the Docker image (prod environment)
.PHONY: build-prod
build-prod:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE_PROD) -f deployments/docker/Dockerfile.prod .

# Build the Docker image (staging environment)
.PHONY: build-stag
build-stag:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE_STAG) -f deployments/docker/Dockerfile.stag .

# Run the Docker container (dev environment)
.PHONY: run-dev
run-dev:
	docker compose -f docker-compose.yml up --build dev

# Run the Docker container (prod environment)
.PHONY: run-prod
run-prod:
	docker compose -f docker-compose.yml up --build prod

# Run just the database
.PHONY: run-db
run-db:
	docker compose -f docker-compose.yml up -d postgres
	@echo "Waiting for PostgreSQL to fully initialize..."
	@sleep 5
	$(MAKE) migrate

# Stop running Docker containers
.PHONY: stop
stop:
	docker compose -f docker-compose.yml down

# Initialize Go module & download deps (inside Docker container)
.PHONY: init
init:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go mod tidy

# Run database migrations
.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	docker compose -f docker-compose.yml exec -T dev go run cmd/server/main.go migrate || \
	docker run --rm --network=host -v $(PWD):/app -w /app golang:$(GO_VERSION) \
		go run cmd/server/main.go migrate

# Create a new migration file
.PHONY: migration-create
migration-create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p internal/migrations; \
	touch internal/migrations/$${timestamp}_$${name}.go; \
	echo "package migrations\n\nfunc init() {\n\tregisterMigration(\"$${timestamp}_$${name}\", up$${timestamp}, down$${timestamp})\n}\n\nfunc up$${timestamp}(tx *gorm.DB) error {\n\t// TODO: Implement migration\n\treturn nil\n}\n\nfunc down$${timestamp}(tx *gorm.DB) error {\n\t// TODO: Implement rollback\n\treturn nil\n}" > internal/migrations/$${timestamp}_$${name}.go; \
	echo "Created migration file: internal/migrations/$${timestamp}_$${name}.go"

# Roll back the last migration
.PHONY: migrate-rollback
migrate-rollback:
	@echo "Rolling back the last migration..."
	docker compose -f docker-compose.yml exec -T dev go run cmd/server/main.go migrate rollback || \
	docker run --rm --network=host -v $(PWD):/app -w /app golang:$(GO_VERSION) \
		go run cmd/server/main.go migrate rollback

# Generate Swagger documentation
.PHONY: swagger
swagger:
	docker run --rm -v $(PWD):/app -w /app quay.io/goswagger/swagger generate spec -o ./internal/docs/swagger.json --scan-models
	@echo "Swagger docs generated at ./internal/docs/swagger.json"

# Run the swag tool to generate Swagger docs (requires swag installed)
.PHONY: swaggo
swaggo:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) sh -c "go install github.com/swaggo/swag/cmd/swag@latest && /go/bin/swag init -g cmd/server/main.go -o internal/docs"
	@echo "Swagger docs generated at ./internal/docs"

# Run tests (inside Docker container)
.PHONY: test
test:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go test ./...

# Run tests with coverage (inside Docker container)
.PHONY: coverage
coverage:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go test -coverprofile=coverage.out ./...
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go tool cover -html=coverage.out

# Format the code (inside Docker container)
.PHONY: fmt
fmt:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go fmt ./...

# Lint (inside Docker container, requires golangci-lint installed)
.PHONY: lint
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint run

# Clean (remove binaries and build cache from the Docker container)
.PHONY: clean
clean:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go clean
