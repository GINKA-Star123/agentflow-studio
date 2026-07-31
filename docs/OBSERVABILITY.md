# OBSERVABILITY.md

> OpenTelemetry 可观测设计。

## 目标

系统必须具备 Trace、Metrics、Logs 三类可观测能力。

重点观测对象：

- HTTP 请求
- Workflow Run
- Node Execution
- LLM 调用
- RAG 检索
- Redis 操作
- PostgreSQL 查询
- Qdrant 查询
- WebSocket 连接

## Trace

### Trace 层级

```text
HTTP Request Span
  -> Workflow Run Span
    -> Node Execution Span
      -> AI Runtime Span
        -> LLM Provider Span
        -> Qdrant Search Span
```

### 必备属性

HTTP Span：

- http.method
- http.route
- http.status_code
- user_id
- workspace_id

Workflow Run Span：

- workflow_id
- workflow_version_id
- run_id
- status

Node Execution Span：

- run_id
- node_id
- node_type
- status
- latency_ms

LLM Span：

- provider
- model
- input_tokens
- output_tokens
- total_tokens
- latency_ms

RAG Span：

- knowledge_base_id
- top_k
- result_count
- latency_ms

## Metrics

### HTTP 指标

- `http_requests_total`
- `http_request_duration_seconds`
- `http_requests_errors_total`

标签：

- method
- route
- status

### Workflow 指标

- `workflow_runs_total`
- `workflow_run_duration_seconds`
- `workflow_run_errors_total`
- `workflow_active_runs`

标签：

- workflow_id
- status

### Node 指标

- `workflow_node_executions_total`
- `workflow_node_duration_seconds`
- `workflow_node_errors_total`

标签：

- node_type
- status

### AI 指标

- `llm_requests_total`
- `llm_request_duration_seconds`
- `llm_tokens_total`
- `rag_queries_total`
- `rag_query_duration_seconds`

标签：

- provider
- model
- status

## Logs

日志必须使用结构化 JSON。

通用字段：

- timestamp
- level
- message
- service
- env
- request_id
- trace_id
- span_id

Workflow 字段：

- workspace_id
- workflow_id
- run_id
- node_id
- node_type

示例：

```json
{
  "level": "info",
  "message": "node execution completed",
  "service": "api",
  "trace_id": "trace-id",
  "workspace_id": "uuid",
  "workflow_id": "uuid",
  "run_id": "uuid",
  "node_id": "llm_1",
  "node_type": "LLM",
  "latency_ms": 1200
}
```

## 工具链

OpenTelemetry：

- 统一 SDK 和语义
- Go 服务、Python Runtime 均接入

Jaeger：

- 查看 Trace
- 调试 Workflow 执行链路

Prometheus：

- 拉取 Metrics

Grafana：

- 展示 Dashboard

## 告警方向

后续可配置：

- API 错误率过高
- Workflow Run 失败率过高
- LLM 调用延迟过高
- Token 用量异常
- Redis 不可用
- PostgreSQL 不可用
- Qdrant 不可用

## 实现约束

- 不允许只打日志而不接 Trace
- Workflow Run 和 Node Execution 必须能在 Trace 中定位
- 日志必须携带 trace_id
- 用户输入、密钥、完整 Prompt 默认不进入普通日志
- 敏感字段必须脱敏
