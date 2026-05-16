package main

import (
	"runtime"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func getOSHint() string {
	if runtime.GOOS == "windows" {
		return "Windows PowerShell"
	}
	return "Linux bash"
}

type CommandResult struct {
	Command string `json:"command"`
	Danger  string `json:"danger"` // green, yellow, orange, red
}

type ToolDefinition struct {
	Type     string     `json:"type"`
	Function ToolFunc   `json:"function"`
}

type ToolFunc struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

type ToolParameters struct {
	Type       string              `json:"type"`
	Properties map[string]ToolProp `json:"properties"`
	Required   []string            `json:"required"`
}

type ToolProp struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type ToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function ToolCallFunction    `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCallArgs struct {
	Command string `json:"command"`
	Purpose string `json:"purpose"`
}

type ChatResponseWithTools struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Config) ConvertToCommand(naturalLang string) (string, string, error) {
	LogInfo("API call | Provider: %s | Model: %s | Endpoint: %s", c.Provider, c.Model, c.Endpoint)

	provider := GetProvider(c)
	return provider.ConvertToCommand(c, naturalLang)
}

