package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Providers []ProviderConfig `json:"providers"`
	Run       RunConfig        `json:"run"`
	Suite     SuiteConfig      `json:"suite"`
}

type ProviderConfig struct {
	Name            string            `json:"name"`
	Type            string            `json:"type,omitempty"`
	BaseURL         string            `json:"base_url,omitempty"`
	BaseURLEnv      string            `json:"base_url_env,omitempty"`
	Endpoint        string            `json:"endpoint,omitempty"`
	Model           string            `json:"model,omitempty"`
	ModelEnv        string            `json:"model_env,omitempty"`
	APIKey          string            `json:"api_key,omitempty"`
	APIKeyEnv       string            `json:"api_key_env,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	TimeoutSeconds  int               `json:"timeout_seconds,omitempty"`
	ExtraBody       map[string]any    `json:"extra_body,omitempty"`
	ReasoningEffort string            `json:"reasoning_effort,omitempty"`
}

type RunConfig struct {
	Repeats        int     `json:"repeats"`
	Output         string  `json:"output,omitempty"`
	Temperature    float64 `json:"temperature"`
	CaptureHeaders bool    `json:"capture_headers"`
}

type SuiteConfig struct {
	Cases []CaseConfig `json:"cases"`
}

type DatasetConfig struct {
	Path    string `json:"path"`
	Name    string `json:"name,omitempty"`
	Split   string `json:"split,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Shuffle bool   `json:"shuffle,omitempty"`
}

type CaseConfig struct {
	Name    string         `json:"name"`
	Enabled bool           `json:"enabled"`
	Params  map[string]any `json:"params,omitempty"`
	Dataset *DatasetConfig `json:"dataset,omitempty"`
}

func Default() Config {
	return Config{
		Run: RunConfig{
			Repeats:        3,
			Temperature:    0,
			CaptureHeaders: true,
		},
		Suite: SuiteConfig{Cases: []CaseConfig{
			{Name: "exact_json", Enabled: true},
			{Name: "exact_line", Enabled: true},
			{Name: "logic_filter", Enabled: true},
			{Name: "chinese_compact", Enabled: true},
			{Name: "nested_json_schema", Enabled: true},
			{Name: "response_format_json_schema", Enabled: false},
			{Name: "go_snippet_output", Enabled: true},
			{Name: "tool_call_echo", Enabled: false},
			{Name: "long_context_needle_small", Enabled: true, Params: map[string]any{"approx_tokens": 4000}},
			{Name: "long_context_needle_medium", Enabled: true, Params: map[string]any{"approx_tokens": 12000}},
		}},
	}
}

func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse json config: %w", err)
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if len(c.Providers) == 0 {
		return fmt.Errorf("at least one provider is required")
	}
	for _, provider := range c.Providers {
		if strings.TrimSpace(provider.Name) == "" {
			return fmt.Errorf("provider name is required")
		}
		if strings.TrimSpace(provider.ResolvedBaseURL()) == "" {
			return fmt.Errorf("provider %s: base_url or base_url_env is required", provider.Name)
		}
		if strings.TrimSpace(provider.ResolvedModel()) == "" {
			return fmt.Errorf("provider %s: model or model_env is required", provider.Name)
		}
		if provider.Type != "" && provider.Type != "eino_openai" {
			return fmt.Errorf("provider %s: unsupported type %q", provider.Name, provider.Type)
		}
	}
	if c.Run.Repeats <= 0 {
		return fmt.Errorf("run.repeats must be > 0")
	}
	if len(c.Suite.Cases) == 0 {
		return fmt.Errorf("suite.cases cannot be empty")
	}
	for _, item := range c.Suite.Cases {
		if item.Dataset != nil && strings.TrimSpace(item.Dataset.Path) == "" {
			return fmt.Errorf("case %s: dataset.path is required when dataset is configured", item.Name)
		}
	}
	return nil
}

func (p ProviderConfig) ResolvedEndpoint() string {
	if strings.TrimSpace(p.Endpoint) != "" {
		return p.Endpoint
	}
	return "/chat/completions"
}

func (p ProviderConfig) ResolvedAPIKey() string {
	if p.APIKey != "" {
		return p.APIKey
	}
	if p.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(p.APIKeyEnv)
}

func (p ProviderConfig) ResolvedBaseURL() string {
	if strings.TrimSpace(p.BaseURL) != "" {
		return p.BaseURL
	}
	if p.BaseURLEnv == "" {
		return ""
	}
	return os.Getenv(p.BaseURLEnv)
}

func (p ProviderConfig) ResolvedModel() string {
	if strings.TrimSpace(p.Model) != "" {
		return p.Model
	}
	if p.ModelEnv == "" {
		return ""
	}
	return os.Getenv(p.ModelEnv)
}

func (p ProviderConfig) ResolvedType() string {
	if strings.TrimSpace(p.Type) == "" {
		return "eino_openai"
	}
	return p.Type
}
