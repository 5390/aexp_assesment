package cli

import (
	"aexp_assesment/domain"
	"aexp_assesment/store"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// helper to capture stdout during command execution
func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = old }()

	err := f()
	w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), err
}

func TestCreateGetListUpdateDelete(t *testing.T) {
	// use a fresh in-memory store
	productStore = store.NewInMemoryStore()

	// create
	out, err := captureOutput(func() error {
		rootCmd.SetArgs([]string{"create", "--name", "TestProd", "--price", "5.5", "--quantity", "2", "--category", "T"})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var p domain.Product
	if err := json.Unmarshal([]byte(out), &p); err != nil {
		t.Fatalf("failed to parse create output: %v", err)
	}
	if p.Name != "TestProd" {
		t.Fatalf("unexpected product name: %s", p.Name)
	}

	// get
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"get", p.ID})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	var gp domain.Product
	if err := json.Unmarshal([]byte(out), &gp); err != nil {
		t.Fatalf("failed to parse get output: %v", err)
	}
	if gp.ID != p.ID {
		t.Fatalf("get returned wrong product id")
	}

	// list
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"list"})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if out == "" {
		t.Fatalf("list returned empty output")
	}

	// update
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"update", p.ID, "--price", "7.75"})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	var up domain.Product
	if err := json.Unmarshal([]byte(out), &up); err != nil {
		t.Fatalf("failed to parse update output: %v", err)
	}
	if up.Price != 7.75 {
		t.Fatalf("update did not change price")
	}

	// delete (force)
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"delete", "--force", p.ID})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// ensure not found after delete
	_, err = productStore.Get(context.Background(), p.ID)
	if err == nil {
		t.Fatalf("expected product to be deleted")
	}
}

func TestImportExport(t *testing.T) {
	productStore = store.NewInMemoryStore()
	// create temporary import file
	tmp := filepath.Join(os.TempDir(), "cli_test_import.json")
	plist := []domain.Product{
		{ID: "imp-1", Name: "I1", Price: 1, Quantity: 1, Category: "C"},
		{ID: "imp-2", Name: "I2", Price: 2, Quantity: 2, Category: "C"},
	}
	b, _ := json.MarshalIndent(plist, "", "  ")
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		t.Fatalf("failed to write tmp import file: %v", err)
	}
	defer os.Remove(tmp)

	// import
	_, err := captureOutput(func() error {
		rootCmd.SetArgs([]string{"import", "--file", tmp})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// list and ensure items present
	out, err := captureOutput(func() error {
		rootCmd.SetArgs([]string{"list", "--output", "json"})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("list failed after import: %v", err)
	}
	var listed []domain.Product
	if err := json.Unmarshal([]byte(out), &listed); err != nil {
		t.Fatalf("failed to parse list json: %v", err)
	}
	if len(listed) < 2 {
		t.Fatalf("expected at least 2 imported products, got %d", len(listed))
	}

	// export
	expf := filepath.Join(os.TempDir(), "cli_test_export.json")
	defer os.Remove(expf)
	_, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"export", "--file", expf})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("export failed: %v", err)
	}
	if _, err := os.Stat(expf); err != nil {
		t.Fatalf("export file not created: %v", err)
	}
}
