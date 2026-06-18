// Package vectorstore/store.go
// 向量存储接口定义
package vectorstore

// Entry 向量存储条目
type Entry struct {
	ID        string    `json:"id"`        // 唯一表示
	DocTitle  string    `json:"title"`     // 来源文档
	ChunkIdx  int       `json:"chunk_idx"` // 块序号
	Content   string    `json:"content"`   // 原始文本
	Embedding []float64 `json:"embedding"` // 向量
}

// Store 向量存储接口
// 定义这个接口是为了以后切换后端(如 Qdrant,Milvus)
type Store interface {
	Add(entry Entry) error
	Search(query []float64, topK int) ([]Entry, error)
	RemoveByDoc(docTitle string) error
	Count() int
}
