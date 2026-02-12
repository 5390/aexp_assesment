package cli

import (
	"aexp_assesment/domain"
	"aexp_assesment/store"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"
)

// capture stdout during cobra execution
func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), err
}

// reset cobra + global state between tests
func resetCLI() {
	rootCmd.SetArgs(nil)
	productStore = nil
}

func TestCreateGetListUpdateDelete(t *testing.T) {
	defer resetCLI()
	productStore = store.NewInMemoryStore()

	// CREATE
	out, err := captureOutput(func() error {
		rootCmd.SetArgs([]string{
			"create",
			"--name", "TestProd",
			"--price", "5.5",
			"--quantity", "2",
			"--category", "T",
		})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var created domain.Product
	if err := json.Unmarshal([]byte(out), &created); err != nil {
		t.Fatalf("invalid create output: %v", err)
	}

	// GET
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"get", created.ID})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}

	// LIST
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"list"})
		return rootCmd.Execute()
	})
	if err != nil || out == "" {
		t.Fatalf("list failed")
	}

	// UPDATE
	out, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{
			"update", created.ID,
			"--price", "7.75",
		})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}

	var updated domain.Product
	_ = json.Unmarshal([]byte(out), &updated)
	if updated.Price != 7.75 {
		t.Fatalf("price not updated")
	}

	// DELETE
	_, err = captureOutput(func() error {
		rootCmd.SetArgs([]string{"delete", "--force", created.ID})
		return rootCmd.Execute()
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = productStore.Get(context.Background(), created.ID)
	if err == nil {
		t.Fatalf("expected product to be deleted")
	}
}

// func TestCreateValidationError(t *testing.T) {
// 	defer resetCLI()
// 	productStore = store.NewInMemoryStore()

// 	_, err := captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"create"})
// 		return rootCmd.Execute()
// 	})

// 	if err == nil {
// 		t.Fatalf("expected validation error")
// 	}
// }

// func TestListJSONOutput(t *testing.T) {
// 	defer resetCLI()
// 	productStore = store.NewInMemoryStore()

// 	_ = productStore.Create(context.Background(), domain.Product{
// 		ID: "p1", Name: "Book", Price: 10, Quantity: 1, Category: "Edu",
// 	})

// 	out, err := captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"list", "--output", "json"})
// 		return rootCmd.Execute()
// 	})
// 	if err != nil {
// 		t.Fatalf("list failed: %v", err)
// 	}

// 	var products []domain.Product
// 	if err := json.Unmarshal([]byte(out), &products); err != nil {
// 		t.Fatalf("invalid json list")
// 	}
// }

// func TestImportExport(t *testing.T) {
// 	defer resetCLI()
// 	productStore = store.NewInMemoryStore()

// 	importFile := filepath.Join(os.TempDir(), "cli_import.json")
// 	exportFile := filepath.Join(os.TempDir(), "cli_export.json")
// 	defer os.Remove(importFile)
// 	defer os.Remove(exportFile)

// 	products := []domain.Product{
// 		{ID: "i1", Name: "Item1", Price: 1, Quantity: 1, Category: "C"},
// 		{ID: "i2", Name: "Item2", Price: 2, Quantity: 2, Category: "C"},
// 	}

// 	b, _ := json.Marshal(products)
// 	_ = os.WriteFile(importFile, b, 0644)

// 	// IMPORT
// 	_, err := captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"import", "--file", importFile})
// 		return rootCmd.Execute()
// 	})
// 	if err != nil {
// 		t.Fatalf("import failed: %v", err)
// 	}

// 	// EXPORT
// 	_, err = captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"export", "--file", exportFile})
// 		return rootCmd.Execute()
// 	})
// 	if err != nil {
// 		t.Fatalf("export failed: %v", err)
// 	}

// 	if _, err := os.Stat(exportFile); err != nil {
// 		t.Fatalf("export file not created")
// 	}
// }

// func TestGetNotFound(t *testing.T) {
// 	defer resetCLI()
// 	productStore = store.NewInMemoryStore()

// 	_, err := captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"get", "unknown"})
// 		return rootCmd.Execute()
// 	})

// 	if err != nil {
// 		t.Fatalf("get should not return error on not found")
// 	}
// }

// func TestDeleteWithoutForceAbort(t *testing.T) {
// 	defer resetCLI()
// 	productStore = store.NewInMemoryStore()

// 	_ = productStore.Create(context.Background(), domain.Product{
// 		ID: "p3", Name: "Mouse", Price: 20, Quantity: 10, Category: "IT",
// 	})

// 	old := os.Stdin
// 	r, w, _ := os.Pipe()
// 	os.Stdin = r
// 	defer func() { os.Stdin = old }()

// 	// user presses Enter → abort
// 	w.WriteString("\n")
// 	w.Close()

// 	_, err := captureOutput(func() error {
// 		rootCmd.SetArgs([]string{"delete", "p3"})
// 		return rootCmd.Execute()
// 	})
// 	if err != nil {
// 		t.Fatalf("delete abort should not error")
// 	}

// 	// product should still exist
// 	if _, err := productStore.Get(context.Background(), "p3"); err != nil {
// 		t.Fatalf("product should not be deleted")
// 	}
// }
