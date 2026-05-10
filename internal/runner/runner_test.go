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
	}, nil)

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
	}, nil)
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
	}, nil)
	for _, warning := range summary.Warnings {
		if warning == "benchmark gpqa is weak (pass rate 50%, starter band=acceptable)" {
			t.Fatalf("did not expect weak warning for acceptable starter band: %#v", summary.Warnings)
		}
	}
}

func TestSummarizeProviderWatermarkDetection(t *testing.T) {
	// pass rate = 4/10 = 40%; reference = 0.80; threshold = 0.80*0.8 = 64% → suspected
	runs := make([]report.CaseRunResult, 10)
	for i := range runs {
		runs[i] = report.CaseRunResult{CaseName: "mmlu-pro-starter", Benchmark: "mmlu_pro", Passed: i < 4, Score: map[bool]float64{true: 1, false: 0}[i < 4]}
	}
	refScores := map[string]float64{"mmlu_pro": 0.80}
	summary := summarizeProvider(runs, refScores)

	if summary.Suspicion != "high" {
		t.Fatalf("expected high suspicion when watermark detected, got %s", summary.Suspicion)
	}
	foundWatermark := false
	for _, bs := range summary.BenchmarkSummaries {
		if bs.Benchmark == "mmlu_pro" && bs.WatermarkSuspected {
			foundWatermark = true
		}
	}
	if !foundWatermark {
		t.Fatal("expected WatermarkSuspected=true on mmlu_pro benchmark summary")
	}
}

func TestSummarizeProviderNoWatermarkWhenAboveThreshold(t *testing.T) {
	// pass rate = 7/10 = 70%; reference = 0.80; threshold = 64% → not suspected
	runs := make([]report.CaseRunResult, 10)
	for i := range runs {
		runs[i] = report.CaseRunResult{CaseName: "mmlu-pro-starter", Benchmark: "mmlu_pro", Passed: i < 7, Score: map[bool]float64{true: 1, false: 0}[i < 7]}
	}
	refScores := map[string]float64{"mmlu_pro": 0.80}
	summary := summarizeProvider(runs, refScores)

	for _, bs := range summary.BenchmarkSummaries {
		if bs.Benchmark == "mmlu_pro" && bs.WatermarkSuspected {
			t.Fatal("did not expect WatermarkSuspected when pass rate is above threshold")
		}
	}
}
