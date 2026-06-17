// Package embedding/embedding.go
// 负责：调用 Embedding API，将文本转为向量
// 支持 Ollama 本地部署 + 任意 OpenAI 兼容嵌入服务
package embedding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client Embedding API 客户端（OpenAI 兼容接口）
type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
}

// NewClient 创建 Embedding 客户端
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		http:    &http.Client{},
	}
}

// buildURL 安全拼接 baseURL 和 path，避免双斜杠
func (c *Client) buildURL(path string) string {
	return strings.TrimRight(c.baseURL, "/") + path
}

// setAuthHeader 仅在 apiKey 非空时设置 Authorization 头
// Ollama 本地部署无需认证
func (c *Client) setAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}

// Embed 将文本转为向量
// 文本 → HTTP POST → Embedding API → 浮点数组（如 768 维）
func (c *Client) Embed(text string) ([]float64, error) {
	// 构造请求
	reqBody := map[string]interface{}{
		"model": c.model,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := c.buildURL("/v1/embeddings")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API 错误 (状态码 %d): %s ", resp.StatusCode, string(respBody))
	}

	// 解析响应: 提取 embedding 数组
	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("API 返回了空 Embedding")
	}

	return result.Data[0].Embedding, nil
}

// EmbedBatch 批量嵌入（一次请求处理多条文本，节省 API 调用）
func (c *Client) EmbedBatch(texts []string) ([][]float64, error) {
	reqBody := map[string]interface{}{
		"model": c.model,
		"input": texts,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}

	url := c.buildURL("/v1/embeddings")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("批量嵌入请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取批量响应失败: %w", err)
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析批量响应失败: %w", err)
	}

	// 按 index 排序返回
	embeddings := make([][]float64, len(texts))
	for _, d := range result.Data {
		embeddings[d.Index] = d.Embedding
	}
	return embeddings, nil
}
