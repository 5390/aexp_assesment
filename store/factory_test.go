package store

import (
	"os"
	"testing"
)

func TestNewStoreFactory_MemoryAndFile(t *testing.T) {
	// memory
	st, err := NewStore("memory", "")
	if err != nil {
		t.Fatalf("NewStore memory failed: %v", err)
	}
	if st == nil {
		t.Fatal("expected non-nil store for memory")
	}

	// file
	path := "testdata/factory_store.json"
	_ = os.Remove(path)
	defer os.Remove(path)
	st2, err := NewStore("file", path)
	if err != nil {
		t.Fatalf("NewStore file failed: %v", err)
	}
	if st2 == nil {
		t.Fatal("expected non-nil store for file")
	}
}
