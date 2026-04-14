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
