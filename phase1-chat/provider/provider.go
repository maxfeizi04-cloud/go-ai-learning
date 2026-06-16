// Package provider/provider.go
// 定义大模型服务商的统一接口
// 未来加新模型（通义千问、文心一言）只需实现这个接口
package provider

// Message 对话消息
// 用 provider 包自己的 Message 类型,避免各包之间的相互依赖
type Message struct {
	Role    string
	Content string
}

// Provider 大模型服务商接口
// 所有模型适配器必须实现这个三个方法
type Provider interface {
	// Name 返回服务商/模型名称,用于显示
	Name() string

	// Chat 同步聊天 -- 等全部回复完再返回
	// 适合后台处理场景
	Chat(messages []Message) (string, error)

	// ChatStream 流式聊天 -- 逐 Token 返回
	ChatStream(messages []Message) (<-chan string, error)
}
