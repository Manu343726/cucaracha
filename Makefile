.PHONY: help test test-quick test-verbose test-coverage test-docker test-workflows generate build build-fast build-docker build-docker-full lint format format-fix vet run repl clean push act-install act-test act-test-quick act-test-docker status log docker-clean clean-all dev dev-quick

# Colors for output
BLUE=\033[0;34m
GREEN=\033[0;32m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Cucaracha Development Makefile$(NC)"
	@echo ""
	@echo "$(GREEN)Local Development:$(NC)"
	@echo "  make generate          - Run code generation (go generate)"
	@echo "  make build             - Run code generation (depends on generate)"
	@echo "  make test              - Run all unit tests locally (depends on build)"
	@echo "  make test-quick        - Run quick unit tests (skip slow ones)"
	@echo "  make lint              - Run linting checks (golangci-lint)"
	@echo "  make format            - Check code formatting (gofmt)"
	@echo "  make vet               - Run go vet analysis"
	@echo "  make run               - Run cucaracha with go run"
	@echo "  make repl              - Start Cucaracha in interactive REPL mode"
	@echo ""
	@echo "$(GREEN)Docker Testing:$(NC)"
	@echo "  make build-docker      - Build Docker image (cucaracha:latest)"
	@echo "  make build-docker-full - Full Docker build with LLVM compile (60+ min)"
	@echo "  make test-docker       - Run tests in Docker container"
	@echo ""
	@echo "$(GREEN)GitHub Actions Local Testing (requires 'act'):$(NC)"
	@echo "  make act-install       - Install 'act' tool for local workflow testing"
	@echo "  make act-test          - Run all workflows locally"
	@echo "  make act-test-quick    - Run quick-tests workflow locally"
	@echo "  make act-test-docker   - Run build-and-test workflow locally (dry-run)"
	@echo ""
	@echo "$(GREEN)Repository:$(NC)"
	@echo "  make push              - Stage, commit, and push all changes"
	@echo "  make clean             - Clean build artifacts and caches"
	@echo ""

# ============================================================================
# Code Generation
# ============================================================================

generate: ## Run code generation (go generate)
	@echo "$(BLUE)Running code generation...$(NC)"
	go generate ./...
	@echo "$(GREEN)Code generation complete$(NC)"

# ============================================================================
# Local Testing
# ============================================================================

test: build ## Run all unit tests (depends on build for code generation)
	@echo "$(BLUE)Running all unit tests...$(NC)"
	go test ./... -v --skip TestIntegration -timeout 30s

test-quick: build ## Run quick unit tests (skip slow tests, depends on build)
	@echo "$(BLUE)Running quick unit tests...$(NC)"
	go test ./... -v --skip TestIntegration -short -timeout 20s

test-verbose: build ## Run tests with verbose output and race condition detection
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	go test ./... -v -race --skip TestIntegration -timeout 30s

test-coverage: build ## Run tests and generate coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test ./... -v -coverprofile=coverage.out --skip TestIntegration -timeout 30s
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

# ============================================================================
# Code Quality
# ============================================================================

lint: ## Run golangci-lint
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	golangci-lint run ./... --timeout=5m

format: ## Check code formatting
	@echo "$(BLUE)Checking code formatting...$(NC)"
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "$(RED)Code formatting issues found:$(NC)"; \
		gofmt -s -d .; \
		exit 1; \
	else \
		echo "$(GREEN)Code formatting OK$(NC)"; \
	fi

format-fix: ## Fix code formatting issues
	@echo "$(BLUE)Fixing code formatting...$(NC)"
	gofmt -s -w .
	@echo "$(GREEN)Code formatting fixed$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...

# ============================================================================
# Building
# ============================================================================

build: generate ## Run code generation (depends on generate)
	@echo "$(BLUE)Build complete (code generation done)$(NC)"

run: build ## Run cucaracha with go run
	@echo "$(BLUE)Running cucaracha...$(NC)"
	go run ./main.go

repl: build ## Run cucaracha in interactive REPL mode
	@echo "$(BLUE)Starting Cucaracha REPL (type 'help' for commands, 'exit' to quit)...$(NC)"
	@go run ./main.go

# ============================================================================
# Docker
# ============================================================================

build-docker: ## Build Docker image (fast, uses cache)
	@echo "$(BLUE)Building Docker image (cucaracha:latest)...$(NC)"
	DOCKER_BUILDKIT=1 docker build -t cucaracha:latest .github/Dockerfile
	@echo "$(GREEN)Docker image built: cucaracha:latest$(NC)"

build-docker-full: ## Full Docker build without cache (slow - 60+ min first time)
	@echo "$(BLUE)Full Docker build (no cache, slow!)...$(NC)"
	DOCKER_BUILDKIT=1 docker build --no-cache -t cucaracha:latest .github/Dockerfile
	@echo "$(GREEN)Docker image built: cucaracha:latest$(NC)"

test-docker: build-docker ## Run tests in Docker container
	@echo "$(BLUE)Running tests in Docker...$(NC)"
	docker run --rm \
		-v $$(pwd):/workspace \
		-w /workspace \
		cucaracha:latest \
		go test ./... -v --skip TestIntegration -timeout 30s

docker-shell: build-docker ## Open interactive shell in Docker container
	@echo "$(BLUE)Starting Docker shell...$(NC)"
	docker run --rm -it \
		-v $$(pwd):/workspace \
		-w /workspace \
		cucaracha:latest \
		bash

docker-cli-tests: build-docker ## Test all CLI commands in Docker
	@echo "$(BLUE)Testing CLI commands in Docker...$(NC)"
	@docker run --rm cucaracha:latest --help
	@echo "---"
	@docker run --rm cucaracha:latest debug --help
	@echo "---"
	@docker run --rm cucaracha:latest cpu --help
	@echo "---"
	@docker run --rm cucaracha:latest tools --help
	@echo "---"
	@docker run --rm cucaracha:latest tui --help
	@echo "$(GREEN)All CLI tests passed$(NC)"

# ============================================================================
# GitHub Actions Local Testing with 'act'
# ============================================================================

act-install: ## Install 'act' tool for local GitHub Actions testing
	@echo "$(BLUE)Installing act...$(NC)"
	@command -v act >/dev/null 2>&1 && echo "$(GREEN)act already installed$(NC)" || \
		(curl https://raw.githubusercontent.com/nektos/act/master/install.sh | bash && \
		echo "$(GREEN)act installed successfully$(NC)")

act-test: act-install ## Run all workflows with act (dry-run)
	@echo "$(BLUE)Running all workflows locally (dry-run)...$(NC)"
	act --dry-run

act-test-quick: act-install ## Run quick-tests workflow locally
	@echo "$(BLUE)Running quick-tests workflow locally...$(NC)"
	act -j quick-tests

act-test-docker: act-install ## Run build-and-test workflow locally (dry-run)
	@echo "$(BLUE)Running build-and-test workflow locally (dry-run)...$(NC)"
	act -j build-docker --dry-run

act-list: act-install ## List all available workflows
	@echo "$(BLUE)Available workflows:$(NC)"
	act -l

# ============================================================================
# Repository Management
# ============================================================================

status: ## Show git status
	@echo "$(BLUE)Git status:$(NC)"
	@git status

log: ## Show recent commits
	@echo "$(BLUE)Recent commits:$(NC)"
	@git log --oneline -10

push: ## Stage all changes, commit, and push
	@echo "$(BLUE)Staging all changes...$(NC)"
	git add -A
	@echo "$(BLUE)Committing changes...$(NC)"
	git commit -m "Update: $(shell date +%Y-%m-%d)" || true
	@echo "$(BLUE)Pushing to origin...$(NC)"
	git push origin main
	@echo "$(GREEN)Push complete$(NC)"

# ============================================================================
# Cleanup
# ============================================================================

clean: ## Clean build artifacts and caches
	@echo "$(BLUE)Cleaning up...$(NC)"
	rm -f cucaracha
	rm -f coverage.out coverage.html
	go clean -cache -testcache
	@echo "$(GREEN)Cleanup complete$(NC)"

docker-clean: ## Remove Docker images
	@echo "$(BLUE)Removing Docker images...$(NC)"
	docker rmi cucaracha:latest -f 2>/dev/null || true
	@echo "$(GREEN)Docker cleanup complete$(NC)"

clean-all: clean docker-clean ## Clean everything
	@echo "$(GREEN)Full cleanup complete$(NC)"

# ============================================================================
# Development Workflow Shortcuts
# ============================================================================

dev: lint format vet test ## Run full local development checks
	@echo "$(GREEN)✓ Development checks complete$(NC)"

dev-quick: format vet test-quick ## Quick development checks
	@echo "$(GREEN)✓ Quick checks complete$(NC)"

# ============================================================================
# Default target
# ============================================================================

.DEFAULT_GOAL := help

# Display help when makefile is called with no arguments
%:
	@$(MAKE) help
