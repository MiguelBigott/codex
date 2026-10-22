param([string]$Version = '1.0.1')
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms

$origen = Split-Path -Parent $MyInvocation.MyCommand.Path
$destino = Join-Path $env:LOCALAPPDATA 'Programs\PdfEpub'
$nombre = 'PdfEpub'

try {
    $abierto = Get-Process -Name PdfEpub -ErrorAction SilentlyContinue
    if ($abierto) {
        throw 'Cierra PdfEpub y vuelve a ejecutar el instalador.'
    }

    New-Item -ItemType Directory -Force -Path $destino | Out-Null
    foreach ($archivo in @('PdfEpub.exe', 'LICENSE', 'README.md', 'THIRD-PARTY-NOTICES.md', 'THIRD-PARTY-LICENSES.txt', 'Desinstalar.ps1')) {
        Copy-Item -LiteralPath (Join-Path $origen $archivo) -Destination $destino -Force
    }

    $powershell = Join-Path $env:WINDIR 'System32\WindowsPowerShell\v1.0\powershell.exe'
    $desinstalador = Join-Path $destino 'Desinstalar.ps1'
    $comandoDesinstalar = '"{0}" -NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File "{1}"' -f $powershell, $desinstalador
    $clave = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\PdfEpub'
    New-Item -Path $clave -Force | Out-Null
    Set-ItemProperty -Path $clave -Name DisplayName -Value $nombre
    Set-ItemProperty -Path $clave -Name DisplayVersion -Value $Version
    Set-ItemProperty -Path $clave -Name Publisher -Value 'PdfEpub contributors'
    Set-ItemProperty -Path $clave -Name InstallLocation -Value $destino
    Set-ItemProperty -Path $clave -Name DisplayIcon -Value (Join-Path $destino 'PdfEpub.exe')
    Set-ItemProperty -Path $clave -Name UninstallString -Value $comandoDesinstalar
    New-ItemProperty -Path $clave -Name NoModify -PropertyType DWord -Value 1 -Force | Out-Null
    New-ItemProperty -Path $clave -Name NoRepair -PropertyType DWord -Value 1 -Force | Out-Null

    $shell = New-Object -ComObject WScript.Shell
    $menu = Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\PdfEpub.lnk'
    $acceso = $shell.CreateShortcut($menu)
    $acceso.TargetPath = Join-Path $destino 'PdfEpub.exe'
    $acceso.WorkingDirectory = $destino
    $acceso.Save()
    $escritorio = Join-Path ([Environment]::GetFolderPath('Desktop')) 'PdfEpub.lnk'
    $acceso = $shell.CreateShortcut($escritorio)
    $acceso.TargetPath = Join-Path $destino 'PdfEpub.exe'
    $acceso.WorkingDirectory = $destino
    $acceso.Save()

    $respuesta = [Windows.Forms.MessageBox]::Show(
        'PdfEpub se instaló correctamente. ¿Quieres abrirlo ahora?',
        'Instalación completada', [Windows.Forms.MessageBoxButtons]::YesNo,
        [Windows.Forms.MessageBoxIcon]::Information)
    if ($respuesta -eq [Windows.Forms.DialogResult]::Yes) {
        Start-Process -FilePath (Join-Path $destino 'PdfEpub.exe')
    }
} catch {
    throw
}
