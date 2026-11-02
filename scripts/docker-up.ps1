$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$ComposeFile = Join-Path $Root "deployments\docker\docker-compose.yml"

Write-Host "构建并启动 Docker Compose 完整环境..." -ForegroundColor Cyan

docker compose -f $ComposeFile up --build -d

Write-Host ""
Write-Host "Docker Compose 环境已启动。" -ForegroundColor Green
Write-Host "Nginx:      http://localhost"
Write-Host "Web:        http://localhost:3000"
Write-Host "Go API:     http://localhost:8080/readyz"
Write-Host "AI Runtime: http://localhost:8090/readyz"
Write-Host "Jaeger:     http://localhost:16686"
Write-Host "Prometheus: http://localhost:9090"
Write-Host "Grafana:    http://localhost:3001"