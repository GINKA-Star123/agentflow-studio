# ROADMAP.md

> 开发路线，严格按 Phase 推进。

## Phase 1：项目初始化

目标：建立可运行、可扩展、可持续演进的基础工程。

任务：

- 初始化仓库目录结构
- 初始化前端 Next.js 应用
- 初始化 Go API 服务
- 初始化 Python AI Runtime 服务
- 提供基础 Docker Compose
- 建立统一配置规范
- 建立日志规范
- 建立基础健康检查

验收标准：

- 前端、Go 服务、Python 服务均可本地启动
- `docker compose up` 可以启动基础依赖
- 项目根目录包含清晰 README 或启动说明
- 每个服务都有独立配置入口

## Phase 2：用户系统

目标：支持用户认证、授权和 Workspace 隔离。

任务：

- 用户注册与登录
- JWT 签发与校验
- 密码哈希存储
- Workspace 创建与切换
- 用户与 Workspace 关系建模
- 基础 RBAC 预留

验收标准：

- 登录后可访问受保护 API
- 不同 Workspace 数据隔离
- 认证失败返回统一错误结构

## Phase 3：Workflow Designer

目标：实现可视化 Workflow 设计器。

任务：

- 使用 React Flow 实现画布
- 支持节点拖拽、连线、删除
- 支持节点属性编辑
- 支持 Workflow 保存和读取
- 支持 Workflow JSON 导入导出
- 支持版本记录

验收标准：

- 用户可以创建一个包含 Start、Prompt、LLM、End 的流程
- 流程可以保存到后端并重新加载
- 前端状态和后端数据结构一致

## Phase 4：Workflow Engine

目标：实现后端 Workflow 执行引擎。

任务：

- 定义 Workflow DAG 数据结构
- 实现节点拓扑校验
- 实现 Start、End、Prompt、LLM 基础节点执行
- 实现 Condition、Loop 控制流
- 实现 HTTP、Tool、Memory、RAG 节点接口
- 记录节点执行状态、输入、输出、错误和耗时
- 支持执行暂停、失败、重试的状态预留

验收标准：

- 后端可以执行保存后的 Workflow
- 每次执行都有 Run ID
- 每个节点都有可查询的执行日志
- 非法 Workflow 会在执行前返回明确错误

## Phase 5：LLM Runtime

目标：接入 AI Runtime，支持模型调用、流式输出和工具调用。

任务：

- Python FastAPI Runtime 服务
- OpenAI Compatible API 适配
- Qwen 适配预留
- Prompt 模板渲染
- Streaming 输出
- Tool Calling 协议设计
- Token 用量统计

验收标准：

- Go 服务可以调用 AI Runtime
- 前端可以看到流式输出
- LLM 节点保存 token、latency 和模型信息

## Phase 6：Knowledge Base

目标：实现知识库、Embedding 和 RAG 节点。

任务：

- 文档上传
- 文档解析
- Chunk 切分
- Embedding 生成
- Qdrant 写入与检索
- RAG 上下文组装
- Hybrid Search 预留

验收标准：

- 用户可以上传文档并创建知识库
- RAG 节点可以从知识库召回上下文
- 检索结果可追踪到原始文档片段

## Phase 7：Redis

目标：把 Redis 集成到状态、缓存、限流和事件通道中。

任务：

- Session 或 Token Cache
- Workflow Run 状态缓存
- Rate Limit
- Pub/Sub 推送执行事件
- 热点数据缓存

验收标准：

- Redis 键名有统一命名规范
- Workflow 执行事件可以通过 Pub/Sub 分发
- 关键缓存有 TTL

## Phase 8：OpenTelemetry

目标：建立 Trace、Metrics、Logs 三位一体的可观测能力。

任务：

- Go 服务接入 OpenTelemetry
- Python Runtime 接入 OpenTelemetry
- HTTP 请求 Trace
- Workflow Run Trace
- 节点执行 Span
- Prometheus 指标
- Jaeger 链路查看
- Grafana 面板预留

验收标准：

- 每次 Workflow Run 都能看到完整 Trace
- 节点执行耗时、错误数和 token 用量可观测
- 日志包含 trace_id 和 run_id

## Phase 9：Docker

目标：完成本地和演示环境的一键部署。

任务：

- 前端 Dockerfile
- Go 服务 Dockerfile
- Python Runtime Dockerfile
- Docker Compose 编排
- Nginx 反向代理
- 环境变量模板

验收标准：

- 一条命令启动完整系统
- 服务间网络、端口和健康检查明确
- 不把密钥写入镜像或仓库

## Phase 10：MCP

目标：支持 MCP 扩展，为外部工具和 Agent 能力开放标准接口。

任务：

- MCP Server 设计
- MCP Tool 注册
- MCP Tool 调用
- 权限和审计
- 与 Workflow Tool 节点集成

验收标准：

- Workflow 可以调用 MCP Tool
- MCP 调用有日志、权限校验和错误处理
- MCP 能力与普通 Tool Calling 统一抽象
