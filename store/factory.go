package store

import (
	"aexp_assesment/domain"
	"fmt"
)

// NewStore constructs a domain.ProductStore by kind: "memory" or "file".
// For file store, provide the file path in path; for memory, path is ignored.
func NewStore(kind, path string) (domain.ProductStore, error) {
	switch kind {
	case "memory", "mem":
		return NewInMemoryStore(), nil
	case "file":
		if path == "" {
			return nil, fmt.Errorf("file path required for file store")
		}
		return NewFileStore(path)
	default:
		return nil, fmt.Errorf("unknown store kind: %s", kind)
	}
}
