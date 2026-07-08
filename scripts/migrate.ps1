$env:DATABASE_URL = if ($env:DATABASE_URL) { $env:DATABASE_URL } else { "postgres://aipass:aipass@localhost:5432/aipass?sslmode=disable" }
$go = Get-Command go -ErrorAction SilentlyContinue
if (-not $go -and (Test-Path "C:\Program Files\Go\bin\go.exe")) {
    $go = "C:\Program Files\Go\bin\go.exe"
}
if (-not $go) {
    throw "Go is not available in PATH and was not found at C:\Program Files\Go\bin\go.exe"
}
& $go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.1 -path migrations -database $env:DATABASE_URL up
