.PHONY: build test clean run docker-build helm-lint help

# Variables
BINARY_NAME=exporter
DOCKER_IMAGE=cph-metro-exporter
DOCKER_TAG=latest
HELM_CHART=./helm/cph-metro-exporter

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o $(BINARY_NAME) ./cmd/exporter

test: ## Run tests
	go test ./... -v -cover

test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

run: build ## Build and run the exporter locally
	PORT=9100 LOG_LEVEL=debug ./$(BINARY_NAME)

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Build and run Docker container
	docker run -p 9100:9100 $(DOCKER_IMAGE):$(DOCKER_TAG)

helm-lint: ## Lint Helm chart
	helm lint $(HELM_CHART)

helm-template: ## Render Helm templates
	helm template cph-metro-exporter $(HELM_CHART)

helm-install: ## Install Helm chart locally
	helm install cph-metro-exporter $(HELM_CHART)

helm-uninstall: ## Uninstall Helm chart
	helm uninstall cph-metro-exporter

deps: ## Download dependencies
	go mod download
	go mod tidy

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

lint: fmt vet ## Run formatters and linters
	@echo "Linting complete"

all: clean deps lint test build ## Run all build steps

compose-up: ## Start docker-compose stack
	docker compose up -d

compose-down: ## Stop docker-compose stack
	docker compose down

compose-logs: ## Show docker-compose logs
	docker compose logs -f

compose-restart: compose-down compose-up ## Restart docker-compose stack

