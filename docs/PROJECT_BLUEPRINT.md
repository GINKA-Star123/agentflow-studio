# AgentFlow-Studio / FlowMind 项目蓝图

> 一个面向 AI Agent 的生产级可视化 Workflow 编排平台。

## 1. 项目定位

FlowMind 是一个支持可视化 Workflow 编排、多 Agent 协同、RAG、Memory、Tool Calling 的 AI Workflow 平台。

参考产品：

- Dify
- Coze
- LangGraph Studio
- n8n

项目目标不是复刻某个现有产品，而是打造一个更偏工程化、更适合作为 AI 全栈项目展示的平台。

## 2. 项目目标

用户可以通过拖拽节点完成 AI Workflow 的设计、执行、调试和版本管理。

首期支持节点：

- Start
- End
- Prompt
- LLM
- Condition
- Loop
- HTTP
- Tool
- Memory
- RAG

所有节点必须逐步支持：

- 实时调试
- Trace
- Token 统计
- Latency 统计
- 执行日志

## 3. 核心价值

项目用于展示以下能力：

- Go 工程能力
- AI Agent 与 Workflow Engine 设计
- Redis 缓存、状态管理、限流和 Pub/Sub
- PostgreSQL 业务建模
- Qdrant 向量检索
- Docker 与 Docker Compose 部署
- OpenTelemetry 链路追踪、指标和日志
- WebSocket 实时通信
- MCP 扩展能力
- 系统设计与可扩展架构能力

## 4. 技术栈概览

前端：

- React
- Next.js
- TypeScript
- TailwindCSS
- React Flow
- shadcn/ui
- Zustand
- TanStack Query

Go 服务：

- Go
- Gin
- GORM
- JWT
- Redis
- WebSocket
- OpenTelemetry
- Zap
- Viper

AI Runtime：

- Python
- FastAPI
- LangGraph，后续可替换为自研执行器
- OpenAI Compatible API
- Qwen
- SentenceTransformer
- Qdrant

数据层：

- PostgreSQL
- Redis
- Qdrant

可观测：

- OpenTelemetry
- Jaeger
- Prometheus
- Grafana

部署：

- Docker
- Docker Compose
- Nginx
- GitHub Actions，后期接入

## 5. 整体架构

```text
Browser
  -> Next.js
  -> Go Gateway / API Service
  -> Workflow Service
  -> Redis
  -> PostgreSQL
  -> Python AI Runtime
  -> LLM / RAG / Memory / Tool
  -> Qdrant
```

## 6. 核心模块

用户中心：

- 登录
- JWT
- 权限
- Workspace
- 用户管理

Workflow Designer：

- 拖拽
- 连线
- 保存
- 导入导出
- 版本管理

Workflow Engine：

- Start
- End
- Prompt
- LLM
- Condition
- Loop
- Tool
- HTTP
- RAG
- Memory

AI Runtime：

- Prompt 渲染
- Memory 管理
- Streaming
- Tool Calling
- 模型管理

Knowledge Base：

- 文档上传
- Chunk
- Embedding
- Qdrant
- Hybrid Search
- RAG

Redis：

- Session
- Workflow 状态
- Cache
- Rate Limit
- Pub/Sub
- Token Cache

Monitoring：

- Trace
- Metrics
- Logs
- Jaeger
- Grafana

Deployment：

- Docker
- Docker Compose
- Nginx

## 7. 开发路线

开发按照 Phase 推进：

1. Phase 1：项目初始化
2. Phase 2：用户系统
3. Phase 3：Workflow Designer
4. Phase 4：Workflow Engine
5. Phase 5：LLM Runtime
6. Phase 6：Knowledge Base
7. Phase 7：Redis
8. Phase 8：OpenTelemetry
9. Phase 9：Docker
10. Phase 10：MCP

## 8. 学习路线

随着项目逐步学习并集成：

1. Redis
2. Docker
3. OpenTelemetry
4. PostgreSQL
5. Workflow Engine
6. LangGraph
7. MCP
8. 系统设计

## 9. 开发原则

1. 工程优先
2. 高内聚、低耦合
3. API First
4. Docker First
5. 配置化
6. 可观测
7. 可扩展
8. 插件化节点
9. 每学习一项新技术即集成到项目

## 10. 最终成果

最终实现一个生产级 AI Workflow 平台，具备：

- Workflow
- RAG
- Memory
- Tool Calling
- Streaming
- Redis
- OpenTelemetry
- Docker
- MCP
- 一键部署

## 11. 项目亮点

- AI Workflow Engine
- 微服务架构
- Redis 缓存
- OpenTelemetry 链路追踪
- Docker Compose 部署
- PostgreSQL + Qdrant
- MCP 扩展
- 系统设计能力
