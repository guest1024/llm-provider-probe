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
			Runs: []CaseRunResult{{
				CaseName:  "commonsenseqa-starter",
				Benchmark: "commonsenseqa",
			}},
		}},
	})
	if err != nil {
		t.Fatalf("HTML render failed: %v", err)
	}
	if !strings.Contains(html, "demo") || !strings.Contains(html, "88.5") || !strings.Contains(html, "commonsenseqa") || !strings.Contains(html, "strong") {
		t.Fatalf("unexpected html output: %s", html)
	}
}

func TestSummarizeBenchmarksIncludesStarterBand(t *testing.T) {
	rows := SummarizeBenchmarks([]CaseRunResult{
		{Benchmark: "webqa", Split: "starter", Passed: true, Score: 1, LatencyMs: 100},
		{Benchmark: "webqa", Split: "starter", Passed: true, Score: 1, LatencyMs: 200},
		{Benchmark: "webqa", Split: "starter", Passed: false, Score: 0, LatencyMs: 300},
	})
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].StarterBaselineBand != "acceptable" {
		t.Fatalf("unexpected starter band: %+v", rows[0])
	}
	if rows[0].PassRate != 2.0/3.0 {
		t.Fatalf("unexpected pass rate: %+v", rows[0])
	}
}
