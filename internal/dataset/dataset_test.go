package dataset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "samples.jsonl")
	payload := "{\"id\":\"s1\",\"prompt\":\"hello\",\"evaluator\":\"exact_match\",\"expected\":\"world\"}\n{\"prompt\":\"next\",\"evaluator\":\"regex_match\",\"regex\":\"ok\"}\n"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := LoadJSONL(path, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[1].ID == "" {
		t.Fatal("expected auto-generated id")
	}
}
