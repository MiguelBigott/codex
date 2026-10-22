param([switch]$SoloExe)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
$raiz = Split-Path $PSScriptRoot -Parent
Set-Location $raiz

$version = (Select-String -Path main.go -Pattern 'const Version = "(.+)"').Matches[0].Groups[1].Value
$go = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $go) { throw 'Instala Go 1.27.3 o posterior para compilar PdfEpub.' }
$gopath = (& $go env GOPATH).Trim()
$winres = Join-Path $gopath 'bin\go-winres.exe'
$python = (Get-Command python -ErrorAction SilentlyContinue).Source
if (-not $python) { throw 'Instala Python 3 y Pillow para generar los iconos.' }
if (-not (Test-Path $winres)) { throw 'Falta go-winres. Instálalo con: go install github.com/tc-hib/go-winres@latest' }

Write-Host "PdfEpub $version"
& $python (Join-Path $PSScriptRoot 'generar_logo.py')
if ($LASTEXITCODE) { throw 'No se pudo generar el icono.' }
& $winres simply --arch amd64 --out rsrc --manifest gui --icon build\app.ico `
    --product-name PdfEpub --file-description 'PdfEpub - convierte PDF a EPUB' `
    --product-version $version --file-version $version `
    --original-filename PdfEpub.exe --copyright 'Copyright (c) 2026 Miguel Bigott'
if ($LASTEXITCODE) { throw 'go-winres falló.' }

New-Item -ItemType Directory -Force dist | Out-Null
& $go build -trimpath -ldflags '-H windowsgui -s -w' -o dist\PdfEpub.exe .
if ($LASTEXITCODE) { throw 'go build falló.' }
Copy-Item LICENSE, THIRD-PARTY-NOTICES.md, THIRD-PARTY-LICENSES.txt, README.md dist -Force

if (-not $SoloExe) {
    & (Join-Path $PSScriptRoot 'empaquetar-setup.ps1') -Version $version
    & (Join-Path $PSScriptRoot 'paquetizar.ps1') -Version $version
}
Write-Host "Ejecutable: $raiz\dist\PdfEpub.exe"
