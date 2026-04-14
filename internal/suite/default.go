package suite

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"model-codex/internal/config"
	"model-codex/internal/provider"
)

type EvalResult struct {
	Passed   bool    `json:"passed"`
	Score    float64 `json:"score"`
	Expected string  `json:"expected"`
	Actual   string  `json:"actual"`
	Warning  string  `json:"warning,omitempty"`
}

type BuiltCase struct {
	Name      string
	Category  string
	Messages  []provider.Message
	Expected  string
	ExtraBody map[string]any
	Evaluate  func(provider.Response) EvalResult
}

func AvailableCases() []string {
	return []string{
		"exact_json",
		"exact_line",
		"logic_filter",
		"chinese_compact",
		"nested_json_schema",
		"response_format_json_schema",
		"go_snippet_output",
		"tool_call_echo",
		"long_context_needle_small",
		"long_context_needle_medium",
		"long_context_needle_large",
	}
}

func Build(caseCfg config.CaseConfig) (BuiltCase, error) {
	switch caseCfg.Name {
	case "exact_json":
		return buildExactJSON(caseCfg), nil
	case "exact_line":
		return buildExactLine(caseCfg), nil
	case "logic_filter":
		return buildLogicFilter(caseCfg), nil
	case "chinese_compact":
		return buildChineseCompact(caseCfg), nil
	case "nested_json_schema":
		return buildNestedJSONSchema(caseCfg), nil
	case "response_format_json_schema":
		return buildResponseFormatJSONSchema(caseCfg), nil
	case "go_snippet_output":
		return buildGoSnippetOutput(caseCfg), nil
	case "tool_call_echo":
		return buildToolCallEcho(caseCfg), nil
	case "long_context_needle_small", "long_context_needle_medium", "long_context_needle_large":
		return buildLongContext(caseCfg), nil
	default:
		return BuiltCase{}, fmt.Errorf("unknown case: %s", caseCfg.Name)
	}
}

func buildExactJSON(caseCfg config.CaseConfig) BuiltCase {
	expected := map[string]any{
		"ok":            true,
		"provider_test": "exact_json",
		"number":        7,
	}
	expectedJSON, _ := json.Marshal(expected)
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "format",
		Messages: []provider.Message{{Role: "system", Content: "You are a strict formatter."}, {Role: "user", Content: "Output valid JSON only with keys ok=true, provider_test='exact_json', number=7. No markdown, no extra text."}},
		Expected: string(expectedJSON),
		Evaluate: func(resp provider.Response) EvalResult {
			var got map[string]any
			if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Content)), &got); err != nil {
				return EvalResult{Passed: false, Score: 0, Expected: string(expectedJSON), Actual: resp.Content, Warning: "invalid JSON"}
			}
			okValue, ok := got["ok"].(bool)
			testValue, testOK := got["provider_test"].(string)
			number := intFromAny(got["number"])
			passed := ok && okValue && testOK && testValue == "exact_json" && number == 7
			score := 0.25
			if ok && okValue {
				score += 0.25
			}
			if testOK && testValue == "exact_json" {
				score += 0.25
			}
			if number == 7 {
				score += 0.25
			}
			return EvalResult{Passed: passed, Score: score, Expected: string(expectedJSON), Actual: strings.TrimSpace(resp.Content)}
		},
	}
}

func buildExactLine(caseCfg config.CaseConfig) BuiltCase {
	expected := "ALPHA|BRAVO|CHARLIE"
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "instruction_following",
		Messages: []provider.Message{{Role: "user", Content: "Return exactly one line with this content and nothing else: ALPHA|BRAVO|CHARLIE"}},
		Expected: expected,
		Evaluate: exactTextEvaluator(expected),
	}
}

func buildLogicFilter(caseCfg config.CaseConfig) BuiltCase {
	expected := "2"
	prompt := `You must count records that satisfy ALL conditions and reply with the integer only.
Records:
1. id=A, region=cn, active=true, score=91
2. id=B, region=us, active=false, score=97
3. id=C, region=cn, active=true, score=84
4. id=D, region=eu, active=true, score=93
5. id=E, region=cn, active=true, score=96
6. id=F, region=cn, active=false, score=99
Conditions:
- region must be cn
- active must be true
- score must be >= 90
How many records match?`
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "reasoning",
		Messages: []provider.Message{{Role: "user", Content: prompt}},
		Expected: expected,
		Evaluate: func(resp provider.Response) EvalResult {
			actual := strings.TrimSpace(resp.Content)
			passed := actual == expected
			score := 0.0
			if passed {
				score = 1.0
			} else if strings.Contains(actual, expected) {
				score = 0.5
			}
			return EvalResult{Passed: passed, Score: score, Expected: expected, Actual: actual}
		},
	}
}

func buildChineseCompact(caseCfg config.CaseConfig) BuiltCase {
	expected := "通过|失败|待定"
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "instruction_following",
		Messages: []provider.Message{{Role: "user", Content: "请只输出一行文本，不要解释，不要标点变化，内容必须严格等于：通过|失败|待定"}},
		Expected: expected,
		Evaluate: exactTextEvaluator(expected),
	}
}

func buildNestedJSONSchema(caseCfg config.CaseConfig) BuiltCase {
	expected := `{"meta":{"lang":"zh","count":2},"items":[{"id":"a1","ok":true},{"id":"b2","ok":false}]}`
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "format",
		Messages: []provider.Message{{Role: "system", Content: "Return JSON only."}, {Role: "user", Content: "输出严格 JSON，不要 markdown，不要解释。结构必须是：meta.lang='zh', meta.count=2, items=[{id:'a1',ok:true},{id:'b2',ok:false}]。"}},
		Expected: expected,
		Evaluate: func(resp provider.Response) EvalResult {
			var got struct {
				Meta struct {
					Lang  string `json:"lang"`
					Count int    `json:"count"`
				} `json:"meta"`
				Items []struct {
					ID string `json:"id"`
					OK bool   `json:"ok"`
				} `json:"items"`
			}
			if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Content)), &got); err != nil {
				return EvalResult{Passed: false, Score: 0, Expected: expected, Actual: resp.Content, Warning: "invalid nested JSON"}
			}
			score := 0.0
			if got.Meta.Lang == "zh" {
				score += 0.25
			}
			if got.Meta.Count == 2 {
				score += 0.25
			}
			if len(got.Items) == 2 {
				score += 0.25
			}
			if len(got.Items) == 2 && got.Items[0].ID == "a1" && got.Items[0].OK && got.Items[1].ID == "b2" && !got.Items[1].OK {
				score += 0.25
			}
			passed := score == 1.0
			return EvalResult{Passed: passed, Score: score, Expected: expected, Actual: strings.TrimSpace(resp.Content)}
		},
	}
}

func buildGoSnippetOutput(caseCfg config.CaseConfig) BuiltCase {
	expected := "12"
	prompt := "Read the Go code and reply with the printed integer only.\n\npackage main\nimport \"fmt\"\nfunc main() {\n    nums := []int{2,3,4}\n    sum := 0\n    for i, v := range nums {\n        sum += i + v\n    }\n    fmt.Println(sum)\n}"
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "code_reasoning",
		Messages: []provider.Message{{Role: "user", Content: prompt}},
		Expected: expected,
		Evaluate: exactTextEvaluator(expected),
	}
}

func buildResponseFormatJSONSchema(caseCfg config.CaseConfig) BuiltCase {
	expected := `{"verdict":"pass","count":3}`
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "structured_output",
		Messages: []provider.Message{{Role: "user", Content: "Return a JSON object with verdict='pass' and count=3. Do not include any extra text."}},
		Expected: expected,
		ExtraBody: map[string]any{
			"response_format": map[string]any{
				"type": "json_schema",
				"json_schema": map[string]any{
					"name":   "provider_probe_schema",
					"strict": true,
					"schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"verdict": map[string]any{"type": "string", "enum": []string{"pass"}},
							"count":   map[string]any{"type": "integer", "enum": []int{3}},
						},
						"required":             []string{"verdict", "count"},
						"additionalProperties": false,
					},
				},
			},
		},
		Evaluate: func(resp provider.Response) EvalResult {
			var got struct {
				Verdict string `json:"verdict"`
				Count   int    `json:"count"`
			}
			if err := json.Unmarshal([]byte(strings.TrimSpace(resp.Content)), &got); err != nil {
				return EvalResult{Passed: false, Score: 0, Expected: expected, Actual: resp.Content, Warning: "invalid response_format JSON"}
			}
			score := 0.0
			if got.Verdict == "pass" {
				score += 0.5
			}
			if got.Count == 3 {
				score += 0.5
			}
			return EvalResult{Passed: score == 1.0, Score: score, Expected: expected, Actual: strings.TrimSpace(resp.Content)}
		},
	}
}

func buildToolCallEcho(caseCfg config.CaseConfig) BuiltCase {
	expected := `probe_echo({"token":"PX-77"})`
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "tool_calling",
		Messages: []provider.Message{{Role: "user", Content: "You must call the tool `probe_echo` with JSON arguments {\"token\":\"PX-77\"}. Do not answer normally."}},
		Expected: expected,
		ExtraBody: map[string]any{
			"tools": []map[string]any{
				{
					"type": "function",
					"function": map[string]any{
						"name":        "probe_echo",
						"description": "Echoes the probe token.",
						"parameters": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"token": map[string]any{"type": "string"},
							},
							"required":             []string{"token"},
							"additionalProperties": false,
						},
					},
				},
			},
			"tool_choice": map[string]any{
				"type": "function",
				"function": map[string]any{
					"name": "probe_echo",
				},
			},
		},
		Evaluate: func(resp provider.Response) EvalResult {
			if len(resp.ToolCalls) == 0 {
				return EvalResult{Passed: false, Score: 0, Expected: expected, Actual: resp.Content, Warning: "no tool call returned"}
			}
			call := resp.ToolCalls[0]
			var args struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
				return EvalResult{Passed: false, Score: 0.25, Expected: expected, Actual: call.Name + "(" + call.Arguments + ")", Warning: "invalid tool call arguments"}
			}
			score := 0.0
			if call.Name == "probe_echo" {
				score += 0.5
			}
			if args.Token == "PX-77" {
				score += 0.5
			}
			actual := call.Name + "(" + call.Arguments + ")"
			return EvalResult{Passed: score == 1.0, Score: score, Expected: expected, Actual: actual}
		},
	}
}

func buildLongContext(caseCfg config.CaseConfig) BuiltCase {
	tokenCount := paramInt(caseCfg.Params, "approx_tokens", 4000)
	needle := fmt.Sprintf("ZXQ-%d-KITE", tokenCount)
	haystack := generateHaystack(tokenCount, needle)
	expected := needle
	return BuiltCase{
		Name:     caseCfg.Name,
		Category: "long_context",
		Messages: []provider.Message{{Role: "system", Content: "You are a precise retriever."}, {Role: "user", Content: haystack + "\n\nQuestion: What is the hidden token? Reply with the token only."}},
		Expected: expected,
		Evaluate: func(resp provider.Response) EvalResult {
			actual := strings.TrimSpace(resp.Content)
			passed := actual == expected
			score := 0.0
			if passed {
				score = 1.0
			} else if strings.Contains(actual, expected) {
				score = 0.5
			}
			warning := ""
			if !passed {
				warning = "needle retrieval failed"
			}
			return EvalResult{Passed: passed, Score: score, Expected: expected, Actual: actual, Warning: warning}
		},
	}
}

func generateHaystack(tokenCount int, needle string) string {
	if tokenCount < 100 {
		tokenCount = 100
	}
	tokens := make([]string, 0, tokenCount+8)
	insertAt := tokenCount * 3 / 5
	for i := 0; i < tokenCount; i++ {
		if i == insertAt {
			tokens = append(tokens, "HIDDEN_TOKEN_START", needle, "HIDDEN_TOKEN_END")
		}
		tokens = append(tokens, fmt.Sprintf("tok%05d", i))
	}
	return "Context:\n" + strings.Join(tokens, " ")
}

func paramInt(params map[string]any, key string, fallback int) int {
	if params == nil {
		return fallback
	}
	value, ok := params[key]
	if !ok {
		return fallback
	}
	return intFromAny(value)
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	default:
		return 0
	}
}

func exactTextEvaluator(expected string) func(provider.Response) EvalResult {
	return func(resp provider.Response) EvalResult {
		actual := strings.TrimSpace(resp.Content)
		passed := actual == expected
		score := 0.0
		if passed {
			score = 1.0
		} else if strings.Contains(actual, expected) {
			score = 0.5
		}
		return EvalResult{Passed: passed, Score: score, Expected: expected, Actual: actual}
	}
}
