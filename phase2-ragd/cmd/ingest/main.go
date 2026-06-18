package main

import (
	"fmt"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/chunker"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/config"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/embedding"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/loader"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/vectorstore"
)

func main() {
	cfg := config.Load()
	embClient := embedding.NewClient(cfg.APIKey, cfg.EmbeddingBaseURL, cfg.EmbeddingModel)
	store := vectorstore.NewLocalStore("data/vectors.json")

	// 1. 加载文档
	docs, err := loader.LoadDir("../../testdata/docs")
	if err != nil {
		fmt.Printf("Error loading docs: %v\n", err)
	}
	fmt.Printf("加载了 %d 个文档\n", len(docs))

	// 2. 分块 -> 向量化 -> 存储
	for _, doc := range docs {
		chunks := chunker.FixedSize(doc.Title, doc.Content, 500, 50)
		stored := 0
		for _, c := range chunks {
			emb, err := embClient.Embed(c.Content)
			if err != nil {
				fmt.Printf("⚠️ 向量化失败: %v\n", err)
				continue
			}
			err = store.Add(vectorstore.Entry{
				ID:        fmt.Sprintf("%s-%d", doc.Title, c.Index),
				DocTitle:  doc.Title,
				ChunkIdx:  c.Index,
				Content:   c.Content,
				Embedding: emb,
			})
			if err != nil {
				fmt.Print(err.Error())
				continue
			}
			stored++
		}
		fmt.Printf("  %s → %d 块 → %d 已存储\n", doc.Title, len(chunks), stored)
	}

	// 3. 搜索测试
	query := "Go 语言怎么处理并发?"
	fmt.Printf("\n🔍 搜索: %s\n\n", query)
	queryEmb, _ := embClient.Embed(query)
	results, _ := store.Search(queryEmb, 3)
	for i, r := range results {
		fmt.Printf("结果 %d [%s] (相似度通过 Search 内部计算)\n", i+1, r.DocTitle)
		fmt.Printf("  %s\n\n", r.Content)
	}
	fmt.Printf("📊 存储中共 %d 个向量\n", store.Count())
}
