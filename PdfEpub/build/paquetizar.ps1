param([Parameter(Mandatory = $true)][string]$Version)
$ErrorActionPreference = 'Stop'
$raiz = Split-Path $PSScriptRoot -Parent
$dist = Join-Path $raiz 'dist'
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem

function Get-RelativePath([string]$Root, [string]$Path) {
    $raizAbs = [IO.Path]::GetFullPath($Root).TrimEnd('\', '/')
    $pathAbs = [IO.Path]::GetFullPath($Path)
    $prefijo = $raizAbs + [IO.Path]::DirectorySeparatorChar
    if (-not $pathAbs.StartsWith($prefijo, [StringComparison]::OrdinalIgnoreCase)) {
        throw "El archivo no está dentro de la carpeta esperada: $Path"
    }
    return $pathAbs.Substring($prefijo.Length)
}

function Add-ZipEntry($Archive, [string]$Path, [string]$Name) {
    $entry = $Archive.CreateEntry($Name.Replace('\', '/'), [IO.Compression.CompressionLevel]::Optimal)
    $entrada = [IO.File]::OpenRead($Path)
    try {
        $salida = $entry.Open()
        try { $entrada.CopyTo($salida) } finally { $salida.Dispose() }
    } finally { $entrada.Dispose() }
}

function New-Zip([string]$Path, [array]$Files, [string]$Root) {
    $destino = [IO.Path]::GetFullPath($Path)
    $distAbs = [IO.Path]::GetFullPath($dist) + [IO.Path]::DirectorySeparatorChar
    if (-not $destino.StartsWith($distAbs, [StringComparison]::OrdinalIgnoreCase)) {
        throw 'La ruta del ZIP debe estar dentro de dist.'
    }
    if (Test-Path -LiteralPath $destino) { Remove-Item -LiteralPath $destino -Force }
    $archivoZip = [IO.Compression.ZipFile]::Open($destino, [IO.Compression.ZipArchiveMode]::Create)
    try {
        foreach ($archivo in $Files) {
            $relativo = $archivo.Entry
            if (-not $relativo) { $relativo = Get-RelativePath $Root $archivo.FullName }
            Add-ZipEntry $archivoZip $archivo.FullName $relativo
        }
    } finally { $archivoZip.Dispose() }
}

$fuentes = Get-ChildItem -LiteralPath $raiz -File -Recurse -Force | Where-Object {
    $rel = (Get-RelativePath $raiz $_.FullName).Replace('/', '\')
    -not ($rel.StartsWith('dist\', [StringComparison]::OrdinalIgnoreCase) -or
          $rel.StartsWith('build\setup\payload\', [StringComparison]::OrdinalIgnoreCase) -or
          $rel.StartsWith('build\iexpress-payload\', [StringComparison]::OrdinalIgnoreCase) -or
          $rel -eq 'build\setup\payload.zip' -or
          $rel -eq 'build\iexpress.sed' -or
          $_.Extension -in @('.exe', '.syso'))
}
$portable = @()
foreach ($nombre in @('PdfEpub.exe', 'LICENSE', 'README.md', 'THIRD-PARTY-NOTICES.md', 'THIRD-PARTY-LICENSES.txt')) {
    $fuente = Join-Path $dist $nombre
    $portable += [pscustomobject]@{ FullName = $fuente; Entry = $nombre }
}
$web = Join-Path $raiz 'web'
foreach ($archivo in Get-ChildItem -LiteralPath $web -File -Recurse -Force) {
    $relativo = Get-RelativePath $web $archivo.FullName
    $portable += [pscustomobject]@{ FullName = $archivo.FullName; Entry = "web\$relativo" }
}

New-Zip (Join-Path $dist "PdfEpub-$Version-source.zip") $fuentes $raiz
New-Zip (Join-Path $dist "PdfEpub-$Version-portable.zip") $portable $dist
Write-Host "Código fuente: $dist\PdfEpub-$Version-source.zip"
Write-Host "Versión portable: $dist\PdfEpub-$Version-portable.zip"
