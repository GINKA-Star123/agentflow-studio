# WORKFLOW_ENGINE.md

> Workflow 引擎设计。

## 设计目标

Workflow Engine 负责把用户在画布中保存的 Workflow Definition 转换为可执行任务，并记录完整执行过程。

引擎必须支持：

- DAG 校验
- 节点执行
- 控制流
- 上下文传递
- 实时事件
- 执行日志
- Trace
- Token 和 Latency 统计

## Workflow Definition

Workflow 定义使用 JSON。

```json
{
  "schema_version": "1.0",
  "nodes": [
    {
      "id": "start_1",
      "type": "Start",
      "name": "Start",
      "config": {}
    }
  ],
  "edges": [
    {
      "id": "edge_1",
      "source": "start_1",
      "target": "prompt_1",
      "condition": null
    }
  ]
}
```

规则：

- 必须且只能有一个 Start 节点
- 至少有一个 End 节点
- 节点 ID 在 Workflow 内唯一
- Edge 的 source 和 target 必须指向存在的节点
- 不允许不可达节点
- Loop 节点必须有最大次数或退出条件
- Condition 节点必须有明确分支规则

## 节点类型

### Start

职责：

- 接收 Workflow 输入
- 初始化执行上下文

输入：Run input

输出：Context initial state

### End

职责：

- 汇总输出
- 结束 Workflow

输入：上游节点输出

输出：Run output

### Prompt

职责：

- 渲染 Prompt 模板
- 从上下文读取变量

配置：

- template
- variables

### LLM

职责：

- 调用 AI Runtime
- 支持 Chat、Streaming、Tool Calling
- 记录 token usage 和 latency

配置：

- provider
- model
- temperature
- max_tokens
- system_prompt
- user_prompt

### Condition

职责：

- 根据表达式选择分支

配置：

- expression
- branches

### Loop

职责：

- 重复执行子流程或目标节点

配置：

- max_iterations
- break_condition

### HTTP

职责：

- 调用外部 HTTP API

配置：

- method
- url
- headers
- body
- timeout_ms

### Tool

职责：

- 调用内部 Tool 或 MCP Tool

配置：

- tool_name
- arguments
- timeout_ms

### Memory

职责：

- 读取或写入会话记忆

配置：

- memory_key
- operation
- scope

### RAG

职责：

- 检索知识库
- 返回上下文片段

配置：

- knowledge_base_id
- top_k
- score_threshold

## 执行状态

Workflow Run 状态：

- pending
- running
- succeeded
- failed
- canceled

Node Execution 状态：

- pending
- running
- succeeded
- failed
- skipped

## 执行流程

```text
Create Run
  -> Load Workflow Version
  -> Validate DAG
  -> Create Execution Context
  -> Execute Start Node
  -> Schedule Ready Nodes
  -> Execute Node
  -> Persist Node Result
  -> Publish Event
  -> Continue Until End / Failed / Canceled
  -> Persist Run Result
```

## 执行上下文

上下文必须包含：

- workspace_id
- workflow_id
- run_id
- user_id
- input
- variables
- node_outputs
- trace_id

上下文原则：

- 节点只能读取上下文和自身配置
- 节点输出写入 node_outputs
- 不允许节点直接修改其他节点输出
- 敏感信息不能写入普通日志

## 节点接口

Go 侧节点执行器应遵循统一接口。

```go
type NodeExecutor interface {
    Type() string
    Validate(config map[string]any) error
    Execute(ctx context.Context, input NodeInput) (NodeOutput, error)
}
```

## 事件

执行过程必须发布事件：

- run_started
- run_completed
- run_failed
- node_started
- node_completed
- node_failed
- node_skipped
- token_usage_updated

事件用途：

- WebSocket 前端实时展示
- Redis Pub/Sub 分发
- OpenTelemetry Span 关联
- 执行日志持久化

## 错误处理

- Validation Error：执行前失败
- Node Error：节点执行失败
- Runtime Error：AI Runtime 失败
- Timeout Error：节点超时
- Canceled：用户取消

错误必须记录：

- code
- message
- node_id
- details
- retryable

## 可观测要求

- 每个 Workflow Run 创建根 Span
- 每个节点执行创建子 Span
- Span 属性包含 workflow_id、run_id、node_id、node_type
- LLM 节点必须记录 provider、model、input_tokens、output_tokens、latency_ms，并在触发 Tool Calling 时保留 tool_call_id
- HTTP 和 Tool 节点必须记录目标、状态码和耗时

## Runtime Execution 当前实现（Phase 5）

Phase 5 在 Phase 4 的同步 Workflow Runtime 基础上，补齐了 AI Runtime、Streaming、Tool Calling 协议和工具桥接预留。当前实现依然以同步 Runner 为主，但 LLM 节点已经能够把 `tool_calls` 透传到 Go Workflow，并生成 `tool` role message 预留后续多轮执行。

### 已实现能力

- Workflow Run 数据表：`workflow_runs`
- 节点执行记录表：`node_executions`
- Workflow Schema Go 类型与 JSON 解析
- Workflow DAG 校验
- NodeExecutor 接口
- ExecutionContext 执行上下文
- NodeExecutionResult 标准节点输出
- ExecutorRegistry 执行器注册表
- Start / End / Prompt / LLM 节点执行器
- AI Runtime Client 骨架
- AI Runtime OpenAI Compatible Provider 与 Streaming 支持
- Workflow Runner Service 同步执行
- Workflow Tool bridge 与 mock tool executor
- Workflow Run API
- RuntimeError 到统一 API 响应转换
- 前端 Designer Run 按钮
- 前端 Run 状态、节点执行列表、LLM 输出、token usage、latency 展示
- Phase 5 验证脚本

### 当前执行流程

```text
Load Workflow
  -> Parse Workflow Schema
  -> Create Workflow Run
  -> Validate DAG
  -> Validate Node Executors
  -> Mark Run Running
  -> Create Execution Context
  -> Build Topological Execution Order
  -> Execute Start Node
  -> Persist Start Node Execution
  -> Execute Prompt Node
  -> Persist Prompt Node Execution
  -> Execute LLM Node
  -> Persist LLM Node Execution
  -> If tool_calls exist, build tool bridge and tool messages
  -> Execute End Node
  -> Persist End Node Execution
  -> Mark Run Succeeded / Failed
```

### 当前节点执行范围

| 节点类型 | 当前状态 | 说明 |
| --- | --- | --- |
| Start | 已实现 | 读取 Run input，初始化输出 |
| Prompt | 已实现 | 使用 `text/template` 渲染 `promptTemplate` |
| LLM | 已实现 | 调用 AI Runtime Chat / Stream，保存 `response_text`、`token_usage`、`latency_ms` 和 `tool_calls` |
| End | 已实现 | 汇总上游输出作为 Workflow output |
| Tool | 已实现（桥接 / Mock） | 从 LLM `tool_calls` 生成 `ToolCallRequest`、mock 结果与 `tool` role message |
| Condition | 预留 | 后续实现分支控制 |
| Loop | 预留 | 后续实现循环控制 |
| HTTP | 预留 | 后续实现外部 HTTP 调用 |
| Memory | 预留 | 后续接入记忆读写 |
| RAG | 预留 | 后续接入知识库检索 |

### 当前 API 范围

- `POST /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}/runs`
- `GET /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}`
- `GET /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}/nodes`
- `POST /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}/cancel`

当前 Runner 是同步执行版本。发起 Run 的接口会等待执行完成或失败后返回。Cancel API 已提供，但在同步 Runner 下主要用于后续异步执行预留。

### 当前权限规则

- owner / admin / member 可以发起 Workflow Run。
- owner / admin / member 可以取消未终态 Run。
- viewer 可以查看 Run 详情和节点执行记录。
- viewer 不能发起或取消 Workflow Run。

### Phase 4 边界

以下能力不纳入 Phase 4 验收范围，进入 Phase 5 及后续阶段：

- AI Runtime 真实 Provider 接入
- LLM Streaming
- Tool Calling
- WebSocket 执行事件推送
- Redis Pub/Sub 事件分发
- 异步 Runner
- 运行中取消正在执行的节点
- Condition / Loop / HTTP / Tool / Memory / RAG 节点执行器
- OpenTelemetry Span 与 Prometheus Metrics
