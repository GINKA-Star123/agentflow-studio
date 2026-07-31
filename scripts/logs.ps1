param(
    [string]$Service = "",
    [switch]$Follow
)

$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$ComposeFile = Join-Path $Root "deployments\docker\docker-compose.yml"

$Args = @(
    "compose",
    "-f",
    $ComposeFile,
    "logs"
)

if ($Follow) {
    $Args += "-f"
}

if ($Service -ne "") {
    $Args += $Service
}

docker @Args