package main

import (
	"strings"
	"testing"
	"time"

	"model-codex/internal/report"
)

func TestHistoryVerdictHighRisk(t *testing.T) {
	got := historyVerdict(historyAggregate{Reports: 5, ScoreMin: 25, ScoreMax: 90, HighCount: 2, ErrorRuns: 6})
	if !strings.Contains(got, "高波动/高风险") {
		t.Fatalf("unexpected verdict: %s", got)
	}
}

func TestBuildHistorySummaryIncludesLatestAndVerdict(t *testing.T) {
	result1 := report.RunResult{
		CompletedAt: time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC),
		Providers:   []report.ProviderResult{{Name: "demo", Summary: report.ProviderSummary{Score: 50, Suspicion: "high", ErrorRuns: 2}}},
	}
	result2 := report.RunResult{
		CompletedAt: time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC),
		Providers:   []report.ProviderResult{{Name: "demo", Summary: report.ProviderSummary{Score: 90, Suspicion: "low", ErrorRuns: 0}}},
	}
	text, items, err := buildHistorySummary([]report.RunResult{result1, result2}, "demo", "", "provider")
	if err != nil {
		t.Fatalf("buildHistorySummary failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 aggregate, got %d", len(items))
	}
	if items[0].LatestScore != 90 || items[0].LatestSuspicion != "low" {
		t.Fatalf("unexpected latest aggregate: %+v", items[0])
	}
	if !strings.Contains(text, "Latest score/pass rate/suspicion") || !strings.Contains(text, "Verdict:") {
		t.Fatalf("summary missing expected fields: %s", text)
	}
}

func TestBuildHistorySummarySupportsBenchmarkGroup(t *testing.T) {
	result := report.RunResult{
		CompletedAt: time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC),
		Providers: []report.ProviderResult{{
			Name: "demo",
			Runs: []report.CaseRunResult{
				{Benchmark: "webqa", Passed: true, Score: 1},
				{Benchmark: "webqa", Passed: false, Score: 0},
			},
		}},
	}
	text, items, err := buildHistorySummary([]report.RunResult{result}, "", "", "benchmark")
	if err != nil {
		t.Fatalf("buildHistorySummary failed: %v", err)
	}
	if len(items) != 1 || items[0].Name != "webqa" {
		t.Fatalf("unexpected benchmark aggregates: %+v", items)
	}
	if !strings.Contains(text, "Group by: benchmark") || !strings.Contains(text, "Pass rate avg/latest") {
		t.Fatalf("unexpected benchmark text: %s", text)
	}
}

func TestBenchmarkSuspicionUsesStarterBand(t *testing.T) {
	if got := benchmarkSuspicion(report.BenchmarkSummary{Benchmark: "gpqa", PassRate: 0.5, AvgScore: 0.5}); got != "medium" {
		t.Fatalf("expected medium for acceptable gpqa band, got %s", got)
	}
	if got := benchmarkSuspicion(report.BenchmarkSummary{Benchmark: "gpqa", PassRate: 0.2, AvgScore: 0.2}); got != "high" {
		t.Fatalf("expected high for weak gpqa band, got %s", got)
	}
}

func TestHistoryVerdictForBenchmarkUsesCoarseSignals(t *testing.T) {
	got := historyVerdict(historyAggregate{
		Benchmark:       "gpqa",
		Reports:         1,
		LatestSuspicion: "medium",
		MediumCount:     1,
		LatestPassRate:  0.5,
		ScoreMin:        50,
		ScoreMax:        50,
	})
	if !strings.Contains(got, "当前较稳定") {
		t.Fatalf("unexpected benchmark verdict: %s", got)
	}
}
