package main

import (
	"fmt"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/chunker"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase2-ragd/loader"
)

func main() {
	docs, err := loader.LoadDir("../../testdata/docs")
	if err != nil {
		panic(err)
	}

	for _, doc := range docs {
		fmt.Printf("📄 %s (%d 字符)\n", doc.Title, len([]rune(doc.Content)))
		chunks := chunker.FixedSize(doc.Title, doc.Content, 300, 30)
		fmt.Printf("   → 切分为 %d 个块\n", len(chunks))
		for _, c := range chunks {
			fmt.Printf("     块 %d: %s...\n", c.Index, string([]rune(c.Content)[:min(50, len([]rune(c.Content)))]))
		}
	}
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
