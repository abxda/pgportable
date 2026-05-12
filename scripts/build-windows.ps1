# Build de PgPortable para Windows x86-64.
# Requiere: Go 1.22+, MSYS2 gcc (C:\msys64\mingw64\bin\gcc.exe), Wails CLI v2.

$ErrorActionPreference = "Stop"

$env:PATH = "C:\msys64\mingw64\bin;C:\Program Files\Go\bin;$env:USERPROFILE\go\bin;$env:PATH"
$env:CGO_ENABLED = "1"

Set-Location (Resolve-Path "$PSScriptRoot\..")

Write-Host "[1/2] go mod tidy"
go mod tidy

Write-Host "[2/2] wails build (windows/amd64)"
wails build -platform windows/amd64 -trimpath

$exe = "build\bin\PgPortable.exe"
if (-not (Test-Path $exe)) { throw "Build failed: $exe no existe" }

$mb = [math]::Round((Get-Item $exe).Length / 1MB, 2)
Write-Host ("OK Built {0} ({1} MB)" -f $exe, $mb)
