# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 模块与环境

- **模块：** `github.com/maxfeizi04-cloud/go-ai-learning`（Go 1.26）
- **无构建脚本、Makefile 或 CI** — 这是一个动手学习项目，直接用 `go run` 运行。
- **项目中暂无测试文件**。

## 项目结构

单 Go 模块下包含两个独立可执行程序，按学习阶段组织：

```
phase1-chat/          # CLI AI 聊天工具（DeepSeek API、SSE 流式输出、多轮对话）
phase2-ragd/          # RAG 流水线（进行中 — embedding 客户端已完成，chunker/rag/vectorstore 为空壳）
```

### Phase 1 — `phase1-chat/`

**运行：** 先设置 `DEEPSEEK_API_KEY`，然后 `cd phase1-chat && go run .`

内置命令：`/help`、`/clear`、`/history`、`/model`、`/save`、`/exit`。

**核心架构 — Provider 接口**（`phase1-chat/provider/provider.go`）：
`Provider` 接口（`Name`、`Chat`、`ChatStream`）是核心抽象。`chat.Session` 持有 `Provider` 并仅通过此接口操作 — 更换后端（如新增模型）只需在新文件中实现 `Provider` 并在 `main.go` 中注入即可，会话逻辑无需任何改动。

**流式输出模式**（`phase1-chat/provider/deepseek.go`）：
`ChatStream` 启动一个 goroutine 从 HTTP 响应体读取 SSE `data:` 行，将 token 发送到缓冲 channel（容量 100）。调用方通过 range 遍历 channel。goroutine 在收到 `[DONE]` 或 EOF 时关闭 channel。所有 Provider 实现都应遵循此基于 channel 的约定。

**会话 Token 管理**（`phase1-chat/chat/chat.go`）：
`Session` 维护 `[]provider.Message` 历史记录，以 system 消息开头。`estimateTokens()` 使用粗略估算（`runeCount/2 + 4` 每条消息）。超出预算时，`trimHistory()` 会丢弃最旧的一对 user+assistant 对话轮次，在每次 `Send()` 前触发。

**配置**（`phase1-chat/config/config.go`）：所有值从环境变量读取并带有默认值（见 `Load()`）。不支持 `.env` 文件 — 必须在 shell 中 export。

### Phase 2 — `phase2-ragd/`

**运行：** `cd phase2-ragd/cmd/similarity && go run main.go`（需要本地运行 Ollama 并已拉取 `nomic-embed-text`，或通过 `DEEPSEEK_EMBEDDING_*` 环境变量指向远程服务）。

**Embedding 客户端**（`phase2-ragd/embedding/embedding.go`）：
兼容 OpenAI 格式的 `POST /v1/embeddings` 客户端。`Embed(text)` 用于单条文本，`EmbedBatch(texts)` 用于批量 — 批量结果按响应的 `index` 排序以保证输入顺序。当 `apiKey` 为空时不发送 `Authorization` 头（适配 Ollama 本地模式）。

**已完成 vs 计划中：**
- ✅ `embedding/` — embedding 客户端已可用
- ✅ `cmd/similarity/` — 余弦相似度演示
- ✅ `loader/` — `Document` 类型已定义，`LoadDir()` 仅为声明（未实现）
- ❌ `chunker/` — 文本分块（空目录）
- ❌ `vectorstore/` — 向量存储/索引（空目录）
- ❌ `rag/` — RAG 编排（空目录）

**Phase 2 配置** 在 Phase 1 模式基础上扩展了 `EmbeddingModel`（默认 `nomic-embed-text`）和 `EmbeddingBaseURL`（默认 `http://localhost:11434`）。
