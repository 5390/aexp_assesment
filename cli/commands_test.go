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
