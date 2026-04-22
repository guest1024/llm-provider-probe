package report

import (
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"
)

type RunResult struct {
	StartedAt   time.Time        `json:"started_at"`
	CompletedAt time.Time        `json:"completed_at"`
	Providers   []ProviderResult `json:"providers"`
}

type ProviderResult struct {
	Name    string          `json:"name"`
	BaseURL string          `json:"base_url"`
	Model   string          `json:"model"`
	Runs    []CaseRunResult `json:"runs"`
	Summary ProviderSummary `json:"summary"`
}

type CaseRunResult struct {
	CaseName           string              `json:"case_name"`
	Category           string              `json:"category"`
	Benchmark          string              `json:"benchmark,omitempty"`
	Split              string              `json:"split,omitempty"`
	SampleID           string              `json:"sample_id,omitempty"`
	Attempt            int                 `json:"attempt"`
	Passed             bool                `json:"passed"`
	Score              float64             `json:"score"`
	Expected           string              `json:"expected"`
	Actual             string              `json:"actual"`
	Warning            string              `json:"warning,omitempty"`
	Error              string              `json:"error,omitempty"`
	LatencyMs          int64               `json:"latency_ms"`
	StatusCode         int                 `json:"status_code"`
	ReturnedModel      string              `json:"returned_model,omitempty"`
	FinishReason       string              `json:"finish_reason,omitempty"`
	PromptTokens       int                 `json:"prompt_tokens,omitempty"`
	CompletionTokens   int                 `json:"completion_tokens,omitempty"`
	TotalTokens        int                 `json:"total_tokens,omitempty"`
	ResponseHeaders    map[string][]string `json:"response_headers,omitempty"`
	RawResponseSnippet string              `json:"raw_response_snippet,omitempty"`
}

type ProviderSummary struct {
	Score                  float64            `json:"score"`
	Suspicion              string             `json:"suspicion"`
	TotalRuns              int                `json:"total_runs"`
	PassedRuns             int                `json:"passed_runs"`
	ErrorRuns              int                `json:"error_runs"`
	PassRate               float64            `json:"pass_rate"`
	BenchmarkSummaries     []BenchmarkSummary `json:"benchmark_summaries,omitempty"`
	DistinctReturnedModels []string           `json:"distinct_returned_models,omitempty"`
	Warnings               []string           `json:"warnings,omitempty"`
}

type BenchmarkSummary struct {
	Benchmark           string  `json:"benchmark"`
	Split               string  `json:"split,omitempty"`
	Attempts            int     `json:"attempts"`
	Passes              int     `json:"passes"`
	Errors              int     `json:"errors"`
	PassRate            float64 `json:"pass_rate"`
	AvgScore            float64 `json:"avg_score"`
	AvgLatencyMs        int64   `json:"avg_latency_ms"`
	StarterBaselineBand string  `json:"starter_baseline_band,omitempty"`
}

func Markdown(result RunResult) string {
	var b strings.Builder
	b.WriteString("# Provider Probe Report\n\n")
	b.WriteString(fmt.Sprintf("- Started: %s\n", result.StartedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- Completed: %s\n", result.CompletedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("- Providers: %d\n\n", len(result.Providers)))

	for _, provider := range result.Providers {
		b.WriteString(fmt.Sprintf("## %s\n\n", provider.Name))
		b.WriteString(fmt.Sprintf("- Base URL: `%s`\n", provider.BaseURL))
		b.WriteString(fmt.Sprintf("- Requested model: `%s`\n", provider.Model))
		b.WriteString(fmt.Sprintf("- Score: **%.1f/100**\n", provider.Summary.Score))
		b.WriteString(fmt.Sprintf("- Suspicion: **%s**\n", provider.Summary.Suspicion))
		b.WriteString(fmt.Sprintf("- Pass rate: **%.1f%%** (%d/%d)\n", provider.Summary.PassRate*100, provider.Summary.PassedRuns, provider.Summary.TotalRuns))
		b.WriteString(fmt.Sprintf("- Error runs: **%d**\n", provider.Summary.ErrorRuns))
		if len(provider.Summary.DistinctReturnedModels) > 0 {
			b.WriteString(fmt.Sprintf("- Returned models: `%s`\n", strings.Join(provider.Summary.DistinctReturnedModels, "`, `")))
		}
		if len(provider.Summary.Warnings) > 0 {
			b.WriteString("- Warnings:\n")
			for _, warning := range provider.Summary.Warnings {
				b.WriteString(fmt.Sprintf("  - %s\n", warning))
			}
		}
		if rows := benchmarkRows(provider); len(rows) > 0 {
			b.WriteString("\n### Benchmark Summary\n\n")
			b.WriteString("| Benchmark | Split | Attempts | Passes | Errors | Pass Rate | Avg Score | Avg Latency(ms) | Starter Band |\n")
			b.WriteString("| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |\n")
			for _, row := range rows {
				b.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d | %.1f%% | %.2f | %d | %s |\n", row.Benchmark, row.Split, row.Attempts, row.Passes, row.Errors, row.PassRate*100, row.AvgScore, row.AvgLatencyMs, row.StarterBaselineBand))
			}
		}
		b.WriteString("\n### Case Summary\n\n")
		b.WriteString("| Case | Benchmark | Attempts | Passes | Errors | Avg Score | Avg Latency(ms) |\n")
		b.WriteString("| --- | --- | ---: | ---: | ---: | ---: | ---: |\n")
		for _, row := range summarizeCases(provider.Runs) {
			b.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d | %.2f | %d |\n", row.Name, row.Benchmark, row.Attempts, row.Passes, row.Errors, row.AvgScore, row.AvgLatencyMs))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func HTML(result RunResult) (string, error) {
	const tpl = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <title>Provider Probe Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; color: #222; }
    table { border-collapse: collapse; width: 100%; margin: 12px 0 24px; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    th { background: #f7f7f7; }
    .low { color: #12711c; font-weight: bold; }
    .medium { color: #9a6700; font-weight: bold; }
    .high { color: #b42318; font-weight: bold; }
    code { background: #f5f5f5; padding: 1px 4px; }
  </style>
</head>
<body>
  <h1>Provider Probe Report</h1>
  <p>Started: {{ .StartedAt.Format "2006-01-02 15:04:05Z07:00" }}</p>
  <p>Completed: {{ .CompletedAt.Format "2006-01-02 15:04:05Z07:00" }}</p>
  {{ range .Providers }}
  <section>
    <h2>{{ .Name }}</h2>
    <ul>
      <li>Base URL: <code>{{ .BaseURL }}</code></li>
      <li>Requested model: <code>{{ .Model }}</code></li>
      <li>Score: <strong>{{ printf "%.1f" .Summary.Score }}/100</strong></li>
      <li>Suspicion: <span class="{{ .Summary.Suspicion }}">{{ .Summary.Suspicion }}</span></li>
      <li>Pass rate: <strong>{{ printf "%.1f" (mul100 .Summary.PassRate) }}%</strong> ({{ .Summary.PassedRuns }}/{{ .Summary.TotalRuns }})</li>
      <li>Error runs: <strong>{{ .Summary.ErrorRuns }}</strong></li>
      {{ if .Summary.DistinctReturnedModels }}<li>Returned models:
        {{ range $i, $m := .Summary.DistinctReturnedModels }}{{ if $i }}, {{ end }}<code>{{ $m }}</code>{{ end }}
      </li>{{ end }}
    </ul>
    {{ if .Summary.Warnings }}
    <h3>Warnings</h3>
    <ul>
      {{ range .Summary.Warnings }}<li>{{ . }}</li>{{ end }}
    </ul>
    {{ end }}
    {{ $benchmarks := benchmarkRows . }}
    {{ if $benchmarks }}
    <h3>Benchmark Summary</h3>
    <table>
      <thead>
        <tr><th>Benchmark</th><th>Split</th><th>Attempts</th><th>Passes</th><th>Errors</th><th>Pass Rate</th><th>Avg Score</th><th>Avg Latency (ms)</th><th>Starter Band</th></tr>
      </thead>
      <tbody>
        {{ range $benchmarks }}
        <tr>
          <td>{{ .Benchmark }}</td>
          <td>{{ .Split }}</td>
          <td>{{ .Attempts }}</td>
          <td>{{ .Passes }}</td>
          <td>{{ .Errors }}</td>
          <td>{{ printf "%.1f%%" (mul100 .PassRate) }}</td>
          <td>{{ printf "%.2f" .AvgScore }}</td>
          <td>{{ .AvgLatencyMs }}</td>
          <td>{{ .StarterBaselineBand }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
    {{ end }}
    <h3>Case Summary</h3>
    <table>
      <thead>
        <tr><th>Case</th><th>Benchmark</th><th>Attempts</th><th>Passes</th><th>Errors</th><th>Avg Score</th><th>Avg Latency (ms)</th></tr>
      </thead>
      <tbody>
        {{ range summarizeCases .Runs }}
        <tr>
          <td>{{ .Name }}</td>
          <td>{{ .Benchmark }}</td>
          <td>{{ .Attempts }}</td>
          <td>{{ .Passes }}</td>
          <td>{{ .Errors }}</td>
          <td>{{ printf "%.2f" .AvgScore }}</td>
          <td>{{ .AvgLatencyMs }}</td>
        </tr>
        {{ end }}
      </tbody>
    </table>
  </section>
  {{ end }}
	</body>
</html>`
	funcs := template.FuncMap{
		"summarizeCases": summarizeCases,
		"benchmarkRows":  benchmarkRows,
		"mul100":         func(v float64) float64 { return v * 100 },
	}
	t, err := template.New("report").Funcs(funcs).Parse(tpl)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, result); err != nil {
		return "", err
	}
	return b.String(), nil
}

type CaseAggregate struct {
	Name         string
	Benchmark    string
	Attempts     int
	Passes       int
	Errors       int
	AvgScore     float64
	AvgLatencyMs int64
}

func summarizeCases(runs []CaseRunResult) []CaseAggregate {
	type acc struct {
		name      string
		benchmark string
		attempts  int
		passes    int
		errors    int
		score     float64
		latency   int64
	}
	stats := map[string]*acc{}
	for _, run := range runs {
		key := run.CaseName + "::" + run.Benchmark
		entry := stats[key]
		if entry == nil {
			entry = &acc{name: run.CaseName, benchmark: run.Benchmark}
			stats[key] = entry
		}
		entry.attempts++
		if run.Passed {
			entry.passes++
		}
		if run.Error != "" {
			entry.errors++
		}
		entry.score += run.Score
		entry.latency += run.LatencyMs
	}
	out := make([]CaseAggregate, 0, len(stats))
	for _, entry := range stats {
		out = append(out, CaseAggregate{
			Name:         entry.name,
			Benchmark:    entry.benchmark,
			Attempts:     entry.attempts,
			Passes:       entry.passes,
			Errors:       entry.errors,
			AvgScore:     entry.score / float64(entry.attempts),
			AvgLatencyMs: entry.latency / int64(entry.attempts),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].Benchmark < out[j].Benchmark
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func benchmarkRows(provider ProviderResult) []BenchmarkSummary {
	if len(provider.Summary.BenchmarkSummaries) > 0 {
		return provider.Summary.BenchmarkSummaries
	}
	return SummarizeBenchmarks(provider.Runs)
}

func SummarizeBenchmarks(runs []CaseRunResult) []BenchmarkSummary {
	type acc struct {
		benchmark string
		split     string
		attempts  int
		passes    int
		errors    int
		score     float64
		latency   int64
	}
	stats := map[string]*acc{}
	for _, run := range runs {
		if strings.TrimSpace(run.Benchmark) == "" {
			continue
		}
		key := run.Benchmark + "::" + run.Split
		entry := stats[key]
		if entry == nil {
			entry = &acc{benchmark: run.Benchmark, split: run.Split}
			stats[key] = entry
		}
		entry.attempts++
		if run.Passed {
			entry.passes++
		}
		if run.Error != "" {
			entry.errors++
		}
		entry.score += run.Score
		entry.latency += run.LatencyMs
	}
	out := make([]BenchmarkSummary, 0, len(stats))
	for _, entry := range stats {
		passRate := float64(entry.passes) / float64(entry.attempts)
		out = append(out, BenchmarkSummary{
			Benchmark:           entry.benchmark,
			Split:               entry.split,
			Attempts:            entry.attempts,
			Passes:              entry.passes,
			Errors:              entry.errors,
			PassRate:            passRate,
			AvgScore:            entry.score / float64(entry.attempts),
			AvgLatencyMs:        entry.latency / int64(entry.attempts),
			StarterBaselineBand: StarterBaselineBand(entry.benchmark, passRate),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Benchmark == out[j].Benchmark {
			return out[i].Split < out[j].Split
		}
		return out[i].Benchmark < out[j].Benchmark
	})
	return out
}

func StarterBaselineBand(benchmark string, passRate float64) string {
	type threshold struct {
		acceptable float64
		strong     float64
	}
	bands := map[string]threshold{
		"commonsenseqa":                 {acceptable: 0.50, strong: 0.80},
		"mmlu_pro":                      {acceptable: 0.45, strong: 0.75},
		"gpqa":                          {acceptable: 0.35, strong: 0.65},
		"logiqa":                        {acceptable: 0.40, strong: 0.70},
		"bbh_logical_deduction":         {acceptable: 0.40, strong: 0.70},
		"bbh_tracking_shuffled_objects": {acceptable: 0.40, strong: 0.70},
		"webqa":                         {acceptable: 0.50, strong: 0.80},
		"cn_brainteaser":                {acceptable: 0.40, strong: 0.70},
		"bfcl_style":                    {acceptable: 0.50, strong: 0.90},
		"ruler_retrieval":               {acceptable: 0.50, strong: 0.80},
	}
	item, ok := bands[benchmark]
	if !ok {
		return "-"
	}
	switch {
	case passRate < item.acceptable:
		return "weak"
	case passRate < item.strong:
		return "acceptable"
	default:
		return "strong"
	}
}
