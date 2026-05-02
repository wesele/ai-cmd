package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey   string `json:"api_key"`
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
		Endpoint: "http://192.168.31.233:4001/v1",
		Model:    "qwen3-next",
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

	return cfg
}
