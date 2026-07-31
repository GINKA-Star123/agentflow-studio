# SYSTEM_ARCHITECTURE.md

> 系统架构。

## 架构目标

系统需要同时满足：

- 可视化 Workflow 编排
- 可执行、可追踪、可调试的 Workflow Engine
- 可替换模型 Provider 的 AI Runtime
- 可扩展节点和工具体系
- 可观测、可部署、可演示

## 总体链路

```text
Browser
  -> Next.js Web App
  -> Go API Gateway
  -> Workflow Service
  -> PostgreSQL
  -> Redis
  -> Python AI Runtime
  -> LLM Provider / Tool / Memory / RAG
  -> Qdrant
```

## 服务划分

### Web App

职责：

- 登录和用户界面
- Workflow Designer
- Workflow 运行态展示
- 实时日志和 Trace 展示
- 知识库管理

边界：

- 不直接访问数据库
- 不保存权威 Workflow 状态
- 不直接调用模型 Provider

### Go API Gateway / Backend

职责：

- 用户认证和权限
- Workspace 隔离
- Workflow 定义管理
- Workflow 执行调度
- Run 状态管理
- WebSocket 推送
- Redis、PostgreSQL、OpenTelemetry 集成

边界：

- 不在 Handler 中编写复杂业务逻辑
- 不直接实现模型推理
- 不把长时间 AI 任务阻塞在普通请求生命周期中

### Workflow Service

职责：

- Workflow DAG 校验
- 节点执行编排
- 执行上下文管理
- 节点状态记录
- 错误处理、重试和中断预留

边界：

- 节点实现通过统一接口接入
- AI 能力通过 AI Runtime 调用
- 数据存储通过 Repository 抽象

### Python AI Runtime

职责：

- LLM 调用
- Prompt 渲染
- Streaming
- Tool Calling 适配
- Embedding
- RAG 上下文组装

边界：

- 不管理用户会话
- 不承担 Workspace 权限判断
- 不保存 Workflow 权威定义

### Data Services

PostgreSQL：

- 用户、Workspace、Workflow、Run、Node Execution、Knowledge Base 元数据

Redis：

- Session / Token Cache
- Workflow Run 临时状态
- Rate Limit
- Pub/Sub 执行事件
- 热点缓存

Qdrant：

- 文档 Chunk 向量
- 语义检索索引

## 核心数据流

### Workflow 保存

```text
Web Designer
  -> Go API: POST /workflows
  -> Validate Workflow Schema
  -> PostgreSQL
  -> Return Workflow
```

### Workflow 执行

```text
Web App
  -> Go API: POST /workflows/{id}/runs
  -> Workflow Engine
  -> Redis: write running state
  -> PostgreSQL: create run records
  -> Python Runtime: execute LLM/RAG/Tool parts
  -> WebSocket: push events
  -> PostgreSQL: persist final state
```

### RAG 检索

```text
Workflow RAG Node
  -> Go Workflow Engine
  -> Python Runtime
  -> Embedding Model
  -> Qdrant Search
  -> Context Assembly
  -> LLM Node
```

## 关键设计约束

- Workflow 定义必须使用稳定 JSON Schema
- 每次 Workflow 执行必须生成唯一 Run ID
- 每个节点执行必须有唯一 Node Execution ID
- 所有跨服务调用必须携带 trace_id
- 前端显示的状态必须来自后端事件或查询结果
- Redis 只作为临时状态和加速层，PostgreSQL 才是业务权威状态
- Qdrant 只保存向量检索数据，不保存业务权限状态

## 扩展方向

- MCP Tool 扩展
- Multi-Agent 编排
- 节点插件市场
- Workflow Template
- 多模型 Provider
- 异步任务队列
