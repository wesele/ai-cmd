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

	fmt.Println("=== AI Command 配置向导 ===")
	fmt.Println()

	cfg := &Config{}

	// Provider selection
	fmt.Println("请选择 AI 提供商:")
	fmt.Println("  1) OpenAI (兼容 API)")
	fmt.Println("  2) 百度文心一言")
	fmt.Print("选择 [1-2, 默认 1]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "2":
		cfg.Provider = "baidu"
		fmt.Println("已选择: 百度文心一言")
	default:
		cfg.Provider = "openai"
		fmt.Println("已选择: OpenAI")
	}

	// API Key
	fmt.Println()
	fmt.Print("请输入 API Key: ")
	apiKey, _ := reader.ReadString('\n')
	cfg.APIKey = strings.TrimSpace(apiKey)

	// Secret (for Baidu)
	if cfg.Provider == "baidu" {
		fmt.Println()
		fmt.Print("请输入 Secret Key (百度需要): ")
		secret, _ := reader.ReadString('\n')
		cfg.Secret = strings.TrimSpace(secret)
	}

	// Endpoint
	fmt.Println()
	defaultEndpoint := cfg.DefaultEndpoint()
	fmt.Printf("请输入 API 端点 [默认: %s]: ", defaultEndpoint)
	endpoint, _ := reader.ReadString('\n')
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		cfg.Endpoint = defaultEndpoint
	} else {
		cfg.Endpoint = endpoint
	}

	// Model
	fmt.Println()
	defaultModel := cfg.DefaultModel()
	fmt.Printf("请输入模型名称 [默认: %s]: ", defaultModel)
	model, _ := reader.ReadString('\n')
	model = strings.TrimSpace(model)
	if model == "" {
		cfg.Model = defaultModel
	} else {
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
