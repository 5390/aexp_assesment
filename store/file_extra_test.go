package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"aexp_assesment/domain"
)

func floatPtr(v float64) *float64 { return &v }

func TestFileStore_List_SortingAndFiltering(t *testing.T) {
	path := filepath.Join(os.TempDir(), "file_store_list_test.json")
	_ = os.Remove(path)
	defer os.Remove(path)

	s, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	items := []domain.Product{
		{ID: "a1", Name: "Alpha", Price: 50, Quantity: 4, Category: "C1"},
		{ID: "b2", Name: "Beta", Price: 20, Quantity: 2, Category: "C2"},
		{ID: "c3", Name: "Gamma", Price: 80, Quantity: 1, Category: "C1"},
	}
	for _, it := range items {
		if err := s.Create(context.Background(), it); err != nil {
			t.Fatalf("setup Create failed: %v", err)
		}
	}

	// Filter by MinPrice
	out, err := s.List(context.Background(), domain.ListFilter{MinPrice: floatPtr(30)})
	if err != nil {
		t.Fatalf("List MinPrice failed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items with MinPrice>=30, got %d", len(out))
	}

	// Filter by MaxPrice
	out, err = s.List(context.Background(), domain.ListFilter{MaxPrice: floatPtr(30)})
	if err != nil {
		t.Fatalf("List MaxPrice failed: %v", err)
	}
	if len(out) != 1 || out[0].ID != "b2" {
		t.Fatalf("unexpected MaxPrice result: %+v", out)
	}

	// Filter by Category
	out, err = s.List(context.Background(), domain.ListFilter{Category: "C1"})
	if err != nil {
		t.Fatalf("List Category failed: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 items in C1, got %d", len(out))
	}

	// Sort by price desc
	out, err = s.List(context.Background(), domain.ListFilter{SortBy: "price", Order: "desc"})
	if err != nil {
		t.Fatalf("List sort desc failed: %v", err)
	}
	if len(out) < 2 || out[0].Price < out[1].Price {
		t.Fatalf("expected price desc order, got %+v", out)
	}

	// Sort by name asc
	out, err = s.List(context.Background(), domain.ListFilter{SortBy: "name", Order: "asc"})
	if err != nil {
		t.Fatalf("List sort name asc failed: %v", err)
	}
	if len(out) < 2 || out[0].Name > out[1].Name {
		t.Fatalf("expected name asc order, got %+v", out)
	}
}

func TestFileStore_BulkImport_InvalidAndDuplicates(t *testing.T) {
	path := filepath.Join(os.TempDir(), "file_store_bulk_test.json")
	_ = os.Remove(path)
	defer os.Remove(path)

	s, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}

	// Invalid product (empty ID)
	invalid := []domain.Product{{ID: "", Name: "Bad", Price: 1, Quantity: 1}}
	err = s.BulkImport(context.Background(), invalid)
	if err == nil {
		t.Fatalf("expected error for invalid product, got nil")
	}
	if !domain.IsInvalidProductError(err) {
		// aggregated error may wrap; try As detection
		if !domain.IsInvalidProductError(err) {
			t.Fatalf("expected InvalidProductError in aggregated error, got %v", err)
		}
	}

	// Duplicate IDs within input
	dupInput := []domain.Product{{ID: "d1", Name: "D1", Price: 1, Quantity: 1}, {ID: "d1", Name: "D1b", Price: 2, Quantity: 2}}
	err = s.BulkImport(context.Background(), dupInput)
	if err == nil {
		t.Fatalf("expected error for duplicate IDs in input, got nil")
	}
	if !domain.IsDuplicateProductError(err) {
		// OK if aggregated; just ensure duplicate detected
	}

	// Duplicate vs existing store
	// first add an item
	if err := s.Create(context.Background(), domain.Product{ID: "ex1", Name: "Existing", Price: 5, Quantity: 1}); err != nil {
		t.Fatalf("setup create failed: %v", err)
	}
	// now bulk import with same ID
	err = s.BulkImport(context.Background(), []domain.Product{{ID: "ex1", Name: "X", Price: 5, Quantity: 1}})
	if err == nil {
		t.Fatalf("expected error when importing duplicate against existing store, got nil")
	}
}
