package runner

import (
	"testing"

	"model-codex/internal/report"
)

func TestSummarizeProviderFlagsEmptyChoicesError(t *testing.T) {
	summary := summarizeProvider([]report.CaseRunResult{
		{
			CaseName: "exact_line",
			Error:    "provider returned 200 but no choices payload",
		},
		{
			CaseName: "exact_line",
			Passed:   true,
			Score:    1,
		},
	})

	if summary.ErrorRuns != 1 {
		t.Fatalf("expected 1 error run, got %d", summary.ErrorRuns)
	}
	if summary.Suspicion != "medium" {
		t.Fatalf("expected medium suspicion, got %s", summary.Suspicion)
	}
	found := false
	for _, warning := range summary.Warnings {
		if warning == "provider sometimes returns HTTP 200 with empty choices payload" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected empty choices warning, got %#v", summary.Warnings)
	}
}

func TestSummarizeProviderFlagsWeakBenchmark(t *testing.T) {
	summary := summarizeProvider([]report.CaseRunResult{
		{CaseName: "commonsenseqa-starter", Benchmark: "commonsenseqa", Passed: false, Score: 0},
		{CaseName: "commonsenseqa-starter", Benchmark: "commonsenseqa", Passed: true, Score: 1},
		{CaseName: "commonsenseqa-starter", Benchmark: "commonsenseqa", Passed: false, Score: 0},
	})
	found := false
	for _, warning := range summary.Warnings {
		if warning == "benchmark commonsenseqa is weak (pass rate 33%, starter band=weak)" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected benchmark weakness warning, got %#v", summary.Warnings)
	}
}

func TestSummarizeProviderDoesNotFlagAcceptableStarterBandAsWeak(t *testing.T) {
	summary := summarizeProvider([]report.CaseRunResult{
		{CaseName: "gpqa-starter", Benchmark: "gpqa", Passed: true, Score: 1},
		{CaseName: "gpqa-starter", Benchmark: "gpqa", Passed: false, Score: 0},
	})
	for _, warning := range summary.Warnings {
		if warning == "benchmark gpqa is weak (pass rate 50%, starter band=acceptable)" {
			t.Fatalf("did not expect weak warning for acceptable starter band: %#v", summary.Warnings)
		}
	}
}
