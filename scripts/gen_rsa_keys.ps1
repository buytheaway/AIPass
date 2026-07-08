$keyDir = "deployments/docker/keys"
New-Item -ItemType Directory -Force -Path $keyDir | Out-Null
openssl genrsa -out "$keyDir/private.pem" 2048
openssl rsa -in "$keyDir/private.pem" -pubout -out "$keyDir/public.pem"
Write-Host "Generated RSA keys in $keyDir"

