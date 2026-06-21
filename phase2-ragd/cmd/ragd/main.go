package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/config"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/embedding"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/loader"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/rag"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/vectorstore"
)

// callChatAPI 调用 DeepSeek Chat API
func callChatAPI(apiKey, baseURL, prompt string) (string, error) {
	cfg := config.Load()
	if cfg.APIKey == "" {
		fmt.Println("请设置 DEEPSEEK_API_KEY")
		os.Exit(1)
	}
	reqBody := map[string]interface{}{
		"model": "deepseek-v4-flash",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.APIKey))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response body found")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  ragd ingest <dir>     导入文档目录")
		fmt.Println("  raged ask <question>  向知识库提问")
		os.Exit(1)
	}

	cfg := config.Load()
	if cfg.APIKey == "" {
		fmt.Println("请设置 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	embClient := embedding.NewClient(cfg.APIKey, cfg.EmbeddingBaseURL, cfg.EmbeddingModel)
	store := vectorstore.NewLocalStore("../../data/vectors.json")
	pipeline := rag.NewPipeline(embClient, store, 500, 50, 3)

	switch os.Args[1] {
	case "ingest":
		dir := os.Args[2]
		docs, err := loader.LoadDir(dir)
		if err != nil {
			fmt.Printf("加载目录失败: %v\n", err)
			os.Exit(1)
		}

		for _, doc := range docs {
			n, err := pipeline.Ingest(doc)
			if err != nil {
				fmt.Printf("⚠️ %s: %v\n", doc.Title, err)
				continue
			}
			fmt.Printf("✅ %s → %d 块\n", doc.Title, n)
		}
		fmt.Printf("\n📊 向量库共 %d 条记录\n", store.Count())
	case "ask":
		query := os.Args[2]
		// 1. 检索
		entries, err := pipeline.Retrieve(query)
		if err != nil {
			fmt.Printf("❌ 检索失败: %v\n", err)
			os.Exit(1)
		}

		// 2. 构建 Prompt
		prompt := rag.BuildPrompt(query, entries)

		// 3. 调用 LLM
		fmt.Printf("🤖 思考中...")
		answer, err := callChatAPI(cfg.APIKey, cfg.BaseURL, prompt)
		if err != nil {
			fmt.Printf("❌ LLM 调用失败: %v\n", err)
			os.Exit(1)
		}

		// 4. 展示结果
		fmt.Println("\n📝 回答:")
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println(answer)
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("\n📚 参考来源:")
		fmt.Println(rag.FormatSources(entries))
	}
}
