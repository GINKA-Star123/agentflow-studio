# CODING_STANDARD.md

> 项目规范，后续代码编写必须遵守。

## 总原则

1. 工程优先
2. 高内聚、低耦合
3. API First
4. Docker First
5. 配置化
6. 可观测
7. 可扩展
8. 插件化节点
9. 每学习一项新技术即集成到项目

## 目录原则

建议目录：

```text
apps/
  web/
services/
  api/
  ai-runtime/
deployments/
  docker/
docs/
scripts/
```

规则：

- 前端代码放在 `apps/web`
- Go API 放在 `services/api`
- Python AI Runtime 放在 `services/ai-runtime`
- 部署文件放在 `deployments`
- 项目文档优先放在根目录或后续迁移到 `docs`

## API 规范

- API 路径以 `/api/v1` 开头
- 内部 Runtime API 路径以 `/internal/v1` 开头
- 请求和响应使用 JSON
- 错误结构统一
- Handler 不直接写复杂业务逻辑
- API 参数必须校验
- Workspace 资源必须做权限校验

## Go 编码规范

分层建议：

```text
cmd/
internal/
  config/
  server/
  middleware/
  handler/
  service/
  repository/
  model/
  workflow/
  telemetry/
```

规则：

- 使用 `context.Context` 传递请求上下文
- Handler 负责协议转换
- Service 负责业务逻辑
- Repository 负责数据访问
- 不在全局变量中保存可变业务状态
- 错误要包装上下文，但返回给用户时要转为统一错误码
- 日志使用结构化字段
- Workflow 相关日志必须包含 workflow_id、run_id、node_id

测试：

- 核心 Service 必须有单元测试
- Workflow Engine 必须覆盖 DAG 校验和节点执行
- Repository 测试可使用测试数据库或接口抽象

## Python 编码规范

分层建议：

```text
app/
  api/
  core/
  providers/
  rag/
  tools/
  memory/
  telemetry/
```

规则：

- FastAPI 路由层只做参数校验和调用服务
- Provider 使用 Adapter 封装
- 外部请求必须设置 timeout
- AI Runtime 错误必须结构化
- 不在日志中输出密钥和完整敏感 Prompt
- RAG 查询必须带 workspace_id 过滤

测试：

- Provider Adapter 使用 mock 测试
- RAG 检索逻辑要覆盖过滤条件
- Prompt 渲染要覆盖变量缺失场景

## 前端编码规范

规则：

- 使用 TypeScript
- UI 组件优先复用 shadcn/ui
- Workflow 画布使用 React Flow
- 服务端状态使用 TanStack Query
- 复杂客户端状态使用 Zustand
- API 请求集中封装
- 不在组件中散落 fetch 逻辑
- 表单必须有校验和错误提示
- 执行态页面必须支持 loading、empty、error、success 状态

设计规则：

- SaaS / 工作台风格以清晰、克制、高效为主
- 不做营销落地页作为主界面
- 常用操作使用明确按钮和图标
- 文本不能在移动端和桌面端溢出容器
- Workflow Designer 优先保证画布、侧边栏、属性面板和执行日志的工作效率

## 数据库规范

- 使用 Migration 管理结构变更
- 表名使用复数 snake_case
- 字段使用 snake_case
- 主键优先 UUID
- 所有 Workspace 资源表必须包含 workspace_id
- JSONB 字段必须有 Schema 约束或代码校验
- 删除业务资源优先软删除

## Redis 规范

Key 格式：

```text
af:{env}:{domain}:{resource}:{id}
```

规则：

- 缓存必须设置 TTL
- 不能把 Redis 作为权威数据库
- Pub/Sub 事件必须有明确 schema
- 限流 key 必须区分用户或 Workspace

## 可观测规范

- 所有 HTTP 请求必须有 request_id
- 所有跨服务调用必须传递 trace_id
- Workflow Run 必须创建 Trace
- Node Execution 必须创建 Span
- 日志必须结构化
- 敏感字段必须脱敏

## 配置规范

- 配置来自环境变量和配置文件
- 本地提供 `.env.example`
- 密钥不得提交
- 默认值只能用于本地开发
- 生产环境必须显式配置密钥和外部服务地址

## Git 与提交规范

建议提交类型：

- feat
- fix
- docs
- refactor
- test
- chore

示例：

```text
feat(workflow): add DAG validation
fix(auth): handle expired token
docs(api): document workflow run endpoints
```

## 禁止事项

- 禁止把密钥提交到仓库
- 禁止在 Handler 中堆叠复杂业务逻辑
- 禁止跳过 Workspace 权限校验
- 禁止用 Redis 替代 PostgreSQL 保存权威业务状态
- 禁止在普通日志中输出完整用户敏感输入
- 禁止无 Trace 的 Workflow 执行
- 禁止新增节点却不补充节点配置、校验、执行和日志规范
