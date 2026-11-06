param([switch]$SkipTests)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
$go = if (Test-Path '.tools/go/bin/go.exe') { (Resolve-Path '.tools/go/bin/go.exe').Path } else { 'go' }
$env:GOCACHE="$PWD/.tools/gocache"
$env:GOMODCACHE="$PWD/.tools/gomodcache"
$env:CGO_ENABLED='0'
if (!$SkipTests) {
    & $go test ./...
    if ($LASTEXITCODE -ne 0) { throw 'Tests failed. See README for the WebAssembly test option.' }
}
foreach ($target in @(@('windows','amd64','MoneyCopy.exe'), @('darwin','amd64','moneycopy-amd64'), @('darwin','arm64','moneycopy-arm64'))) {
    $env:GOOS=$target[0];$env:GOARCH=$target[1]
    $out="artifacts/build/$($target[2])"
    & $go build -trimpath '-ldflags=-s -w' -o $out ./cmd/moneycopy
    if ($LASTEXITCODE -ne 0) {throw "Build failed: $target"}
    Write-Output "Built $out"
}
Remove-Item Env:GOOS,Env:GOARCH
& python scripts/package.py
if ($LASTEXITCODE -ne 0) { throw 'Packaging failed' }
