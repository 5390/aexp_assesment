# Inventory CLI (inventory-cli)

## Overview
This project is a small CLI application in Go for managing a product inventory. It demonstrates:

- Idiomatic Go code and clean architecture with organized packages
- Custom error types for clear error handling
- A thread-safe in-memory store and a JSON file-backed store
- Concurrency patterns (worker pool) and context cancellation
- A Cobra-based CLI with common product operations
- Comprehensive tests and benchmarks

## Project Structure

```
aexp_assesment/
├── cmd/inventory/          # Application entry point (main.go)
├── cli/                    # CLI commands & interactive shell (commands.go)
├── domain/                 # Core types & business logic
│   ├── product.go         # Product struct, ListFilter, ProductStore interface
│   ├── error.go           # Custom error types
│   └── error_test.go      # Error type tests
├── store/                 # Store implementations
│   ├── factory.go         # NewStore() factory for DI
│   ├── memory.go          # InMemoryStore (thread-safe with RWMutex)
│   ├── memory_test.go     # InMemoryStore tests
│   ├── file.go            # FileStore (JSON persistence)
│   └── file_test.go       # FileStore tests
├── util/                  # Utilities
│   └── uuid.go            # UUID v4 generation
├── data/products.json     # Sample product data
└── go.mod, go.sum, etc.   # Build configuration & dependencies
```

## Package Organization

- **domain/**: Core business types (`Product`, `ProductStore` interface) and error definitions
- **store/**: Storage implementations (in-memory and file-backed) with factory pattern
- **cli/**: Cobra CLI command handlers and interactive REPL shell
- **util/**: Utility functions (UUID generation)
- **cmd/inventory/**: Application entry point linking everything together

## Business Logic
---------------
Products have the following fields:

- `id` (string, UUID v4)
- `name` (string)
- `price` (float64)
- `quantity` (int)
- `category` (string)

Validation rules:

- `id` must be non-empty when inserting via store constructors (CLI generates ids for `create`)
- `name` must be non-empty
- `price` must be >= 0
- `quantity` must be >= 0

## Errors
---
The project defines custom errors (`ProductNotFoundError`, `InvalidProductError`, `DuplicateProductError`) implemented to work with `errors.Is`/`errors.As`.

## Stores & Dependency Injection
---
There is a `ProductStore` interface with two concrete implementations:

- In-memory: `NewInMemoryStore()` — fast, thread-safe using `sync.RWMutex`.
- File-backed: `NewFileStore(path string)` — persists products in JSON, safe for concurrent writes using mutex and atomic rename.

Use the `NewStore(kind, path)` factory to obtain a `ProductStore` by configuration.

## Concurrency & Bulk Import
---
`BulkImport` uses a worker pool (up to 10 workers) and channels to process products concurrently. It is context-aware and will stop work and return when the provided `context` is cancelled or reaches its deadline. Partial failures are aggregated and returned as a wrapped error.

## CLI (Cobra)

### Build and Run

```bash
# Option 1: Run directly from project root
go run ./cmd/inventory <command> [flags]

# Option 2: Build an executable first
go build -o inventory-cli.exe ./cmd/inventory
./inventory-cli.exe <command>

# Option 3: Run with Make (if make is installed)
make build
make run ARGS="<command> [flags]"
```

Global persistent flags (available before subcommand):

- `--store` — `memory` (default) or `file`
- `--store-file` — path for JSON file store (default `data/products.json`)

## Commands and Usage

### 1) Create

Create a product (CLI generates `id` automatically):

```bash
go run ./cmd/inventory create --name "Laptop" --price 999.99 --quantity 10 --category "Electronics"
```

### 2) Get

Retrieve product by id (prints JSON):

```bash
go run ./cmd/inventory get <product-id>
```

### 3) List

List with optional filters and sorting:

```bash
go run ./cmd/inventory list --category "Electronics" --min-price 100 --sort-by price --order desc
go run ./cmd/inventory list --output json
```

### 4) Update

Partial updates via flags:

```bash
go run ./cmd/inventory update <product-id> --price 899.99 --quantity 15
```

### 5) Delete

Prompted confirmation (use `--force` to skip):

```bash
go run ./cmd/inventory delete <product-id>
go run ./cmd/inventory delete --force <product-id>
```

### 6) Import

Import products from JSON. Supported input formats:
- JSON array of products (standard),
- single JSON object, or
- newline-delimited JSON (NDJSON).

Example (file-backed store):

```bash
go run ./cmd/inventory --store file --store-file data/products.json import --file data/products.json
```

### 7) Export

Export filtered products to a file:

```bash
go run ./cmd/inventory --store file --store-file data/products.json export --file exported.json --category Electronics
```

### 8) Shell (REPL)

Start an interactive prompt to run multiple commands without restarting:

```bash
go run ./cmd/inventory shell
# then inside prompt:
# inventory> create --name "Desk" --price 49.99 --quantity 5 --category Office
# inventory> list
# inventory> exit
# inventory> import --file data/products.json
# inventory> export --file exported.json
```

## Sample Data
---
`data/products.json` is included with sample products. Use it as import source or as the file store location.

## Testing
---
Run unit tests:

```bash
go test ./...
```

Run benchmarks:

```bash
go test -bench . -run '^$'
```

**Race detector note:**
The Go race detector on Windows requires CGO (a C compiler) to be enabled. Options:

- Install a C toolchain and run:

```powershell
$env:CGO_ENABLED = "1"
$env:CC = "gcc"
go test -race ./...
```

- Or run tests from WSL/Linux (recommended on Windows). If you cannot enable CGO locally, run tests without `-race`.

## Design Notes & Trade-offs
---
- The in-memory store uses `sync.RWMutex` for simplicity and good read concurrency.
- The file store stores the entire product list as JSON and atomically writes via a temporary file + rename. This is simple but not optimized for large datasets.
- `BulkImport` demonstrates concurrent processing and context propagation. Errors are collected per-item and aggregated.
- The CLI REPL uses basic whitespace splitting for simplicity; it does not parse quoted arguments. For production, replace with a proper shell parser.

## Next Steps (Optional)
---
- Add a Dockerfile (multi-stage) for building a small runtime image.
- Improve REPL argument parsing to handle quoted strings.
- Add more robust file-store locking for multi-process access (e.g., file lock).
- Add more CLI tests using mocked `ProductStore` implementations.

## Contact
---
If you want me to run tests or produce a Dockerfile, tell me and I'll add them.
