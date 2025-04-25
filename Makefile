# Makefile for managing the Go project

# Variables
APP_NAME := httpclient
GO_FILES := $(shell find . -name '*.go' -not -path "./vendor/*")

# Default target
.PHONY: all
all: build lint test

# Build the application
.PHONY: build
build:
	@echo "Building the application..."
	@go build -o $(APP_NAME) ./...

# Run unit tests
.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -v ./...

# Run linting
.PHONY: lint
lint:
	@echo "Running linting..."
	@golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Tidy up the module
.PHONY: tidy
tidy:
	@echo "Tidying up the module..."
	@go mod tidy

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(APP_NAME)

# Run all tools
.PHONY: tools
tools: fmt vet lint
	@echo "All tools executed."

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  test    - Run unit tests"
	@echo "  lint    - Run linting"
	@echo "  fmt     - Format the code"
	@echo "  vet     - Run go vet"
	@echo "  tidy    - Tidy up the module"
	@echo "  deps    - Install dependencies"
	@echo "  clean   - Clean up build artifacts"
	@echo "  tools   - Run all tools (fmt, vet, lint)"
	@echo "  help    - Show this help message"