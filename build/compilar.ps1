# Compila Codex.exe y su instalador en dist\.
# Requiere: Go, NSIS (makensis) y go-winres (go install github.com/tc-hib/go-winres@latest).
$ErrorActionPreference = 'Stop'
$raiz = Split-Path $PSScriptRoot -Parent
Set-Location $raiz

$version = (Select-String -Path main.go -Pattern 'const Version = "(.+)"').Matches[0].Groups[1].Value
$goExe = (Get-Command go -ErrorAction SilentlyContinue).Source
if (-not $goExe) { $goExe = 'C:\Program Files\Go\bin\go.exe' }
$winres = Join-Path (& $goExe env GOPATH) 'bin\go-winres.exe'
$makensis = 'C:\Program Files (x86)\NSIS\makensis.exe'

Write-Host "Codex $version"

# Icono (recurso "APP", el que carga la ventana), manifiesto y datos de versión.
& $winres simply --arch amd64 --out rsrc --manifest gui `
    --icon build\app.ico `
    --product-name 'Codex' --file-description 'Codex' `
    --product-version $version --file-version $version `
    --original-filename 'Codex.exe' --copyright 'Copyright (c) 2026 Miguel Bigott'
if ($LASTEXITCODE) { throw 'go-winres falló' }

New-Item -ItemType Directory -Force dist | Out-Null
& $goExe build -trimpath -ldflags '-H windowsgui -s -w' -o 'dist\Codex.exe' .
if ($LASTEXITCODE) { throw 'go build falló' }

# ---------- Licencias ----------
# LICENSE.txt es la licencia MIT de la app. LICENCIAS-TERCEROS.txt junta las de todo lo que va
# dentro del .exe (Go, sus módulos y las tipografías), que deben acompañar al programa.
Copy-Item LICENSE (Join-Path $raiz 'dist\LICENSE.txt')
$terceros = @("Licencias de terceros incluidas en Codex.exe.")
$terceros += "==== $(& $goExe env GOVERSION): biblioteca estándar y runtime de Go ====`n" + (Get-Content -Raw (Join-Path (& $goExe env GOROOT) 'LICENSE'))
$modulos = & $goExe list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}@{{.Version}}|{{.Dir}}{{end}}{{end}}' . | Sort-Object -Unique
if ($LASTEXITCODE) { throw 'go list falló' }
foreach ($m in $modulos) {
    $nombre, $dir = $m -split '\|', 2
    Get-ChildItem $dir -Recurse -Depth 3 -File | Where-Object { $_.Name -match '^(LICEN[CS]E|COPYING|NOTICE)' } | ForEach-Object {
        $terceros += "==== $nombre ($($_.FullName.Substring($dir.Length + 1))) ====`n" + (Get-Content -Raw $_.FullName)
    }
}
Get-ChildItem web\fonts -Filter 'OFL-*.txt' | ForEach-Object {
    $terceros += "==== Tipografía $($_.BaseName.Substring(4)) ====`n" + (Get-Content -Raw $_.FullName)
}
Set-Content (Join-Path $raiz 'dist\LICENCIAS-TERCEROS.txt') ($terceros -join "`n`n") -Encoding utf8

# El .nsi está en UTF-8 sin BOM: sin /INPUTCHARSET, makensis lo lee como ANSI y rompe las tildes.
& $makensis /V2 /INPUTCHARSET UTF8 "/DVERSION=$version" build\instalador.nsi
if ($LASTEXITCODE) { throw 'makensis falló' }

Get-ChildItem dist -File | Format-Table Name, @{n = 'MB'; e = { [math]::Round($_.Length / 1MB, 2) } }
