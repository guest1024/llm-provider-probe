package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"model-codex/internal/config"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model       string         `json:"model"`
	Messages    []Message      `json:"messages"`
	Temperature float64        `json:"temperature,omitempty"`
	ExtraBody   map[string]any `json:"-"`
}

type Response struct {
	StatusCode       int
	Headers          map[string][]string
	RawBody          []byte
	Content          string
	ReturnedModel    string
	FinishReason     string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	ToolCalls        []ToolCall
}

type ToolCall struct {
	ID        string
	Type      string
	Name      string
	Arguments string
}

type Client struct {
	httpClient *http.Client
	provider   config.ProviderConfig
}

func NewClient(provider config.ProviderConfig) *Client {
	timeout := time.Duration(provider.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
		provider:   provider,
	}
}

func (c *Client) Do(ctx context.Context, req Request) (Response, error) {
	payload := map[string]any{
		"model":       req.Model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
	}
	for key, value := range c.provider.ExtraBody {
		payload[key] = value
	}
	for key, value := range req.ExtraBody {
		payload[key] = value
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return Response{}, err
	}

	url := strings.TrimRight(c.provider.BaseURL, "/") + c.provider.ResolvedEndpoint()
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey := c.provider.ResolvedAPIKey(); apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}
	for key, value := range c.provider.Headers {
		httpReq.Header.Set(key, value)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return Response{}, err
	}

	result := Response{StatusCode: httpResp.StatusCode, Headers: httpResp.Header.Clone(), RawBody: body}
	if httpResp.StatusCode >= 400 {
		return result, fmt.Errorf("http %d: %s", httpResp.StatusCode, string(body))
	}

	var parsed struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return result, fmt.Errorf("parse provider response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return result, fmt.Errorf("provider returned 200 but no choices payload")
	}
	result.Content = parsed.Choices[0].Message.Content
	result.FinishReason = parsed.Choices[0].FinishReason
	for _, item := range parsed.Choices[0].Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        item.ID,
			Type:      item.Type,
			Name:      item.Function.Name,
			Arguments: item.Function.Arguments,
		})
	}
	result.ReturnedModel = parsed.Model
	result.PromptTokens = parsed.Usage.PromptTokens
	result.CompletionTokens = parsed.Usage.CompletionTokens
	result.TotalTokens = parsed.Usage.TotalTokens
	return result, nil
}
