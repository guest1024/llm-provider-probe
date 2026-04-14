package suite

import (
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
