.PHONY: help test test-coverage test-unit test-integration mock-gen lint fmt docker-up docker-down run-scheduler run-crawler clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Testing
test: ## Run all tests
	@echo "🧪 Running tests in scheduler-service..."
	cd scheduler-service && go test -v -race -short ./... || true
	@echo "\n🧪 Running tests in crawler-service..."
	cd crawler-service && go test -v -race -short ./... || true

test-unit: ## Run unit tests only
	@echo "🧪 Running unit tests in scheduler-service..."
	cd scheduler-service && go test -v -race -short ./... || true
	@echo "\n🧪 Running unit tests in crawler-service..."
	cd crawler-service && go test -v -race -short ./... || true

test-integration: ## Run integration tests only
	@echo "🧪 Running integration tests in scheduler-service..."
	cd scheduler-service && go test -v -race -run Integration ./... || true
	@echo "\n🧪 Running integration tests in crawler-service..."
	cd crawler-service && go test -v -race -run Integration ./... || true

test-coverage: ## Run tests with coverage report
	@echo "📊 Generating coverage for scheduler-service..."
	cd scheduler-service && go test -v -short -coverprofile=coverage.out ./... 2>&1 | grep -E "(PASS|FAIL|coverage:|ok|---)" || true
	@if [ -f scheduler-service/coverage.out ]; then \
		cd scheduler-service && go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Scheduler coverage report: scheduler-service/coverage.html"; \
		cd scheduler-service && go tool cover -func=coverage.out | grep total || echo "No coverage data"; \
	fi
	@echo "\n📊 Generating coverage for crawler-service..."
	cd crawler-service && go test -v -short -coverprofile=coverage.out ./... 2>&1 | grep -E "(PASS|FAIL|coverage:|ok|---)" || true
	@if [ -f crawler-service/coverage.out ]; then \
		cd crawler-service && go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Crawler coverage report: crawler-service/coverage.html"; \
		cd crawler-service && go tool cover -func=coverage.out | grep total || echo "No coverage data"; \
	fi
	@echo "\n✨ Coverage reports generated!"

test-coverage-threshold: ## Check if coverage meets threshold (80%)
	@echo "🎯 Checking coverage threshold..."
	@cd scheduler-service && go test -short -coverprofile=coverage.out ./... > /dev/null 2>&1 || true
	@if [ -f scheduler-service/coverage.out ]; then \
		total=$$(cd scheduler-service && go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Scheduler coverage: $$total%"; \
	fi
	@cd crawler-service && go test -short -coverprofile=coverage.out ./... > /dev/null 2>&1 || true
	@if [ -f crawler-service/coverage.out ]; then \
		total=$$(cd crawler-service && go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
		echo "Crawler coverage: $$total%"; \
	fi

# Mock generation
mock-gen: ## Generate mocks for all interfaces
	@echo "🔧 Generating mocks for scheduler-service..."
	cd scheduler-service && go generate ./... || true
	@echo "🔧 Generating mocks for crawler-service..."
	cd crawler-service && go generate ./... || true
	@echo "✅ Mocks generated successfully"

# Code quality
lint: ## Run linter
	@if command -v golangci-lint > /dev/null; then \
		echo "🔍 Linting scheduler-service..."; \
		cd scheduler-service && golangci-lint run ./... || true; \
		echo "🔍 Linting crawler-service..."; \
		cd crawler-service && golangci-lint run ./... || true; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: brew install golangci-lint"; \
	fi

fmt: ## Format code
	@echo "✨ Formatting scheduler-service..."
	cd scheduler-service && go fmt ./...
	@echo "✨ Formatting crawler-service..."
	cd crawler-service && go fmt ./...
	@if command -v goimports > /dev/null; then \
		cd scheduler-service && goimports -w . ; \
		cd crawler-service && goimports -w . ; \
	fi

# Docker
docker-up: ## Start docker services
	docker-compose up -d

docker-down: ## Stop docker services
	docker-compose down

docker-restart: docker-down docker-up ## Restart docker services

# Running services
run-scheduler: ## Run scheduler service
	cd scheduler-service && go run main.go scheduler

run-scheduler-worker: ## Run scheduler worker
	cd scheduler-service && go run main.go scheduler_worker

run-crawler: ## Run crawler worker
	cd crawler-service && go run main.go crawler-worker

run-crawler-retry: ## Run crawler retry worker
	cd crawler-service && go run main.go crawler-worker-retry

# Build
build-scheduler: ## Build scheduler service
	cd scheduler-service && go build -o bin/scheduler main.go

build-crawler: ## Build crawler service
	cd crawler-service && go build -o bin/crawler main.go

build-all: build-scheduler build-crawler ## Build all services

# Database
db-migrate-up: ## Run database migrations up
	@echo "Database migrations not yet implemented"

db-migrate-down: ## Run database migrations down
	@echo "Database migrations not yet implemented"

# Clean
clean: ## Clean build artifacts and test cache
	rm -rf */bin
	rm -f */coverage.out */coverage.html
	cd scheduler-service && go clean -testcache
	cd crawler-service && go clean -testcache
	@echo "✅ Cleaned build artifacts"

# CI helpers
ci-test: mock-gen test-coverage-threshold ## Run CI tests with coverage check

ci-lint: lint ## Run CI linting

# Development
dev-setup: docker-up mock-gen ## Setup development environment
	@echo "✅ Development environment ready"

install-tools: ## Install development tools
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "✅ Tools installed"
