// Package loader/loader.go
// 负责：从文件系统加载文档，支持多种格式
package loader

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ledongthuc/pdf"
)

// Document 加载后的文档
type Document struct {
	Path    string // 文件完整路径
	Title   string // 文件名(用于展示和引用)
	Content string // 提取的纯文本内容
}

// LoadDir 加载目录中所有支持的文档
// 支持的格式：.txt, .md, .pdf
func LoadDir(dir string) ([]*Document, error) {
	var docs []*Document
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".txt", ".md":
			doc, err := loadTextFile(path)
			if err != nil {
				return nil // 跳过无法读取的文件
			}
			docs = append(docs, doc)
		case ".pdf":
			doc, err := loadPDF(path)
			if err != nil {
				fmt.Printf("⚠️ 跳过 PDF %s: %v\n", path, err)
				return nil
			}
			docs = append(docs, doc)
		}
		return nil
	})
	return docs, err
}

// loadTextFile 加载纯文本/Markdown 文件
func loadTextFile(path string) (*Document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	return &Document{
		Path:    path,
		Title:   filepath.Base(path),
		Content: string(data),
	}, nil
}

// loadPDF 从 PDF 中提取文本
func loadPDF(path string) (*Document, error) {
	f, reader, err := pdf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("打开 PDF 失败")
	}
	defer f.Close()

	var buf bytes.Buffer
	numPages := reader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue // 通过无法解析的页
		}
		buf.WriteString(text)
		buf.WriteString("\n")
	}

	return &Document{
		Path:    path,
		Title:   filepath.Base(path),
		Content: buf.String(),
	}, nil
}
