# Load environment variables from .env file
include .env

# Project Variables
APP_NAME := $(APP_NAME)
GO_VERSION := $(GO_VERSION)

# Docker image names
DOCKER_IMAGE_DEV := $(APP_NAME):dev
DOCKER_IMAGE_PROD := $(APP_NAME):prod

# Default target
.PHONY: all
all: build

# Build the Docker image (dev environment)
.PHONY: build-dev
build-dev:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE_DEV) -f deployments/docker/Dockerfile.dev .

# Build the Docker image (prod environment)
.PHONY: build-prod
build-prod:
	docker build --build-arg GO_VERSION=$(GO_VERSION) -t $(DOCKER_IMAGE_PROD) -f deployments/docker/Dockerfile.prod .

# Run the Docker container (dev environment)
.PHONY: run-dev
run-dev:
	docker compose -f docker-compose.yml up --build dev

# Run the Docker container (prod environment)
.PHONY: run-prod
run-prod:
	docker compose -f docker-compose.yml up --build prod

# Stop running Docker containers
.PHONY: stop
stop:
	docker compose -f docker-compose.yml down

# Initialize Go module & download deps (inside Docker container)
.PHONY: init
init:
	docker run --rm -v $(PWD):/app -w /app golang:$(GO_VERSION) go mod tidy

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
