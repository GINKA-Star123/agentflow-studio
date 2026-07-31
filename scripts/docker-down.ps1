param(
    [switch]$Volumes
)

$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$ComposeFile = Join-Path $Root "deployments\docker\docker-compose.yml"

if ($Volumes) {
    Write-Host "停止 Docker Compose 环境并删除数据卷..." -ForegroundColor Yellow
    docker compose -f $ComposeFile down -v
} else {
    Write-Host "停止 Docker Compose 环境..." -ForegroundColor Cyan
    docker compose -f $ComposeFile down
}

Write-Host ""
Write-Host "Docker Compose 环境已停止。" -ForegroundColor Green