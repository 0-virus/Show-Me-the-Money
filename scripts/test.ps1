param([switch]$Wasm)
$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
$go = if (Test-Path '.tools/go/bin/go.exe') { (Resolve-Path '.tools/go/bin/go.exe').Path } else { 'go' }
$env:GOCACHE="$PWD/.tools/gocache"
$env:GOMODCACHE="$PWD/.tools/gomodcache"
$savedGoOS=$env:GOOS
$savedGoArch=$env:GOARCH
try {
if ($Wasm) {
    $env:GOOS='js'; $env:GOARCH='wasm'
    $goroot = & $go env GOROOT
    $packages = & $go list ./internal/...
    foreach($package in $packages) {
        $name = ($package -split '/')[-1]
        & $go test -c -o ".tools/$name.test.wasm" $package
        if ($LASTEXITCODE -ne 0) { throw "Test compilation failed: $package" }
        & node "$goroot/lib/wasm/wasm_exec_node.js" ".tools/$name.test.wasm" '-test.v'
        if ($LASTEXITCODE -ne 0) { throw "Tests failed: $package" }
    }
} else {
    & $go test ./... -count=1
    if ($LASTEXITCODE -ne 0) { throw 'Tests failed' }
}
} finally {
    $env:GOOS=$savedGoOS
    $env:GOARCH=$savedGoArch
}
