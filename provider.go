package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Provider interface {
	ConvertToCommand(cfg *Config, naturalLang string) (string, string, error)
	ChatWithTools(cfg *Config, messages []Message) (*ChatResponseWithTools, error)
}

type OpenAIProvider struct{}

func (p *OpenAIProvider) ConvertToCommand(cfg *Config, naturalLang string) (string, string, error) {
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
		Model: cfg.Model,
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

	endpoint := cfg.Endpoint
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = endpoint + "/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
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
	if err := extractJSON(raw, &result); err != nil {
		return raw, "orange", nil
	}

	return result.Command, result.Danger, nil
}

func (p *OpenAIProvider) ChatWithTools(cfg *Config, messages []Message) (*ChatResponseWithTools, error) {
	tools := []ToolDefinition{
		{
			Type: "function",
			Function: ToolFunc{
				Name:        "execute_command",
				Description: "Execute a read-only system command to gather information. Only safe, non-modifying commands are allowed.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProp{
						"command": {
							Type:        "string",
							Description: "The read-only shell command to execute",
						},
						"purpose": {
							Type:        "string",
							Description: "Brief explanation of why this command is being run",
						},
					},
					Required: []string{"command", "purpose"},
				},
			},
		},
	}

	chatReq := ChatRequestWithTools{
		Model:       cfg.Model,
		Messages:    messages,
		Temperature: 0,
		Tools:       tools,
	}

	body, err := json.Marshal(chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := cfg.Endpoint
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = endpoint + "/chat/completions"
	}

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var chatResp ChatResponseWithTools
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	return &chatResp, nil
}

type ChatRequestWithTools struct {
	Model       string           `json:"model"`
	Messages    []Message        `json:"messages"`
	Temperature float64          `json:"temperature"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
}

func GetProvider(cfg *Config) Provider {
	switch strings.ToLower(cfg.Provider) {
	case "openai":
		return &OpenAIProvider{}
	case "baidu":
		return &BaiduProvider{}
	default:
		return &OpenAIProvider{}
	}
}

type BaiduProvider struct{}

func (p *BaiduProvider) ConvertToCommand(cfg *Config, naturalLang string) (string, string, error) {
	// Baidu Wenxin API implementation
	// API endpoint: https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions
	// Auth: need to get access_token using API Key and Secret Key

	accessToken, err := getBaiduAccessToken(cfg.APIKey, cfg.Secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to get Baidu access token: %w", err)
	}

	systemPrompt := fmt.Sprintf(`You convert natural language into a single shell command for %s.

Rules:
- Output ONLY a JSON object with "command" and "danger" keys
- "danger" levels: green, yellow, orange, red
- No markdown, no backticks, no explanations
- Use appropriate commands for the target OS`, getOSHint())

	reqBody := map[string]interface{}{
		"messages": []Message{
			{Role: "user", Content: systemPrompt + "\n\nUser: " + naturalLang},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
	}
	endpoint = endpoint + "?access_token=" + accessToken

	httpReq, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	var baiduResp struct {
		Result string `json:"result"`
		Error  *struct {
			ErrorCode int    `json:"error_code"`
			ErrorMsg  string `json:"error_msg"`
		} `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&baiduResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if baiduResp.Error != nil {
		return "", "", fmt.Errorf("API error: %s (code: %d)", baiduResp.Error.ErrorMsg, baiduResp.Error.ErrorCode)
	}

	if baiduResp.Result == "" {
		return "", "", fmt.Errorf("no response from API")
	}

	raw := strings.TrimSpace(baiduResp.Result)
	var result CommandResult
	if err := extractJSON(raw, &result); err != nil {
		return raw, "orange", nil
	}

	return result.Command, result.Danger, nil
}

func (p *BaiduProvider) ChatWithTools(cfg *Config, messages []Message) (*ChatResponseWithTools, error) {
	return nil, fmt.Errorf("Baidu provider does not support tool calling mode")
}

func getBaiduAccessToken(apiKey, secretKey string) (string, error) {
	url := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s",
		apiKey, secretKey)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return "", fmt.Errorf("token error: %s", tokenResp.Error)
	}

	return tokenResp.AccessToken, nil
}

func extractJSON(data string, v interface{}) error {
	start := strings.Index(data, "{")
	if start == -1 {
		return fmt.Errorf("no JSON object found")
	}
	decoder := json.NewDecoder(strings.NewReader(data[start:]))
	return decoder.Decode(v)
}
