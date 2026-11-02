param(
    [string]$BaseUrl = "http://localhost:8080/api/v1"
)

$ErrorActionPreference = "Stop"

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
        throw "断言失败：$Message"
    }
}

function Invoke-Api {
    param(
        [string]$Method,
        [string]$Path,
        [object]$Body = $null,
        [string]$Token = "",
        [switch]$AllowError
    )

    $Headers = @{}

    if ($Token -ne "") {
        $Headers["Authorization"] = "Bearer $Token"
    }

    $Params = @{
        Uri         = "$BaseUrl$Path"
        Method      = $Method
        Headers     = $Headers
        TimeoutSec  = 10
        ContentType = "application/json"
    }

    if ($null -ne $Body) {
        $Params["Body"] = ($Body | ConvertTo-Json -Depth 10)
    }

    try {
        return Invoke-RestMethod @Params
    } catch {
        if ($AllowError) {
            $Response = $_.Exception.Response

            if ($null -ne $Response) {
                return [pscustomobject]@{
                    StatusCode = [int]$Response.StatusCode
                    Error      = $_
                }
            }

            return [pscustomobject]@{
                StatusCode = 0
                Error      = $_
            }
        }

        throw
    }
}

$Suffix = [guid]::NewGuid().ToString("N").Substring(0, 8)

$OwnerEmail = "owner-$Suffix@example.com"
$TargetEmail = "target-$Suffix@example.com"
$ThirdEmail = "third-$Suffix@example.com"

Write-Step "检查未携带 Token 访问 /auth/me 应返回 401"

$MissingTokenResult = Invoke-Api `
    -Method "GET" `
    -Path "/auth/me" `
    -AllowError

Assert-True ($MissingTokenResult.StatusCode -eq 401) "未携带 Token 应返回 401"

Write-Step "注册 owner 用户"

$OwnerResult = Invoke-Api `
    -Method "POST" `
    -Path "/auth/register" `
    -Body @{
        email          = $OwnerEmail
        password       = "password123"
        display_name   = "Owner $Suffix"
        workspace_name = "Workspace $Suffix"
    }

$OwnerToken = $OwnerResult.data.access_token
$WorkspaceID = $OwnerResult.data.current_workspace.id
$OwnerUserID = $OwnerResult.data.user.id

Assert-True ($OwnerToken -ne "") "owner token 不能为空"
Assert-True ($WorkspaceID -ne "") "workspace id 不能为空"
Assert-True ($OwnerUserID -ne "") "owner user id 不能为空"

Write-Step "注册 target 用户"

$TargetResult = Invoke-Api `
    -Method "POST" `
    -Path "/auth/register" `
    -Body @{
        email          = $TargetEmail
        password       = "password123"
        display_name   = "Target $Suffix"
        workspace_name = "Target Workspace $Suffix"
    }

$TargetToken = $TargetResult.data.access_token
$TargetUserID = $TargetResult.data.user.id

Assert-True ($TargetToken -ne "") "target token 不能为空"
Assert-True ($TargetUserID -ne "") "target user id 不能为空"

Write-Step "注册 third 用户"

$ThirdResult = Invoke-Api `
    -Method "POST" `
    -Path "/auth/register" `
    -Body @{
        email          = $ThirdEmail
        password       = "password123"
        display_name   = "Third $Suffix"
        workspace_name = "Third Workspace $Suffix"
    }

$ThirdUserID = $ThirdResult.data.user.id

Assert-True ($ThirdUserID -ne "") "third user id 不能为空"

Write-Step "owner 调用 /auth/me"

$MeResult = Invoke-Api `
    -Method "GET" `
    -Path "/auth/me" `
    -Token $OwnerToken

Assert-True ($MeResult.data.user.email -eq $OwnerEmail) "/auth/me 返回的 owner 邮箱不一致"

Write-Step "owner 查询 Workspace 列表"

$WorkspaceList = Invoke-Api `
    -Method "GET" `
    -Path "/workspaces" `
    -Token $OwnerToken

Assert-True (@($WorkspaceList.data.items).Count -ge 1) "owner 至少应该有一个 workspace"

Write-Step "owner 查询成员列表"

$MemberList = Invoke-Api `
    -Method "GET" `
    -Path "/workspaces/$WorkspaceID/members" `
    -Token $OwnerToken

Assert-True (@($MemberList.data.items).Count -ge 1) "成员列表至少应该包含 owner"

Write-Step "owner 添加 target 为 member"

$AddMemberResult = Invoke-Api `
    -Method "POST" `
    -Path "/workspaces/$WorkspaceID/members" `
    -Token $OwnerToken `
    -Body @{
        email = $TargetEmail
        role  = "member"
    }

Assert-True ($AddMemberResult.data.email -eq $TargetEmail) "添加成员返回邮箱不一致"
Assert-True ($AddMemberResult.data.role -eq "member") "添加成员角色应该是 member"

Write-Step "重复添加 target 应返回 409"

$DuplicateAddResult = Invoke-Api `
    -Method "POST" `
    -Path "/workspaces/$WorkspaceID/members" `
    -Token $OwnerToken `
    -Body @{
        email = $TargetEmail
        role  = "member"
    } `
    -AllowError

Assert-True ($DuplicateAddResult.StatusCode -eq 409) "重复添加成员应返回 409"

Write-Step "owner 将 target 角色更新为 viewer"

$UpdateRoleResult = Invoke-Api `
    -Method "PATCH" `
    -Path "/workspaces/$WorkspaceID/members/$TargetUserID/role" `
    -Token $OwnerToken `
    -Body @{
        role = "viewer"
    }

Assert-True ($UpdateRoleResult.data.role -eq "viewer") "target 角色应该更新为 viewer"

Write-Step "viewer 尝试添加 third 成员应返回 403"

$ViewerAddResult = Invoke-Api `
    -Method "POST" `
    -Path "/workspaces/$WorkspaceID/members" `
    -Token $TargetToken `
    -Body @{
        email = $ThirdEmail
        role  = "viewer"
    } `
    -AllowError

Assert-True ($ViewerAddResult.StatusCode -eq 403) "viewer 添加成员应返回 403"

Write-Step "owner 移除 target 成员"

$RemoveResult = Invoke-Api `
    -Method "DELETE" `
    -Path "/workspaces/$WorkspaceID/members/$TargetUserID" `
    -Token $OwnerToken

Assert-True ($RemoveResult.data.removed -eq $true) "移除成员结果应该为 true"

Write-Step "target 被移除后访问成员列表应失败"

$RemovedMemberAccessResult = Invoke-Api `
    -Method "GET" `
    -Path "/workspaces/$WorkspaceID/members" `
    -Token $TargetToken `
    -AllowError

Assert-True (
    $RemovedMemberAccessResult.StatusCode -eq 403 -or
    $RemovedMemberAccessResult.StatusCode -eq 404
) "被移除成员访问 Workspace 应返回 403 或 404"

Write-Host ""
Write-Host "Phase 2 接口验证通过。" -ForegroundColor Green