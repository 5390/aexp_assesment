package store

import (
	"aexp_assesment/domain"
	"context"
	"os"
	"testing"
)

func TestFileStore_CreateGetUpdateDelete(t *testing.T) {
	path := "testdata/store_test.json"
	_ = os.Remove(path)
	s, err := NewFileStore(path)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	ctx := context.Background()

	p := domain.Product{ID: "f1", Name: "FileProd", Price: 3.14, Quantity: 2, Category: "F"}
	if err := s.Create(ctx, p); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	got, err := s.Get(ctx, "f1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.Name != p.Name {
		t.Fatalf("unexpected name")
	}

	if err := s.Update(ctx, "f1", domain.Product{Name: "FileProd2", Price: 4, Quantity: 1}); err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if err := s.Delete(ctx, "f1"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	_ = os.Remove(path)
}
