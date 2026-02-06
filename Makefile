.PHONY: build run dev test clean docker-build docker-up docker-down docker-dev-up docker-dev-down lint

# Variables
APP_NAME=ticket-service
MAIN_PATH=./cmd/main.go
BIN_PATH=./bin/$(APP_NAME)

# Build
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BIN_PATH) $(MAIN_PATH)
	@echo "Build complete: $(BIN_PATH)"

# Run
run: build
	@echo "Running $(APP_NAME)..."
	@$(BIN_PATH)

# Development with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@air -c .air.toml || go run $(MAIN_PATH)

# Test
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean
clean:
	@echo "Cleaning..."
	@rm -rf ./bin
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Lint (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Format
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Generate Swagger docs
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/main.go -o docs --parseDependency --parseInternal
	@echo "Swagger docs generated in docs/"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .

# Docker compose up
docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d

# Docker compose down
docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

# Docker compose dev up
docker-dev-up:
	@echo "Starting development Docker containers..."
	@docker-compose -f docker-compose.dev.yml up -d

# Docker compose dev down
docker-dev-down:
	@echo "Stopping development Docker containers..."
	@docker-compose -f docker-compose.dev.yml down

# Docker logs
docker-logs:
	@docker-compose logs -f $(APP_NAME)

# Generate swagger docs (requires swag)
swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/main.go -o ./docs

# Help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Build and run the application"
	@echo "  dev             - Run with hot reload (requires air)"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  lint            - Run linter (requires golangci-lint)"
	@echo "  fmt             - Format code"
	@echo "  tidy            - Tidy dependencies"
	@echo "  deps            - Download dependencies"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-up       - Start Docker containers"
	@echo "  docker-down     - Stop Docker containers"
	@echo "  docker-dev-up   - Start development Docker containers"
	@echo "  docker-dev-down - Stop development Docker containers"
	@echo "  docker-logs     - Show Docker logs"
	@echo "  swagger         - Generate Swagger documentation"
	@echo "  help            - Show this help message"
