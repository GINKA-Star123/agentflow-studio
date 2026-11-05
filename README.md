# AgentFlow Studio

**面向 AI Agent 的可视化 Workflow 编排平台** —— 拖拽编排 → DAG 校验 → 节点调度 → 模型调用 → 流式回显 → 全链路可观测。

> 一个可本地一键运行、带完整可观测性与单元测试的 AI Agent 编排引擎实现，用于展示「前端编排 + 后端执行 + AI 运行时 + 云原生部署」的端到端工程能力。

![Go](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go&logoColor=white)
![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=next.js&logoColor=white)
![FastAPI](https://img.shields.io/badge/FastAPI-Python-009688?logo=fastapi&logoColor=white)
![Docker](https://img.shields.io/badge/Docker%20Compose-10%20services-2496ED?logo=docker&logoColor=white)

---

## 一、架构总览

```mermaid
flowchart TB
  subgraph Web["前端 apps/web · Next.js + React Flow"]
    Canvas["可视化编排画布<br/>节点拖拽 / 连线 / 属性面板"]
    RunView["运行态视图<br/>SSE 流式 token 渲染"]
  end

  subgraph API["Go API services/api · Gin + GORM"]
    Auth["JWT 鉴权<br/>RBAC 权限中间件"]
    WF["Workflow CRUD<br/>Schema 版本管理"]
    DAG["DAG 校验<br/>环检测 / 连通性 / 必填参数"]
    Node["节点执行器注册表<br/>Start / Prompt / LLM / Tool / Code / Knowledge"]
    Sched["Run 调度<br/>状态持久化 / 失败回放"]
  end

  subgraph RT["AI Runtime services/ai-runtime · FastAPI"]
    Provider["OpenAI Compatible Provider<br/>多模型参数归一化 / 超时重试"]
    ToolCall["Tool Calling<br/>JSON Schema 描述 → 调用 → 结果归一化"]
  end

  PG[("PostgreSQL<br/>用户 / 工作流 / 运行记录")]
  Redis[("Redis<br/>会话 / 限流")]
  Qdrant[("Qdrant<br/>知识库向量")]
  LLM["OpenAI / DeepSeek / Qwen"]
  Obs["Prometheus · Grafana · Jaeger"]

  Canvas -->|REST| Auth --> WF --> DAG --> Node --> Sched
  Sched -->|HTTP + SSE| Provider --> LLM
  Sched --> ToolCall
  WF --> PG
  Sched --> PG
  Auth --> Redis
  ToolCall --> Qdrant
  API -.埋点.-> Obs
  RT -.埋点.-> Obs
```

## 二、核心能力

| 模块 | 说明 |
| --- | --- |
| 可视化编排 | React Flow 画布，节点拖拽 / 连线校验 / 属性面板动态表单，输出可执行 DAG Schema |
| DAG 校验 | 提交阶段做环检测、连通性、必填参数校验，非法工作流不进执行队列 |
| 节点体系 | 节点执行器注册表，6 类节点（Start / Prompt / LLM / Tool / Code / Knowledge）插件化注册 |
| AI Runtime | 统一多模型接入层，参数规范、超时重试、错误归一化、SSE 流式输出 |
| Tool Calling | JSON Schema 描述工具，串联「调用 → 参数校验 → 结果回填 → 归一化」 |
| 知识库 / RAG | 文档解析、切块、向量化，Qdrant 检索后注入上下文 |
| 多租户与权限 | JWT 鉴权 + Workspace 成员隔离 + RBAC 角色权限 |
| 可观测性 | Prometheus 指标、Grafana 面板、Jaeger 链路追踪，运行轨迹可回放 |

## 三、技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | Next.js 16 · React 19 · TypeScript · TailwindCSS · React Flow · Zustand · TanStack Query |
| Go API | Go · Gin · GORM · JWT · Redis · Zap · Viper |
| AI Runtime | Python · FastAPI · Pydantic · SSE · OpenAI Compatible SDK |
| 存储 | PostgreSQL · Redis · Qdrant |
| 可观测 | Prometheus · Grafana · Jaeger |
| 部署 | Docker · Docker Compose（10 服务）· Nginx |

## 四、快速开始

```bash
git clone https://github.com/GINKA-Star123/agentflow-studio.git
cd agentflow-studio
cp .env.example .env

# 一条命令拉起 Web / Go API / AI Runtime / PostgreSQL / Redis / Qdrant / Nginx / Prometheus / Grafana / Jaeger
docker compose up -d

# 健康检查
curl http://localhost:8080/api/v1/health
```

访问 http://localhost:3000 打开编排画布。

## 五、仓库结构

```text
agentflow-studio/
├── apps/web/               # Next.js 前端：编排画布、工作台、运行态视图
├── services/api/           # Go 服务：鉴权、Workflow、DAG 校验、节点调度
│   ├── cmd/                # 启动入口
│   ├── internal/           # auth / workspace / workflow / engine / handler
│   └── migrations/         # 数据库迁移
├── services/ai-runtime/    # Python FastAPI：多模型接入、Tool Calling、流式输出
├── deployments/docker/     # Prometheus / Grafana / Jaeger 配置
├── docs/                   # 架构、API、数据库、可观测性设计文档
└── scripts/                # 本地开发、健康检查、容器编排脚本
```

## 六、可观测性

- **指标**：Go API 与 AI Runtime 暴露 Prometheus 指标（请求量、延迟、错误率、token_usage）
- **链路**：Jaeger 追踪一次 Workflow Run 的完整调用链，定位到具体节点
- **看板**：Grafana 预置面板，查看 P95 延迟与失败率
- **回放**：运行记录持久化节点状态，失败 Run 可按轨迹逐步回放

## 七、测试

- Go 侧 **205 个测试函数**，覆盖鉴权、密码、DAG 校验、节点调度与错误分支
- Python 侧覆盖接口、知识库解析与工具调用
- 全部用例可离线运行，不依赖真实模型 API

## 八、设计文档

| 文档 | 内容 |
| --- | --- |
| [SYSTEM_ARCHITECTURE.md](docs/SYSTEM_ARCHITECTURE.md) | 系统分层与模块边界 |
| [WORKFLOW_ENGINE.md](docs/WORKFLOW_ENGINE.md) | DAG 执行与节点调度设计 |
| [AI_RUNTIME.md](docs/AI_RUNTIME.md) | 模型接入层与 Tool Calling |
| [DATABASE_DESIGN.md](docs/DATABASE_DESIGN.md) | 数据模型与表结构 |
| [OBSERVABILITY.md](docs/OBSERVABILITY.md) | 指标、链路与告警设计 |
| [API_DESIGN.md](docs/API_DESIGN.md) | 接口规范与错误码 |
| [DEPLOYMENT.md](docs/DEPLOYMENT.md) | 部署与运维 |

---

<details>
<summary>📜 开发阶段记录（Phase 1–5，点击展开）</summary>

AgentFlow Studio 是一个面向 AI Agent 的生产级可视化 Workflow 编排平台。

当前阶段：Phase 5 AI Runtime / Tools / Streaming 已完成，准备进入 Phase 6 Knowledge Base。

Phase 1 的目标是完成基础工程、基础服务、Docker 编排、健康检查和本地开发流程，为后续用户系统、Workflow Designer、Workflow Engine、AI Runtime、RAG、Redis、OpenTelemetry、MCP 等阶段打基础。

Phase 2 的目标是完成用户注册、用户登录、JWT 签发与校验、当前用户接口、Workspace 创建、Workspace 成员关系、Workspace 权限校验和基础 RBAC 行为。

</details>

## 1. 项目定位

AgentFlow Studio 支持：

- 可视化 Workflow 编排
- 多 Agent 协同
- RAG
- Memory
- Tool Calling
- Streaming
- Workflow Trace
- Token 统计
- Latency 统计
- 执行日志
- Docker 一键部署
- OpenTelemetry 可观测

参考产品：

- Dify
- Coze
- LangGraph Studio
- n8n

项目目标不是复刻现有产品，而是打造一个更偏工程化、更适合作为 AI 全栈项目展示的平台。

## 2. 技术栈

前端：

- React
- Next.js
- TypeScript
- TailwindCSS
- React Flow
- shadcn/ui
- Zustand
- TanStack Query

Go API：

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
- LangGraph，后续可替换为自研 Runtime
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

## 3. 环境要求

本地需要安装：

- Node.js >= 20
- npm >= 10
- Go >= 1.22
- Python >= 3.11
- Docker Desktop
- Docker Compose v2
- PowerShell

检查命令：

```powershell
node -v
npm -v
go version
python --version
pip --version
docker --version
docker compose version
git --version
```

## 4. 项目结构

```text
agentflow-studio/
  apps/
    web/                         # Next.js 前端应用

  services/
    api/                         # Go API 服务
    ai-runtime/                  # Python AI Runtime 服务

  deployments/
    docker/                      # Docker Compose、Nginx、Prometheus 配置

  scripts/                       # 本地启动、Docker 启动、健康检查、日志脚本

  docs/
    PROJECT_BLUEPRINT.md         # 项目蓝图
    ROADMAP.md                   # 开发路线
    TECH_STACK.md                # 技术栈说明
    SYSTEM_ARCHITECTURE.md       # 系统架构
    DATABASE_DESIGN.md           # 数据库设计
    API_DESIGN.md                # API 设计
    WORKFLOW_ENGINE.md           # Workflow 引擎设计
    AI_RUNTIME.md                # AI Runtime 设计
    OBSERVABILITY.md             # 可观测设计
    DEPLOYMENT.md                # 部署设计
    CODING_STANDARD.md           # 编码规范
    TODO.md                      # 开发任务清单
```

## 5. 环境变量配置

Phase 1 统一使用项目根目录 `.env` 或 `.env.example` 维护配置。

建议先创建 `.env.example`，再复制为 `.env`：

```powershell
Copy-Item .env.example .env
```

### 5.1 统一 `.env.example`

```env
# App
APP_ENV=dev

# Web
NEXT_PUBLIC_APP_NAME=AgentFlow Studio
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8080

# Next.js Server Side
API_BASE_URL=http://localhost:8080
AI_RUNTIME_BASE_URL=http://localhost:8090

# Go API
API_HTTP_PORT=8080
DATABASE_URL=postgres://agentflow:agentflow@localhost:5432/agentflow?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
JWT_SECRET=change-me
JWT_ISSUER=agentflow-studio
JWT_ACCESS_TOKEN_TTL=2h
AI_RUNTIME_URL=http://localhost:8090
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# AI Runtime
AI_RUNTIME_HTTP_PORT=8090
OPENAI_COMPATIBLE_BASE_URL=https://api.openai.com/v1
OPENAI_COMPATIBLE_API_KEY=change-me
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=

# PostgreSQL
POSTGRES_DB=agentflow
POSTGRES_USER=agentflow
POSTGRES_PASSWORD=agentflow

# Redis
REDIS_PORT=6379

# Qdrant
QDRANT_PORT=6333
```

### 5.2 前端环境变量说明

```env
NEXT_PUBLIC_APP_NAME=AgentFlow Studio
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_BASE_URL=ws://localhost:8080

API_BASE_URL=http://localhost:8080
AI_RUNTIME_BASE_URL=http://localhost:8090
```

说明：

- `NEXT_PUBLIC_APP_NAME`：浏览器端可读取的应用名称。
- `NEXT_PUBLIC_API_BASE_URL`：浏览器端可读取的 API 基础地址。
- `NEXT_PUBLIC_WS_BASE_URL`：浏览器端可读取的 WebSocket 基础地址。
- `API_BASE_URL`：Next.js 服务端访问 Go API 使用，不暴露给浏览器。
- `AI_RUNTIME_BASE_URL`：Next.js 服务端访问 AI Runtime 使用，不暴露给浏览器。

### 5.3 Go API 环境变量说明

```env
APP_ENV=dev
API_HTTP_PORT=8080
DATABASE_URL=postgres://agentflow:agentflow@localhost:5432/agentflow?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
JWT_SECRET=change-me
JWT_ISSUER=agentflow-studio
JWT_ACCESS_TOKEN_TTL=2h
AI_RUNTIME_URL=http://localhost:8090
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

说明：

- `APP_ENV`：运行环境，常用值为 `dev`、`docker`、`prod`。
- `API_HTTP_PORT`：Go API HTTP 服务端口。
- `DATABASE_URL`：PostgreSQL 连接地址。
- `REDIS_ADDR`：Redis 地址。
- `REDIS_PASSWORD`：Redis 密码，本地开发可以为空。
- `JWT_SECRET`：JWT 签名密钥，生产环境必须替换。
- `JWT_ISSUER`：JWT 签发者，用于校验 Token 来源。
- `JWT_ACCESS_TOKEN_TTL`：Access Token 有效期，示例值为 `2h`。
- `AI_RUNTIME_URL`：Python AI Runtime 服务地址。
- `OTEL_EXPORTER_OTLP_ENDPOINT`：OpenTelemetry OTLP 上报地址。

### 5.4 AI Runtime 环境变量说明

```env
APP_ENV=dev
AI_RUNTIME_HTTP_PORT=8090
OPENAI_COMPATIBLE_BASE_URL=https://api.openai.com/v1
OPENAI_COMPATIBLE_API_KEY=change-me
QDRANT_URL=http://localhost:6333
QDRANT_API_KEY=
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

说明：

- `APP_ENV`：运行环境。
- `AI_RUNTIME_HTTP_PORT`：AI Runtime HTTP 服务端口。
- `OPENAI_COMPATIBLE_BASE_URL`：OpenAI Compatible API 地址。
- `OPENAI_COMPATIBLE_API_KEY`：模型服务 API Key。
- `QDRANT_URL`：Qdrant 服务地址。
- `QDRANT_API_KEY`：Qdrant API Key，本地开发可以为空。
- `OTEL_EXPORTER_OTLP_ENDPOINT`：OpenTelemetry OTLP 上报地址。

## 6. 本地开发启动

### 6.1 启动基础依赖

基础依赖包括：

- PostgreSQL
- Redis
- Qdrant
- Jaeger

```powershell
.\scripts\dev-deps.ps1
```

等价命令：

```powershell
docker compose -f deployments\docker\docker-compose.yml up -d postgres redis qdrant jaeger
```

### 6.2 启动 Web 前端

```powershell
cd apps\web
npm run dev
```

访问地址：

```text
http://localhost:3000
```

### 6.3 启动 Go API

```powershell
cd services\api
go run .\cmd\api
```

健康检查：

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/readyz
```

### 6.4 启动 AI Runtime

```powershell
cd services\ai-runtime
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload
```

健康检查：

```powershell
Invoke-RestMethod http://localhost:8090/healthz
Invoke-RestMethod http://localhost:8090/readyz
```

### 6.5 一键启动本地开发服务

```powershell
.\scripts\dev-all.ps1
```

跳过基础依赖，只启动应用服务：

```powershell
.\scripts\dev-all.ps1 -SkipDeps
```

## 7. Docker 启动

### 7.1 构建并启动完整环境

```powershell
.\scripts\docker-up.ps1
```

等价命令：

```powershell
docker compose -f deployments\docker\docker-compose.yml up --build -d
```

### 7.2 查看服务状态

```powershell
docker compose -f deployments\docker\docker-compose.yml ps
```

### 7.3 查看日志

查看全部服务日志：

```powershell
.\scripts\logs.ps1
```

持续查看全部服务日志：

```powershell
.\scripts\logs.ps1 -Follow
```

查看指定服务日志：

```powershell
.\scripts\logs.ps1 -Service web -Follow
.\scripts\logs.ps1 -Service api -Follow
.\scripts\logs.ps1 -Service ai-runtime -Follow
```

### 7.4 停止 Docker 环境

```powershell
.\scripts\docker-down.ps1
```

停止并删除数据卷：

```powershell
.\scripts\docker-down.ps1 -Volumes
```

等价命令：

```powershell
docker compose -f deployments\docker\docker-compose.yml down
docker compose -f deployments\docker\docker-compose.yml down -v
```

## 8. 健康检查

### 8.1 本地模式健康检查

```powershell
.\scripts\health-check.ps1 -Mode local
```

检查内容：

- Web 前端
- Go API
- AI Runtime

### 8.2 Docker 模式健康检查

```powershell
.\scripts\health-check.ps1 -Mode docker
```

检查内容：

- Nginx 聚合入口
- Web 前端
- Go API
- AI Runtime

### 8.3 手动健康检查命令

```powershell
Invoke-RestMethod http://localhost:3000/api/health
Invoke-RestMethod http://localhost:8080/readyz
Invoke-RestMethod http://localhost:8090/readyz
Invoke-RestMethod http://localhost/api/health
```

## 9. 常用访问地址

本地开发：

```text
Web 前端：
http://localhost:3000

Go API：
http://localhost:8080

AI Runtime：
http://localhost:8090

Jaeger：
http://localhost:16686
```

Docker 模式：

```text
Nginx 统一入口：
http://localhost

Web 前端：
http://localhost:3000

Go API：
http://localhost:8080

AI Runtime：
http://localhost:8090

PostgreSQL：
localhost:5432

Redis：
localhost:6379

Qdrant：
http://localhost:6333

Jaeger：
http://localhost:16686

Prometheus：
http://localhost:9090

Grafana：
http://localhost:3001
```

## 10. Phase 2 启动与验证

Phase 2 主要后端能力位于 `services/api`。

### 10.1 启动基础依赖

```powershell
.\scripts\dev-deps.ps1
```

等价命令：

```powershell
docker compose -f deployments\docker\docker-compose.yml up -d postgres redis qdrant jaeger
```

### 10.2 执行数据库迁移

```powershell
Get-Content -Raw services\api\migrations\000001_create_auth_workspace_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow
```

检查表结构：

```powershell
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\dt"
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\d users"
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\d workspaces"
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\d workspace_members"
```

### 10.3 启动 Go API

```powershell
cd services\api
go run .\cmd\api
```

### 10.4 运行后端测试

```powershell
cd services\api
go test ./...
```

只运行 Auth 测试：

```powershell
cd services\api
go test ./internal/auth
```

只运行 Workspace RBAC 测试：

```powershell
cd services\api
go test ./internal/service -run Workspace
```

### 10.5 运行 Phase 2 接口验证脚本

```powershell
.\scripts\verify-phase2.ps1
```

指定 API 地址：

```powershell
.\scripts\verify-phase2.ps1 -BaseUrl "http://localhost:8080/api/v1"
```

验证脚本覆盖：

- 未携带 Token 访问受保护接口失败
- 用户注册
- 用户登录
- `/auth/me`
- Workspace 列表
- Workspace 创建
- 成员列表
- 添加已有用户为成员
- 重复添加成员失败
- 成员角色更新
- viewer 无法添加成员
- owner 移除成员
- 被移除成员无法继续访问原 Workspace

## 11. Phase 2 API 快速验证

### 11.1 注册用户

```powershell
$registerBody = @{
  email = "demo@example.com"
  password = "password123"
  display_name = "Demo"
  workspace_name = "Demo Workspace"
} | ConvertTo-Json

$registerResult = Invoke-RestMethod `
  -Uri http://localhost:8080/api/v1/auth/register `
  -Method Post `
  -Body $registerBody `
  -ContentType "application/json"

$token = $registerResult.data.access_token
$workspaceID = $registerResult.data.current_workspace.id
```

### 11.2 登录用户

```powershell
$loginBody = @{
  email = "demo@example.com"
  password = "password123"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri http://localhost:8080/api/v1/auth/login `
  -Method Post `
  -Body $loginBody `
  -ContentType "application/json"
```

### 11.3 获取当前用户

```powershell
Invoke-RestMethod `
  -Uri http://localhost:8080/api/v1/auth/me `
  -Method Get `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.4 查询 Workspace 列表

```powershell
Invoke-RestMethod `
  -Uri http://localhost:8080/api/v1/workspaces `
  -Method Get `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.5 创建 Workspace

```powershell
$workspaceBody = @{
  name = "Second Workspace"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri http://localhost:8080/api/v1/workspaces `
  -Method Post `
  -Body $workspaceBody `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.6 查询 Workspace 成员

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/workspaces/$workspaceID/members" `
  -Method Get `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.7 添加已有用户为成员

```powershell
$memberBody = @{
  email = "member@example.com"
  role = "member"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/workspaces/$workspaceID/members" `
  -Method Post `
  -Body $memberBody `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.8 更新成员角色

```powershell
$roleBody = @{
  role = "viewer"
} | ConvertTo-Json

Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/workspaces/$workspaceID/members/$targetUserID/role" `
  -Method Patch `
  -Body $roleBody `
  -ContentType "application/json" `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

### 11.9 移除成员

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/workspaces/$workspaceID/members/$targetUserID" `
  -Method Delete `
  -Headers @{
    Authorization = "Bearer $token"
  }
```

## 12. Phase 2 最终验收清单

基础文档：

- [x] `docs/PROJECT_BLUEPRINT.md` 已完成
- [x] `docs/ROADMAP.md` 已完成
- [x] `docs/TECH_STACK.md` 已完成
- [x] `docs/SYSTEM_ARCHITECTURE.md` 已完成
- [x] `docs/DATABASE_DESIGN.md` 已完成
- [x] `docs/API_DESIGN.md` 已更新 Phase 2 API
- [x] `docs/WORKFLOW_ENGINE.md` 已完成
- [x] `docs/AI_RUNTIME.md` 已完成
- [x] `docs/OBSERVABILITY.md` 已完成
- [x] `docs/DEPLOYMENT.md` 已完成
- [x] `docs/CODING_STANDARD.md` 已完成
- [x] `docs/TODO.md` 已更新 Phase 2 状态

后端能力：

- [x] 用户表、Workspace 表、Workspace 成员表已设计
- [x] PostgreSQL 迁移已准备
- [x] 用户注册已实现
- [x] 用户登录已实现
- [x] 密码 bcrypt 哈希与校验已实现
- [x] JWT 签发与校验已实现
- [x] JWT Middleware 已实现
- [x] `/api/v1/auth/me` 当前用户接口已实现
- [x] Workspace 创建已实现
- [x] Workspace 列表已实现
- [x] Workspace 成员列表已实现
- [x] 添加已有用户为 Workspace 成员已实现
- [x] 成员角色更新已实现
- [x] 成员移除已实现
- [x] owner/admin/member/viewer 基础 RBAC 已实现

权限规则：

- [x] 受保护接口必须携带 Bearer Token
- [x] 非 Workspace 成员不能访问 Workspace 成员资源
- [x] owner 可以管理 admin/member/viewer
- [x] admin 可以管理 member/viewer
- [x] member 和 viewer 不能管理成员
- [x] 不能修改自己的角色
- [x] 不能移除自己
- [x] 不能通过普通接口设置 owner
- [x] 不能移除 owner

测试与验证：

- [x] 密码哈希测试已准备
- [x] JWT 测试已准备
- [x] Workspace RBAC 测试已准备
- [x] Phase 2 接口验证脚本已准备
- [ ] `go test ./...` 通过
- [ ] `.\scripts\verify-phase2.ps1` 通过

最终验证命令：

```powershell
.\scripts\dev-deps.ps1
```

```powershell
Get-Content -Raw services\api\migrations\000001_create_auth_workspace_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow
```

```powershell
cd services\api
go test ./...
go run .\cmd\api
```

```powershell
.\scripts\verify-phase2.ps1
```

## 13. Phase 3 Workflow Designer 启动与验证

Phase 3 已完成，当前 Workflow Designer 已支持：

- 节点拖拽添加
- 节点连线、选择、删除
- 节点属性编辑
- Prompt / LLM 配置编辑
- Workflow Schema 生成与基础校验
- Workflow 列表读取、详情加载和保存

### 13.1 本地启动顺序

1. 启动基础依赖：

```powershell
.\scripts\dev-deps.ps1
```

2. 执行 Workflow 数据库迁移：

```powershell
Get-Content -Raw services\api\migrations\000002_create_workflows_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow
```

3. 启动 Go API：

```powershell
cd services\api
go run .\cmd\api
```

4. 启动前端：

```powershell
cd apps\web
npm run dev
```

### 13.2 Phase 3 验证命令

```powershell
cd services\api
go test ./...
```

```powershell
cd apps\web
npm run lint
npm run build
```

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:8080/readyz
Invoke-RestMethod http://localhost:3000/api/health
```

```powershell
Invoke-RestMethod `
  -Uri "http://localhost:8080/api/v1/workspaces/<workspace_id>/workflows" `
  -Headers @{ Authorization = "Bearer <access_token>" }
```

### 13.3 Phase 3 最终验收清单

- [x] 画布可拖拽添加节点
- [x] 画布可连线、选择、删除
- [x] 右侧属性面板可编辑 `label` 和 `description`
- [x] Prompt 节点支持 `promptTemplate` 和 `variables`
- [x] LLM 节点支持 `provider`、`model`、`temperature`、`maxTokens`、`systemPrompt`
- [x] Workflow Schema 可从 `nodes / edges` 生成
- [x] Workflow Schema 可做基础校验
- [x] Workflow 列表、加载和保存可用
- [x] 前后端 Workflow API 设计与实现保持一致
- [x] Phase 3 文档已同步更新

### 13.4 Phase 4 Runtime Execution 进入条件

- [x] Phase 3 验收清单全部完成
- [x] `go test ./...` 通过
- [x] `npm run lint` 通过
- [x] `npm run build` 通过
- [x] Workflow 保存、加载、列表稳定可用
- [x] `docs/API_DESIGN.md` 与代码保持一致
- [x] `docs/TODO.md` 已标记 Phase 3 完成
- [x] 准备开始 Workflow Engine / Runtime Execution

## 14. Phase 4 Runtime Execution 启动与验证

Phase 4 已完成 Workflow Runtime Execution 的第一版闭环，当前支持：

- Workflow Run 数据表与节点执行记录表
- Workflow Schema Go 类型与 JSON 解析
- Workflow DAG 校验
- Start / End / Prompt / LLM 节点执行器
- AI Runtime Client 骨架
- Workflow Runner Service 同步执行
- Workflow Run API
- 前端 Designer 运行入口
- Run 状态、节点执行列表、LLM 输出、token usage、latency 展示

### 14.1 本地启动顺序

1. 启动基础依赖：

```powershell
.\scripts\dev-deps.ps1
```

2. 执行数据库迁移：

```powershell
Get-Content -Raw services\api\migrations\000001_create_auth_workspace_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow

Get-Content -Raw services\api\migrations\000002_create_workflows_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow

Get-Content -Raw services\api\migrations\000003_create_workflow_runtime_tables.up.sql |
docker exec -i agentflow-postgres psql -U agentflow -d agentflow
```

3. 启动 AI Runtime：

```powershell
cd services\ai-runtime
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload
```

4. 启动 Go API：

```powershell
cd services\api
go run .\cmd\api
```

5. 启动 Web 前端：

```powershell
cd apps\web
npm run dev
```

### 14.2 Phase 4 API 快速验证

发起 Workflow Run：

```powershell
$runBody = @{
  input = @{
    message = "hello from phase 4"
  }
  trace_id = "trace_phase4_manual"
} | ConvertTo-Json -Depth 10

$runResult = Invoke-RestMethod `
  -Method POST `
  -Uri "http://localhost:8080/api/v1/workspaces/<workspace_id>/workflows/<workflow_id>/runs" `
  -Headers @{ Authorization = "Bearer <access_token>" } `
  -ContentType "application/json" `
  -Body $runBody

$runID = $runResult.data.id
```

查询 Run 详情：

```powershell
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8080/api/v1/workspaces/<workspace_id>/workflow-runs/$runID" `
  -Headers @{ Authorization = "Bearer <access_token>" }
```

查询节点执行记录：

```powershell
Invoke-RestMethod `
  -Method GET `
  -Uri "http://localhost:8080/api/v1/workspaces/<workspace_id>/workflow-runs/$runID/nodes" `
  -Headers @{ Authorization = "Bearer <access_token>" }
```

取消 Run：

```powershell
Invoke-RestMethod `
  -Method POST `
  -Uri "http://localhost:8080/api/v1/workspaces/<workspace_id>/workflow-runs/$runID/cancel" `
  -Headers @{ Authorization = "Bearer <access_token>" }
```

### 14.3 前端验证入口

保存 Workflow 后，进入 Designer 页面时带上 Workflow ID：

```text
http://localhost:3000/dashboard/workflows?workflowId=<workflow_id>
```

在 Designer 中可以：

- 输入 Run input JSON
- 点击运行
- 查看 Run 状态
- 查看节点执行列表
- 查看 LLM `response_text`
- 查看 `token_usage`
- 查看 `latency_ms`

### 14.4 Phase 4 最终验收清单

- [x] `workflow_runs` 表已创建
- [x] `node_executions` 表已创建
- [x] Workflow Schema Go 类型已实现
- [x] Workflow Schema JSON 解析已实现
- [x] Workflow DAG 校验已实现
- [x] NodeExecutor 接口已实现
- [x] ExecutionContext 已实现
- [x] ExecutorRegistry 已实现
- [x] Start Executor 已实现
- [x] End Executor 已实现
- [x] Prompt Executor 已实现
- [x] AI Runtime Client 已实现
- [x] LLM Executor 骨架已实现
- [x] Workflow Runner Service 已实现
- [x] Run 状态持久化已实现
- [x] Node Execution 持久化已实现
- [x] Workflow Run API 已实现
- [x] RuntimeError 到 API 响应转换已实现
- [x] 前端 Run 按钮已实现
- [x] 前端 Run 状态展示已实现
- [x] 前端节点执行列表展示已实现
- [x] 前端 LLM 输出、token usage、latency 展示已实现

### 14.5 Phase 5 AI Runtime / Tools / Streaming 进入条件（已满足）

- [x] Phase 4 Runtime Execution 后端主链路完成
- [x] Phase 4 Runtime Execution 前端运行入口完成
- [x] Workflow Run API 已写入 `docs/API_DESIGN.md`
- [x] Runtime 当前实现已写入 `docs/WORKFLOW_ENGINE.md`
- [x] Phase 4 TODO 已勾选
- [x] 最终验证命令已整理
- [x] AI Runtime `/internal/v1/llm/chat` 已准备进入真实 Provider 接入
- [x] Streaming、Tool Calling、事件推送进入 Phase 5 实现范围

### 14.6 Phase 4 当前边界

Phase 4 的目标是打通 Runtime Execution 第一版闭环，而不是完成完整生产级执行系统。当前边界如下：

- Runner 当前为同步执行，API 请求会等待本次 Workflow 执行完成或失败。
- Workflow Run 已持久化到 `workflow_runs`，节点执行已持久化到 `node_executions`。
- 当前可执行节点为 Start / End / Prompt / LLM。
- LLM 节点已通过 Go API 的 AI Runtime Client 调用 `/internal/v1/llm/chat`，真实 Provider、流式输出和 Tool Calling 进入 Phase 5。
- Cancel API 已提供，但同步 Runner 下主要用于后续异步执行预留。
- WebSocket 事件推送、Redis Pub/Sub、异步队列、运行中取消和实时前端事件流进入 Phase 5 及后续阶段。

## 15. Phase 5 AI Runtime / Tools / Streaming 启动与验证

Phase 5 已完成当前版本的 AI Runtime 闭环能力，包含：

- FastAPI AI Runtime 服务
- OpenAI Compatible Provider
- `POST /internal/v1/llm/chat`
- `POST /internal/v1/llm/stream`
- Tool Calling 数据结构与 `tool` role message
- Go Workflow Tool bridge 与 mock tool executor
- 前端流式输出面板
- Phase 5 验证脚本

### 15.1 本地启动顺序

1. 启动基础依赖：

```powershell
.\scripts\dev-deps.ps1
```

2. 启动 AI Runtime：

```powershell
cd services\ai-runtime
.\.venv\Scripts\Activate.ps1
uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload
```

3. 启动 Go API：

```powershell
cd services\api
go run .\cmd\api
```

4. 启动 Web 前端：

```powershell
cd apps\web
npm run dev
```

### 15.2 Phase 5 验证命令

推荐执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\verify-phase5.ps1
```

单项验证：

```powershell
cd services\api
go test ./internal/workflowruntime ./internal/service

cd services\ai-runtime
.\.venv\Scripts\python.exe -m unittest discover -s tests -p "test_*.py"
```

### 15.3 Phase 5 最终验收清单

- [x] AI Runtime FastAPI 服务已启动
- [x] OpenAI Compatible Provider 已注册
- [x] `POST /internal/v1/llm/chat` 已返回 `response_text`、`message`、`tool_calls`、`token_usage`
- [x] `POST /internal/v1/llm/stream` 已返回 `start`、`delta`、`tool_call_delta`、`usage`、`done` 事件
- [x] Tool Calling 数据结构与 `tool` role message 已定义
- [x] Go Workflow Tool bridge 已接入
- [x] `tool_calls` 已在 LLM 节点输出中保留
- [x] 前端流式输出面板已接入
- [x] `scripts/verify-phase5.ps1` 已补充

### 15.4 Phase 6 Knowledge Base 进入条件

- [x] Phase 5 核心链路已完成
- [x] Phase 5 文档与验证脚本已整理
- [ ] Knowledge Base 表结构设计完成
- [ ] 文档上传、解析、Chunk 切分方案完成
- [ ] Embedding API 与 Qdrant 接入方案完成
- [ ] RAG 节点上下文组装方案完成

## 16. 常见问题

### 16.1 PowerShell 无法执行脚本

执行：

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

### 16.2 端口被占用

检查端口：

```powershell
netstat -ano | findstr :3000
netstat -ano | findstr :8080
netstat -ano | findstr :8090
```

常用端口：

```text
3000：Web 前端
8080：Go API
8090：AI Runtime
5432：PostgreSQL
6379：Redis
6333：Qdrant
16686：Jaeger
9090：Prometheus
3001：Grafana
80：Nginx
```

### 16.3 Docker 服务启动失败

查看服务状态：

```powershell
docker compose -f deployments\docker\docker-compose.yml ps
```

查看日志：

```powershell
.\scripts\logs.ps1 -Follow
```

重新构建：

```powershell
docker compose -f deployments\docker\docker-compose.yml up --build -d
```

### 16.4 前端健康页显示后端异常

需要确认 Go API 和 AI Runtime 已启动：

```powershell
Invoke-RestMethod http://localhost:8080/readyz
Invoke-RestMethod http://localhost:8090/readyz
```

如果这两个接口正常，再检查前端 `.env.local` 或 Docker Compose 中的：

```env
API_BASE_URL
AI_RUNTIME_BASE_URL
```

### 16.5 Phase 2 验证脚本失败

先确认 API 已启动：

```powershell
Invoke-RestMethod http://localhost:8080/readyz
```

再确认数据库表已迁移：

```powershell
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\dt"
```

如果邮箱冲突，重新执行 `.\scripts\verify-phase2.ps1` 即可。脚本会使用随机邮箱后缀，正常情况下可以重复运行。