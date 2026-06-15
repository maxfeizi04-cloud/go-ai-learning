// Package llm/client.go
// 负责：封装 HTTP 通信，对外暴露简洁的 Chat 方法
package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/config"
)

// Message 表示一条对话消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage 是 Token 用量信息
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     //输入的 Token 数
	CompletionTokens int `json:"completion_tokens"` // 完成的 Token 数
	TotalTokens      int `json:"total_tokens"`      // 总共消耗
}

// Client 封装了 DeepSeek API 的 HTTP 通信
// 所有和 API 的交互都通过这个结构体完成
type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewClient 创建一个新的 API 客户端
func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// Chat 发送消息并获取回复
// 这是对外暴露的核心方法 -- 调用者不需要知道 HTTP 细节
func (c *Client) Chat(messages []Message) (string, Usage, error) {
	// 构造请求体 （把我们的 Message 转成 API 需要的格式）
	apiMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		apiMessages[i] = map[string]string{
			"role":    m.Role,
			"content": m.Content,
		}
	}

	reqBody := map[string]interface{}{
		"model":    c.cfg.Model,
		"messages": apiMessages,
		"stream":   false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", Usage{}, fmt.Errorf("序列号请求失败: %w", err)
	}

	// 创建 HTTP 请求
	url := c.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", Usage{}, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", Usage{}, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", Usage{}, fmt.Errorf("API 返回错误 (状态码 %d): %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return "", Usage{}, fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("API 返回了空的回复")
	}

	return result.Choices[0].Message.Content, result.Usage, nil
}
