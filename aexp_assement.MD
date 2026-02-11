# Instructions

Welcome, and thank you for applying! Please read the instructions carefully and submit your completed assessment.

*Instructions:*
- Write your code using Go 1.21 or higher.
- Use only standard Go libraries and the following allowed dependencies:
  - [Cobra](https://github.com/spf13/cobra) for CLI framework
  - [Viper](https://github.com/spf13/viper) for configuration management (optional)
  - Standard testing libraries for unit tests
- Do not use heavy frameworks or ORMs. Keep dependencies minimal.
- Push your code to a public GitHub repository so interviewers are able to access and review your submission.
- Include a README.md with clear instructions on how to build, test, and run your application.

# Scenario

You are joining a team responsible for building robust CLI tools and backend services in Go. The team values idiomatic Go code, proper error handling, concurrency patterns, and comprehensive testing. Your task is to implement a production-grade CLI application for managing a product inventory system using modern Go practices, Cobra for command-line interface, and demonstrating proficiency in goroutines, channels, and context management.

---

## Main Task: Product Inventory CLI Application

Design and implement a CLI-based product inventory management system in Go 1.21 or higher. The application should support CRUD operations for products, where each product has:
- `id` (string, UUID format)
- `name` (string)
- `price` (float64)
- `quantity` (int)
- `category` (string)

### Task 1: Custom Error Types

Implement custom error types for your application:
- `ProductNotFoundError` - returned when a product with the given ID is not found
- `InvalidProductError` - returned when product validation fails (e.g., negative price, empty name)
- `DuplicateProductError` - returned when attempting to create a product with an existing ID

Your error types should:
- Implement the `error` interface
- Include relevant context (product ID, validation details, etc.)
- Be easily distinguishable using `errors.Is()` or `errors.As()` patterns
- Provide meaningful error messages

### Task 2: Thread-Safe In-Memory Store

Implement a thread-safe in-memory product store that supports concurrent access:
- Use appropriate synchronization primitives (mutex, RWMutex, or channels)
- Support concurrent reads and writes
- Demonstrate proper use of defer for cleanup
- Implement the following interface:

```go
type ProductStore interface {
    Create(ctx context.Context, product Product) error
    Get(ctx context.Context, id string) (Product, error)
    Update(ctx context.Context, id string, product Product) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ListFilter) ([]Product, error)
    BulkImport(ctx context.Context, products []Product) error
}
```

### Task 3: Comprehensive Testing

Write comprehensive tests using Go's testing package:
- Unit tests for all store methods covering positive and negative cases
- Table-driven tests where appropriate
- Concurrent access tests using goroutines to verify thread safety
- Benchmark tests for critical operations
- Mock implementations for testing CLI commands in isolation
- Achieve at least 80% code coverage
- Use `t.Run()` for test subtests

### Task 4: Dependency Injection & Interfaces

Refactor your application to demonstrate proper dependency injection:
- Define clear interfaces for your data store
- Allow injection of different store implementations (in-memory, file-based)
- Create a factory pattern or constructor functions for creating store instances
- Demonstrate the Liskov Substitution Principle by implementing at least two store types:
  - In-memory store (primary implementation)
  - JSON file-based store (secondary implementation with file persistence)

---

## Task 5: Cobra CLI Implementation

Build a complete CLI application using the Cobra framework. Your CLI should include the following commands and subcommands:

### Root Command
```bash
inventory-cli - A product inventory management system
```

### Commands Structure

#### 1. `create` command
```bash
inventory-cli create --name "Laptop" --price 999.99 --quantity 10 --category "Electronics"
```
- Accept flags for all product fields
- Generate UUID for product ID automatically
- Validate input (positive price, non-empty name, etc.)
- Print confirmation with created product details

#### 2. `get` command
```bash
inventory-cli get <product-id>
```
- Accept product ID as positional argument
- Display product details in a formatted table
- Handle ProductNotFoundError gracefully

#### 3. `list` command
```bash
inventory-cli list [--category "Electronics"] [--min-price 100] [--max-price 2000]
```
- Support optional filtering by category and price range
- Display results in a formatted table
- Support sorting by name, price, or quantity
- Implement `--output json` flag for JSON output format

#### 4. `update` command
```bash
inventory-cli update <product-id> [--name "New Name"] [--price 899.99] [--quantity 15]
```
- Accept product ID as positional argument
- Support partial updates via flags
- Validate updated fields
- Print updated product details

#### 5. `delete` command
```bash
inventory-cli delete <product-id>
```
- Accept product ID as positional argument
- Prompt for confirmation (or use `--force` flag to skip)
- Print deletion confirmation

#### 6. `import` command
```bash
inventory-cli import --file products.json
```
- Read products from a JSON file
- Support bulk import with concurrent processing
- Use goroutines and worker pools for processing
- Report success/failure statistics
- Implement proper error handling for malformed JSON

#### 7. `export` command
```bash
inventory-cli export --file products.json [--category "Electronics"]
```
- Export products to JSON file
- Support optional filtering
- Create file with proper permissions

### Additional Requirements for Cobra Implementation

- Implement persistent flags (e.g., `--verbose`, `--config`)
- Add command aliases for common operations
- Implement proper command help text and examples
- Use `PreRun` hooks for validation and setup
- Use `PostRun` hooks for cleanup
- Demonstrate proper flag binding with Viper (optional)
- Implement command completion generation (bash, zsh, fish)

---

## Task 6: Concurrency & Context Management

Demonstrate advanced Go concurrency patterns:

### Concurrent Bulk Processing
Implement a bulk import feature that:
- Processes products concurrently using a worker pool pattern
- Limits concurrent operations (e.g., max 10 workers)
- Uses channels for work distribution
- Properly propagates context for cancellation
- Collects and aggregates results
- Handles partial failures gracefully

### Context Cancellation
- Accept context in all store methods
- Implement timeout handling for long-running operations
- Demonstrate proper context cancellation propagation
- Handle `context.Canceled` and `context.DeadlineExceeded` errors

### Example Implementation Pattern
```go
func (s *Store) BulkImport(ctx context.Context, products []Product) error {
    // Implement worker pool with context cancellation
    // Use select statement to handle context.Done()
    // Report progress through channels
}
```

---

## Task 7: Configuration Management

Implement configuration management using Viper (optional but recommended):
- Support configuration from:
  - Config file (YAML or JSON)
  - Environment variables
  - Command-line flags (with precedence)
- Configuration should include:
  - Storage backend selection (memory/file)
  - File path for file-based storage
  - Default output format
  - Logging level
- Demonstrate proper configuration precedence

---

## Task 8: Logging & Observability

Implement structured logging:
- Use Go's `log/slog` package (Go 1.21+) or a minimal logging library
- Log important operations (create, update, delete)
- Include structured fields (product_id, operation, duration)
- Support different log levels (DEBUG, INFO, WARN, ERROR)
- Make logging configurable via flags or config file

---

## Task 9: Docker Containerization

Create a multi-stage Dockerfile for your CLI application:

### Requirements
- Use multi-stage build to minimize final image size
- Build stage should compile the Go binary
- Runtime stage should use a minimal base image (alpine or distroless)
- Binary should be statically compiled (CGO_ENABLED=0)
- Include health check if applicable
- Support configuration via environment variables

### Example Usage
```bash
# Build the Docker image
docker build -t inventory-cli:latest .

# Run the CLI in a container
docker run --rm inventory-cli:latest list

# Run with volume for persistent storage
docker run --rm -v $(pwd)/data:/data inventory-cli:latest create --name "Product"
```

### Dockerfile Best Practices
- Use appropriate base images
- Leverage build cache
- Create non-root user for security
- Set appropriate WORKDIR
- Use .dockerignore file
- Document environment variables

---

## Task 10: Documentation & Code Quality

### README.md
Include a comprehensive README with:
- Project overview and features
- Prerequisites and dependencies
- Build instructions
- Usage examples for all commands
- Testing instructions
- Docker instructions
- Project structure explanation
- Design decisions and trade-offs

### Code Quality
- Follow Go best practices and idioms
- Use `go fmt` and `go vet`
- Consider using `golangci-lint` for additional checks
- Add meaningful comments for exported functions
- Use meaningful variable and function names
- Keep functions small and focused
- Demonstrate SOLID principles where applicable

### Go Modules
- Use Go modules for dependency management
- Include `go.mod` and `go.sum` files
- Pin dependency versions appropriately

---

# References

**Go Documentation:**
- [Go Official Documentation](https://go.dev/doc/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

**Cobra Framework:**
- [Cobra GitHub Repository](https://github.com/spf13/cobra)
- [Cobra User Guide](https://github.com/spf13/cobra/blob/main/user_guide.md)

**Viper Configuration:**
- [Viper GitHub Repository](https://github.com/spf13/viper)

**Testing & Quality:**
- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Race Detector](https://go.dev/doc/articles/race_detector)

**Concurrency:**
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [Context Package](https://pkg.go.dev/context)

---

# Submission Checklist

Before submitting, ensure your repository includes:

- [ ] Complete source code with proper package structure
- [ ] `go.mod` and `go.sum` files
- [ ] Comprehensive test suite with `go test -v -race ./...` passing
- [ ] `README.md` with build and usage instructions
- [ ] `Dockerfile` and `.dockerignore`
- [ ] All commands working as specified
- [ ] Custom error types implemented
- [ ] Thread-safe store with concurrent access tests
- [ ] Cobra CLI with all specified commands
- [ ] Configuration management (at minimum via flags)
- [ ] Proper error handling throughout
- [ ] No race conditions (verified with race detector)
- [ ] Code formatted with `go fmt`
- [ ] Examples of running commands in README

---

**Good luck! We look forward to reviewing your submission.**