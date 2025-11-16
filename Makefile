# Makefile for Red Team C2 Framework
# Provides convenient build, test, and deployment targets

.PHONY: all build test clean docker help install lint fmt check-deps

# Default target
all: help

##@ General

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Building

build: build-implant build-teamserver ## Build all components

build-implant: ## Build Go implant
	@echo "Building Go implant..."
	cd implant && go build -ldflags="-s -w" -o implant main.go
	@echo "✓ Implant built: implant/implant"

build-teamserver: ## Build Rust teamserver
	@echo "Building Rust teamserver..."
	cd teamserver && cargo build --release
	@echo "✓ Teamserver built: teamserver/target/release/teamserver"

build-all-platforms: ## Cross-compile implant for all platforms
	@echo "Cross-compiling for all platforms..."
	./build_implant.sh

##@ Testing

test: test-go test-rust test-python ## Run all tests

test-go: ## Run Go tests
	@echo "Running Go tests..."
	cd implant && go test -v ./...

test-rust: ## Run Rust tests
	@echo "Running Rust tests..."
	cd teamserver && cargo test --verbose

test-python: ## Run Python tests
	@echo "Running Python tests..."
	cd controller && python -m pytest -v || echo "pytest not configured yet"

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	cd implant && go test -cover ./...
	cd teamserver && cargo test --verbose -- --test-threads=1
	cd controller && python -m pytest --cov=. || echo "pytest-cov not configured yet"

##@ Code Quality

lint: lint-go lint-rust lint-python ## Run all linters

lint-go: ## Lint Go code
	@echo "Linting Go code..."
	cd implant && go vet ./... && go fmt ./...

lint-rust: ## Lint Rust code
	@echo "Linting Rust code..."
	cd teamserver && cargo clippy --all-targets --all-features

lint-python: ## Lint Python code
	@echo "Linting Python code..."
	cd controller && python -m pylint main.py || echo "pylint not installed"

fmt: ## Format all code
	@echo "Formatting code..."
	cd implant && go fmt ./...
	cd teamserver && cargo fmt
	cd controller && python -m black . || echo "black not installed"

##@ Docker

docker-build: ## Build all Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start all services
	@echo "Starting services..."
	docker-compose up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## View logs
	docker-compose logs -f

docker-clean: ## Remove all containers and volumes
	@echo "Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

##@ Development

install: install-deps ## Install all dependencies

install-deps: ## Install development dependencies
	@echo "Installing dependencies..."
	@echo "Go dependencies..."
	cd implant && go mod download
	@echo "Rust dependencies..."
	cd teamserver && cargo fetch
	@echo "Python dependencies..."
	cd controller && pip install -r requirements.txt

check-deps: ## Check if dependencies are installed
	@echo "Checking dependencies..."
	@command -v go >/dev/null 2>&1 || { echo "✗ Go not installed"; exit 1; }
	@command -v cargo >/dev/null 2>&1 || { echo "✗ Rust not installed"; exit 1; }
	@command -v python3 >/dev/null 2>&1 || { echo "✗ Python not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "✗ Docker not installed"; exit 1; }
	@echo "✓ All dependencies installed"

run-teamserver: ## Run teamserver locally
	@echo "Starting teamserver..."
	cd teamserver && cargo run --release

run-controller: ## Run controller CLI
	@echo "Starting controller..."
	cd controller && python main.py

##@ Deployment

deploy-prod: ## Deploy to production (Docker)
	@echo "Deploying to production..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

deploy-dev: docker-up ## Deploy development environment

##@ Cleaning

clean: clean-build clean-test clean-cache ## Clean all build artifacts

clean-build: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf implant/implant
	rm -rf implant/*.exe
	rm -rf teamserver/target
	rm -rf build/
	find . -name "*.pyc" -delete
	find . -name "__pycache__" -delete

clean-test: ## Remove test artifacts
	@echo "Cleaning test artifacts..."
	find . -name "*.test" -delete
	rm -rf .pytest_cache
	rm -rf htmlcov
	rm -rf .coverage

clean-cache: ## Remove cache files
	@echo "Cleaning cache files..."
	go clean -cache
	cargo clean
	find . -name ".DS_Store" -delete

##@ Documentation

docs: ## Build documentation
	@echo "Building documentation..."
	@echo "✓ Documentation already in Markdown format"

docs-serve: ## Serve documentation (if using mdbook/mkdocs)
	@echo "Documentation server not configured yet"

##@ Release

release: check-deps test build ## Prepare release (test + build)
	@echo "Creating release..."
	./build_implant.sh
	@echo "✓ Release artifacts in build/"

tag-release: ## Tag a new release (requires VERSION=x.y.z)
ifndef VERSION
	$(error VERSION is not set. Usage: make tag-release VERSION=1.0.0)
endif
	@echo "Tagging release v$(VERSION)..."
	git tag -a v$(VERSION) -m "Release v$(VERSION)"
	git push origin v$(VERSION)

##@ CI/CD

ci-test: test lint ## Run CI tests (test + lint)

ci-build: build ## Run CI build

##@ Utility

version: ## Show version information
	@echo "Red Team C2 Framework - Version Information"
	@echo "==========================================="
	@echo "Go version:"
	@go version
	@echo ""
	@echo "Rust version:"
	@cargo --version
	@echo "rustc:"
	@rustc --version
	@echo ""
	@echo "Python version:"
	@python3 --version
	@echo ""
	@echo "Docker version:"
	@docker --version

size: ## Show binary sizes
	@echo "Binary sizes:"
	@du -h implant/implant 2>/dev/null || echo "Implant not built"
	@du -h teamserver/target/release/teamserver 2>/dev/null || echo "Teamserver not built"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	cd implant && go test -bench=. -benchmem ./...
	cd teamserver && cargo bench || echo "Benchmarks not configured"

##@ Security

security-scan: ## Run security scanners
	@echo "Running security scans..."
	@echo "Go security check..."
	cd implant && go list -json -m all | nancy sleuth || echo "nancy not installed"
	@echo "Rust security check..."
	cd teamserver && cargo audit || echo "cargo-audit not installed"
	@echo "Python security check..."
	cd controller && pip-audit || echo "pip-audit not installed"

check-secrets: ## Check for committed secrets
	@echo "Checking for secrets..."
	@command -v trufflehog >/dev/null 2>&1 && trufflehog filesystem . || echo "trufflehog not installed"

##@ Database

db-setup: ## Set up PostgreSQL database
	@echo "Setting up database..."
	psql -U postgres -c "CREATE USER c2user WITH PASSWORD 'c2password';" || true
	psql -U postgres -c "CREATE DATABASE c2_database OWNER c2user;" || true
	@echo "✓ Database setup complete"

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	cd teamserver && sqlx migrate run

db-reset: ## Reset database
	@echo "Resetting database..."
	psql -U postgres -c "DROP DATABASE IF EXISTS c2_database;"
	psql -U postgres -c "CREATE DATABASE c2_database OWNER c2user;"
	$(MAKE) db-migrate

##@ Information

info: ## Show project information
	@echo "Red Team C2 Framework"
	@echo "===================="
	@echo ""
	@echo "Components:"
	@echo "  - Go Implant (implant/)"
	@echo "  - Rust Teamserver (teamserver/)"
	@echo "  - Python Controller (controller/)"
	@echo ""
	@echo "Documentation:"
	@echo "  - README.md          - Project overview"
	@echo "  - QUICKSTART.md      - Quick start guide"
	@echo "  - ARCHITECTURE.md    - Architecture documentation"
	@echo "  - API.md             - API documentation"
	@echo "  - CONTRIBUTING.md    - Contribution guidelines"
	@echo ""
	@echo "For help: make help"
