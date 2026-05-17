param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$RegenerateMaterial,
    [int]$TimeoutSeconds = 240
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$certDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth-prod\certs"))
$composeFile = Join-Path $PSScriptRoot "docker-compose.sasl-ssl-prod.yml"
$projectName = "simpletrack-kafka-auth-prod"
$probeContainer = "simpletrack-kafka-auth-prod-1"

& (Join-Path $PSScriptRoot "New-KafkaSaslSslProdMaterial.ps1") -RepoRoot $resolvedRepo -Force:$RegenerateMaterial

$env:KAFKA_AUTH_PROD_CERTS_DIR = $certDir.Replace("\", "/")
$composeEnv = Join-Path $certDir "compose-env.ps1"
if (-not (Test-Path -LiteralPath $composeEnv)) {
    throw "Kafka production drill compose env was not generated: $composeEnv"
}
. $composeEnv
docker compose -f $composeFile --project-name $projectName up -d --remove-orphans

$deadline = (Get-Date).AddSeconds($TimeoutSeconds)
$lastOutput = ""
while ((Get-Date) -lt $deadline) {
    $output = docker exec $probeContainer kafka-topics --bootstrap-server kafka-auth-prod-1:9094 --command-config /etc/kafka/secrets/client.properties --list 2>&1
    $lastOutput = ($output | Out-String).Trim()
    if ($LASTEXITCODE -eq 0) {
        $brokerList = "127.0.0.1:39193,127.0.0.1:39194,127.0.0.1:39195"
        Write-Host "Kafka production SASL_SSL drill cluster is ready at $brokerList"
        Write-Host "Load env with: . $($certDir)\analytics-core-auth-prod-env.ps1"
        return
    }
    Start-Sleep -Seconds 5
}

docker ps --filter "name=simpletrack-kafka-auth-prod" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
docker logs --tail 120 $probeContainer
throw "Kafka production SASL_SSL drill cluster did not become ready within $TimeoutSeconds seconds. Last probe: $lastOutput"
