package main

// Acceso directo en el escritorio de las apps propias recién instaladas. El instalador también
// lo crea, pero corre como administrador: si la cuenta de Windows no es administradora, Windows
// pide la clave de otra cuenta y el acceso directo queda en el escritorio de esa otra cuenta.
// Codex corre como el usuario de verdad, así que lo crea en el suyo si falta.

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const crearSinVentana = 0x08000000 // CREATE_NO_WINDOW: sin consola visible

// comillasPS pone un texto entre comillas simples de PowerShell.
func comillasPS(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// crearAccesoEscritorio crea «<nombre>.lnk» en el escritorio del usuario, salvo que ya exista
// ahí o en el escritorio público.
func crearAccesoEscritorio(nombre, exe string) error {
	script := `$s = New-Object -ComObject WScript.Shell
$n = ` + comillasPS(nombre+".lnk") + `
$d = [Environment]::GetFolderPath('Desktop')
$p = [Environment]::GetFolderPath('CommonDesktopDirectory')
if ((Test-Path -LiteralPath (Join-Path $d $n)) -or (Test-Path -LiteralPath (Join-Path $p $n))) { exit 0 }
$l = $s.CreateShortcut((Join-Path $d $n))
$l.TargetPath = ` + comillasPS(exe) + `
$l.WorkingDirectory = ` + comillasPS(filepath.Dir(exe)) + `
$l.Save()`
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: crearSinVentana}
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("acceso directo de %s: %v %s", nombre, err, out)
		return err
	}
	return nil
}

// accesoDeAppInstalada busca el .exe de la app que se acaba de instalar y le crea su acceso
// directo. Solo para las apps con instalador NSIS (las de Codex): los MSI crean el suyo.
func accesoDeAppInstalada(app AppCatalogo) {
	if app.EsMSI() {
		return
	}
	inst := buscarInstalada(app.Registro())
	if inst == nil {
		return
	}
	if exe := ejecutableDe(app, inst); exe != "" {
		_ = crearAccesoEscritorio(app.NombreCompleto(), exe)
	}
}
