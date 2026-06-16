package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/chat"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/config"
	"github.com/maxfeizi04-cloud/go-ai-learning/phase1-chat/provider"
)

func main() {
	// 加载配置
	cfg := config.Load()
	if cfg.APIKey == "" {
		fmt.Println("❌ 请设置环境变量 DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	// 创建 Provider 和 Session
	prov := provider.NewDeepSeek(cfg)
	session := chat.NewSession(prov, "你是一个乐于助人的编程助手，回答简洁准确.", 4096)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("╔══════════════════════════════╗")
	fmt.Println("║    🤖 AI Chat 命令行聊天      ║")
	fmt.Println("║   输入 /exit 退出             ║")
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
			fmt.Printf("👋 再见!")
			break
		}

		// 先检查是不是内置命令
		if session.HandleCommand(input) {
			continue
		}

		// 发送消息并显示流式回复
		fmt.Print("🤖 AI: ")
		stream, err := session.Send(input)
		if err != nil {
			fmt.Printf("\n❌ %v\n", err)
			continue
		}

		var reply strings.Builder
		for token := range stream {
			fmt.Print(token)
			reply.WriteString(token)
		}
		fmt.Println()

		// 归档完整回复
		session.CollectReply(reply.String())

	}
}
