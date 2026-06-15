package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// ==========================================
// 数据结构定义 —— 映射 DeepSeek API 的 JSON 格式
// 参考：https://platform.deepseek.com/api-docs/
// ==========================================

// Message 表示一条对话消息
type Message struct {
	Role    string `json:"role"`    // "system" | "user" | "assistant"
	Content string `json:"content"` // 消息正文
}

// ChatRequest 是发送给 API 的请求体
type ChatRequest struct {
	Model    string    `json:"model"`    // 模型名称,如 "deepseek-chat"
	Messages []Message `json:"messages"` // 对话历史
	Stream   bool      `json:"stream"`   // 是否流式输出
}

// Choice 是 API 返回的候选项
type Choice struct {
	Message Message `json:"message"`
}

// Usage 是 Token 用量信息
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     //输入的 Token 数
	CompletionTokens int `json:"completion_tokens"` // 完成的 Token 数
	TotalTokens      int `json:"total_tokens"`      // 总共消耗
}

// ChatResponse 是 API 返回的完整响应体
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ---------- API 调用函数（把昨天的逻辑封装成函数） ----------

// callAPI 发送对话历史到 DeepSeek，返回 AI 的回复文本
// 这是整个程序最核心的函数 —— 把 HTTP 细节封装起来
func callAPI(apiKey string, message []Message) (string, Usage, error) {
	// -------------------------------------------------
	// 第1步：构造请求体
	// -------------------------------------------------
	reqBody := ChatRequest{
		Model:    "deepseek-v4-flash",
		Messages: message,
		Stream:   false,
	}

	// json.Marshal 把 Go 结构体序列化成 JSON 字节
	body, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("❌ JSON 序列化失败: %v\n", err)
		os.Exit(1)
	}

	// -------------------------------------------------
	// 第2步：创建 HTTP 请求
	// -------------------------------------------------
	req, err := http.NewRequest(
		"POST",
		"https://api.deepseek.com/chat/completions",
		bytes.NewBuffer(body), // 请求体
	)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		os.Exit(1)
	}

	// 设置请求头 -- 这一步很容易漏!
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// -------------------------------------------------
	// 第3步：发送请求
	// -------------------------------------------------
	fmt.Println("🚀 正在调用 DeepSeek API...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求发送失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// -------------------------------------------------
	// 第4步：读取并解析响应
	// -------------------------------------------------
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应失败: %v\n", err)
		os.Exit(1)
	}

	// 先检查 HTTP 状态码
	if resp.StatusCode != 200 {
		fmt.Printf("❌ API 返回错误 (状态码 %d):\n%s\n",
			resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		fmt.Printf("❌ JSON 解析失败: %v\n", err)
		os.Exit(1)
	}

	if len(chatResp.Choices) == 0 {
		return "", Usage{}, fmt.Errorf("API 返回了空的回复")
	}

	return chatResp.Choices[0].Message.Content, chatResp.Usage, nil
}

func main() {
	// -------------------------------------------------
	// 第1步：从环境变量读取 API Key
	// -------------------------------------------------
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 请先设置环境变量 DEEPSEEK_API_KEY")
		fmt.Println("  export DEEPSEEK_API_KEY=\\\"sk-xxx\\\"")
		os.Exit(1)
	}

	// 初始化对话历史,第一条是 system prompt
	messages := []Message{
		{
			Role:    "system",
			Content: "你是一个乐于助人的编程助手,回答简洁准确.",
		},
	}

	totalTokens := 0 // 累计 Token 消耗

	fmt.Println("╔══════════════════════════════╗")
	fmt.Println("║    🤖 AI Chat 命令行聊天      ║")
	fmt.Println("║       输入 /exit 退出         ║")
	fmt.Println("╚══════════════════════════════╝")
	fmt.Println()

	// 创建输入读取器
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// 打印提示符
		fmt.Print("👤 你: ")
		if !scanner.Scan() {
			break // Ctrl+D 或读取错误
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue // 忽略空输入
		}

		// 处理退出命令
		if userInput == "/exit" || userInput == "/quit" {
			fmt.Printf("\n👋 再见！本次会话共消耗 %d Token\n", totalTokens)
			break
		}

		// 把用户消息加入历史
		messages = append(messages, Message{
			Role:    "user",
			Content: userInput,
		})

		// 调用 API
		fmt.Print("🤖 AI: ")
		reply, usage, err := callAPI(apiKey, messages)
		if err != nil {
			fmt.Printf("\n❌ 错误: %v\n", err)
			messages = messages[:len(messages)-1]
			continue
		}

		// 打印回复并累计 Token
		fmt.Println(reply)
		totalTokens += usage.TotalTokens

		// 把 AI 回复加入历史
		messages = append(messages, Message{
			Role:    "assistant",
			Content: reply,
		})

		fmt.Printf("(本次消耗 %d Token | 累计 %d)\n\n",
			usage.TotalTokens, totalTokens)
	}
}
