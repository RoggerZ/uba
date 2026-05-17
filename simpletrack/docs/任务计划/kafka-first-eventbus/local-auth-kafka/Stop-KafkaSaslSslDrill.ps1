param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$RemoveVolumes
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$certDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth\certs"))
$composeFile = Join-Path $PSScriptRoot "docker-compose.sasl-ssl.yml"
$projectName = "simpletrack-kafka-auth"

$env:KAFKA_AUTH_CERTS_DIR = $certDir.Replace("\", "/")
if ($RemoveVolumes) {
    docker compose -f $composeFile --project-name $projectName down --volumes --remove-orphans
} else {
    docker compose -f $composeFile --project-name $projectName down --remove-orphans
}
