package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
	Secret   string `json:"secret,omitempty"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, ".ai-cmd.json"), nil
}

func maskKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

func LoadConfig() *Config {
	cfg := &Config{
		Provider: "openai",
		Endpoint: "http://10.203.106.209:80/v1",
		Model:    "Qwen3.6-35B-A3B-UD-Q5_K_S.gguf",
	}

	path, err := getConfigPath()
	if err == nil {
		data, err := os.ReadFile(path)
		if err == nil {
			json.Unmarshal(data, cfg)
		}
	}

	if v := os.Getenv("AI_CMD_API_KEY"); v != "" {
		cfg.APIKey = v
	}
	if v := os.Getenv("AI_CMD_ENDPOINT"); v != "" {
		cfg.Endpoint = v
	}
	if v := os.Getenv("AI_CMD_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("AI_CMD_PROVIDER"); v != "" {
		cfg.Provider = v
	}

	return cfg
}

func RunConfigWizard() {
	reader := bufio.NewReader(os.Stdin)

	cfg := LoadConfig()

	fmt.Println("=== AI Command 配置向导 ===")
	fmt.Println()

	// Provider selection
	fmt.Println("请选择 AI 提供商:")
	fmt.Println("  1) OpenAI (兼容 API)")
	fmt.Println("  2) 百度文心一言")
	if cfg.Provider == "baidu" {
		fmt.Printf("选择 [1-2, 默认 2]: ")
	} else {
		fmt.Printf("选择 [1-2, 默认 1]: ")
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "2":
		cfg.Provider = "baidu"
		fmt.Println("已选择: 百度文心一言")
	case "1":
		cfg.Provider = "openai"
		fmt.Println("已选择: OpenAI")
	default:
		fmt.Printf("已选择: %s (保持当前设置)\n", cfg.Provider)
	}

	// API Key
	fmt.Println()
	currentKey := maskKey(cfg.APIKey)
	if currentKey == "" {
		currentKey = "(未设置)"
	}
	fmt.Printf("请输入 API Key [当前: %s]: ", currentKey)
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)
	if apiKey != "" {
		cfg.APIKey = apiKey
	}

	// Secret (for Baidu)
	if cfg.Provider == "baidu" {
		fmt.Println()
		currentSecret := maskKey(cfg.Secret)
		if currentSecret == "" {
			currentSecret = "(未设置)"
		}
		fmt.Printf("请输入 Secret Key [当前: %s]: ", currentSecret)
		secret, _ := reader.ReadString('\n')
		secret = strings.TrimSpace(secret)
		if secret != "" {
			cfg.Secret = secret
		}
	}

	// Endpoint
	fmt.Println()
	fmt.Printf("请输入 API 端点 [当前: %s]: ", cfg.Endpoint)
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint != "" {
		cfg.Endpoint = endpoint
	}

	// Model
	fmt.Println()
	fmt.Printf("请输入模型名称 [当前: %s]: ", cfg.Model)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model != "" {
		cfg.Model = model
	}

	// Save
	fmt.Println()
	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "保存配置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ 配置已保存到", cfg.getConfigPath())
}

func (c *Config) DefaultEndpoint() string {
	switch c.Provider {
	case "baidu":
		return "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions"
	default:
		return "https://api.openai.com/v1"
	}
}

func (c *Config) DefaultModel() string {
	switch c.Provider {
	case "baidu":
		return "ernie-4.0-turbo-8k"
	default:
		return "gpt-4o-mini"
	}
}

func (c *Config) getConfigPath() string {
	path, _ := getConfigPath()
	return path
}

func (c *Config) Save() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
