# TODO.md

> 开发任务清单。

## Phase 1：项目初始化

- [x] 创建 `apps/web` 前端工程
- [x] 创建 `services/api` Go 服务
- [x] 创建 `services/ai-runtime` Python 服务
- [x] 创建 `deployments/docker` 部署目录
- [x] 添加基础 `.env.example`
- [x] 添加前端启动脚本
- [x] 添加 Go 服务启动脚本
- [x] 添加 Python Runtime 启动脚本
- [x] 添加基础 Docker Compose
- [x] 添加 PostgreSQL、Redis、Qdrant 服务
- [x] 添加健康检查端点
- [x] 添加基础日志配置

## Phase 2：用户系统

- [x] 设计用户表
- [x] 设计 Workspace 表
- [x] 设计 Workspace 成员表
- [x] 实现 PostgreSQL 迁移
- [x] 实现密码 bcrypt 哈希与校验
- [x] 实现 JWT 签发与校验
- [x] 实现 Auth 领域错误定义
- [x] 实现用户注册
- [x] 实现用户登录
- [x] 实现 JWT 中间件
- [x] 实现当前用户接口 `/api/v1/auth/me`
- [x] 实现 Workspace 创建
- [x] 实现 Workspace 列表
- [x] 实现 Workspace 成员关系
- [x] 实现 Workspace 成员列表
- [x] 实现添加已有用户为 Workspace 成员
- [x] 实现 Workspace 成员角色更新
- [x] 实现 Workspace 成员移除
- [x] 实现 Workspace 权限校验
- [x] 实现 owner/admin/member/viewer 基础 RBAC
- [x] 补充认证测试
- [x] 补充 JWT 测试
- [x] 补充 Workspace RBAC 测试
- [x] 补充 Phase 2 接口验证脚本
- [ ] 执行最终验证：`go test ./...`
- [ ] 执行最终验证：`.\scripts\verify-phase2.ps1`

### Phase 2 最终验收清单

- [x] 用户可以注册
- [x] 注册后自动创建默认 Workspace
- [x] 注册用户自动成为默认 Workspace owner
- [x] 用户可以登录
- [x] 登录成功返回 Bearer Token
- [x] 受保护接口必须携带 JWT
- [x] 当前用户接口可返回用户和 Workspace 信息
- [x] 用户可以创建 Workspace
- [x] 用户可以查看自己的 Workspace 列表
- [x] Workspace 成员可以查看成员列表
- [x] owner/admin 可以添加已有用户为 member/viewer
- [x] owner 可以更新 admin/member/viewer 角色
- [x] admin 可以更新 member/viewer 角色
- [x] owner 可以移除 admin/member/viewer
- [x] admin 可以移除 member/viewer
- [x] member/viewer 不能管理成员
- [x] owner 角色受到保护
- [x] 自己不能修改自己的角色
- [x] 自己不能移除自己
- [x] 后端测试文件已补充
- [x] 接口验证脚本已补充

## Phase 3：Workflow Designer

- [x] 初始化 React Flow 画布
- [x] 实现节点面板和拖拽添加
- [x] 实现 Start 节点
- [x] 实现 End 节点
- [x] 实现 Prompt 节点
- [x] 实现 LLM 节点
- [x] 实现 Condition 节点
- [x] 实现 Loop 节点
- [x] 实现 HTTP 节点
- [x] 实现 Tool 节点
- [x] 实现 Memory 节点
- [x] 实现 RAG 节点
- [x] 实现节点选择、连线、删除
- [x] 实现节点属性面板
- [x] 实现 Prompt / LLM 配置编辑
- [x] 实现 Workflow Schema 生成与基础校验
- [x] 实现 Workflow 列表展示
- [x] 实现 Workflow 加载
- [x] 实现 Workflow 保存
- [x] 实现 Workflow API 对接

> 说明：导入导出和版本历史留给后续迭代，不纳入本阶段验收范围。

## Phase 4：Workflow Engine

- [ ] 定义 Workflow JSON Schema
- [ ] 实现 Workflow DAG 校验
- [ ] 实现节点执行器接口
- [ ] 实现 Start 节点执行器
- [ ] 实现 End 节点执行器
- [ ] 实现 Prompt 节点执行器
- [ ] 实现 LLM 节点执行器
- [ ] 实现 Condition 节点执行器
- [ ] 实现 Loop 节点执行器
- [ ] 实现 HTTP 节点执行器
- [ ] 实现 Tool 节点执行器
- [ ] 实现 Memory 节点执行器
- [ ] 实现 RAG 节点执行器
- [ ] 实现 Run 状态记录
- [ ] 实现 Node Execution 记录
- [ ] 实现 WebSocket 事件推送
- [ ] 补充 DAG 校验测试
- [ ] 补充节点执行测试

## Phase 5：LLM Runtime

- [ ] 初始化 FastAPI 服务
- [ ] 实现 OpenAI Compatible Provider
- [ ] 实现 Chat API
- [ ] 实现 Streaming API
- [ ] 实现 Prompt 模板渲染
- [ ] 实现 Tool Calling 数据结构
- [ ] 实现 Token 用量统计
- [ ] Go 服务接入 AI Runtime
- [ ] 前端展示流式输出

## Phase 6：Knowledge Base

- [ ] 设计知识库表
- [ ] 设计文档表
- [ ] 设计 Chunk 表
- [ ] 实现文档上传
- [ ] 实现文档解析
- [ ] 实现 Chunk 切分
- [ ] 实现 Embedding API
- [ ] 实现 Qdrant 写入
- [ ] 实现 Qdrant 检索
- [ ] 实现 RAG 节点上下文组装
- [ ] 前端实现知识库管理页面

## Phase 7：Redis

- [ ] 设计 Redis Key 规范
- [ ] 实现 Redis 客户端
- [ ] 实现 Session / Token Cache
- [ ] 实现 Workflow Run 状态缓存
- [ ] 实现 Rate Limit
- [ ] 实现 Pub/Sub 执行事件
- [ ] 实现热点数据缓存

## Phase 8：OpenTelemetry

- [ ] Go 服务接入 OpenTelemetry
- [ ] Python Runtime 接入 OpenTelemetry
- [ ] HTTP 请求 Trace
- [ ] Workflow Run Trace
- [ ] Node Execution Span
- [ ] LLM 调用 Span
- [ ] RAG 查询 Span
- [ ] Prometheus Metrics
- [ ] Jaeger 本地服务
- [ ] Grafana Dashboard 预留

## Phase 9：Docker

- [ ] 添加 Web Dockerfile
- [ ] 添加 Go API Dockerfile
- [ ] 添加 AI Runtime Dockerfile
- [ ] 完善 Docker Compose
- [ ] 添加 Nginx 配置
- [ ] 配置服务健康检查
- [ ] 配置 Volume
- [ ] 验证一键启动

## Phase 10：MCP

- [ ] 设计 MCP Server 模块
- [ ] 设计 MCP Tool Schema
- [ ] 实现 MCP Tool 注册
- [ ] 实现 MCP Tool 调用
- [ ] 与 Workflow Tool 节点集成
- [ ] 添加 MCP 调用审计
- [ ] 添加 MCP 权限控制

## 当前优先级

1. 开始 Phase 4：Workflow Engine / Runtime Execution
2. 定义 Workflow JSON Schema
3. 实现 Workflow DAG 校验
4. 实现节点执行器接口
5. 推进 Start / End / Prompt / LLM 节点执行器

## Phase 4 进入条件

- [x] PostgreSQL 迁移已在本地执行
- [x] Go API 可正常启动
- [x] `go test ./...` 通过
- [x] `npm run lint` 通过
- [x] `npm run build` 通过
- [x] Auth / Workspace / Workflow API 与 `docs/API_DESIGN.md` 保持一致
- [x] Phase 2、Phase 3 相关任务在本文件中已勾选
- [x] 前端可访问后端健康状态
- [x] Workflow Designer 页面可用
- [x] 准备开始 Phase 4：Workflow Engine / Runtime Execution
