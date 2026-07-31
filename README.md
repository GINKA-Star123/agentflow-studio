# AgentFlow Studio

AgentFlow Studio 是一个面向 AI Agent 的生产级可视化 Workflow 编排平台。

当前阶段：Phase 3 Workflow Designer 已完成，准备进入 Phase 4 Runtime Execution。

Phase 1 的目标是完成基础工程、基础服务、Docker 编排、健康检查和本地开发流程，为后续用户系统、Workflow Designer、Workflow Engine、AI Runtime、RAG、Redis、OpenTelemetry、MCP 等阶段打基础。

Phase 2 的目标是完成用户注册、用户登录、JWT 签发与校验、当前用户接口、Workspace 创建、Workspace 成员关系、Workspace 权限校验和基础 RBAC 行为。

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

## 14. 常见问题

### 14.1 PowerShell 无法执行脚本

执行：

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
```

### 14.2 端口被占用

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

### 14.3 Docker 服务启动失败

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

### 14.4 前端健康页显示后端异常

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

### 14.5 Phase 2 验证脚本失败

先确认 API 已启动：

```powershell
Invoke-RestMethod http://localhost:8080/readyz
```

再确认数据库表已迁移：

```powershell
docker exec -it agentflow-postgres psql -U agentflow -d agentflow -c "\dt"
```

如果邮箱冲突，重新执行 `.\scripts\verify-phase2.ps1` 即可。脚本会使用随机邮箱后缀，正常情况下可以重复运行。
