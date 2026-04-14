package suite

import (
	"testing"

	"model-codex/internal/config"
	"model-codex/internal/provider"
)

func TestChineseCompactCasePasses(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "chinese_compact", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{Content: built.Expected})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
}

func TestNestedJSONSchemaCasePasses(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "nested_json_schema", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	payload := `{"meta":{"lang":"zh","count":2},"items":[{"id":"a1","ok":true},{"id":"b2","ok":false}]}`
	result := built.Evaluate(provider.Response{Content: payload})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
}

func TestResponseFormatJSONSchemaCasePasses(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "response_format_json_schema", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{Content: `{"verdict":"pass","count":3}`})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
}

func TestToolCallEchoCasePasses(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "tool_call_echo", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{
		ToolCalls: []provider.ToolCall{{
			Name:      "probe_echo",
			Arguments: `{"token":"PX-77"}`,
		}},
	})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
}

func TestGoSnippetOutputCasePasses(t *testing.T) {
	built, err := Build(config.CaseConfig{Name: "go_snippet_output", Enabled: true})
	if err != nil {
		t.Fatalf("build case: %v", err)
	}
	result := built.Evaluate(provider.Response{Content: built.Expected})
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
	if built.Expected != "12" {
		t.Fatalf("expected locked regression value 12, got %q", built.Expected)
	}
}
