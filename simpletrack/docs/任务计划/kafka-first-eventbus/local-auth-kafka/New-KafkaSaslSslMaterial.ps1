param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$runtimeDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth"))
$certDir = [System.IO.Path]::GetFullPath((Join-Path $runtimeDir "certs"))
$expectedPrefix = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp"))

if (-not $runtimeDir.StartsWith($expectedPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to write Kafka drill material outside repo .tmp: $runtimeDir"
}

$keystore = Join-Path $certDir "broker.keystore.p12"
if ((Test-Path -LiteralPath $keystore) -and -not $Force) {
    Write-Host "Kafka SASL_SSL material already exists: $certDir"
    Write-Host "Use -Force to regenerate disposable certificates."
    return
}

if ((Test-Path -LiteralPath $certDir) -and $Force) {
    $resolvedCertDir = [System.IO.Path]::GetFullPath($certDir)
    if (-not $resolvedCertDir.StartsWith($expectedPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to remove path outside repo .tmp: $resolvedCertDir"
    }
    Remove-Item -LiteralPath $certDir -Recurse -Force
}

New-Item -ItemType Directory -Force -Path $certDir | Out-Null

$bash = @'
set -euo pipefail
cd /work/certs

cat > broker.cnf <<'EOF'
[req]
distinguished_name=req_distinguished_name
req_extensions=v3_req
prompt=no

[req_distinguished_name]
CN=localhost

[v3_req]
subjectAltName=@alt_names

[alt_names]
DNS.1=localhost
DNS.2=kafka-auth
DNS.3=simpletrack-kafka-auth
DNS.4=simpletrack-kafka-auth-kafka
IP.1=127.0.0.1
EOF

openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.pem -subj "/CN=SimpleTrack Local Kafka CA"
openssl genrsa -out broker.key 2048
openssl req -new -key broker.key -out broker.csr -config broker.cnf
openssl x509 -req -in broker.csr -CA ca.pem -CAkey ca.key -CAcreateserial -out broker.pem -days 365 -sha256 -extensions v3_req -extfile broker.cnf
openssl pkcs12 -export -in broker.pem -inkey broker.key -certfile ca.pem -out broker.keystore.p12 -password pass:changeit -name broker
keytool -importcert -alias ca -file ca.pem -keystore broker.truststore.p12 -storetype PKCS12 -storepass changeit -noprompt

cat > kafka_server_jaas.conf <<'EOF'
KafkaServer {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="broker"
  password="broker-secret"
  user_broker="broker-secret"
  user_simpletrack="simpletrack-secret";
};
EOF

cat > client.properties <<'EOF'
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="simpletrack" password="simpletrack-secret";
ssl.truststore.location=/etc/kafka/secrets/broker.truststore.p12
ssl.truststore.password=changeit
ssl.truststore.type=PKCS12
ssl.endpoint.identification.algorithm=https
EOF

cat > analytics-core-auth-env.ps1 <<'EOF'
$env:ANALYTICS_CORE_KAFKA_INTEGRATION = "1"
$env:ANALYTICS_CORE_KAFKA_BROKERS = "127.0.0.1:39093"
$env:ANALYTICS_CORE_KAFKA_TLS_ENABLED = "true"
$env:ANALYTICS_CORE_KAFKA_TLS_SERVER_NAME = "localhost"
$env:ANALYTICS_CORE_KAFKA_TLS_CA_FILE = "__CA_FILE__"
$env:ANALYTICS_CORE_KAFKA_SASL_ENABLED = "true"
$env:ANALYTICS_CORE_KAFKA_SASL_MECHANISM = "plain"
$env:ANALYTICS_CORE_KAFKA_SASL_USERNAME = "simpletrack"
$env:ANALYTICS_CORE_KAFKA_SASL_PASSWORD = "simpletrack-secret"
$env:ANALYTICS_CORE_KAFKA_SASL_HANDSHAKE = "true"
EOF
'@

$scriptPath = Join-Path $runtimeDir "generate-material.sh"
[System.IO.File]::WriteAllText($scriptPath, $bash, [System.Text.UTF8Encoding]::new($false))

$dockerRuntime = $runtimeDir.Replace("\", "/")
docker run --rm -v "${dockerRuntime}:/work" confluentinc/cp-kafka:7.5.0 bash /work/generate-material.sh

$envFile = Join-Path $certDir "analytics-core-auth-env.ps1"
$caFile = (Join-Path $certDir "ca.pem").Replace("\", "\\")
$envContent = [System.IO.File]::ReadAllText($envFile, [System.Text.Encoding]::UTF8)
$envContent = $envContent.Replace("__CA_FILE__", $caFile)
[System.IO.File]::WriteAllText($envFile, $envContent, [System.Text.UTF8Encoding]::new($false))

Write-Host "Generated disposable Kafka SASL_SSL material in $certDir"
Write-Host "Do not commit files under .tmp/kafka-auth; they contain local private keys and test passwords."
