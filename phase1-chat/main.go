package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/config"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/llm"
)

func main() {
	// 加载配置
	cfg := config.Load()
	if cfg.APIKey == "" {
		fmt.Println("❌ 请设置环境变量 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	// 创建 LLM 客户端
	client := llm.NewClient(cfg)

	// 初始化对话历史
	messages := []llm.Message{
		{Role: "system", Content: "你是一个乐于助人的编程助手，回答简洁准确."},
	}

	totalTokens := 0
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("╔══════════════════════════════╗")
	fmt.Println("║    🤖 AI Chat 命令行聊天    ║")
	fmt.Println("║  输入 /exit 退出            ║")
	fmt.Println("╚══════════════════════════════╝")

	for {
		fmt.Print("\n👤 你: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "/exit" || input == "/quit" {
			fmt.Printf("👋 再见！累计消耗 %d Token\n", totalTokens)
			break
		}

		// 加入用户消息
		messages = append(messages, llm.Message{Role: "user", Content: input})

		// 调用 LLM
		fmt.Print("🤖 AI: ")
		reply, usage, err := client.Chat(messages)
		if err != nil {
			fmt.Printf("\n❌ %v\n", err)
			messages = messages[:len(messages)-1]
			continue
		}

		fmt.Println(reply)
		totalTokens += usage.TotalTokens

		// 加入 AI 回复
		messages = append(messages, llm.Message{Role: "assistant", Content: reply})
		fmt.Printf("(%d Token)\n", totalTokens)

	}
}
