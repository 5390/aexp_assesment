package cli

import (
	"aexp_assesment/store"
	"io/ioutil"
	"os"
	"testing"
)

// capture error return of Execute for commands expecting failure
func TestPersistentPreRun_FileStoreMissingPath(t *testing.T) {
	productStore = nil
	// attempt to use file store but pass empty path
	rootCmd.PersistentFlags().Set("store", "file")
	rootCmd.PersistentFlags().Set("store-file", "")
	rootCmd.SetArgs([]string{"--store", "file", "--store-file", "", "create", "--name", "X"})
	if err := Execute(); err == nil {
		t.Fatalf("expected error when file store path is empty, got nil")
	}
}

func TestImport_UnsupportedFormat(t *testing.T) {
	productStore = store.NewInMemoryStore()
	rootCmd.PersistentFlags().Set("store", "memory")
	rootCmd.PersistentFlags().Set("store-file", "")
	// create a temp file with invalid JSON
	tmp, err := ioutil.TempFile("", "bad_import_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString("this is not json")
	tmp.Close()

	rootCmd.SetArgs([]string{"import", "--file", tmp.Name()})
	if err := Execute(); err == nil {
		t.Fatalf("expected error for unsupported import format, got nil")
	}
}

func TestImport_NDJSON(t *testing.T) {
	productStore = store.NewInMemoryStore()
	rootCmd.PersistentFlags().Set("store", "memory")
	rootCmd.PersistentFlags().Set("store-file", "")
	// create NDJSON file
	tmp, err := ioutil.TempFile("", "ndjson_import_*.ndjson")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	_, _ = tmp.WriteString("{\"id\":\"n1\",\"name\":\"N1\",\"price\":1,\"quantity\":1}\n")
	_, _ = tmp.WriteString("{\"id\":\"n2\",\"name\":\"N2\",\"price\":2,\"quantity\":2}\n")
	tmp.Close()

	rootCmd.SetArgs([]string{"import", "--file", tmp.Name()})
	if err := Execute(); err != nil {
		t.Fatalf("expected successful NDJSON import, got error: %v", err)
	}

	// list to verify
	rootCmd.SetArgs([]string{"list", "--output", "json"})
	if err := Execute(); err != nil {
		t.Fatalf("list failed after NDJSON import: %v", err)
	}
}

func TestUnknownStoreKind(t *testing.T) {
	productStore = nil
	// leave store flag set to unknown to validate error path
	rootCmd.PersistentFlags().Set("store", "unknown")
	rootCmd.PersistentFlags().Set("store-file", "")
	rootCmd.SetArgs([]string{"--store", "unknown", "create", "--name", "X"})
	if err := Execute(); err == nil {
		t.Fatalf("expected error for unknown store kind, got nil")
	}
}

func TestExport_NoFileFlag(t *testing.T) {
	productStore = store.NewInMemoryStore()
	rootCmd.PersistentFlags().Set("store", "memory")
	rootCmd.PersistentFlags().Set("store-file", "")
	// ensure export subcommand flag is empty (clear any previous test state)
	for _, c := range rootCmd.Commands() {
		if c.Name() == "export" {
			c.Flags().Set("file", "")
			break
		}
	}
	rootCmd.SetArgs([]string{"export"})
	if err := Execute(); err == nil {
		t.Fatalf("expected error when export --file missing, got nil")
	}
}
