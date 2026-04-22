package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	einoOpenAI "github.com/cloudwego/eino-ext/components/model/openai"
	einoModel "github.com/cloudwego/eino/components/model"
	einoSchema "github.com/cloudwego/eino/schema"

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
	provider config.ProviderConfig
}

func NewClient(provider config.ProviderConfig) *Client {
	return &Client{provider: provider}
}

func (c *Client) Do(ctx context.Context, req Request) (Response, error) {
	baseModel, err := einoOpenAI.NewChatModel(ctx, &einoOpenAI.ChatModelConfig{
		APIKey:          c.provider.ResolvedAPIKey(),
		Model:           firstNonEmpty(req.Model, c.provider.ResolvedModel()),
		BaseURL:         c.provider.ResolvedBaseURL(),
		Timeout:         time.Duration(c.provider.TimeoutSeconds) * time.Second,
		ReasoningEffort: parseReasoningEffort(c.provider.ReasoningEffort),
	})
	if err != nil {
		return Response{}, err
	}

	messages := toEinoMessages(req.Messages)
	options, tools, err := c.buildOptions(req)
	if err != nil {
		return Response{}, err
	}

	modelToUse := einoModel.BaseChatModel(baseModel)
	if len(tools) > 0 {
		toolModel, err := baseModel.WithTools(tools)
		if err != nil {
			return Response{}, err
		}
		modelToUse = toolModel
	}

	var rawBody []byte
	options = append(options, einoOpenAI.WithResponseMessageModifier(func(ctx context.Context, msg *einoSchema.Message, body []byte) (*einoSchema.Message, error) {
		rawBody = append([]byte(nil), body...)
		if len(body) > 0 {
			var meta struct {
				Model string `json:"model"`
			}
			if err := json.Unmarshal(body, &meta); err == nil && meta.Model != "" {
				if msg.Extra == nil {
					msg.Extra = map[string]any{}
				}
				msg.Extra["returned_model"] = meta.Model
			}
		}
		return msg, nil
	}))

	msg, err := modelToUse.Generate(ctx, messages, options...)
	if err != nil {
		return Response{StatusCode: statusCodeFromError(err)}, err
	}

	resp := Response{
		StatusCode: 200,
		RawBody:    rawBody,
		Content:    msg.Content,
	}
	if msg.ResponseMeta != nil {
		resp.FinishReason = msg.ResponseMeta.FinishReason
		if msg.ResponseMeta.Usage != nil {
			resp.PromptTokens = msg.ResponseMeta.Usage.PromptTokens
			resp.CompletionTokens = msg.ResponseMeta.Usage.CompletionTokens
			resp.TotalTokens = msg.ResponseMeta.Usage.TotalTokens
		}
	}
	if msg.Extra != nil {
		if returnedModel, ok := msg.Extra["returned_model"].(string); ok {
			resp.ReturnedModel = returnedModel
		}
	}
	if resp.ReturnedModel == "" {
		resp.ReturnedModel = firstNonEmpty(req.Model, c.provider.ResolvedModel())
	}
	for _, item := range msg.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			ID:        item.ID,
			Type:      item.Type,
			Name:      item.Function.Name,
			Arguments: item.Function.Arguments,
		})
	}
	if len(resp.RawBody) > 0 && strings.Contains(string(resp.RawBody), `"choices":null`) && resp.Content == "" && len(resp.ToolCalls) == 0 {
		return resp, fmt.Errorf("provider returned 200 but no choices payload")
	}
	return resp, nil
}

func (c *Client) buildOptions(req Request) ([]einoModel.Option, []*einoSchema.ToolInfo, error) {
	options := make([]einoModel.Option, 0, 8)
	if req.Model != "" {
		options = append(options, einoModel.WithModel(req.Model))
	}
	temperature := float32(req.Temperature)
	options = append(options, einoModel.WithTemperature(temperature))
	if len(c.provider.Headers) > 0 {
		options = append(options, einoOpenAI.WithExtraHeader(c.provider.Headers))
	}

	mergedExtra := map[string]any{}
	for k, v := range c.provider.ExtraBody {
		mergedExtra[k] = v
	}
	for k, v := range req.ExtraBody {
		mergedExtra[k] = v
	}

	tools, toolChoice, allowedToolNames, extraFields, err := splitOpenAIExtraFields(mergedExtra)
	if err != nil {
		return nil, nil, err
	}
	if len(extraFields) > 0 {
		options = append(options, einoOpenAI.WithExtraFields(extraFields))
	}
	if toolChoice != nil {
		options = append(options, einoModel.WithToolChoice(*toolChoice, allowedToolNames...))
	}
	return options, tools, nil
}

func splitOpenAIExtraFields(extra map[string]any) ([]*einoSchema.ToolInfo, *einoSchema.ToolChoice, []string, map[string]any, error) {
	if len(extra) == 0 {
		return nil, nil, nil, nil, nil
	}
	cloned := map[string]any{}
	for k, v := range extra {
		cloned[k] = v
	}

	var tools []*einoSchema.ToolInfo
	if rawTools, ok := cloned["tools"]; ok {
		parsed, err := parseTools(rawTools)
		if err != nil {
			return nil, nil, nil, nil, err
		}
		tools = parsed
		delete(cloned, "tools")
	}

	var toolChoice *einoSchema.ToolChoice
	var allowed []string
	if rawChoice, ok := cloned["tool_choice"]; ok {
		choice, names := parseToolChoice(rawChoice)
		toolChoice = choice
		allowed = names
		delete(cloned, "tool_choice")
	}
	return tools, toolChoice, allowed, cloned, nil
}

func parseTools(raw any) ([]*einoSchema.ToolInfo, error) {
	items, ok := raw.([]any)
	if !ok {
		if typed, ok := raw.([]map[string]any); ok {
			items = make([]any, len(typed))
			for i := range typed {
				items[i] = typed[i]
			}
		} else {
			return nil, fmt.Errorf("tools must be an array")
		}
	}
	out := make([]*einoSchema.ToolInfo, 0, len(items))
	for _, item := range items {
		toolMap, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool item must be an object")
		}
		function, ok := toolMap["function"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("tool item missing function object")
		}
		name, _ := function["name"].(string)
		desc, _ := function["description"].(string)
		params, err := paramsFromJSONSchema(function["parameters"])
		if err != nil {
			return nil, err
		}
		out = append(out, &einoSchema.ToolInfo{
			Name:        name,
			Desc:        desc,
			ParamsOneOf: params,
		})
	}
	return out, nil
}

func paramsFromJSONSchema(raw any) (*einoSchema.ParamsOneOf, error) {
	schemaMap, ok := raw.(map[string]any)
	if !ok || len(schemaMap) == 0 {
		return nil, nil
	}
	properties, _ := schemaMap["properties"].(map[string]any)
	requiredSet := map[string]struct{}{}
	if required, ok := schemaMap["required"].([]any); ok {
		for _, item := range required {
			requiredSet[fmt.Sprint(item)] = struct{}{}
		}
	}
	params := map[string]*einoSchema.ParameterInfo{}
	for key, value := range properties {
		propertyMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		info, err := parameterInfoFromMap(propertyMap)
		if err != nil {
			return nil, err
		}
		_, info.Required = requiredSet[key]
		params[key] = info
	}
	return einoSchema.NewParamsOneOfByParams(params), nil
}

func parameterInfoFromMap(raw map[string]any) (*einoSchema.ParameterInfo, error) {
	info := &einoSchema.ParameterInfo{
		Type: einoSchema.DataType(firstNonEmpty(fmt.Sprint(raw["type"]), string(einoSchema.String))),
		Desc: fmt.Sprint(raw["description"]),
	}
	if enumItems, ok := raw["enum"].([]any); ok {
		info.Enum = make([]string, 0, len(enumItems))
		for _, item := range enumItems {
			info.Enum = append(info.Enum, fmt.Sprint(item))
		}
	}
	if info.Type == einoSchema.Array {
		if itemsMap, ok := raw["items"].(map[string]any); ok {
			elem, err := parameterInfoFromMap(itemsMap)
			if err != nil {
				return nil, err
			}
			info.ElemInfo = elem
		}
	}
	if info.Type == einoSchema.Object {
		if subProps, ok := raw["properties"].(map[string]any); ok {
			requiredSet := map[string]struct{}{}
			if required, ok := raw["required"].([]any); ok {
				for _, item := range required {
					requiredSet[fmt.Sprint(item)] = struct{}{}
				}
			}
			info.SubParams = map[string]*einoSchema.ParameterInfo{}
			for key, value := range subProps {
				valueMap, ok := value.(map[string]any)
				if !ok {
					continue
				}
				sub, err := parameterInfoFromMap(valueMap)
				if err != nil {
					return nil, err
				}
				_, sub.Required = requiredSet[key]
				info.SubParams[key] = sub
			}
		}
	}
	return info, nil
}

func parseToolChoice(raw any) (*einoSchema.ToolChoice, []string) {
	switch typed := raw.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(typed)) {
		case "none":
			choice := einoSchema.ToolChoiceForbidden
			return &choice, nil
		case "required":
			choice := einoSchema.ToolChoiceForced
			return &choice, nil
		default:
			choice := einoSchema.ToolChoiceAllowed
			return &choice, nil
		}
	case map[string]any:
		if fn, ok := typed["function"].(map[string]any); ok {
			if name, ok := fn["name"].(string); ok && name != "" {
				choice := einoSchema.ToolChoiceForced
				return &choice, []string{name}
			}
		}
	}
	choice := einoSchema.ToolChoiceAllowed
	return &choice, nil
}

func toEinoMessages(items []Message) []*einoSchema.Message {
	out := make([]*einoSchema.Message, 0, len(items))
	for _, item := range items {
		role := einoSchema.User
		switch strings.ToLower(strings.TrimSpace(item.Role)) {
		case "system":
			role = einoSchema.System
		case "assistant":
			role = einoSchema.Assistant
		case "tool":
			role = einoSchema.Tool
		}
		out = append(out, &einoSchema.Message{Role: role, Content: item.Content})
	}
	return out
}

func parseReasoningEffort(value string) einoOpenAI.ReasoningEffortLevel {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return einoOpenAI.ReasoningEffortLevelLow
	case "high":
		return einoOpenAI.ReasoningEffortLevelHigh
	default:
		return einoOpenAI.ReasoningEffortLevelMedium
	}
}

func statusCodeFromError(err error) int {
	if apiErr, ok := err.(*einoOpenAI.APIError); ok && apiErr.HTTPStatusCode > 0 {
		return apiErr.HTTPStatusCode
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
