package config

import "testing"

func TestDefaultValidate(t *testing.T) {
	cfg := Default()
	cfg.Providers = []ProviderConfig{{Name: "demo", BaseURL: "https://api.example.com/v1", Model: "demo-model"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}
