# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix

# Tool paths
MIGRATE_CMD=$(shell go env GOPATH)/bin/migrate
AIR_CMD=$(shell go env GOPATH)/bin/air

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) -v ./cmd/server

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v ./cmd/server

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
run: build
	./bin/$(BINARY_NAME)

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	$(AIR_CMD)

# Start development environment
dev-env:
	docker-compose up -d postgres redis

# Stop development environment
dev-env-stop:
	docker-compose down

# Reset development environment
dev-env-reset:
	docker-compose down -v
	docker-compose up -d postgres redis

# Run migrations (requires migrate tool)
migrate-up:
	$(MIGRATE_CMD) -path migrations -database "postgres://postgres:password@localhost:5432/p4rsec_dev?sslmode=disable" up

migrate-down:
	$(MIGRATE_CMD) -path migrations -database "postgres://postgres:password@localhost:5432/p4rsec_dev?sslmode=disable" down

migrate-create:
	$(MIGRATE_CMD) create -ext sql -dir migrations -seq $(name)

# Docker commands
docker-build:
	docker build -t p4rsec-server .

docker-run:
	docker run -p 8080:8080 --env-file .env p4rsec-server

# Linting (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Security scan (requires go vet and custom checks)
security:
	go vet ./...
	@echo "For comprehensive security scanning, install: go install github.com/securecodewarrior/gosec/cmd/gosec@latest"

# Install tools
install-tools:
	go install github.com/air-verse/air@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Development tools installed successfully!"
	@echo "Note: For security scanning, manually install gosec if needed"

.PHONY: build build-linux clean test test-coverage deps run dev dev-env dev-env-stop dev-env-reset migrate-up migrate-down migrate-create docker-build docker-run lint fmt security install-tools 