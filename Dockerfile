# Multi-stage Dockerfile for inventory-cli
# - Builder: official Go image
# - Runtime: minimal Alpine image with non-root user
# - Uses build cache mounts for faster incremental builds (BuildKit)

# -------- Build stage --------
FROM golang:1.21 AS builder
WORKDIR /src

# Cache Go modules and build cache to speed up rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary for linux/amd64 for checking race condition
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -ldflags='-s -w' -o /out/inventory-cli ./cmd/inventory


# -------- Runtime stage --------
FROM alpine:3.18 AS runtime

# Create non-root user and group for security
RUN addgroup -S app && adduser -S -G app app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /out/inventory-cli /app/inventory-cli
RUN chown app:app /app/inventory-cli

# Run as non-root
USER app

# Environment variables (can be overridden at runtime):
#  - INVENTORY_STORE: storage backend, "memory" or "file" (default: memory)
#  - INVENTORY_STORE_FILE: path to JSON file for file store (default: /data/products.json)
#  - INVENTORY_LOG_LEVEL: logging level (debug|info|warn|error)
ENV INVENTORY_STORE=memory \
    INVENTORY_STORE_FILE=/data/products.json \
    INVENTORY_LOG_LEVEL=info

# Persistent folder for file-backed store
VOLUME ["/data"]

ENTRYPOINT ["/app/inventory-cli"]
