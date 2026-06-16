# chat — Go 命令行 AI 聊天工具

基于 DeepSeek API 的终端多轮对话工具，支持流式打字输出。

## 快速开始

```bash
export DEEPSEEK_API_KEY="sk-xxx"
go run .
```

## 功能

- 流式打字输出（逐 token 显示）
- 多轮对话（自动维护上下文）
- 内置命令：
  - `/help` — 查看帮助
  - `/clear` — 清除历史
  - `/history` — 查看历史
  - `/model` — 当前模型
  - `/save` — 保存对话
  - `/exit` — 退出
- 支持调整 Temperature（`DEEPSEEK_TEMPERATURE=0.7`）
- 自动 Token 预算管理

## 项目结构
```
chat/
├── main.go             # CLI 入口
├── config/config.go    # 配置加载
├── provider/
│   ├── provider.go     # Provider 接口
│   └── deepseek.go     # DeepSeek 适配
└── chat/chat.go        # 会话管理
```

## 扩展新模型
1. 新建 `provider/qwen.go`
2. 实现 `Provider` 接口
3. `main.go` 中替换 Provider 即可