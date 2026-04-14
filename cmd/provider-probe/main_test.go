package main

import (
	"strings"
	"testing"
	"time"

	"model-codex/internal/report"
)

func TestHistoryVerdictHighRisk(t *testing.T) {
	got := historyVerdict(historyAggregate{Runs: 5, ScoreMin: 25, ScoreMax: 90, HighCount: 2, ErrorRuns: 6})
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
	text, items := buildHistorySummary([]report.RunResult{result1, result2}, "demo")
	if len(items) != 1 {
		t.Fatalf("expected 1 aggregate, got %d", len(items))
	}
	if items[0].LatestScore != 90 || items[0].LatestSuspicion != "low" {
		t.Fatalf("unexpected latest aggregate: %+v", items[0])
	}
	if !strings.Contains(text, "Latest score/suspicion") || !strings.Contains(text, "Verdict:") {
		t.Fatalf("summary missing expected fields: %s", text)
	}
}
