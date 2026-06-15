// Package config/config.go
// 负责：从环境变量加载配置，提供默认值
package config

import "os"

// Config 集中管理所有配置项
type Config struct {
	APIKey  string // DeepSeek API Key
	Model   string // 使用的模型名称
	BaseURL string // API 地址
}

// Load 从环境变量加载配置,未设置的项使用默认值
func Load() *Config {
	return &Config{
		APIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		Model:   getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		BaseURL: getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/"),
	}
}

// getEnv 读取环境变量，不存在时返回默认值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
