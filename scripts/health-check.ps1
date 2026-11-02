param(
    [ValidateSet("local", "docker")]
    [string]$Mode = "local"
)

$ErrorActionPreference = "Stop"

if ($Mode -eq "docker") {
    $Targets = @(
        @{
            Name = "Nginx 聚合入口"
            Url  = "http://localhost/api/health"
        },
        @{
            Name = "Web 前端"
            Url  = "http://localhost:3000/api/health"
        },
        @{
            Name = "Go API"
            Url  = "http://localhost:8080/readyz"
        },
        @{
            Name = "AI Runtime"
            Url  = "http://localhost:8090/readyz"
        }
    )
} else {
    $Targets = @(
        @{
            Name = "Web 前端"
            Url  = "http://localhost:3000/api/health"
        },
        @{
            Name = "Go API"
            Url  = "http://localhost:8080/readyz"
        },
        @{
            Name = "AI Runtime"
            Url  = "http://localhost:8090/readyz"
        }
    )
}

$Results = @()
$FailedCount = 0

foreach ($Target in $Targets) {
    $StartedAt = Get-Date

    try {
        Invoke-RestMethod -Uri $Target.Url -Method Get -TimeoutSec 5 | Out-Null

        $FinishedAt = Get-Date
        $LatencyMs = [math]::Round(($FinishedAt - $StartedAt).TotalMilliseconds, 2)

        $Results += [pscustomobject]@{
            服务 = $Target.Name
            状态 = "正常"
            地址 = $Target.Url
            耗时毫秒 = $LatencyMs
        }
    } catch {
        $FinishedAt = Get-Date
        $LatencyMs = [math]::Round(($FinishedAt - $StartedAt).TotalMilliseconds, 2)
        $FailedCount++

        $Results += [pscustomobject]@{
            服务 = $Target.Name
            状态 = "异常"
            地址 = $Target.Url
            耗时毫秒 = $LatencyMs
        }
    }
}

$Results | Format-Table -AutoSize

if ($FailedCount -gt 0) {
    Write-Host ""
    Write-Host "健康检查失败数量：$FailedCount" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "全部服务健康检查通过。" -ForegroundColor Green