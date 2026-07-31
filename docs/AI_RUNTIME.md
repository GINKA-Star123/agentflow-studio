# AI_RUNTIME.md

> AI Runtime 设计。

## 设计目标

AI Runtime 是独立 Python FastAPI 服务，负责 AI 能力执行。

职责：

- LLM 调用
- Streaming
- Prompt 渲染
- Tool Calling 适配
- Embedding
- RAG 查询
- Memory 能力预留

## 服务边界

AI Runtime 只处理 AI 执行能力，不处理：

- 用户登录
- Workspace 权限判断
- Workflow 权威定义存储
- 前端会话管理

这些能力由 Go 服务负责。

## 模块划分

```text
ai-runtime
  app
    api
    core
    providers
    prompts
    tools
    rag
    memory
    telemetry
```

### api

FastAPI 路由层。

职责：

- 请求参数校验
- 调用核心服务
- 返回统一结构

### providers

模型 Provider 适配层。

首期：

- OpenAI Compatible API
- Qwen 预留

Provider 必须统一为内部 Chat 接口。

### prompts

Prompt 模板渲染。

要求：

- 支持变量插值
- 缺失变量必须返回明确错误
- 渲染后的 Prompt 可追踪但不能泄露敏感变量

### tools

Tool Calling 适配。

要求：

- Tool Schema 使用 JSON Schema
- Tool 调用必须有 timeout
- Tool 错误必须结构化返回
- MCP Tool 后续接入到同一抽象

### rag

RAG 能力。

流程：

```text
Query
  -> Embedding
  -> Qdrant Search
  -> Filter by workspace_id and knowledge_base_id
  -> Return Chunks
  -> Assemble Context
```

### memory

Memory 能力预留。

初期可只定义接口，后续接入：

- Conversation memory
- Summary memory
- Vector memory

## Internal API

### POST /internal/v1/llm/chat

Request：

```json
{
  "provider": "openai_compatible",
  "model": "qwen-plus",
  "messages": [
    {
      "role": "user",
      "content": "Hello"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "trace": {
    "trace_id": "trace-id",
    "run_id": "uuid",
    "node_id": "llm_1"
  }
}
```

Response：

```json
{
  "content": "Hello, how can I help?",
  "tool_calls": [],
  "usage": {
    "input_tokens": 10,
    "output_tokens": 20,
    "total_tokens": 30
  },
  "model": "qwen-plus",
  "finish_reason": "stop"
}
```

### POST /internal/v1/llm/stream

返回 Server-Sent Events 或内部流式响应，Go 服务再转发给 WebSocket。

事件：

- message_delta
- tool_call_delta
- usage
- done
- error

### POST /internal/v1/embeddings

Request：

```json
{
  "model": "sentence-transformer",
  "texts": ["content"]
}
```

Response：

```json
{
  "vectors": [[0.1, 0.2]],
  "model": "sentence-transformer",
  "dimension": 768
}
```

### POST /internal/v1/rag/query

Request：

```json
{
  "workspace_id": "uuid",
  "knowledge_base_id": "uuid",
  "query": "question",
  "top_k": 5,
  "score_threshold": 0.5
}
```

Response：

```json
{
  "chunks": [
    {
      "chunk_id": "uuid",
      "document_id": "uuid",
      "content": "matched content",
      "score": 0.88,
      "metadata": {}
    }
  ]
}
```

## 配置

Runtime 配置通过环境变量注入。

必要配置：

- `APP_ENV`
- `HTTP_PORT`
- `OPENAI_COMPATIBLE_BASE_URL`
- `OPENAI_COMPATIBLE_API_KEY`
- `QDRANT_URL`
- `QDRANT_API_KEY`
- `OTEL_EXPORTER_OTLP_ENDPOINT`

## 错误结构

```json
{
  "code": "MODEL_PROVIDER_ERROR",
  "message": "provider request failed",
  "details": {},
  "retryable": true
}
```

## 可观测要求

- 每次 Runtime API 调用必须创建 Span
- 模型调用必须记录 provider、model、latency、token usage
- RAG 查询必须记录 top_k、召回数量和耗时
- 错误日志必须包含 trace_id、run_id、node_id
