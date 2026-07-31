# DEPLOYMENT.md

> Docker 部署。

## 部署目标

项目必须支持本地和演示环境一键启动。

基础命令目标：

```bash
docker compose up -d
```

## 服务清单

| 服务 | 说明 | 默认端口 |
| --- | --- | --- |
| web | Next.js 前端 | 3000 |
| api | Go API 服务 | 8080 |
| ai-runtime | Python FastAPI Runtime | 8090 |
| postgres | PostgreSQL | 5432 |
| redis | Redis | 6379 |
| qdrant | Qdrant | 6333 |
| jaeger | Trace 查看 | 16686 |
| prometheus | Metrics | 9090 |
| grafana | Dashboard | 3001 |
| nginx | 反向代理 | 80 |

## Dockerfile 规则

每个服务独立 Dockerfile：

- `apps/web/Dockerfile`
- `services/api/Dockerfile`
- `services/ai-runtime/Dockerfile`

规则：

- 使用多阶段构建
- 不把 `.env` 和密钥复制进镜像
- 镜像内使用非 root 用户优先
- 暴露健康检查端点
- 构建依赖和运行依赖分层

## Docker Compose 规则

Compose 必须包含：

- 服务网络
- Volume
- 环境变量
- 健康检查
- depends_on 条件

依赖关系：

```text
web -> api
api -> postgres, redis, ai-runtime
ai-runtime -> qdrant
prometheus -> api, ai-runtime
grafana -> prometheus
jaeger <- api, ai-runtime
```

## 环境变量

统一使用 `.env.example` 说明配置。

Go API：

- `APP_ENV`
- `HTTP_PORT`
- `DATABASE_URL`
- `REDIS_ADDR`
- `JWT_SECRET`
- `AI_RUNTIME_URL`
- `OTEL_EXPORTER_OTLP_ENDPOINT`

AI Runtime：

- `APP_ENV`
- `HTTP_PORT`
- `OPENAI_COMPATIBLE_BASE_URL`
- `OPENAI_COMPATIBLE_API_KEY`
- `QDRANT_URL`
- `QDRANT_API_KEY`
- `OTEL_EXPORTER_OTLP_ENDPOINT`

Web：

- `NEXT_PUBLIC_API_BASE_URL`
- `NEXT_PUBLIC_WS_BASE_URL`

## Nginx 路由

建议路由：

```text
/              -> web
/api/          -> api
/ws/           -> api websocket
/runtime/      -> ai-runtime internal only, 默认不对公网开放
```

## 数据持久化

必须使用 Volume：

- PostgreSQL data
- Redis data，按环境决定是否开启
- Qdrant storage
- Grafana data

## 健康检查

每个应用服务提供：

- `/healthz`：进程存活
- `/readyz`：依赖可用

Docker Compose 中必须配置 healthcheck。

## 安全要求

- 密钥不得提交到 Git
- 默认演示环境也要更改 JWT_SECRET
- AI Runtime 内部接口不直接暴露公网
- 上传文件大小必须有限制
- Nginx 配置请求体大小和超时

## 后续 CI/CD

GitHub Actions 后续接入：

- Lint
- Test
- Build
- Docker Image Build
- Security Scan
- Deploy
