package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"model-codex/internal/config"
	"model-codex/internal/report"
	"model-codex/internal/runner"
	"model-codex/internal/suite"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "compare" {
		return runCompare(os.Args[2:])
	}
	if len(os.Args) > 1 && os.Args[1] == "history" {
		return runHistory(os.Args[2:])
	}

	var (
		configPath   = flag.String("config", "", "Path to JSON config file")
		outputPath   = flag.String("out", "", "Path to JSON report output")
		markdownPath = flag.String("md-out", "", "Path to Markdown report output")
		htmlPath     = flag.String("html-out", "", "Path to HTML report output")
		failOn       = flag.String("fail-on", "", "Exit non-zero if any provider suspicion reaches this level: medium or high")
		provider     = flag.String("provider", "", "Only run one provider by name")
		caseNames    = flag.String("cases", "", "Comma-separated case names to run")
		repeat       = flag.Int("repeat", 0, "Override repeat count")
		timeoutSec   = flag.Int("timeout", 0, "Override provider timeout seconds")
		listCases    = flag.Bool("list-cases", false, "List available built-in cases and exit")
		baseURL      = flag.String("base-url", "", "Single-provider mode: base URL, e.g. https://api.openai.com/v1")
		model        = flag.String("model", "", "Single-provider mode: model name")
		apiKey       = flag.String("api-key", "", "Single-provider mode: API key")
		apiKeyEnv    = flag.String("api-key-env", "OPENAI_API_KEY", "Single-provider mode: env var containing API key")
		providerName = flag.String("name", "cli-provider", "Single-provider mode: provider name")
	)
	flag.Parse()

	if *listCases {
		for _, item := range suite.AvailableCases() {
			fmt.Println(item)
		}
		return 0
	}

	cfg, err := loadConfig(*configPath, *baseURL, *model, *apiKey, *apiKeyEnv, *providerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}

	if *provider != "" {
		cfg.Providers = filterProviders(cfg.Providers, *provider)
		if len(cfg.Providers) == 0 {
			fmt.Fprintf(os.Stderr, "provider %q not found in config\n", *provider)
			return 1
		}
	}
	if strings.TrimSpace(*caseNames) != "" {
		cfg.Suite.Cases = filterCases(cfg.Suite.Cases, strings.Split(*caseNames, ","))
		if len(cfg.Suite.Cases) == 0 {
			fmt.Fprintf(os.Stderr, "no matching cases found in %q\n", *caseNames)
			return 1
		}
	}
	if *repeat > 0 {
		cfg.Run.Repeats = *repeat
	}
	if *timeoutSec > 0 {
		for i := range cfg.Providers {
			cfg.Providers[i].TimeoutSeconds = *timeoutSec
		}
	}
	if *outputPath != "" {
		cfg.Run.Output = *outputPath
	}
	if cfg.Run.Output == "" {
		cfg.Run.Output = filepath.Join("artifacts", fmt.Sprintf("provider-probe-%s.json", time.Now().Format("20060102-150405")))
	}
	if *markdownPath == "" {
		*markdownPath = strings.TrimSuffix(cfg.Run.Output, filepath.Ext(cfg.Run.Output)) + ".md"
	}
	if *htmlPath == "" {
		*htmlPath = strings.TrimSuffix(cfg.Run.Output, filepath.Ext(cfg.Run.Output)) + ".html"
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		return 1
	}

	ctx := context.Background()
	result, err := runner.Run(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run probe: %v\n", err)
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Run.Output), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir: %v\n", err)
		return 1
	}
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal report: %v\n", err)
		return 1
	}
	if err := os.WriteFile(cfg.Run.Output, payload, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*markdownPath, []byte(report.Markdown(result)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write markdown report: %v\n", err)
		return 1
	}
	htmlPayload, err := report.HTML(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render html report: %v\n", err)
		return 1
	}
	if err := os.WriteFile(*htmlPath, []byte(htmlPayload), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write html report: %v\n", err)
		return 1
	}

	printSummary(result, cfg.Run.Output, *markdownPath, *htmlPath)
	if shouldFail(result, *failOn) {
		return 3
	}
	return 0
}

func runCompare(args []string) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	leftPath := fs.String("left", "", "Left JSON report path")
	rightPath := fs.String("right", "", "Right JSON report path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if strings.TrimSpace(*leftPath) == "" || strings.TrimSpace(*rightPath) == "" {
		fmt.Fprintln(os.Stderr, "compare requires -left and -right")
		return 2
	}

	left, err := loadRunResult(*leftPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load left report: %v\n", err)
		return 1
	}
	right, err := loadRunResult(*rightPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load right report: %v\n", err)
		return 1
	}

	fmt.Print(compareReports(left, right))
	return 0
}

func runHistory(args []string) int {
	fsFlags := flag.NewFlagSet("history", flag.ContinueOnError)
	dir := fsFlags.String("dir", "artifacts", "Directory containing JSON reports")
	providerName := fsFlags.String("provider", "", "Only summarize one provider")
	mdOut := fsFlags.String("md-out", "", "Optional markdown summary output path")
	htmlOut := fsFlags.String("html-out", "", "Optional HTML summary output path")
	if err := fsFlags.Parse(args); err != nil {
		return 2
	}

	paths, err := collectJSONReports(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "collect reports: %v\n", err)
		return 1
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "no json reports found")
		return 1
	}
	results := make([]report.RunResult, 0, len(paths))
	for _, path := range paths {
		item, err := loadRunResult(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip bad report %s: %v\n", path, err)
			continue
		}
		results = append(results, item)
	}
	summary, aggregates := buildHistorySummary(results, *providerName)
	fmt.Print(summary)
	if *mdOut != "" {
		if err := os.WriteFile(*mdOut, []byte(historyMarkdown(aggregates)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write history markdown: %v\n", err)
			return 1
		}
	}
	if *htmlOut != "" {
		payload, err := historyHTML(aggregates)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render history html: %v\n", err)
			return 1
		}
		if err := os.WriteFile(*htmlOut, []byte(payload), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write history html: %v\n", err)
			return 1
		}
	}
	return 0
}

func loadConfig(configPath, baseURL, model, apiKey, apiKeyEnv, providerName string) (config.Config, error) {
	if configPath != "" {
		return config.Load(configPath)
	}
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(model) == "" {
		return config.Config{}, errors.New("provide either -config or both -base-url and -model")
	}
	cfg := config.Default()
	cfg.Providers = []config.ProviderConfig{{
		Name:           providerName,
		BaseURL:        baseURL,
		Model:          model,
		APIKey:         apiKey,
		APIKeyEnv:      apiKeyEnv,
		TimeoutSeconds: 60,
	}}
	return cfg, nil
}

func filterProviders(items []config.ProviderConfig, name string) []config.ProviderConfig {
	out := make([]config.ProviderConfig, 0, len(items))
	for _, item := range items {
		if item.Name == name {
			out = append(out, item)
		}
	}
	return out
}

func filterCases(items []config.CaseConfig, names []string) []config.CaseConfig {
	allowed := map[string]struct{}{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			allowed[name] = struct{}{}
		}
	}
	out := make([]config.CaseConfig, 0, len(items))
	for _, item := range items {
		if _, ok := allowed[item.Name]; ok {
			item.Enabled = true
			out = append(out, item)
		}
	}
	return out
}

func printSummary(result runner.RunResult, outputPath, markdownPath, htmlPath string) {
	fmt.Printf("Probe completed at %s\n", result.CompletedAt.Format(time.RFC3339))
	fmt.Printf("Report saved to %s\n\n", outputPath)
	fmt.Printf("Markdown report saved to %s\n\n", markdownPath)
	fmt.Printf("HTML report saved to %s\n\n", htmlPath)
	for _, provider := range result.Providers {
		fmt.Printf("Provider: %s\n", provider.Name)
		fmt.Printf("  Score: %.1f/100\n", provider.Summary.Score)
		fmt.Printf("  Suspicion: %s\n", provider.Summary.Suspicion)
		fmt.Printf("  Pass rate: %.1f%% (%d/%d)\n", provider.Summary.PassRate*100, provider.Summary.PassedRuns, provider.Summary.TotalRuns)
		fmt.Printf("  Error runs: %d\n", provider.Summary.ErrorRuns)
		if len(provider.Summary.DistinctReturnedModels) > 0 {
			fmt.Printf("  Returned models: %s\n", strings.Join(provider.Summary.DistinctReturnedModels, ", "))
		}
		if len(provider.Summary.Warnings) > 0 {
			fmt.Println("  Warnings:")
			for _, warning := range provider.Summary.Warnings {
				fmt.Printf("    - %s\n", warning)
			}
		}
		fmt.Println()
	}
}

func shouldFail(result runner.RunResult, threshold string) bool {
	threshold = strings.ToLower(strings.TrimSpace(threshold))
	if threshold == "" {
		return false
	}
	rank := map[string]int{"low": 1, "medium": 2, "high": 3}
	want, ok := rank[threshold]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown -fail-on value %q, ignore\n", threshold)
		return false
	}
	for _, provider := range result.Providers {
		if rank[strings.ToLower(provider.Summary.Suspicion)] >= want {
			return true
		}
	}
	return false
}

func loadRunResult(path string) (report.RunResult, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return report.RunResult{}, err
	}
	var result report.RunResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return report.RunResult{}, err
	}
	return result, nil
}

func compareReports(left, right report.RunResult) string {
	var b strings.Builder
	b.WriteString("Provider Probe Compare\n\n")
	b.WriteString(fmt.Sprintf("Left:  %s (%d provider(s))\n", left.CompletedAt.Format(time.RFC3339), len(left.Providers)))
	b.WriteString(fmt.Sprintf("Right: %s (%d provider(s))\n\n", right.CompletedAt.Format(time.RFC3339), len(right.Providers)))

	leftMap := map[string]report.ProviderResult{}
	for _, item := range left.Providers {
		leftMap[item.Name] = item
	}
	for _, item := range right.Providers {
		prev, ok := leftMap[item.Name]
		if !ok {
			b.WriteString(fmt.Sprintf("- %s only exists on right\n", item.Name))
			continue
		}
		scoreDelta := item.Summary.Score - prev.Summary.Score
		passDelta := (item.Summary.PassRate - prev.Summary.PassRate) * 100
		errorDelta := item.Summary.ErrorRuns - prev.Summary.ErrorRuns
		b.WriteString(fmt.Sprintf("## %s\n", item.Name))
		b.WriteString(fmt.Sprintf("- Score: %.1f -> %.1f (%+.1f)\n", prev.Summary.Score, item.Summary.Score, scoreDelta))
		b.WriteString(fmt.Sprintf("- Pass rate: %.1f%% -> %.1f%% (%+.1f%%)\n", prev.Summary.PassRate*100, item.Summary.PassRate*100, passDelta))
		b.WriteString(fmt.Sprintf("- Error runs: %d -> %d (%+d)\n", prev.Summary.ErrorRuns, item.Summary.ErrorRuns, errorDelta))
		b.WriteString(fmt.Sprintf("- Suspicion: %s -> %s\n", prev.Summary.Suspicion, item.Summary.Suspicion))
		perCase := compareCaseAverages(prev.Runs, item.Runs)
		if len(perCase) > 0 {
			b.WriteString("- Case deltas:\n")
			for _, line := range perCase {
				b.WriteString(fmt.Sprintf("  - %s\n", line))
			}
		}
		b.WriteString("\n")
		delete(leftMap, item.Name)
	}
	for name := range leftMap {
		b.WriteString(fmt.Sprintf("- %s only exists on left\n", name))
	}
	return b.String()
}

func collectJSONReports(root string) ([]string, error) {
	entries := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".json") {
			entries = append(entries, path)
		}
		return nil
	})
	sort.Strings(entries)
	return entries, err
}

type historyAggregate struct {
	Name            string
	Runs            int
	ScoreSum        float64
	ScoreMin        float64
	ScoreMax        float64
	MediumCount     int
	HighCount       int
	ErrorRuns       int
	LatestCompleted time.Time
	LatestScore     float64
	LatestSuspicion string
}

func buildHistorySummary(results []report.RunResult, onlyProvider string) (string, []historyAggregate) {
	type acc struct {
		runs            int
		scoreSum        float64
		scoreMin        float64
		scoreMax        float64
		mediumCount     int
		highCount       int
		errorRuns       int
		latestCompleted time.Time
		latestScore     float64
		latestSuspicion string
	}
	stats := map[string]*acc{}
	for _, result := range results {
		for _, provider := range result.Providers {
			if onlyProvider != "" && provider.Name != onlyProvider {
				continue
			}
			item := stats[provider.Name]
			if item == nil {
				item = &acc{scoreMin: provider.Summary.Score, scoreMax: provider.Summary.Score}
				stats[provider.Name] = item
			}
			item.runs++
			item.scoreSum += provider.Summary.Score
			if provider.Summary.Score < item.scoreMin {
				item.scoreMin = provider.Summary.Score
			}
			if provider.Summary.Score > item.scoreMax {
				item.scoreMax = provider.Summary.Score
			}
			if provider.Summary.Suspicion == "medium" {
				item.mediumCount++
			}
			if provider.Summary.Suspicion == "high" {
				item.highCount++
			}
			item.errorRuns += provider.Summary.ErrorRuns
			if result.CompletedAt.After(item.latestCompleted) {
				item.latestCompleted = result.CompletedAt
				item.latestScore = provider.Summary.Score
				item.latestSuspicion = provider.Summary.Suspicion
			}
		}
	}
	names := make([]string, 0, len(stats))
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Provider Probe History Summary\n\n")
	aggregates := make([]historyAggregate, 0, len(names))
	for _, name := range names {
		item := stats[name]
		avg := item.scoreSum / float64(item.runs)
		aggregates = append(aggregates, historyAggregate{
			Name:            name,
			Runs:            item.runs,
			ScoreSum:        item.scoreSum,
			ScoreMin:        item.scoreMin,
			ScoreMax:        item.scoreMax,
			MediumCount:     item.mediumCount,
			HighCount:       item.highCount,
			ErrorRuns:       item.errorRuns,
			LatestCompleted: item.latestCompleted,
			LatestScore:     item.latestScore,
			LatestSuspicion: item.latestSuspicion,
		})
		b.WriteString(fmt.Sprintf("## %s\n", name))
		b.WriteString(fmt.Sprintf("- Reports: %d\n", item.runs))
		b.WriteString(fmt.Sprintf("- Score avg/min/max: %.1f / %.1f / %.1f\n", avg, item.scoreMin, item.scoreMax))
		b.WriteString(fmt.Sprintf("- Latest score/suspicion: %.1f / %s (%s)\n", item.latestScore, item.latestSuspicion, item.latestCompleted.Format(time.RFC3339)))
		b.WriteString(fmt.Sprintf("- Volatility: %.1f\n", item.scoreMax-item.scoreMin))
		b.WriteString(fmt.Sprintf("- Suspicion medium/high counts: %d / %d\n", item.mediumCount, item.highCount))
		b.WriteString(fmt.Sprintf("- Total error runs: %d\n\n", item.errorRuns))
		b.WriteString(fmt.Sprintf("- Verdict: %s\n\n", historyVerdict(historyAggregate{
			Name:            name,
			Runs:            item.runs,
			ScoreSum:        item.scoreSum,
			ScoreMin:        item.scoreMin,
			ScoreMax:        item.scoreMax,
			MediumCount:     item.mediumCount,
			HighCount:       item.highCount,
			ErrorRuns:       item.errorRuns,
			LatestCompleted: item.latestCompleted,
			LatestScore:     item.latestScore,
			LatestSuspicion: item.latestSuspicion,
		})))
	}
	return b.String(), aggregates
}

func historySummary(results []report.RunResult, onlyProvider string) string {
	s, _ := buildHistorySummary(results, onlyProvider)
	return s
}

func historyMarkdown(items []historyAggregate) string {
	var b strings.Builder
	b.WriteString("# Provider Probe History Summary\n\n")
	b.WriteString("| Provider | Reports | Avg Score | Min Score | Max Score | Latest | Volatility | Medium | High | Error Runs | Verdict |\n")
	b.WriteString("| --- | ---: | ---: | ---: | ---: | --- | ---: | ---: | ---: | ---: | --- |\n")
	for _, item := range items {
		avg := item.ScoreSum / float64(item.Runs)
		b.WriteString(fmt.Sprintf("| %s | %d | %.1f | %.1f | %.1f | %.1f/%s | %.1f | %d | %d | %d | %s |\n", item.Name, item.Runs, avg, item.ScoreMin, item.ScoreMax, item.LatestScore, item.LatestSuspicion, item.ScoreMax-item.ScoreMin, item.MediumCount, item.HighCount, item.ErrorRuns, historyVerdict(item)))
	}
	return b.String()
}

func historyHTML(items []historyAggregate) (string, error) {
	const tpl = `<!doctype html>
<html lang="zh-CN">
<head><meta charset="utf-8"><title>Provider Probe History</title>
<style>body{font-family:Arial,sans-serif;margin:24px}table{border-collapse:collapse;width:100%}th,td{border:1px solid #ddd;padding:8px}th{background:#f7f7f7}</style>
</head>
<body>
<h1>Provider Probe History Summary</h1>
<table>
<thead><tr><th>Provider</th><th>Reports</th><th>Avg Score</th><th>Min</th><th>Max</th><th>Latest</th><th>Volatility</th><th>Medium</th><th>High</th><th>Error Runs</th><th>Verdict</th></tr></thead>
<tbody>
{{ range . }}
<tr><td>{{ .Name }}</td><td>{{ .Runs }}</td><td>{{ printf "%.1f" (avg .ScoreSum .Runs) }}</td><td>{{ printf "%.1f" .ScoreMin }}</td><td>{{ printf "%.1f" .ScoreMax }}</td><td>{{ printf "%.1f/%s" .LatestScore .LatestSuspicion }}</td><td>{{ printf "%.1f" (volatility .ScoreMin .ScoreMax) }}</td><td>{{ .MediumCount }}</td><td>{{ .HighCount }}</td><td>{{ .ErrorRuns }}</td><td>{{ verdict . }}</td></tr>
{{ end }}
</tbody>
</table>
</body></html>`
	t, err := template.New("history").Funcs(template.FuncMap{
		"avg": func(sum float64, runs int) float64 {
			if runs == 0 {
				return 0
			}
			return sum / float64(runs)
		},
		"volatility": func(min, max float64) float64 { return max - min },
		"verdict":    historyVerdict,
	}).Parse(tpl)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	if err := t.Execute(&b, items); err != nil {
		return "", err
	}
	return b.String(), nil
}

func compareCaseAverages(leftRuns, rightRuns []report.CaseRunResult) []string {
	type acc struct {
		score float64
		total int
	}
	left := map[string]acc{}
	right := map[string]acc{}
	for _, run := range leftRuns {
		item := left[run.CaseName]
		item.score += run.Score
		item.total++
		left[run.CaseName] = item
	}
	for _, run := range rightRuns {
		item := right[run.CaseName]
		item.score += run.Score
		item.total++
		right[run.CaseName] = item
	}
	names := make([]string, 0, len(right))
	for name := range right {
		if _, ok := left[name]; ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	lines := make([]string, 0, len(names))
	for _, name := range names {
		l := left[name]
		r := right[name]
		ls := l.score / float64(l.total)
		rs := r.score / float64(r.total)
		lines = append(lines, fmt.Sprintf("%s avg score %.2f -> %.2f (%+.2f)", name, ls, rs, rs-ls))
	}
	return lines
}

func historyVerdict(item historyAggregate) string {
	volatility := item.ScoreMax - item.ScoreMin
	switch {
	case item.HighCount >= 2 || item.ErrorRuns >= item.Runs || volatility >= 40:
		return "高波动/高风险，优先做官方基线 A/B 并持续监控"
	case item.HighCount >= 1 || item.MediumCount >= 2 || volatility >= 20:
		return "存在不稳定信号，建议扩大样本和时段复测"
	default:
		return "当前较稳定，但仍建议保留周期性抽检"
	}
}
