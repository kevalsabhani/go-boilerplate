.PHONY: all build run test coverage lint fmt clean migrate-up migrate-down migrate-force migrate-create help

# Database configuration for migrations
DB_URL ?= postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable
MIGRATIONS_DIR ?= ./migrations

# Default target
all: build

## build: Build all commands from cmd/ into bin/
build:
	@mkdir -p bin
	@for dir in cmd/*/; do \
		name=$$(basename $$dir); \
		echo "Building $$name..."; \
		go build -o bin/$$name ./$$dir; \
	done

## run: Run the application
run:
	@go run ./cmd/go-boilerplate

## test: Run all tests
test:
	go test ./...

## coverage: Run tests with coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code using golangci-lint formatters
fmt:
	golangci-lint run --fix ./...

## tidy: Run go mod tidy
tidy: fmt
	go mod tidy
	go mod verify

## check: Runs fmt, lint and test
check: fmt lint test

## clean: Remove build artifacts and test output
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f *.test *.out *.prof
	rm -rf cpu.prof mem.prof

## migrate-up: Apply all database migrations
migrate-up:
	@echo "Applying all database migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

## migrate-down: Revert all database migrations
migrate-down:
	@echo "Reverting all database migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down

## migrate-force: Force database migration to a specific version (e.g., make migrate-force V=1)
migrate-force:
	@if [ -z "$(V)" ]; then \
		echo "Error: V (version) is required. Usage: make migrate-force V=1"; \
		exit 1; \
	fi
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(V)

## migrate-create: Create a new migration file (e.g., make migrate-create NAME=init)
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=init"; \
		exit 1; \
	fi
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'