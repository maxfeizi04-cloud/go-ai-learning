// Package config/config.go
// 负责：从环境变量加载配置，提供默认值
package config

import (
	"os"
	"strconv"
)

// Config 集中管理所有配置项
type Config struct {
	APIKey      string  // DeepSeek API Key
	Model       string  // 使用的模型名称
	BaseURL     string  // API 地址
	Temperature float64 // 0.0 ~ 2.0,控制输出随机性
}

// Load 从环境变量加载配置,未设置的项使用默认值
func Load() *Config {
	return &Config{
		APIKey:      os.Getenv("DEEPSEEK_API_KEY"),
		Model:       getEnv("DEEPSEEK_MODEL", "deepseek-v4-flash"),
		BaseURL:     getEnv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/"),
		Temperature: getEnvFloat("DEEPSEEK_TEMPERATURE", 0.7),
	}
}

// getEnv 读取环境变量，不存在时返回默认值
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// 参数：
// - key: 要读取的环境变量名称（如 "TIMEOUT_SECONDS"）
// - fallback: 如果环境变量不存在或解析失败，返回的默认值
func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
