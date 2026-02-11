# Makefile for inventory-cli

.PHONY: build run run-shell test test-cover bench clean fmt vet

# Build the binary
build:
	go build -o inventory-cli.exe ./cmd/inventory

# Run CLI directly (requires arguments)
run:
	go run ./cmd/inventory $(ARGS)

# Run in interactive shell mode
run-shell:
	go run ./cmd/inventory shell

# Run all tests
test:
	go test -v ./...

# Run tests with coverage report
test-cover:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	go test -bench . -run '^$' ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -f inventory-cli inventory-cli.exe coverage.out coverage.html
	rm -rf testdata/ 

# Installation for development
setup:
	go mod download
	go mod tidy

# All checks
check: fmt vet test

# Quick example: create a product
example-create:
	go run . create --name "Example Product" --price 99.99 --quantity 5 --category "Demo"

# Quick example: list all products
example-list:
	go run . list

# Quick example: import sample data (memory store)
example-import:
	go run . import --file data/products.json

# Quick example: import to file store
example-import-file:
	go run . --store file --store-file /tmp/inventory.json import --file data/products.json

# Quick example: export products
example-export:
	go run . export --file exported.json
