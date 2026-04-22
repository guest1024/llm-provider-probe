package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"model-codex/internal/config"
)

func TestDoRejectsEmptyChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"demo","choices":null,"usage":{"prompt_tokens":0,"completion_tokens":0,"total_tokens":0}}`))
	}))
	defer srv.Close()

	client := NewClient(config.ProviderConfig{
		Name:    "demo",
		BaseURL: srv.URL,
		Model:   "demo-model",
	})

	resp, err := client.Do(context.Background(), Request{Model: "demo-model", Messages: []Message{{Role: "user", Content: "hi"}}})
	if err == nil {
		t.Fatalf("expected error, got nil response=%+v", resp)
	}
}

func TestDoParsesToolCalls(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model":"demo",
			"choices":[{"message":{"content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"probe_echo","arguments":"{\"token\":\"PX-77\"}"}}]},"finish_reason":"tool_calls"}],
			"usage":{"prompt_tokens":12,"completion_tokens":3,"total_tokens":15}
		}`))
	}))
	defer srv.Close()

	client := NewClient(config.ProviderConfig{Name: "demo", BaseURL: srv.URL, Model: "demo-model"})
	resp, err := client.Do(context.Background(), Request{
		Model:    "demo-model",
		Messages: []Message{{Role: "user", Content: "hi"}},
		ExtraBody: map[string]any{
			"tools": []map[string]any{{
				"type": "function",
				"function": map[string]any{
					"name":        "probe_echo",
					"description": "Echo",
					"parameters": map[string]any{
						"type":       "object",
						"properties": map[string]any{"token": map[string]any{"type": "string"}},
						"required":   []string{"token"},
					},
				},
			}},
			"tool_choice": map[string]any{"function": map[string]any{"name": "probe_echo"}},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %+v", resp.ToolCalls)
	}
	if resp.ToolCalls[0].Name != "probe_echo" || resp.ToolCalls[0].Arguments != "{\"token\":\"PX-77\"}" {
		t.Fatalf("unexpected tool call: %+v", resp.ToolCalls[0])
	}
	if resp.ReturnedModel != "demo" {
		t.Fatalf("expected returned model demo, got %q", resp.ReturnedModel)
	}
}
