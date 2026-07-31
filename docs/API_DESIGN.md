# API_DESIGN.md

> 所有接口设计。

## API 原则

- REST API 使用 JSON
- 所有业务接口以 `/api/v1` 开头
- 受保护接口必须校验 JWT
- Workspace 资源必须校验成员关系
- 错误返回统一结构
- 分页接口统一使用 page、page_size
- Workflow 执行事件通过 WebSocket 推送

## 通用响应

成功：

```json
{
  "data": {},
  "request_id": "req_xxx"
}
```

失败：

```json
{
  "error": {
    "code": "INVALID_ARGUMENT",
    "message": "invalid workflow definition",
    "details": {}
  },
  "request_id": "req_xxx"
}
```

## 错误码

| Code | 说明 |
| --- | --- |
| INVALID_ARGUMENT | 参数错误 |
| UNAUTHORIZED | 未认证 |
| FORBIDDEN | 无权限 |
| NOT_FOUND | 资源不存在 |
| CONFLICT | 资源冲突 |
| RATE_LIMITED | 请求限流 |
| INTERNAL | 服务内部错误 |
| AI_RUNTIME_ERROR | AI Runtime 错误 |
| WORKFLOW_VALIDATION_ERROR | Workflow 校验失败 |
| INVALID_CREDENTIALS | 邮箱或密码错误 |
| EMAIL_ALREADY_EXISTS | 邮箱已注册 |
| USER_DISABLED | 用户已禁用 |
| MISSING_TOKEN | 缺少 Token |
| INVALID_TOKEN | Token 无效 |
| EXPIRED_TOKEN | Token 已过期 |
| WORKSPACE_INVALID_INPUT | Workspace 参数错误 |
| WORKSPACE_INVALID_ROLE | Workspace 角色错误 |
| WORKSPACE_NOT_FOUND | Workspace 不存在 |
| WORKSPACE_MEMBER_NOT_FOUND | Workspace 成员不存在 |
| WORKSPACE_MEMBER_ALREADY_EXISTS | Workspace 成员已存在 |
| WORKSPACE_USER_NOT_FOUND | 要添加的用户不存在 |
| WORKSPACE_PERMISSION_DENIED | Workspace 权限不足 |
| WORKSPACE_OWNER_OPERATION_NOT_ALLOWED | 不允许操作 owner |
| WORKFLOW_INVALID_INPUT | Workflow 参数错误 |
| WORKFLOW_INVALID_SCHEMA | Workflow Schema 错误 |
| WORKFLOW_NOT_FOUND | Workflow 不存在 |
| WORKFLOW_CREATE_FAILED | Workflow 创建失败 |
| WORKFLOW_UPDATE_FAILED | Workflow 更新失败 |

## Auth API

Auth API 用于用户注册、登录和当前用户查询。

公开接口：

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`

受保护接口：

- `GET /api/v1/auth/me`

### POST /api/v1/auth/register

注册用户，并自动创建默认 Workspace，同时把当前用户设为该 Workspace 的 owner。

Request：

```json
{
  "email": "user@example.com",
  "password": "password123",
  "display_name": "User",
  "workspace_name": "User Workspace"
}
```

Response：

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "display_name": "User",
      "status": "active"
    },
    "access_token": "jwt",
    "token_type": "Bearer",
    "expires_at": "2026-01-01T00:00:00Z",
    "current_workspace": {
      "id": "uuid",
      "name": "User Workspace",
      "owner_id": "uuid",
      "role": "owner"
    },
    "workspaces": [
      {
        "id": "uuid",
        "name": "User Workspace",
        "owner_id": "uuid",
        "role": "owner"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

错误：

- `INVALID_ARGUMENT`：请求参数格式错误
- `WEAK_PASSWORD`：密码强度不足
- `PASSWORD_TOO_LONG`：密码超过 bcrypt 安全长度
- `EMAIL_ALREADY_EXISTS`：邮箱已注册

### POST /api/v1/auth/login

登录。

Request：

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

Response：

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "display_name": "User",
      "status": "active"
    },
    "access_token": "jwt",
    "token_type": "Bearer",
    "expires_at": "2026-01-01T00:00:00Z",
    "current_workspace": {
      "id": "uuid",
      "name": "User Workspace",
      "owner_id": "uuid",
      "role": "owner"
    },
    "workspaces": [
      {
        "id": "uuid",
        "name": "User Workspace",
        "owner_id": "uuid",
        "role": "owner"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

错误：

- `INVALID_ARGUMENT`：请求参数格式错误
- `INVALID_CREDENTIALS`：邮箱或密码错误
- `USER_DISABLED`：用户已禁用

### GET /api/v1/auth/me

获取当前用户。

Header：

```http
Authorization: Bearer <access_token>
```

Response：

```json
{
  "data": {
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "display_name": "User",
      "status": "active"
    },
    "current_workspace": {
      "id": "uuid",
      "name": "User Workspace",
      "owner_id": "uuid",
      "role": "owner"
    },
    "workspaces": [
      {
        "id": "uuid",
        "name": "User Workspace",
        "owner_id": "uuid",
        "role": "owner"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

错误：

- `MISSING_TOKEN`：缺少 Authorization Bearer Token
- `INVALID_TOKEN`：Token 无效
- `EXPIRED_TOKEN`：Token 已过期
- `USER_DISABLED`：用户已禁用

## Workspace API

Workspace API 全部为受保护接口，必须携带：

```http
Authorization: Bearer <access_token>
```

### GET /api/v1/workspaces

获取当前用户可访问的 Workspace 列表。

Response：

```json
{
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "My Workspace",
        "owner_id": "uuid",
        "role": "owner"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

### POST /api/v1/workspaces

创建 Workspace。

Request：

```json
{
  "name": "My Workspace"
}
```

Response：

```json
{
  "data": {
    "id": "uuid",
    "name": "My Workspace",
    "owner_id": "uuid",
    "role": "owner"
  },
  "request_id": "req_xxx"
}
```

### GET /api/v1/workspaces/{workspace_id}/members

获取 Workspace 成员列表。

权限：

- owner：允许
- admin：允许
- member：允许
- viewer：允许

Response：

```json
{
  "data": {
    "items": [
      {
        "user_id": "uuid",
        "email": "user@example.com",
        "display_name": "User",
        "role": "owner",
        "joined_at": "2026-01-01T00:00:00Z"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

### POST /api/v1/workspaces/{workspace_id}/members

把已有用户添加为 Workspace 成员。

权限：

- owner：允许添加 `member`、`viewer`
- admin：允许添加 `member`、`viewer`
- member：不允许
- viewer：不允许

Request：

```json
{
  "email": "member@example.com",
  "role": "member"
}
```

Response：

```json
{
  "data": {
    "user_id": "uuid",
    "email": "member@example.com",
    "display_name": "Member",
    "role": "member",
    "joined_at": "2026-01-01T00:00:00Z"
  },
  "request_id": "req_xxx"
}
```

错误：

- `WORKSPACE_USER_NOT_FOUND`：用户不存在
- `WORKSPACE_MEMBER_ALREADY_EXISTS`：用户已经是成员
- `WORKSPACE_INVALID_ROLE`：角色不是 `member` 或 `viewer`
- `WORKSPACE_PERMISSION_DENIED`：当前用户无权添加成员

### PATCH /api/v1/workspaces/{workspace_id}/members/{user_id}/role

更新 Workspace 成员角色。

权限：

- owner：可以修改 `admin`、`member`、`viewer`
- admin：可以修改 `member`、`viewer`
- member：不允许
- viewer：不允许

限制：

- 不能修改自己的角色
- 不能修改 owner 角色
- 不能通过该接口设置 owner 角色
- admin 不能修改其他 admin
- admin 不能把成员提升为 admin

Request：

```json
{
  "role": "viewer"
}
```

Response：

```json
{
  "data": {
    "user_id": "uuid",
    "email": "member@example.com",
    "display_name": "Member",
    "role": "viewer",
    "joined_at": "2026-01-01T00:00:00Z"
  },
  "request_id": "req_xxx"
}
```

### DELETE /api/v1/workspaces/{workspace_id}/members/{user_id}

移除 Workspace 成员。

权限：

- owner：可以移除 `admin`、`member`、`viewer`
- admin：可以移除 `member`、`viewer`
- member：不允许
- viewer：不允许

限制：

- 不能移除自己
- 不能移除 owner
- admin 不能移除其他 admin

Response：

```json
{
  "data": {
    "removed": true
  },
  "request_id": "req_xxx"
}
```

## Workspace RBAC 规则

| 角色 | 查看成员 | 添加 member/viewer | 修改角色 | 移除成员 |
| --- | --- | --- | --- | --- |
| owner | 允许 | 允许 | 可修改 admin/member/viewer | 可移除 admin/member/viewer |
| admin | 允许 | 允许 | 仅可修改 member/viewer | 仅可移除 member/viewer |
| member | 允许 | 不允许 | 不允许 | 不允许 |
| viewer | 允许 | 不允许 | 不允许 | 不允许 |

owner 保护规则：

- 不能修改 owner 角色
- 不能通过普通成员接口设置 owner
- 不能移除 owner
- 不能修改自己的角色
- 不能移除自己

## Workflow API

Workflow API 全部为受保护接口，必须携带 Bearer Token，并且必须属于目标 Workspace。
当前阶段仅实现列表、创建、详情读取和更新。删除、版本历史和运行编排保留给后续阶段。

### GET /api/v1/workspaces/{workspace_id}/workflows

获取当前 Workspace 下的 Workflow 列表。

权限：
- owner：允许
- admin：允许
- member：允许
- viewer：允许

Response：

```json
{
  "data": {
    "items": [
      {
        "id": "uuid",
        "workspace_id": "uuid",
        "name": "Customer Support Bot",
        "schema_version": "1.0",
        "node_count": 4,
        "edge_count": 3,
        "created_at": "2026-01-01T00:00:00Z",
        "updated_at": "2026-01-01T00:00:00Z"
      }
    ]
  },
  "request_id": "req_xxx"
}
```

### POST /api/v1/workspaces/{workspace_id}/workflows

创建 Workflow。

权限：
- owner：允许
- admin：允许
- member：允许
- viewer：不允许

Request：

```json
{
  "name": "Customer Support Bot",
  "schema": {
    "schema_version": "1.0",
    "name": "Customer Support Bot",
    "summary": {
      "node_count": 4,
      "edge_count": 3,
      "start_count": 1,
      "end_count": 1
    },
    "nodes": [
      {
        "id": "start_1",
        "type": "Start",
        "label": "Start",
        "description": "流程入口",
        "position": {
          "x": 80,
          "y": 220
        },
        "config": {}
      }
    ],
    "edges": []
  }
}
```

Response：

```json
{
  "data": {
    "id": "uuid",
    "workspace_id": "uuid",
    "name": "Customer Support Bot",
    "schema_version": "1.0",
    "node_count": 2,
    "edge_count": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "schema": {
      "schema_version": "1.0",
      "name": "Customer Support Bot",
      "summary": {
        "node_count": 2,
        "edge_count": 1,
        "start_count": 1,
        "end_count": 1
      },
      "nodes": [
        {
          "id": "start_1",
          "type": "Start",
          "label": "Start",
          "description": "流程入口",
          "position": {
            "x": 80,
            "y": 220
          },
          "config": {}
        },
        {
          "id": "end_1",
          "type": "End",
          "label": "End",
          "description": "流程出口",
          "position": {
            "x": 360,
            "y": 220
          },
          "config": {}
        }
      ],
      "edges": [
        {
          "id": "edge_start_end",
          "source": "start_1",
          "target": "end_1",
          "type": "smoothstep"
        }
      ]
    }
  },
  "request_id": "req_xxx"
}
```

### GET /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}

获取 Workflow 详情。

权限：
- owner：允许
- admin：允许
- member：允许
- viewer：允许

Response：

```json
{
  "data": {
    "id": "uuid",
    "workspace_id": "uuid",
    "name": "Customer Support Bot",
    "schema_version": "1.0",
    "node_count": 2,
    "edge_count": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "schema": {
      "schema_version": "1.0",
      "name": "Customer Support Bot",
      "summary": {
        "node_count": 2,
        "edge_count": 1,
        "start_count": 1,
        "end_count": 1
      },
      "nodes": [
        {
          "id": "start_1",
          "type": "Start",
          "label": "Start",
          "description": "流程入口",
          "position": {
            "x": 80,
            "y": 220
          },
          "config": {}
        },
        {
          "id": "end_1",
          "type": "End",
          "label": "End",
          "description": "流程出口",
          "position": {
            "x": 360,
            "y": 220
          },
          "config": {}
        }
      ],
      "edges": [
        {
          "id": "edge_start_end",
          "source": "start_1",
          "target": "end_1",
          "type": "smoothstep"
        }
      ]
    }
  },
  "request_id": "req_xxx"
}
```

### PUT /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}

更新 Workflow 名称和当前定义。

权限：
- owner：允许
- admin：允许
- member：允许
- viewer：不允许

Request：

```json
{
  "name": "Customer Support Bot",
  "schema": {
    "schema_version": "1.0",
    "name": "Customer Support Bot",
    "summary": {
      "node_count": 4,
      "edge_count": 3,
      "start_count": 1,
      "end_count": 1
    },
    "nodes": [],
    "edges": []
  }
}
```

Response：

```json
{
  "data": {
    "id": "uuid",
    "workspace_id": "uuid",
    "name": "Customer Support Bot",
    "schema_version": "1.0",
    "node_count": 2,
    "edge_count": 1,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z",
    "schema": {
      "schema_version": "1.0",
      "name": "Customer Support Bot",
      "summary": {
        "node_count": 2,
        "edge_count": 1,
        "start_count": 1,
        "end_count": 1
      },
      "nodes": [
        {
          "id": "start_1",
          "type": "Start",
          "label": "Start",
          "description": "流程入口",
          "position": {
            "x": 80,
            "y": 220
          },
          "config": {}
        },
        {
          "id": "end_1",
          "type": "End",
          "label": "End",
          "description": "流程出口",
          "position": {
            "x": 360,
            "y": 220
          },
          "config": {}
        }
      ],
      "edges": [
        {
          "id": "edge_start_end",
          "source": "start_1",
          "target": "end_1",
          "type": "smoothstep"
        }
      ]
    }
  },
  "request_id": "req_xxx"
}
```

### 后续阶段预留

- `DELETE /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}`
- `POST /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}/versions`
- `GET /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}/versions`

## Workflow Run API

### POST /api/v1/workspaces/{workspace_id}/workflows/{workflow_id}/runs

执行 Workflow。

Request：

```json
{
  "input": {
    "message": "hello"
  },
  "version_id": "uuid"
}
```

Response：

```json
{
  "data": {
    "run_id": "uuid",
    "status": "pending"
  },
  "request_id": "req_xxx"
}
```

### GET /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}

获取执行详情。

### GET /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}/nodes

获取节点执行列表。

### POST /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}/cancel

取消执行。

### GET /api/v1/workspaces/{workspace_id}/workflow-runs/{run_id}/events

WebSocket 连接，用于接收执行事件。

事件格式：

```json
{
  "event": "node_completed",
  "run_id": "uuid",
  "node_id": "llm_1",
  "timestamp": "2026-01-01T00:00:00Z",
  "payload": {}
}
```

## Knowledge Base API

### GET /api/v1/workspaces/{workspace_id}/knowledge-bases

获取知识库列表。

### POST /api/v1/workspaces/{workspace_id}/knowledge-bases

创建知识库。

Request：

```json
{
  "name": "Product Docs",
  "description": "Product documentation"
}
```

### POST /api/v1/workspaces/{workspace_id}/knowledge-bases/{knowledge_base_id}/documents

上传文档。

Content-Type：`multipart/form-data`

### GET /api/v1/workspaces/{workspace_id}/knowledge-bases/{knowledge_base_id}/documents

获取文档列表。

### POST /api/v1/workspaces/{workspace_id}/knowledge-bases/{knowledge_base_id}/search

检索知识库。

Request：

```json
{
  "query": "How to reset password?",
  "top_k": 5
}
```

## AI Runtime Internal API

内部接口由 Go 服务调用，不直接暴露给浏览器。

### POST /internal/v1/llm/chat

LLM Chat 调用。

### POST /internal/v1/llm/stream

LLM 流式调用。

### POST /internal/v1/embeddings

生成 Embedding。

### POST /internal/v1/rag/query

RAG 查询。

### POST /internal/v1/tools/call

Tool Calling。

## 健康检查

### GET /healthz

进程存活检查。

### GET /readyz

依赖可用检查。
