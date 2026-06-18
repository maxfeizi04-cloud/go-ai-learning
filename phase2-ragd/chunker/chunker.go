// Package chunker/chunker.go
// 负责：将长文档切分成适合 Embedding 的小块
package chunker

import "strings"

// Chunk 文档分块
type Chunk struct {
	DocTitle string // 来源文档标题
	Content  string // 分块文本
	Index    int    // 在文档中的序号
}

// FixedSize 按固定字符数分块,相邻块有重叠
// size: 每块最大字符数(建议 300-500)
// overlap: 相邻块重叠的字符数(建议 size 的 10%)
// 重叠很重要! 防止关键信息刚好卡在分块边界上
func FixedSize(docTitle, text string, size, overlap int) []Chunk {
	runes := []rune(text) // 用 rune 正确处理中文等多字节字符
	var chunks []Chunk
	idx := 0
	for start := 0; start < len(runes); start += size - overlap {
		end := start + overlap
		if end > len(runes) {
			end = len(runes)
		}

		content := strings.TrimSpace(string(runes[start:end]))
		if content != "" {
			chunks = append(chunks, Chunk{
				DocTitle: docTitle,
				Content:  content,
				Index:    idx,
			})
			idx++
		}
		if end >= len(runes) {
			break
		}
	}
	return chunks
}

// Paragraph 按自然段落分块,超长段落在切分
func Paragraph(docTitle, text string, maxLen int) []Chunk {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []Chunk
	idx := 0

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// 段落不超过 maxLen 就直接用
		if len([]rune(p)) <= maxLen {
			chunks = append(chunks, Chunk{
				DocTitle: docTitle,
				Content:  p,
				Index:    idx,
			})
			idx++
		} else {
			// 超长段落用 FixedSize 再切
			subChunks := FixedSize(docTitle, text, maxLen, maxLen/10)
			for _, sc := range subChunks {
				sc.Index = idx
				chunks = append(chunks, sc)
				idx++
			}
		}
	}
	return chunks
}
