param(
    [string]$GoApiRoot = (Join-Path $PSScriptRoot "..\services\api"),
    [string]$AIRuntimeBaseUrl = "http://localhost:8090",
    [switch]$SkipGoTests,
    [switch]$SkipPythonTests
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

function Write-Step {
    param(
        [string]$Message
    )

    Write-Host ""
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )

    if (-not $Condition) {
        throw "断言失败: $Message"
    }
}

function Read-ResponseText {
    param(
        $Response
    )

    if ($null -eq $Response) {
        return ""
    }

    $stream = $null
    $reader = $null

    try {
        $stream = $Response.GetResponseStream()
        if ($null -eq $stream) {
            return ""
        }

        $reader = New-Object System.IO.StreamReader($stream)
        return $reader.ReadToEnd()
    } finally {
        if ($null -ne $reader) {
            $reader.Close()
        }

        if ($null -ne $stream) {
            $stream.Close()
        }
    }
}

function Invoke-JsonRequest {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Method,

        [Parameter(Mandatory = $true)]
        [string]$Uri,

        [object]$Body = $null,
        [string]$Accept = "application/json"
    )

    $headers = @{
        Accept = $Accept
    }

    $params = @{
        Uri         = $Uri
        Method      = $Method
        Headers     = $headers
        TimeoutSec  = 30
        ErrorAction = "Stop"
    }

    if ($null -ne $Body) {
        $params["Body"] = ($Body | ConvertTo-Json -Depth 20)
        $params["ContentType"] = "application/json"
    }

    try {
        $response = Invoke-WebRequest @params
        $bodyText = $response.Content
        $parsedBody = $null

        if ($Accept -eq "application/json" -and -not [string]::IsNullOrWhiteSpace($bodyText)) {
            $parsedBody = $bodyText | ConvertFrom-Json
        }

        return [pscustomobject]@{
            StatusCode = [int]$response.StatusCode
            Headers    = $response.Headers
            BodyText   = $bodyText
            Body       = $parsedBody
        }
    } catch {
        $response = $_.Exception.Response
        if ($null -eq $response) {
            throw
        }

        $bodyText = Read-ResponseText -Response $response
        $parsedBody = $null

        if ($Accept -eq "application/json" -and -not [string]::IsNullOrWhiteSpace($bodyText)) {
            $parsedBody = $bodyText | ConvertFrom-Json
        }

        return [pscustomobject]@{
            StatusCode = [int]$response.StatusCode
            Headers    = $response.Headers
            BodyText   = $bodyText
            Body       = $parsedBody
        }
    }
}

function Get-SseEventNames {
    param(
        [string]$Content
    )

    if ([string]::IsNullOrWhiteSpace($Content)) {
        return @()
    }

    $matches = [regex]::Matches($Content, '(?m)^event:\s*([^\r\n]+)$')
    $events = @()

    foreach ($match in $matches) {
        $events += $match.Groups[1].Value.Trim()
    }

    return $events
}

$ProjectRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$AIRuntimeRoot = Join-Path $ProjectRoot "services\ai-runtime"
$PythonExe = Join-Path $AIRuntimeRoot ".venv\Scripts\python.exe"

if (-not (Test-Path $PythonExe)) {
    $PythonExe = "python"
}

if (-not $SkipGoTests) {
    Write-Step "运行 Go 侧工具桥接测试"
    Push-Location $GoApiRoot
    try {
        go test ./internal/workflowruntime ./internal/service
    } finally {
        Pop-Location
    }
}

if (-not $SkipPythonTests) {
    Write-Step "运行 AI Runtime 协议单测"
    Push-Location $AIRuntimeRoot
    try {
        & $PythonExe -m unittest discover -s tests -p "test_*.py"
    } finally {
        Pop-Location
    }
}

Write-Step "检查 AI Runtime 健康状态"

$HealthResult = Invoke-JsonRequest `
    -Method "GET" `
    -Uri "$AIRuntimeBaseUrl/readyz"

Assert-True ($HealthResult.StatusCode -eq 200) "readyz 应返回 200"
Assert-True ($HealthResult.Body.status -eq "ready") "readyz 状态应为 ready"

Write-Step "校验 /internal/v1/llm/chat"

$ChatPayload = @{
    provider = "mock"
    model    = "phase5-smoke"
    messages = @(
        @{
            role    = "user"
            content = "你好，Phase 5"
        }
    )
    metadata = @{
        trace_id = "phase5-smoke"
        run_id   = "phase5-smoke"
        node_id  = "phase5-smoke"
    }
}

$ChatResult = Invoke-JsonRequest `
    -Method "POST" `
    -Uri "$AIRuntimeBaseUrl/internal/v1/llm/chat" `
    -Body $ChatPayload

Assert-True ($ChatResult.StatusCode -eq 200) "chat 接口应返回 200"
Assert-True (-not [string]::IsNullOrWhiteSpace($ChatResult.Body.data.text)) "chat 返回 text 不应为空"
Assert-True ($ChatResult.Body.data.message.role -eq "assistant") "chat message role 应为 assistant"
Assert-True ($ChatResult.Body.data.token_usage.total_tokens -gt 0) "chat token_usage.total_tokens 应大于 0"

Write-Step "校验 /internal/v1/llm/stream"

$StreamResult = Invoke-JsonRequest `
    -Method "POST" `
    -Uri "$AIRuntimeBaseUrl/internal/v1/llm/stream" `
    -Body $ChatPayload `
    -Accept "text/event-stream"

$EventNames = Get-SseEventNames -Content $StreamResult.BodyText

Assert-True ($StreamResult.StatusCode -eq 200) "stream 接口应返回 200"
Assert-True ($EventNames.Count -gt 0) "stream 应至少返回一个 SSE 事件"
Assert-True ($EventNames -contains "start") "stream 应包含 start 事件"
Assert-True ($EventNames -contains "delta") "stream 应包含 delta 事件"
Assert-True ($EventNames -contains "usage") "stream 应包含 usage 事件"
Assert-True ($EventNames[-1] -eq "done") "stream 最后一个事件应为 done"

Write-Step "校验 /internal/v1/tools/call"

$ToolResult = Invoke-JsonRequest `
    -Method "POST" `
    -Uri "$AIRuntimeBaseUrl/internal/v1/tools/call" `
    -Body @{
        tool_call_id = "call_phase5_smoke"
        tool_name    = "search_docs"
        arguments    = @{
            query = "phase 5"
        }
        workspace_id = "workspace_smoke"
        workflow_id  = "workflow_smoke"
        run_id       = "run_smoke"
        node_id      = "node_smoke"
        timeout_ms   = 30000
        metadata     = @{
            source = "phase5-smoke"
        }
    }

Assert-True ($ToolResult.StatusCode -eq 501) "tool/call 应返回 501"
Assert-True ($ToolResult.Body.error.code -eq "AI_RUNTIME_TOOL_EXECUTION_NOT_IMPLEMENTED") "tool/call 错误码应正确"

Write-Host ""
Write-Host "Phase 5 验证脚本执行通过。" -ForegroundColor Green
