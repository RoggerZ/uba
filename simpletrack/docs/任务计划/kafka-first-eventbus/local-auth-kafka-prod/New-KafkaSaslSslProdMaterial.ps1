param(
    [string]$RepoRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot "..\..\..\..\..")).Path,
    [switch]$Force
)

$ErrorActionPreference = "Stop"

$resolvedRepo = [System.IO.Path]::GetFullPath($RepoRoot)
$runtimeDir = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp\kafka-auth-prod"))
$certDir = [System.IO.Path]::GetFullPath((Join-Path $runtimeDir "certs"))
$expectedPrefix = [System.IO.Path]::GetFullPath((Join-Path $resolvedRepo ".tmp"))

if (-not $runtimeDir.StartsWith($expectedPrefix, [System.StringComparison]::OrdinalIgnoreCase)) {
    throw "Refusing to write Kafka production drill material outside repo .tmp: $runtimeDir"
}

$keystore = Join-Path $certDir "broker-1.keystore.p12"
if ((Test-Path -LiteralPath $keystore) -and -not $Force) {
    Write-Host "Kafka production SASL_SSL material already exists: $certDir"
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

function New-DisposableSecret {
    $bytes = [byte[]]::new(24)
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    return [Convert]::ToBase64String($bytes).TrimEnd("=").Replace("+", "A").Replace("/", "B")
}

$storePassword = New-DisposableSecret
$brokerPassword = New-DisposableSecret
$clientPassword = New-DisposableSecret

$bash = @'
set -euo pipefail
cd /work/certs

openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 365 -out ca.pem -subj "/CN=SimpleTrack Local Kafka Prod CA"

for broker_id in 1 2 3; do
  broker_host="kafka-auth-prod-${broker_id}"
  cat > "broker-${broker_id}.cnf" <<EOF
[req]
distinguished_name=req_distinguished_name
req_extensions=v3_req
prompt=no

[req_distinguished_name]
CN=${broker_host}

[v3_req]
subjectAltName=@alt_names

[alt_names]
DNS.1=localhost
DNS.2=${broker_host}
DNS.3=simpletrack-kafka-auth-prod-${broker_id}
DNS.4=simpletrack-kafka-auth-prod-${broker_id}-kafka
IP.1=127.0.0.1
EOF

  openssl genrsa -out "broker-${broker_id}.key" 2048
  openssl req -new -key "broker-${broker_id}.key" -out "broker-${broker_id}.csr" -config "broker-${broker_id}.cnf"
  openssl x509 -req -in "broker-${broker_id}.csr" -CA ca.pem -CAkey ca.key -CAcreateserial -out "broker-${broker_id}.pem" -days 365 -sha256 -extensions v3_req -extfile "broker-${broker_id}.cnf"
  openssl pkcs12 -export -in "broker-${broker_id}.pem" -inkey "broker-${broker_id}.key" -certfile ca.pem -out "broker-${broker_id}.keystore.p12" -password pass:__STORE_PASSWORD__ -name "broker-${broker_id}"
done

keytool -importcert -alias ca -file ca.pem -keystore broker.truststore.p12 -storetype PKCS12 -storepass __STORE_PASSWORD__ -noprompt

cat > kafka_server_jaas.conf <<'EOF'
KafkaServer {
  org.apache.kafka.common.security.plain.PlainLoginModule required
  username="broker"
  password="__BROKER_PASSWORD__"
  user_broker="__BROKER_PASSWORD__"
  user_simpletrack="__CLIENT_PASSWORD__";
};
EOF

cat > client.properties <<'EOF'
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="simpletrack" password="__CLIENT_PASSWORD__";
ssl.truststore.location=/etc/kafka/secrets/broker.truststore.p12
ssl.truststore.password=__STORE_PASSWORD__
ssl.truststore.type=PKCS12
ssl.endpoint.identification.algorithm=https
EOF

cat > analytics-core-auth-prod-env.ps1 <<'EOF'
$env:ANALYTICS_CORE_KAFKA_INTEGRATION = "1"
$env:ANALYTICS_CORE_KAFKA_REPLICATED_INTEGRATION = "1"
$env:ANALYTICS_CORE_KAFKA_OUTAGE_INTEGRATION = "1"
$env:ANALYTICS_CORE_KAFKA_OUTAGE_STOP_CONTAINER = "simpletrack-kafka-auth-prod-3"
$env:ANALYTICS_CORE_KAFKA_BROKERS = "127.0.0.1:39193,127.0.0.1:39194,127.0.0.1:39195"
$env:ANALYTICS_CORE_KAFKA_TOPIC_PARTITIONS = "1"
$env:ANALYTICS_CORE_KAFKA_TOPIC_REPLICATION_FACTOR = "3"
$env:ANALYTICS_CORE_KAFKA_TOPIC_MIN_INSYNC_REPLICAS = "2"
$env:ANALYTICS_CORE_KAFKA_TLS_ENABLED = "true"
$env:ANALYTICS_CORE_KAFKA_TLS_SERVER_NAME = "localhost"
$env:ANALYTICS_CORE_KAFKA_TLS_CA_FILE = "__CA_FILE__"
$env:ANALYTICS_CORE_KAFKA_SASL_ENABLED = "true"
$env:ANALYTICS_CORE_KAFKA_SASL_MECHANISM = "plain"
$env:ANALYTICS_CORE_KAFKA_SASL_USERNAME = "simpletrack"
$env:ANALYTICS_CORE_KAFKA_SASL_PASSWORD = "__CLIENT_PASSWORD__"
$env:ANALYTICS_CORE_KAFKA_SASL_HANDSHAKE = "true"
EOF

cat > compose-env.ps1 <<'EOF'
$env:KAFKA_AUTH_PROD_STORE_PASSWORD = "__STORE_PASSWORD__"
EOF
'@

$bash = $bash.Replace("__STORE_PASSWORD__", $storePassword).Replace("__BROKER_PASSWORD__", $brokerPassword).Replace("__CLIENT_PASSWORD__", $clientPassword)

$scriptPath = Join-Path $runtimeDir "generate-prod-material.sh"
[System.IO.File]::WriteAllText($scriptPath, $bash, [System.Text.UTF8Encoding]::new($false))

$dockerRuntime = $runtimeDir.Replace("\", "/")
docker run --rm -v "${dockerRuntime}:/work" confluentinc/cp-kafka:7.5.0 bash /work/generate-prod-material.sh

$envFile = Join-Path $certDir "analytics-core-auth-prod-env.ps1"
$caFile = (Join-Path $certDir "ca.pem").Replace("\", "\\")
$envContent = [System.IO.File]::ReadAllText($envFile, [System.Text.Encoding]::UTF8)
$envContent = $envContent.Replace("__CA_FILE__", $caFile)
[System.IO.File]::WriteAllText($envFile, $envContent, [System.Text.UTF8Encoding]::new($false))

Write-Host "Generated disposable Kafka production SASL_SSL material in $certDir"
Write-Host "Do not commit files under .tmp/kafka-auth-prod; they contain local private keys and test passwords."
