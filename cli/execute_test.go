package cli

import (
	"testing"
)

func TestExecuteWrapper(t *testing.T) {
	// set a fresh in-memory store so PersistentPreRunE will no-op
	productStore = nil
	// ensure persistent flags are sane for the test
	rootCmd.PersistentFlags().Set("store", "memory")
	rootCmd.PersistentFlags().Set("store-file", "")
	rootCmd.SetArgs([]string{"create", "--name", "ExecTest"})
	if err := Execute(); err != nil {
		t.Fatalf("Execute wrapper failed: %v", err)
	}
}
