package util

import (
	"regexp"
	"testing"
)

func TestGenerateUUID_Format(t *testing.T) {
	u := GenerateUUID()
	if u == "" {
		t.Fatal("expected non-empty UUID")
	}
	// simple regex for UUID v4 format
	r := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !r.MatchString(u) {
		t.Fatalf("UUID %s does not match v4 format", u)
	}
}
