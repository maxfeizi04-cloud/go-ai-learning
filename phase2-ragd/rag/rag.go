// Package rag/rag.go
// RAG 管道：检索 + 增强 + 生成
package rag

import (
	"fmt"
	"strings"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/chunker"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/embedding"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/loader"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/vectorstore"
)

// Pipeline RAG 管道,编排整个流程
type Pipeline struct {
	embedder     *embedding.Client
	store        vectorstore.Store
	chunkSize    int
	chunkOverlap int
	topK         int
}

// NewPipeline 创建 RAG 管道
func NewPipeline(embedder *embedding.Client, store vectorstore.Store, chunkSize, chunkOverlap, topK int) *Pipeline {
	return &Pipeline{
		embedder:     embedder,
		store:        store,
		chunkSize:    chunkSize,
		chunkOverlap: chunkOverlap,
		topK:         topK,
	}
}

// Ingest 摄入一个文档(加载 -> 分块 -> 向量化 -> 存储)
func (p *Pipeline) Ingest(doc *loader.Document) (int, error) {
	// 先删除旧数据(支持增量更新)
	err := p.store.RemoveByDoc(doc.Title)
	if err != nil {
		return 0, fmt.Errorf("error removing %s: %w", doc.Title, err)
	}

	chunks := chunker.FixedSize(doc.Title, doc.Content, p.chunkSize, p.chunkOverlap)
	for _, c := range chunks {
		emb, err := p.embedder.Embed(c.Content)
		if err != nil {
			return 0, fmt.Errorf("向量化失败 [%s-%d]: %w", doc.Title, c.Index, err)
		}
		p.store.Add(vectorstore.Entry{
			ID:        fmt.Sprintf("%s-%d", doc.Title, c.Index),
			DocTitle:  doc.Title,
			ChunkIdx:  c.Index,
			Content:   c.Content,
			Embedding: emb,
		})
	}
	return len(chunks), nil
}

// Retrieve 检索与查询最相关的文档片段
func (p *Pipeline) Retrieve(query string) ([]vectorstore.Entry, error) {
	queryEmb, err := p.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}
	return p.store.Search(queryEmb, p.topK)
}

// BuildPrompt 将检索结果组装成发给 LLM 的 Prompt
// 这个 RAG 的关键步骤: 把"外挂知识"注入到对话上下文中
func BuildPrompt(query string, entries []vectorstore.Entry) string {
	// 拼装参考资料
	var contextParts []string
	for i, e := range entries {
		contextParts = append(contextParts, fmt.Sprintf("[参考资料 %d,来源: %s]\n%s", i+1, e.DocTitle, e.Content))
	}
	context := strings.Join(contextParts, "\n\n")

	// 构造最终 Prompt
	return fmt.Sprintf(`你是一个基于参考资料的问答助手。请根据以下资料回答用户的问题。

【规则】
1. 如果资料中有相关信息，请基于资料回答，并在回答末尾注明引用的资料来源
2. 如果资料中没有相关信息，请明确说"根据现有资料，我无法回答这个问题"
3. 不要编造资料中没有的信息

【参考资料】
%s

【用户问题】
%s

【回答】`, context, query)
}

// FormatSources 格式化检索来源,用于展示给用户
func FormatSources(entries []vectorstore.Entry) string {
	var parts []string
	for i, e := range entries {
		preview := string([]rune(e.Content)[:min(80, len([]rune(e.Content)))])
		parts = append(parts, fmt.Sprintf("  [%d] %s — %s...", i+1, e.DocTitle, preview))
	}
	return strings.Join(parts, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
