package main

import (
	"runtime"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

func (c *Config) ConvertToCommand(naturalLang string) (string, string, error) {
	LogInfo("API call | Provider: %s | Model: %s | Endpoint: %s", c.Provider, c.Model, c.Endpoint)

	provider := GetProvider(c)
	return provider.ConvertToCommand(c, naturalLang)
}

