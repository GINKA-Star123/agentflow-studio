# PROJECT_BLUEPRINT.md

# AgentFlow-Studio ------ AI Workflow 平台项目蓝图

> 一个面向 AI Agent 的生产级可视化 Workflow 编排平台。

------------------------------------------------------------------------

# 一、项目定位

## 项目简介

FlowMind 是一个支持可视化 Workflow 编排、多 Agent
协同、RAG、Memory、Tool Calling 的 AI Workflow 平台。

参考产品： - Dify - Coze - LangGraph Studio - n8n

目标不是复刻，而是打造一个更偏工程化、更适合作为 AI 全栈项目展示的平台。

------------------------------------------------------------------------

# 二、项目目标

实现用户通过拖拽节点即可完成 AI Workflow 的设计、执行、调试。

支持节点：

-   Start
-   End
-   Prompt
-   LLM
-   Condition
-   Loop
-   HTTP
-   Tool
-   Memory
-   RAG

所有节点支持：

-   实时调试
-   Trace
-   Token统计
-   Latency统计
-   执行日志

------------------------------------------------------------------------

# 三、项目核心价值

展示以下能力：

-   Go 工程能力
-   AI Agent
-   Workflow Engine
-   Redis
-   PostgreSQL
-   Docker
-   OpenTelemetry
-   WebSocket
-   MCP
-   系统设计

------------------------------------------------------------------------

# 四、技术栈

## 前端

-   React
-   Next.js
-   TypeScript
-   TailwindCSS
-   React Flow
-   shadcn/ui
-   Zustand
-   TanStack Query

## Go 服务

-   Go
-   Gin
-   GORM
-   JWT
-   Redis
-   WebSocket
-   OpenTelemetry
-   Zap
-   Viper

## AI Runtime

-   Python
-   FastAPI
-   LangGraph（可替换为自研）
-   OpenAI Compatible API
-   Qwen
-   SentenceTransformer
-   Qdrant

## 数据层

-   PostgreSQL
-   Redis
-   Qdrant

## 可观测

-   OpenTelemetry
-   Jaeger
-   Prometheus
-   Grafana

## 部署

-   Docker
-   Docker Compose
-   Nginx
-   GitHub Actions（后期）

------------------------------------------------------------------------

# 五、整体架构

Browser ↓ Next.js ↓ Go Gateway ↓ Workflow Service ↓ Redis ↓ PostgreSQL ↓
Python Runner ↓ LLM / RAG / Memory / Tool ↓ Qdrant

------------------------------------------------------------------------

# 六、核心模块

## 用户中心

-   登录
-   JWT
-   权限
-   Workspace
-   用户管理

## Workflow Designer

-   拖拽
-   连线
-   保存
-   导入导出
-   版本管理

## Workflow Engine

负责执行：

-   Start
-   End
-   Prompt
-   LLM
-   Condition
-   Loop
-   Tool
-   HTTP
-   RAG
-   Memory

后续支持：

-   MCP
-   Multi-Agent

## AI Runtime

-   Prompt
-   Memory
-   Streaming
-   Tool Calling
-   模型管理

## Knowledge Base

-   文档上传
-   Chunk
-   Embedding
-   Qdrant
-   Hybrid Search
-   RAG

## Redis

负责：

-   Session
-   Workflow 状态
-   Cache
-   Rate Limit
-   Pub/Sub
-   Token Cache

## Monitoring

-   Trace
-   Metrics
-   Logs
-   Jaeger
-   Grafana

## Deployment

-   Docker
-   Docker Compose
-   Nginx

------------------------------------------------------------------------

# 七、开发路线

Phase 1：项目初始化

↓

Phase 2：用户系统

↓

Phase 3：Workflow Designer

↓

Phase 4：Workflow Engine

↓

Phase 5：LLM Runtime

↓

Phase 6：Knowledge Base

↓

Phase 7：Redis

↓

Phase 8：OpenTelemetry

↓

Phase 9：Docker

↓

Phase 10：MCP

------------------------------------------------------------------------

# 八、学习路线

随着项目逐步学习：

Redis

↓

Docker

↓

OpenTelemetry

↓

PostgreSQL

↓

Workflow Engine

↓

LangGraph

↓

MCP

↓

系统设计

------------------------------------------------------------------------

# 九、开发原则

1.  工程优先
2.  高内聚、低耦合
3.  API First
4.  Docker First
5.  配置化
6.  可观测
7.  可扩展
8.  插件化节点
9.  每学习一项新技术即集成到项目

------------------------------------------------------------------------

# 十、最终成果

实现一个生产级 AI Workflow 平台，具备：

-   Workflow
-   RAG
-   Memory
-   Tool Calling
-   Streaming
-   Redis
-   OpenTelemetry
-   Docker
-   MCP
-   一键部署

------------------------------------------------------------------------

# 十一、项目亮点

-   AI Workflow Engine
-   微服务架构
-   Redis 缓存
-   OpenTelemetry 链路追踪
-   Docker Compose 部署
-   PostgreSQL + Qdrant
-   MCP 扩展
-   系统设计能力
