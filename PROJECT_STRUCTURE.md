# Project Structure for inventory-cli

## Package Organization

### `domain/` - Core Business Logic
- **product.go**: `Product` struct, `ListFilter`, `ProductStore` interface definition
- **error.go**: Custom error types with `errors.Is()` support
  - `ProductNotFoundError`
  - `InvalidProductError`
  - `DuplicateProductError`
- **error_test.go**: Comprehensive error type tests

### `store/` - Storage Implementation
- **factory.go**: `NewStore(kind, path)` factory for dependency injection
- **memory.go**: `InMemoryStore` - thread-safe in-memory storage
  - Uses `sync.RWMutex` for reader/writer locking
  - `BulkImport` with worker pool pattern (max 10 workers)
  - Context cancellation support
- **memory_test.go**: Unit tests, table-driven tests, concurrency tests, benchmarks
- **file.go**: `FileStore` - JSON file-backed persistent storage
  - Atomic writes via temp file + rename
  - Loads from file on initialization
  - Same `ProductStore` interface as InMemoryStore
- **file_test.go**: FileStore CRUD tests

### `cli/` - Command-Line Interface
- **commands.go**: Cobra CLI implementation
  - Root command with persistent flags (`--store`, `--store-file`)
  - 8 subcommands: create, get, list, update, delete, import, export, shell
  - Interactive REPL shell mode
  - JSON import (array, single object, NDJSON formats)
  - JSON export with optional filtering

### `cmd/inventory/` - Application Entry Point
- **main.go**: Simple entry point that calls `cli.Execute()`

### `util/` - Utilities
- **uuid.go**: `GenerateUUID()` for RFC4122 v4 UUID generation

### `data/` - Sample Data
- **products.json**: 6 sample products for testing import functionality

## Code Organization Principles

### Clean Architecture Layers
1. **Domain Layer** (`domain/`): Pure business types without dependencies
2. **Store Layer** (`store/`): Implementations of `ProductStore` interface
3. **CLI Layer** (`cli/`): User interaction via Cobra
4. **Entry Point** (`cmd/inventory/`): Minimal main function

### SOLID Principles Applied
- **Single Responsibility**: Each file has one reason to change
- **Open/Closed**: `ProductStore` interface allows new implementations without modifying existing code
- **Liskov Substitution**: Both `InMemoryStore` and `FileStore` are drop-in replacements
- **Interface Segregation**: `ProductStore` interface defines minimal required methods
- **Dependency Inversion**: CLI depends on `ProductStore` interface, not concrete types

### Design Patterns
- **Factory Pattern**: `store.NewStore(kind, path)` for flexible store creation
- **Strategy Pattern**: Pluggable store implementations via `ProductStore` interface
- **Worker Pool Pattern**: `BulkImport` uses controlled concurrency with channels
- **Table-Driven Testing**: Reusable test cases for validation logic

## Building & Running

```bash
# Build to ./cmd/inventory
go build -o inventory-cli.exe ./cmd/inventory

# Run using Makefile
make build
make run
make test
make bench

# Direct execution
go run ./cmd/inventory create --name "Product" --price 99.99
go run ./cmd/inventory shell
```

## Testing Coverage

- **Domain Tests** (`domain/error_test.go`): Error type validation, `errors.Is/As()` support
- **Store Tests** (`store/memory_test.go`): CRUD, validation, filtering, sorting, concurrent access, timeouts
- **File Store Tests** (`store/file_test.go`): Persistence, JSON round-trip
- **Benchmarks**: Create, Get operations on InMemoryStore
- **Concurrency Tests**: 100+ goroutines accessing store simultaneously
- **Context Tests**: Timeout and cancellation handling

## Import Structure

All packages import using module path `aexp_assesment/`:
```go
import (
    "aexp_assesment/domain"
    "aexp_assesment/store"
    "aexp_assesment/util"
)
```

## Future Enhancements

- REST/gRPC API server
- Database backends (PostgreSQL, SQLite)
- Structured logging (log/slog)
- Configuration management (Viper)
- Docker deployment
- Multi-user authentication

