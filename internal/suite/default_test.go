package suite

import (
	"os"
	"path/filepath"
	"testing"

	"model-codex/internal/config"
	"model-codex/internal/provider"
)

func TestLongContextCasePassesOnExactNeedle(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "long_context_needle_small", Enabled: true, Params: map[string]any{"approx_tokens": 200}})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{Content: built.Expected})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
}

func TestExactJSONRejectsInvalidJSON(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "exact_json", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{Content: "not json"})
	if result.Passed {
		t.Fatalf("expected fail, got %+v", result)
	}
}

func TestBuildManyDatasetCase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mini.jsonl")
	payload := "{\"id\":\"s1\",\"benchmark\":\"commonsenseqa\",\"prompt\":\"2+2?\",\"evaluator\":\"exact_match\",\"expected\":\"4\"}\n{\"id\":\"s2\",\"benchmark\":\"commonsenseqa\",\"prompt\":\"3+3?\",\"evaluator\":\"exact_match\",\"expected\":\"6\"}\n"
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := BuildMany(config.CaseConfig{Name: "math-mini", Enabled: true, Dataset: &config.DatasetConfig{Path: path}})
	if err != nil {
		t.Fatalf("BuildMany failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].SampleID != "s1" || items[0].Benchmark != "commonsenseqa" {
		t.Fatalf("unexpected metadata: %+v", items[0])
	}
}
