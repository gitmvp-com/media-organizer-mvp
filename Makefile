# Media Organizer MVP Makefile

.PHONY: build run clean test help

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME=media-organizer

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  clean    - Remove build artifacts"
	@echo "  test     - Run tests"
	@echo "  install  - Download dependencies"

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

## run: Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run main.go

## clean: Remove build artifacts and database
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe
	rm -rf data/
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	go test -v ./...

## install: Download Go module dependencies
install:
	@echo "Downloading dependencies..."
	go mod download
	@echo "Dependencies installed"

## build-all: Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-linux-amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)-windows-amd64.exe
	@echo "Multi-platform build complete"
