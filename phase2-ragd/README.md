# raqd — 文本语义相似度计算工具

基于 Embedding 向量的命令行工具，利用**余弦相似度**量化两段文本的语义接近程度。

## 快速开始

```bash
# 1. 安装 Ollama 并拉取嵌入模型
ollama pull nomic-embed-text

# 2. 直接运行（无需 API Key）
cd phase2-ragd/cmd/similarity
go run main.go
```

## 输出示例

```
文本1: Go 语言的 goroutine 让并发编程变得非常简单高效
文本2: Golang 的并发模型基于 CSP 理论，通过 channel 通信
 -> 相似度: 0.5966 (预期: 高)

文本3: 今天天气真好，适合去公园散步野餐
 -> 相似度: 0.4883 (预期: 低)
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DEEPSEEK_EMBEDDING_BASE_URL` | `http://localhost:11434` | 嵌入 API 地址（Ollama 默认端口） |
| `DEEPSEEK_EMBEDDING_MODEL` | `nomic-embed-text` | 嵌入模型名称 |
| `DEEPSEEK_API_KEY` | *(空)* | API Key（Ollama 不需要；远程服务时设置） |

### 切换到更好的中文模型

```bash
ollama pull bge-m3
DEEPSEEK_EMBEDDING_MODEL=bge-m3 go run main.go
```

### 切换到远程 API（如硅基流动）

```bash
export DEEPSEEK_API_KEY="sk-xxx"
export DEEPSEEK_EMBEDDING_BASE_URL="https://api.siliconflow.cn/v1"
export DEEPSEEK_EMBEDDING_MODEL="BAAI/bge-large-zh-v1.5"
go run main.go
```

## 项目结构

```
phase2-ragd/
├── config/config.go          # 配置加载（环境变量 + 默认值）
├── embedding/embedding.go    # 嵌入客户端（OpenAI 兼容接口）
│   ├── Embed()               #   单条文本 → 向量
│   └── EmbedBatch()          #   批量文本 → 向量数组
└── cmd/similarity/main.go    # 演示：余弦相似度对比
```

## 核心概念

### 余弦相似度 (Cosine Similarity)

```
sim(A, B) = (A · B) / (|A| × |B|)
```

- 值域 `[-1, 1]`，越接近 `1` 越相似
- 语义相近的文本向量方向接近 → 相似度高
- 语义无关的文本向量方向正交 → 相似度低

### 嵌入 (Embedding)

将自然语言文本映射为固定维度的浮点向量，语义相近的文本在向量空间中距离更近。

## 支持的嵌入服务

本项目的 Embedding 客户端遵循 **OpenAI 兼容接口**规范 (`POST /v1/embeddings`)，支持：

- **Ollama** — 本地免费，`nomic-embed-text` / `bge-m3`
- **硅基流动** — `BAAI/bge-large-zh-v1.5`
- **OpenAI** — `text-embedding-3-small`
- 任何兼容 OpenAI `/v1/embeddings` 格式的服务
