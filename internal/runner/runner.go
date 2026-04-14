package runner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"model-codex/internal/config"
	"model-codex/internal/provider"
	"model-codex/internal/report"
	"model-codex/internal/suite"
)

type RunResult = report.RunResult

func Run(ctx context.Context, cfg config.Config) (report.RunResult, error) {
	started := time.Now()
	result := report.RunResult{StartedAt: started}

	for _, providerCfg := range cfg.Providers {
		providerResult, err := runProvider(ctx, providerCfg, cfg)
		if err != nil {
			return report.RunResult{}, err
		}
		result.Providers = append(result.Providers, providerResult)
	}

	attachCrossProviderSignals(&result)
	result.CompletedAt = time.Now()
	return result, nil
}

func runProvider(ctx context.Context, providerCfg config.ProviderConfig, cfg config.Config) (report.ProviderResult, error) {
	client := provider.NewClient(providerCfg)
	out := report.ProviderResult{Name: providerCfg.Name, BaseURL: providerCfg.BaseURL, Model: providerCfg.Model}

	for _, caseCfg := range cfg.Suite.Cases {
		if !caseCfg.Enabled {
			continue
		}
		built, err := suite.Build(caseCfg)
		if err != nil {
			return report.ProviderResult{}, err
		}

		for attempt := 1; attempt <= cfg.Run.Repeats; attempt++ {
			req := provider.Request{Model: providerCfg.Model, Messages: built.Messages, Temperature: cfg.Run.Temperature, ExtraBody: built.ExtraBody}
			started := time.Now()
			resp, err := client.Do(ctx, req)
			latency := time.Since(started)

			run := report.CaseRunResult{
				CaseName:  built.Name,
				Category:  built.Category,
				Attempt:   attempt,
				Expected:  built.Expected,
				LatencyMs: latency.Milliseconds(),
			}
			if cfg.Run.CaptureHeaders {
				run.ResponseHeaders = resp.Headers
			}
			run.RawResponseSnippet = truncate(string(resp.RawBody), 400)
			run.StatusCode = resp.StatusCode
			run.ReturnedModel = resp.ReturnedModel
			run.FinishReason = resp.FinishReason
			run.PromptTokens = resp.PromptTokens
			run.CompletionTokens = resp.CompletionTokens
			run.TotalTokens = resp.TotalTokens

			if err != nil {
				run.Error = err.Error()
			} else {
				eval := built.Evaluate(resp)
				run.Passed = eval.Passed
				run.Score = eval.Score
				run.Expected = eval.Expected
				run.Actual = eval.Actual
				run.Warning = eval.Warning
			}
			out.Runs = append(out.Runs, run)
		}
	}

	out.Summary = summarizeProvider(out.Runs)
	return out, nil
}

func summarizeProvider(runs []report.CaseRunResult) report.ProviderSummary {
	summary := report.ProviderSummary{TotalRuns: len(runs)}
	if len(runs) == 0 {
		summary.Suspicion = "unknown"
		summary.Warnings = []string{"no runs executed"}
		return summary
	}

	var totalScore float64
	modelSet := map[string]struct{}{}
	caseStats := map[string]struct{ total, passed int }{}

	for _, run := range runs {
		totalScore += run.Score
		if run.Passed {
			summary.PassedRuns++
		}
		if run.Error != "" {
			summary.ErrorRuns++
		}
		if run.ReturnedModel != "" {
			modelSet[run.ReturnedModel] = struct{}{}
		}
		stat := caseStats[run.CaseName]
		stat.total++
		if run.Passed {
			stat.passed++
		}
		caseStats[run.CaseName] = stat
	}

	summary.Score = totalScore / float64(len(runs)) * 100
	summary.PassRate = float64(summary.PassedRuns) / float64(len(runs))
	summary.DistinctReturnedModels = sortedKeys(modelSet)

	if len(summary.DistinctReturnedModels) > 1 {
		summary.Warnings = append(summary.Warnings, fmt.Sprintf("returned model IDs are unstable: %s", strings.Join(summary.DistinctReturnedModels, ", ")))
	}
	if summary.ErrorRuns > 0 {
		summary.Warnings = append(summary.Warnings, fmt.Sprintf("provider returned %d errored runs", summary.ErrorRuns))
	}
	if hasErrorSubstring(runs, "no choices payload") {
		summary.Warnings = append(summary.Warnings, "provider sometimes returns HTTP 200 with empty choices payload")
	}
	for name, stat := range caseStats {
		rate := float64(stat.passed) / float64(stat.total)
		switch {
		case strings.Contains(name, "long_context") && rate < 0.67:
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("long-context retrieval is weak on %s (pass rate %.0f%%)", name, rate*100))
		case name == "exact_json" && rate < 0.67:
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("JSON compliance is weak on %s (pass rate %.0f%%)", name, rate*100))
		case name == "logic_filter" && rate < 0.67:
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("reasoning consistency is weak on %s (pass rate %.0f%%)", name, rate*100))
		}
	}

	switch {
	case summary.ErrorRuns >= 2 || len(summary.Warnings) >= 3 || summary.Score < 50:
		summary.Suspicion = "high"
	case summary.ErrorRuns >= 1 || len(summary.Warnings) >= 1 || summary.Score < 75:
		summary.Suspicion = "medium"
	default:
		summary.Suspicion = "low"
	}
	return summary
}

func attachCrossProviderSignals(result *report.RunResult) {
	if len(result.Providers) < 2 {
		return
	}
	best := result.Providers[0].Summary.Score
	for _, provider := range result.Providers[1:] {
		if provider.Summary.Score > best {
			best = provider.Summary.Score
		}
	}
	for i := range result.Providers {
		gap := best - result.Providers[i].Summary.Score
		if gap >= 15 {
			result.Providers[i].Summary.Warnings = append(result.Providers[i].Summary.Warnings, fmt.Sprintf("score trails best provider by %.1f points in same test matrix", gap))
			if result.Providers[i].Summary.Suspicion == "low" {
				result.Providers[i].Summary.Suspicion = "medium"
			}
		}
	}
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func truncate(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "..."
}

func hasErrorSubstring(runs []report.CaseRunResult, needle string) bool {
	for _, run := range runs {
		if strings.Contains(run.Error, needle) {
			return true
		}
	}
	return false
}
