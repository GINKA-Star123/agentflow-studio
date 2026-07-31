# TECH_STACK.md

> 技术栈说明。

## 前端

| 技术 | 用途 |
| --- | --- |
| React | UI 组件模型 |
| Next.js | 前端应用框架和路由 |
| TypeScript | 类型安全 |
| TailwindCSS | 样式系统 |
| React Flow | Workflow 可视化画布 |
| shadcn/ui | 基础 UI 组件 |
| Zustand | 局部客户端状态 |
| TanStack Query | 服务端状态、缓存、请求管理 |

前端原则：

- 业务数据通过 API 获取，不在前端伪造长期状态
- Workflow 画布状态必须能序列化为后端 Workflow 定义
- 组件优先复用 shadcn/ui 和项目内组件
- API 类型应从统一契约生成或集中维护

## Go 服务

| 技术 | 用途 |
| --- | --- |
| Go | 后端主语言 |
| Gin | HTTP API 框架 |
| GORM | PostgreSQL ORM |
| JWT | 用户认证 |
| Redis | 缓存、状态、限流、Pub/Sub |
| WebSocket | 实时执行事件推送 |
| OpenTelemetry | Trace、Metrics、Logs |
| Zap | 结构化日志 |
| Viper | 配置管理 |

Go 服务原则：

- API First，先定义契约再实现业务
- Handler 只做协议转换和参数校验，业务逻辑放到 Service
- Repository 负责数据访问，不在 Handler 中直接操作数据库
- 错误返回统一结构
- 日志必须包含 request_id，Workflow 相关日志必须包含 workflow_id 和 run_id

## AI Runtime

| 技术 | 用途 |
| --- | --- |
| Python | AI Runtime 主语言 |
| FastAPI | Runtime HTTP 服务 |
| LangGraph | Agent / Workflow 执行能力，后续可替换 |
| OpenAI Compatible API | 模型调用统一协议 |
| Qwen | 模型接入方向 |
| SentenceTransformer | Embedding |
| Qdrant | 向量数据库 |

AI Runtime 原则：

- 与 Go 服务通过稳定 HTTP API 通信
- 模型 Provider 通过 Adapter 抽象
- Prompt、Tool Calling、Memory、RAG 的输入输出结构必须可追踪
- Runtime 不直接承担用户权限判断，权限由 Go 服务完成

## 数据层

| 技术 | 用途 |
| --- | --- |
| PostgreSQL | 核心业务数据 |
| Redis | 缓存、状态、消息通道、限流 |
| Qdrant | 向量索引和语义检索 |

数据层原则：

- PostgreSQL 保存权威状态
- Redis 保存短期状态、缓存和事件
- Qdrant 保存向量，不替代业务数据库
- 所有跨存储数据必须有稳定 ID 关联

## 可观测

| 技术 | 用途 |
| --- | --- |
| OpenTelemetry | 统一埋点标准 |
| Jaeger | Trace 查看 |
| Prometheus | Metrics 采集 |
| Grafana | Dashboard |

可观测原则：

- 所有外部请求必须有 Trace
- Workflow Run 必须有根 Span
- 每个节点执行必须有子 Span
- Metrics 必须覆盖请求量、错误率、延迟、Token 用量和节点执行状态

## 部署

| 技术 | 用途 |
| --- | --- |
| Docker | 服务镜像 |
| Docker Compose | 本地和演示环境编排 |
| Nginx | 反向代理 |
| GitHub Actions | 后期 CI/CD |

部署原则：

- Docker First
- 配置通过环境变量注入
- 密钥不入库
- 每个服务必须提供健康检查
