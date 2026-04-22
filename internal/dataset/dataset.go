package dataset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Sample struct {
	ID                 string             `json:"id"`
	Benchmark          string             `json:"benchmark,omitempty"`
	Split              string             `json:"split,omitempty"`
	Category           string             `json:"category,omitempty"`
	SystemPrompt       string             `json:"system_prompt,omitempty"`
	Prompt             string             `json:"prompt"`
	Context            string             `json:"context,omitempty"`
	Choices            []string           `json:"choices,omitempty"`
	Expected           string             `json:"expected,omitempty"`
	AcceptableAnswers  []string           `json:"acceptable_answers,omitempty"`
	Regex              string             `json:"regex,omitempty"`
	Evaluator          string             `json:"evaluator,omitempty"`
	RequiredSubstrings []string           `json:"required_substrings,omitempty"`
	Tools              []map[string]any   `json:"tools,omitempty"`
	ToolChoice         any                `json:"tool_choice,omitempty"`
	ExpectedToolCalls  []ExpectedToolCall `json:"expected_tool_calls,omitempty"`
	Metadata           map[string]any     `json:"metadata,omitempty"`
}

type ExpectedToolCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

func LoadJSONL(path string, limit int, shuffle bool) ([]Sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	items := make([]Sample, 0)
	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var sample Sample
		if err := json.Unmarshal(line, &sample); err != nil {
			return nil, fmt.Errorf("parse %s line %d: %w", path, lineNo, err)
		}
		if sample.ID == "" {
			sample.ID = fmt.Sprintf("line-%d", lineNo)
		}
		items = append(items, sample)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if shuffle {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
	}
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}
