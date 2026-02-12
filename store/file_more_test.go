package store

import (
	"aexp_assesment/domain"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestFileStore_List_BulkImport_LoadFromFile(t *testing.T) {
	path := filepath.Join(os.TempDir(), "file_store_more_test.json")
	_ = os.Remove(path)
	defer os.Remove(path)

	s, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	products := []domain.Product{
		{ID: "f1", Name: "P1", Price: 10, Quantity: 1, Category: "A"},
		{ID: "f2", Name: "P2", Price: 20, Quantity: 2, Category: "B"},
	}
	// BulkImport should add products
	if err := s.BulkImport(context.Background(), products); err != nil {
		t.Fatalf("BulkImport failed: %v", err)
	}

	// List all
	all, err := s.List(context.Background(), domain.ListFilter{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 products, got %d", len(all))
	}

	// Filter by category
	out, err := s.List(context.Background(), domain.ListFilter{Category: "A"})
	if err != nil {
		t.Fatalf("List with filter failed: %v", err)
	}
	if len(out) != 1 || out[0].ID != "f1" {
		t.Fatalf("unexpected filter result: %+v", out)
	}

	// Save file was written by BulkImport; create a new store to load it
	s2, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore (load) failed: %v", err)
	}
	loaded, err := s2.List(context.Background(), domain.ListFilter{})
	if err != nil {
		t.Fatalf("List after load failed: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 after load, got %d", len(loaded))
	}

	// Test BulkImport duplicate detection: try importing duplicate id
	dupProducts := []domain.Product{{ID: "f1", Name: "P1", Price: 10, Quantity: 1, Category: "A"}}
	err = s.BulkImport(context.Background(), dupProducts)
	if err == nil {
		t.Fatalf("expected error when bulk importing duplicate id")
	}

	// Ensure file contains JSON array
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	var arr []domain.Product
	if err := json.Unmarshal(b, &arr); err != nil {
		t.Fatalf("file content is not JSON array: %v", err)
	}
}
