package config

import "testing"

func TestDefaultValidate(t *testing.T) {
	cfg := Default()
	cfg.Providers = []ProviderConfig{{Name: "demo", BaseURL: "https://api.example.com/v1", Model: "demo-model"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidateAllowsEnvResolvedBaseURLAndModel(t *testing.T) {
	t.Setenv("PROBE_BASE_URL", "https://example.com/v1")
	t.Setenv("PROBE_MODEL", "demo-model")
	cfg := Default()
	cfg.Providers = []ProviderConfig{{Name: "demo", BaseURLEnv: "PROBE_BASE_URL", ModelEnv: "PROBE_MODEL"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected env-backed config to validate, got %v", err)
	}
}

func TestResolvedFieldsPreferLiteralThenEnv(t *testing.T) {
	t.Setenv("PROBE_BASE_URL", "https://env.example.com/v1")
	t.Setenv("PROBE_MODEL", "env-model")
	t.Setenv("PROBE_KEY", "env-key")
	p := ProviderConfig{
		BaseURL:    "https://literal.example.com/v1",
		BaseURLEnv: "PROBE_BASE_URL",
		ModelEnv:   "PROBE_MODEL",
		APIKeyEnv:  "PROBE_KEY",
	}
	if got := p.ResolvedBaseURL(); got != "https://literal.example.com/v1" {
		t.Fatalf("unexpected base url: %s", got)
	}
	if got := p.ResolvedModel(); got != "env-model" {
		t.Fatalf("unexpected model: %s", got)
	}
	if got := p.ResolvedAPIKey(); got != "env-key" {
		t.Fatalf("unexpected api key: %s", got)
	}
}
