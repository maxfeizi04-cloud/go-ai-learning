// Package provider/deepseek.go
// DeepSeek API 适配器 —— 实现 Provider 接口
package provider

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

// DeepSeek 是 DeepSeek API 的 Provider 实现
type DeepSeek struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewDeepSeek 创建 DeepSeek Provider
func NewDeepSeek(cfg *config.Config) *DeepSeek {
	return &DeepSeek{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// Name 实现 Provider 接口
func (d *DeepSeek) Name() string {
	return fmt.Sprintf("DeepSeek(%s)", d.cfg.Model)
}

// Chat 同步聊天
func (d *DeepSeek) Chat(messages []Message) (string, error) {
	reqBody := d.buildRequest(messages, false)
	respBody, err := d.doRequest(reqBody)
	if err != nil {
		return "", err
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("API 返回了空的回复")
	}
	return result.Choices[0].Message.Content, nil
}

// ChatStream 流式聊天
func (d *DeepSeek) ChatStream(messages []Message) (<-chan string, error) {
	reqBody := d.buildRequest(messages, true)
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := d.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.cfg.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 错误 (状态码 %d): %s", resp.StatusCode, string(errBody))
	}

	ch := make(chan string, 100)
	go d.readStream(resp, ch)
	return ch, nil
}

// ---------- 内部辅助方法 ----------

// buildRequest 构造 API 请求体
func (d *DeepSeek) buildRequest(messages []Message, stream bool) map[string]interface{} {
	apiMessages := make([]map[string]string, len(messages))
	for i, m := range messages {
		apiMessages[i] = map[string]string{
			"role":    m.Role,
			"content": m.Content,
		}
	}
	return map[string]interface{}{
		"model":       d.cfg.Model,
		"messages":    apiMessages,
		"stream":      stream,
		"temperature": d.cfg.Temperature,
		"max_tokens":  d.cfg.MaxTokens,
	}
}

// doRequest 发送同步请求
func (d *DeepSeek) doRequest(reqBody map[string]interface{}) ([]byte, error) {
	body, _ := json.Marshal(reqBody)
	url := d.cfg.BaseURL + "/chat/completions"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer"+d.cfg.APIKey)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求发送失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 错误 (状态码 %d): %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// readStream 在 goroutine 中读取 SSE 流
func (d *DeepSeek) readStream(resp *http.Response, ch chan<- string) {
	defer resp.Body.Close()
	defer close(ch)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		json.Unmarshal([]byte(data), &chunk)
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			ch <- chunk.Choices[0].Delta.Content
		}
	}
}
