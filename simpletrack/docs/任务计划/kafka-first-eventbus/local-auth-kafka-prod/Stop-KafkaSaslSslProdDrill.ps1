param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$RemoveVolumes
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$certDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth-prod\certs"))
$composeFile = Join-Path $PSScriptRoot "docker-compose.sasl-ssl-prod.yml"
$projectName = "simpletrack-kafka-auth-prod"

$env:KAFKA_AUTH_PROD_CERTS_DIR = $certDir.Replace("\", "/")
$composeEnv = Join-Path $certDir "compose-env.ps1"
if (Test-Path -LiteralPath $composeEnv) {
    . $composeEnv
}
if ($RemoveVolumes) {
    docker compose -f $composeFile --project-name $projectName down --volumes --remove-orphans
} else {
    docker compose -f $composeFile --project-name $projectName down --remove-orphans
}
