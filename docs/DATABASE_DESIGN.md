# DATABASE_DESIGN.md

> PostgreSQL、Redis、Qdrant 设计。

## 设计原则

- PostgreSQL 保存业务权威数据
- Redis 保存短期状态、缓存、限流和 Pub/Sub 事件
- Qdrant 保存向量索引，不保存权限和业务主数据
- 所有数据必须通过 workspace_id 做租户隔离
- 所有主表使用 UUID 作为主键
- 所有重要表保留 created_at、updated_at，软删除表保留 deleted_at

## PostgreSQL

### users

用户表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| email | varchar | 登录邮箱，唯一 |
| password_hash | varchar | 密码哈希 |
| display_name | varchar | 显示名称 |
| status | varchar | active、disabled |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### workspaces

Workspace 表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| name | varchar | Workspace 名称 |
| owner_id | uuid | 所有者用户 ID |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### workspace_members

Workspace 成员表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workspace_id | uuid | Workspace ID |
| user_id | uuid | 用户 ID |
| role | varchar | owner、admin、member、viewer |
| created_at | timestamptz | 创建时间 |

约束：

- `(workspace_id, user_id)` 唯一

### workflows

Workflow 定义表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workspace_id | uuid | Workspace ID |
| name | varchar | Workflow 名称 |
| description | text | 描述 |
| status | varchar | draft、published、archived |
| current_version_id | uuid | 当前版本 ID |
| created_by | uuid | 创建人 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| deleted_at | timestamptz | 软删除时间 |

### workflow_versions

Workflow 版本表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workflow_id | uuid | Workflow ID |
| version | integer | 版本号 |
| definition | jsonb | Workflow JSON 定义 |
| schema_version | varchar | Definition Schema 版本 |
| created_by | uuid | 创建人 |
| created_at | timestamptz | 创建时间 |

约束：

- `(workflow_id, version)` 唯一

### workflow_runs

Workflow 执行记录表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | Run ID |
| workspace_id | uuid | Workspace ID |
| workflow_id | uuid | Workflow ID |
| workflow_version_id | uuid | Workflow 版本 ID |
| status | varchar | pending、running、succeeded、failed、canceled |
| input | jsonb | 执行输入 |
| output | jsonb | 执行输出 |
| error | jsonb | 错误信息 |
| started_at | timestamptz | 开始时间 |
| finished_at | timestamptz | 结束时间 |
| created_at | timestamptz | 创建时间 |

### node_executions

节点执行记录表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 节点执行 ID |
| run_id | uuid | Workflow Run ID |
| node_id | varchar | Workflow Definition 中的节点 ID |
| node_type | varchar | 节点类型 |
| status | varchar | pending、running、succeeded、failed、skipped |
| input | jsonb | 节点输入 |
| output | jsonb | 节点输出 |
| error | jsonb | 错误信息 |
| token_usage | jsonb | Token 用量 |
| latency_ms | integer | 节点耗时 |
| started_at | timestamptz | 开始时间 |
| finished_at | timestamptz | 结束时间 |

### knowledge_bases

知识库表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workspace_id | uuid | Workspace ID |
| name | varchar | 知识库名称 |
| description | text | 描述 |
| created_by | uuid | 创建人 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### documents

文档表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workspace_id | uuid | Workspace ID |
| knowledge_base_id | uuid | 知识库 ID |
| filename | varchar | 文件名 |
| mime_type | varchar | MIME 类型 |
| storage_path | varchar | 存储路径 |
| status | varchar | uploaded、processing、ready、failed |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |

### document_chunks

文档 Chunk 表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | uuid | 主键 |
| workspace_id | uuid | Workspace ID |
| document_id | uuid | 文档 ID |
| chunk_index | integer | Chunk 序号 |
| content | text | Chunk 内容 |
| metadata | jsonb | 元数据 |
| qdrant_point_id | uuid | Qdrant Point ID |
| created_at | timestamptz | 创建时间 |

## Redis

### Key 命名

统一格式：

```text
af:{env}:{domain}:{resource}:{id}
```

示例：

```text
af:dev:session:{token_id}
af:dev:workflow_run:{run_id}:state
af:dev:workflow_run:{run_id}:events
af:dev:rate_limit:user:{user_id}
af:dev:token_cache:{provider}:{model}:{hash}
```

### 用途

Session / Token Cache：

- Key：`af:{env}:session:{token_id}`
- TTL：跟 JWT 或 Refresh Token 策略保持一致

Workflow Run 状态：

- Key：`af:{env}:workflow_run:{run_id}:state`
- TTL：建议 24 小时
- Value：当前状态、当前节点、进度、更新时间

Pub/Sub：

- Channel：`af:{env}:workflow_run:{run_id}:events`
- Event：node_started、node_completed、node_failed、run_completed

Rate Limit：

- Key：`af:{env}:rate_limit:user:{user_id}`
- TTL：按窗口设置

Token Cache：

- Key：`af:{env}:token_cache:{provider}:{model}:{hash}`
- TTL：按业务策略设置

## Qdrant

### Collection 命名

```text
agentflow_{env}_kb_chunks
```

### Point Payload

每个向量 Point 必须包含：

```json
{
  "workspace_id": "uuid",
  "knowledge_base_id": "uuid",
  "document_id": "uuid",
  "chunk_id": "uuid",
  "chunk_index": 0,
  "source": "filename.pdf"
}
```

### 检索约束

- 查询必须按 workspace_id 过滤
- 查询知识库时必须按 knowledge_base_id 过滤
- 检索结果返回 chunk_id，再由 PostgreSQL 查询权威内容和权限

## Migration 规则

- 数据库变更必须通过 Migration
- Migration 文件必须可重复在新环境执行
- 破坏性变更必须先做兼容迁移，再清理旧字段
- 索引必须根据查询路径设计，不盲目添加
