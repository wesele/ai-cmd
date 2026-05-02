package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
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
	systemPrompt := fmt.Sprintf(`You convert natural language into a single shell command for %s.

Rules:
- Output ONLY a JSON object with "command" and "danger" keys
- "danger" levels:
  - "green": Harmless, read-only/informational (e.g., ls, dir, cat, Get-Process)
  - "yellow": Low danger, creates content (e.g., mkdir, touch, New-Item)
  - "orange": Medium danger, modifies/deletes individual objects (e.g., rm file.txt, mv, Set-Content)
  - "red": High danger, modifies/deletes batch objects or system-wide changes (e.g., rm -rf, del /s, format)
- No markdown, no backticks, no explanations
- Use appropriate and concise commands for the target OS
- Prefer common idiomatic commands if compatible
- If unclear, output exactly: {"command": "ERROR: unclear request", "danger": "red"}
- If dangerous beyond reason, output exactly: {"command": "ERROR: dangerous command", "danger": "red"}`, getOSHint())

	reqBody := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: naturalLang},
		},
		Temperature: 0,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.Endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		return "", "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", "", fmt.Errorf("no response from API")
	}

	raw := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result CommandResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Fallback for non-JSON response
		return raw, "orange", nil
	}

	return result.Command, result.Danger, nil
}

