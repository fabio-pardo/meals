.PHONY: help docs search-funcs search-routes search-models dev-setup test clean build run docker-up docker-down

# Default target
help:
	@echo "Meals App Development Commands"
	@echo "============================="
	@echo ""
	@echo "Documentation:"
	@echo "  docs          Generate API documentation"
	@echo "  docs-serve    Serve documentation locally"
	@echo ""
	@echo "Code Search:"
	@echo "  search-funcs  List all function definitions"
	@echo "  search-routes List all route definitions"
	@echo "  search-models List all model definitions"
	@echo "  search-errors List error handling patterns"
	@echo "  search-db     List database operations"
	@echo ""
	@echo "Development:"
	@echo "  dev-setup     Set up development environment"
	@echo "  run           Run the application"
	@echo "  test          Run all tests"
	@echo "  test-verbose  Run tests with verbose output"
	@echo "  build         Build the application"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up     Start PostgreSQL and Redis"
	@echo "  docker-down   Stop all containers"
	@echo "  docker-logs   View container logs"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint          Run linter"
	@echo "  format        Format code"
	@echo "  vet           Run go vet"
	@echo ""

# Documentation generation
docs:
	@echo "Generating API documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init; \
		echo "Documentation generated in docs/"; \
	else \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init; \
		echo "Documentation generated in docs/"; \
	fi

docs-serve:
	@echo "Serving documentation..."
	@if command -v swagger-ui-serve >/dev/null 2>&1; then \
		swagger-ui-serve docs/api/openapi.yaml; \
	else \
		echo "Install swagger-ui-serve: npm install -g swagger-ui-serve"; \
		echo "Or view docs/api/openapi.yaml in Swagger Editor"; \
	fi

# Code search commands
search-funcs:
	@echo "=== All Functions ==="
	@rg "^func [A-Z]" --type go --no-heading --line-number

search-routes:
	@echo "=== All Routes ==="
	@rg "router\.(GET|POST|PUT|DELETE)" --type go --no-heading --line-number

search-models:
	@echo "=== All Models ==="
	@rg "^type.*struct" --type go --no-heading --line-number

search-errors:
	@echo "=== Error Handling ==="
	@rg "RespondWithError|HandleAppError|ErrorResponse" --type go --no-heading --line-number

search-db:
	@echo "=== Database Operations ==="
	@rg "store\.DB\.|WithTransaction|\.Create\(|\.Save\(|\.Delete\(" --type go --no-heading --line-number

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing Go tools..."
	@go mod tidy
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Checking for required tools..."
	@if ! command -v docker >/dev/null 2>&1; then \
		echo "Warning: Docker not found. Install Docker for local development."; \
	fi
	@if ! command -v docker-compose >/dev/null 2>&1; then \
		echo "Warning: docker-compose not found. Install docker-compose for local development."; \
	fi
	@echo "Development environment setup complete!"

# Application commands
run:
	@echo "Starting Meals application..."
	@go run main.go

build:
	@echo "Building application..."
	@go build -o bin/meals main.go
	@echo "Binary created: bin/meals"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf docs/swagger/
	@go clean
	@echo "Clean complete!"

# Testing
test:
	@echo "Running tests..."
	@go test ./...

test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker commands
docker-up:
	@echo "Starting PostgreSQL and Redis containers..."
	@docker-compose up -d
	@echo "Containers started. Use 'make docker-logs' to view logs."

docker-down:
	@echo "Stopping all containers..."
	@docker-compose down
	@echo "Containers stopped."

docker-logs:
	@echo "Container logs:"
	@docker-compose logs -f

docker-reset:
	@echo "Resetting Docker environment..."
	@docker-compose down -v
	@docker-compose up -d
	@echo "Docker environment reset complete."

# Code quality
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

format:
	@echo "Formatting code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	else \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
		goimports -w .; \
	fi

vet:
	@echo "Running go vet..."
	@go vet ./...

# Database commands
db-migrate:
	@echo "Running database migrations..."
	@go run main.go --migrate-only
	@echo "Migrations complete."

db-reset:
	@echo "Resetting database..."
	@make docker-down
	@make docker-up
	@sleep 5
	@make db-migrate
	@echo "Database reset complete."

# Context generation
generate-context:
	@echo "Generating code context..."
	@echo "# Meals App Code Context" > CONTEXT.md
	@echo "Generated: $$(date)" >> CONTEXT.md
	@echo "" >> CONTEXT.md
	@echo "## Project Structure" >> CONTEXT.md
	@if command -v tree >/dev/null 2>&1; then \
		tree -I 'bin|vendor|node_modules|.git|*.log' >> CONTEXT.md; \
	else \
		find . -type f -name "*.go" | head -20 >> CONTEXT.md; \
	fi
	@echo "" >> CONTEXT.md
	@echo "## All Functions" >> CONTEXT.md
	@rg "^func [A-Z]" --type go >> CONTEXT.md
	@echo "" >> CONTEXT.md
	@echo "## All Routes" >> CONTEXT.md
	@rg "router\.(GET|POST|PUT|DELETE)" --type go >> CONTEXT.md
	@echo "" >> CONTEXT.md
	@echo "## All Models" >> CONTEXT.md
	@rg "^type.*struct" --type go >> CONTEXT.md
	@echo "Context generated: CONTEXT.md"

# Development workflow
dev: docker-up
	@echo "Starting development environment..."
	@sleep 3
	@make run

dev-test: docker-up
	@echo "Running tests in development environment..."
	@sleep 3
	@make test

# Production build
prod-build:
	@echo "Building for production..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/meals main.go
	@echo "Production binary created: bin/meals"

# Install development dependencies
install-deps:
	@echo "Installing development dependencies..."
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/air-verse/air@latest
	@echo "Dependencies installed!"

# Live reload development
dev-watch:
	@echo "Starting development with live reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Installing air for live reload..."; \
		go install github.com/air-verse/air@latest; \
		air; \
	fi 