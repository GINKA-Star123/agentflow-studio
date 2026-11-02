param(
    [switch]$SkipDeps
)

$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")

function Start-DevWindow {
    param(
        [string]$Title,
        [string]$Command
    )

    $RootPath = $Root.Path.Replace("'", "''")
    $WindowTitle = $Title.Replace("'", "''")
    $FullCommand = "cd '$RootPath'; `$Host.UI.RawUI.WindowTitle = '$WindowTitle'; $Command"

    Start-Process powershell.exe -ArgumentList @(
        "-NoExit",
        "-ExecutionPolicy",
        "Bypass",
        "-Command",
        $FullCommand
    )
}

if (-not $SkipDeps) {
    & (Join-Path $PSScriptRoot "dev-deps.ps1")
}

Write-Host "启动本地开发服务..." -ForegroundColor Cyan

Start-DevWindow `
    -Title "AgentFlow Web" `
    -Command "cd apps\web; npm run dev"

Start-DevWindow `
    -Title "AgentFlow API" `
    -Command "cd services\api; go run .\cmd\api"

Start-DevWindow `
    -Title "AgentFlow AI Runtime" `
    -Command "cd services\ai-runtime; . .\.venv\Scripts\Activate.ps1; uvicorn app.main:app --host 0.0.0.0 --port 8090 --reload"

Write-Host ""
Write-Host "本地开发服务启动命令已发送到独立 PowerShell 窗口。" -ForegroundColor Green
Write-Host "Web:        http://localhost:3000"
Write-Host "Go API:     http://localhost:8080/readyz"
Write-Host "AI Runtime: http://localhost:8090/readyz"