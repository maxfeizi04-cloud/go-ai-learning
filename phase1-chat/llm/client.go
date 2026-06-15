// Package llm/client.go
// 负责：封装 HTTP 通信，对外暴露简洁的 Chat 方法
package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// ChatStream 发送消息并以流式方式接收回复
// 返回值是一个只读 channel,调用方用 for range 读取即可
// channel 会在流结束后自动关闭
func (c *Client) ChatStream(messages []Message) (<-chan string, error) {
	// 构造请求体 (和 Chat 一样，只是 stream 改为 true)
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
		"stream":   true, // 开启流式输出
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}
	url := c.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream") // 告诉服务器我们要 SSE

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 错误 (状态码 %d): %s", resp.StatusCode, string(errBody))
	}

	// 创建带缓冲的 channel, 缓冲防止 goroutine 阻塞
	ch := make(chan string, 100)

	// 启动 goroutine 在后台读取流
	go func() {
		defer resp.Body.Close()
		defer close(ch) // goroutine 结束时关闭 channel

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// SSE 数据行以 "data: " 开头
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// 流结束标记
			if data == "[DONE]" {
				return
			}

			// 解析 JSON 获取 delta.content
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue // 解析失败就跳过这一行
			}

			if len(chunk.Choices) > 0 {
				token := chunk.Choices[0].Delta.Content
				if token != "" {
					ch <- token // 发送 token 到 channel
				}
			}
		}
	}()
	return ch, nil
}
