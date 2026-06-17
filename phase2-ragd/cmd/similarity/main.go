package main

import (
	"fmt"
	"math"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/config"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/embedding"
)

// cosineSimilarity 计算两个向量的余弦相似度
// 值域 [-1, 1],越接近 1 越相识
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func main() {
	cfg := config.Load()
	if cfg.APIKey == "" {
		fmt.Println("提示: 未设置 DEEPSEEK_API_KEY，将使用本地 Ollama（无需认证）")
	}
	client := embedding.NewClient(cfg.APIKey, cfg.EmbeddingBaseURL, cfg.EmbeddingModel)

	// 三组文本: 两个语义相近,一个无关
	text1 := "Go 语言的 goroutine 让并发编程变得非常简单高效"
	text2 := "Golang 的并发模型基于 CSP 理论，通过 channel 通信"
	text3 := "今天天气真好，适合去公园散步野餐"

	emb1, err := client.Embed(text1)
	if err != nil {
		fmt.Println(err)
	}
	emb2, _ := client.Embed(text2)
	emb3, _ := client.Embed(text3)

	sim12 := cosineSimilarity(emb1, emb2) // 应该高
	sim13 := cosineSimilarity(emb1, emb3) // 应该低

	fmt.Printf("文本1: %s\n", text1)
	fmt.Printf("文本2: %s\n", text2)
	fmt.Printf(" -> 相似度: %.4f (预期: 高)\n", sim12)
	fmt.Printf("文本3: %s\n", text3)
	fmt.Printf(" -> 相似度: %.4f (预期: 低)\n", sim13)
}
