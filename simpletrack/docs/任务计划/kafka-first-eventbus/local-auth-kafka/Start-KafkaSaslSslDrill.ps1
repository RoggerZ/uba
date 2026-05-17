param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$RegenerateMaterial,
    [int]$TimeoutSeconds = 180
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$certDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth\certs"))
$composeFile = Join-Path $PSScriptRoot "docker-compose.sasl-ssl.yml"
$projectName = "simpletrack-kafka-auth"
$containerName = "simpletrack-kafka-auth"

& (Join-Path $PSScriptRoot "New-KafkaSaslSslMaterial.ps1") -RepoRoot $resolvedRepo -Force:$RegenerateMaterial

$env:KAFKA_AUTH_CERTS_DIR = $certDir.Replace("\", "/")
docker compose -f $composeFile --project-name $projectName up -d --remove-orphans

$deadline = (Get-Date).AddSeconds($TimeoutSeconds)
$lastOutput = ""
while ((Get-Date) -lt $deadline) {
    $output = docker exec $containerName kafka-topics --bootstrap-server kafka-auth:9094 --command-config /etc/kafka/secrets/client.properties --list 2>&1
    $lastOutput = ($output | Out-String).Trim()
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Kafka SASL_SSL drill broker is ready at 127.0.0.1:39093"
        Write-Host "Load env with: . $($certDir)\analytics-core-auth-env.ps1"
        return
    }
    Start-Sleep -Seconds 5
}

docker logs --tail 120 $containerName
throw "Kafka SASL_SSL drill broker did not become ready within $TimeoutSeconds seconds. Last probe: $lastOutput"
