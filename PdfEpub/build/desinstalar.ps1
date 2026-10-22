$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms

$destino = Split-Path -Parent $MyInvocation.MyCommand.Path
$respuesta = [Windows.Forms.MessageBox]::Show(
    '¿Quieres desinstalar PdfEpub de este equipo?',
    'Desinstalar PdfEpub', [Windows.Forms.MessageBoxButtons]::YesNo,
    [Windows.Forms.MessageBoxIcon]::Question)
if ($respuesta -ne [Windows.Forms.DialogResult]::Yes) { exit 0 }

$abierto = Get-Process -Name PdfEpub -ErrorAction SilentlyContinue
if ($abierto) {
    foreach ($proceso in $abierto) { [void]$proceso.CloseMainWindow() }
    $abierto | Wait-Process -Timeout 5 -ErrorAction SilentlyContinue
    if (Get-Process -Name PdfEpub -ErrorAction SilentlyContinue) {
        [void][Windows.Forms.MessageBox]::Show(
            'Cierra PdfEpub y vuelve a iniciar la desinstalación.',
            'PdfEpub sigue abierto', [Windows.Forms.MessageBoxButtons]::OK,
            [Windows.Forms.MessageBoxIcon]::Information)
        exit 1
    }
}

Remove-Item -LiteralPath (Join-Path $env:APPDATA 'Microsoft\Windows\Start Menu\Programs\PdfEpub.lnk') -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath (Join-Path ([Environment]::GetFolderPath('Desktop')) 'PdfEpub.lnk') -Force -ErrorAction SilentlyContinue
Remove-Item -LiteralPath 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\PdfEpub' -Recurse -Force -ErrorAction SilentlyContinue

$comando = "Start-Sleep -Seconds 2; Remove-Item -LiteralPath '$($destino.Replace("'", "''"))' -Recurse -Force -ErrorAction SilentlyContinue"
$bytes = [Text.Encoding]::Unicode.GetBytes($comando)
$codificado = [Convert]::ToBase64String($bytes)
$powershell = Join-Path $env:WINDIR 'System32\WindowsPowerShell\v1.0\powershell.exe'
Start-Process -FilePath $powershell -WindowStyle Hidden -ArgumentList @('-NoProfile', '-WindowStyle', 'Hidden', '-ExecutionPolicy', 'Bypass', '-EncodedCommand', $codificado)
