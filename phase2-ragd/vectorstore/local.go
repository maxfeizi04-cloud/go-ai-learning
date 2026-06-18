// Package vectorstore/local.go
// 基于 JSON 文件的本地向量存储实现
package vectorstore

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sync"
)

// LocalStore 本地文件向量存储
type LocalStore struct {
	mu      sync.RWMutex
	entries []Entry
	path    string // 持久化文件路径
}

// NewLocalStore 创建或加载本地存储
// path: 存储文件路径,如 "data/vectors.json"
func NewLocalStore(path string) *LocalStore {
	s := &LocalStore{path: path}
	s.load() // 文件恢复
	return s
}

// Add 添加一个向量条目
func (s *LocalStore) Add(entry Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
	return s.save()
}

// Search 按余弦相似度搜索 topK 个最相似的条目
func (s *LocalStore) Search(query []float64, topK int) ([]Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 计算每个条目的相似度
	type scored struct {
		entry Entry
		score float64
	}
	var results []scored
	for _, e := range s.entries {
		sim := cosineSimilarity(query, e.Embedding)
		results = append(results, scored{e, sim})
	}

	// 找 topK (简单选择排序,数据流 < 10000 时够用)
	if topK > len(results) {
		topK = len(results)
	}
	top := make([]Entry, topK)
	// 每次选出最大的
	for i := 0; i < topK; i++ {
		best := i
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[best].score {
				best = j
			}
		}
		results[i], results[best] = results[best], results[i]
		top[i] = results[i].entry
	}
	return top, nil
}

// RemoveByDoc 删除指定文档的所有条目
func (s *LocalStore) RemoveByDoc(docTitle string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var filtered []Entry
	for _, e := range s.entries {
		if e.DocTitle != docTitle {
			filtered = append(filtered, e)
		}
	}
	s.entries = filtered
	return s.save()
}

// Count 返回条目总数
func (s *LocalStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// ---------- 内部方法 ----------

func (s *LocalStore) save() error {
	// 确保目录存在
	if dir := filepath.Dir(s.path); dir != "." {
		os.MkdirAll(dir, 0755)
	}
	data, err := json.Marshal(s.entries)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *LocalStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return // 文件不存在时从空开始
	}
	err = json.Unmarshal(data, &s.entries)
	if err != nil {
		return
	}
}

// cosineSimilarity 余弦相似度
// 返回值范围: [-1, 1]，越接近1表示越相似
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0 // 维度不匹配，无法计算
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0 // 零向量与其他向量相似度定义为0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
