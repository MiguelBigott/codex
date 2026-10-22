param([Parameter(Mandatory = $true)][string]$Version)
$ErrorActionPreference = 'Stop'
$raiz = Split-Path $PSScriptRoot -Parent
Set-Location $raiz
$destino = Join-Path $raiz 'dist'
$setupDir = Join-Path $PSScriptRoot 'setup'
$zip = Join-Path $setupDir 'payload.zip'
$salida = Join-Path $destino "PdfEpub-Setup-$Version.exe"
$carga = Join-Path ([IO.Path]::GetTempPath()) "PdfEpub-setup-$Version-$([guid]::NewGuid().ToString('N'))"
$go = (Get-Command go -ErrorAction Stop).Source
$gopath = (& $go env GOPATH).Trim()
$winres = Join-Path $gopath 'bin\go-winres.exe'

try {
    New-Item -ItemType Directory -Force -Path $carga | Out-Null
    foreach ($archivo in @('PdfEpub.exe', 'LICENSE', 'README.md', 'THIRD-PARTY-NOTICES.md', 'THIRD-PARTY-LICENSES.txt')) {
        $origen = Join-Path $destino $archivo
        if (-not (Test-Path -LiteralPath $origen)) { throw "Falta el archivo requerido: $origen" }
        Copy-Item -LiteralPath $origen -Destination $carga -Force
    }
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'instalar.ps1') -Destination (Join-Path $carga 'instalar.ps1') -Force
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'desinstalar.ps1') -Destination (Join-Path $carga 'Desinstalar.ps1') -Force

    if (Test-Path -LiteralPath $zip) { Remove-Item -LiteralPath $zip -Force }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    [IO.Compression.ZipFile]::CreateFromDirectory($carga, $zip, [IO.Compression.CompressionLevel]::Optimal, $false)

    if (-not (Test-Path -LiteralPath $winres)) { throw 'Falta go-winres; instálalo con: go install github.com/tc-hib/go-winres@latest' }
    & $winres simply --arch amd64 --out build\setup\rsrc --manifest gui --icon build\app.ico `
        --product-name 'PdfEpub Setup' --file-description 'Instalador de PdfEpub' `
        --product-version $Version --file-version $Version `
        --original-filename "PdfEpub-Setup-$Version.exe" --copyright 'Copyright (c) 2026 PdfEpub contributors'
    if ($LASTEXITCODE) { throw 'No se pudo asignar el icono al instalador.' }

    & $go build -trimpath -ldflags "-H windowsgui -s -w -X main.version=$Version" -o $salida ./build/setup
    if ($LASTEXITCODE) { throw 'No se pudo compilar el instalador.' }
    Write-Host "Instalador: $salida"
} finally {
    if (Test-Path -LiteralPath $carga) { Remove-Item -LiteralPath $carga -Recurse -Force }
}
