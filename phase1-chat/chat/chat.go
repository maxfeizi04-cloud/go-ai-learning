// Package chat/chat.go
// 对话会话管理 —— 维护历史、处理内置命令
package chat

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/provider"
)

// Session 表示一次对话会话
type Session struct {
	prov             provider.Provider
	history          []provider.Message
	totalTokens      int
	maxContextTokens int // 对话历史的最大 Token 数
}

// NewSession 创建新的对话会话
func NewSession(prov provider.Provider, systemPrompt string, maxContextTokens int) *Session {
	return &Session{
		prov: prov,
		history: []provider.Message{
			{Role: "system", Content: systemPrompt},
		},
		maxContextTokens: maxContextTokens,
	}
}

// estimateTokens 估算消息列表的 Token 数
// 粗略估算: 中文 ~ 1字 1.5 token,英文 ~ 1词 0.75 token
// 这里简化：总字符数 / 2(偏保守的估计)
func (s *Session) estimateTokens() int {
	total := 0
	for _, m := range s.history {
		// 角色标签也占 Token,加少量开销
		total += len([]rune(m.Content))/2 + 4
	}
	return total
}

// trimHistory 当历史过长是自动裁剪
// 策略：保留 system prompt + 最近 N 条消息，中间用摘要代替
func (s *Session) trimHistory() {
	const minKeep = 4 // 至少保留 system(1) + 最近对话(3)
	for s.estimateTokens() >= s.maxContextTokens && len(s.history) > minKeep {
		// 找到 system prompt 之后最早的用户消息
		idx := 1
		if s.history[idx].Role == "assistant" {
			idx = 2
		}
		if idx+1 < len(s.history) {
			// 删除该轮对话 (user + assistant)
			s.history = append(s.history[:idx], s.history[idx+2:]...)
		} else {
			break
		}
	}
}

// Send 发送用户消息,返回 AI 流式回复
// 自动维护对话历史
func (s *Session) Send(userInput string) (<-chan string, error) {
	// 加入用户消息
	s.history = append(s.history, provider.Message{Role: "user", Content: userInput})

	// 发送前检查并裁剪
	s.trimHistory()

	// 调用 LLM 流式接口
	stream, err := s.prov.ChatStream(s.history)
	if err != nil {
		// 调用失败,移除用户消息
		s.history = s.history[:len(s.history)-1]
		return nil, err
	}

	// 返回原始的流 channel
	// 调用方负责读取完毕后调用 CollectReply 归档
	return stream, nil
}

// CollectReply 收集完整回复并加入历史
// 调用方在读取完 stream 后调用此方法
func (s *Session) CollectReply(reply string) {
	s.history = append(s.history, provider.Message{Role: "assistant", Content: reply})
}

// HandleCommand 处理内部命令,返回 true 表示已处理
func (s *Session) HandleCommand(input string) bool {
	switch strings.TrimSpace(input) {
	case "/clear":
		sysMsg := s.history[0]
		s.history = []provider.Message{sysMsg}
		fmt.Println("✅ 对话历史已清除")
		return true
	case "/history":
		fmt.Println("\n📜 对话历史:")
		fmt.Println(strings.Repeat("-", 40))
		for _, msg := range s.history[1:] { // 跳过 system prompt
			role := map[string]string{
				"user": "👤 你", "assistant": "🤖 AI",
			}[msg.Role]
			fmt.Printf("%s: %s\n", role, msg.Content)
		}
		fmt.Println(strings.Repeat("-", 40))
		return true

	case "/model":
		fmt.Printf("🔧 当前模型: %s\n", s.prov.Name())
		return true

	case "/save":
		s.saveToFele()
		return true

	case "/help":
		fmt.Println("\n可用命令:")
		fmt.Println("  /clear   - 清除对话历史")
		fmt.Println("  /history - 查看对话历史")
		fmt.Println("  /model   - 显示当前模型")
		fmt.Println("  /save    - 保存对话到文件")
		fmt.Println("  /exit    - 退出")
		return true

	default:
		return false
	}
}

// saveToFile 保存对话历史到文件
func (s *Session) saveToFele() {
	filename := fmt.Sprintf("chat_%s.txt", time.Now().Format("20060102_150405"))

	var b strings.Builder
	b.WriteString(fmt.Sprintf("模型: %s\n", s.prov.Name()))
	b.WriteString(fmt.Sprintf("保存时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(strings.Repeat("-", 40) + "\n\n")
	for _, msg := range s.history[1:] {
		b.WriteString(fmt.Sprintf("[%s]\n%s\n\n", msg.Role, msg.Content))
	}
	err := os.WriteFile(filename, []byte(b.String()), 0644)
	if err != nil {
		fmt.Printf("文件创建失败: %s\n", err)
	}
	fmt.Printf("✅ 对话已保存到 %s\n", filename)
}
