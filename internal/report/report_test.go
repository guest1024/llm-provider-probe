package report

import (
	"strings"
	"testing"
	"time"
)

func TestHTMLRendersProviderSummary(t *testing.T) {
	html, err := HTML(RunResult{
		StartedAt:   time.Unix(0, 0),
		CompletedAt: time.Unix(1, 0),
		Providers: []ProviderResult{{
			Name:    "demo",
			BaseURL: "https://example.com/v1",
			Model:   "demo-model",
			Summary: ProviderSummary{Score: 88.5, Suspicion: "medium", PassedRuns: 7, TotalRuns: 8},
		}},
	})
	if err != nil {
		t.Fatalf("HTML render failed: %v", err)
	}
	if !strings.Contains(html, "demo") || !strings.Contains(html, "88.5") {
		t.Fatalf("unexpected html output: %s", html)
	}
}
