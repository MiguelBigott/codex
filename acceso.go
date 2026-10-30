package main

// Accesos directos en el escritorio para todas las apps del catálogo. Codex los crea como el
// usuario actual después de instalar (también si el instalador se elevó) y repara los que falten
// al iniciar, tanto para instalaciones NSIS como MSI.

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const crearSinVentana = 0x08000000 // CREATE_NO_WINDOW: sin consola visible

// comillasPS pone un texto entre comillas simples de PowerShell.
func comillasPS(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// crearAccesoEscritorio asegura «<nombre>.lnk» en el escritorio visible para el usuario. Repara
// el enlace personal si apunta a otra ruta y respeta un enlace público si ya apunta al ejecutable.
func crearAccesoEscritorio(nombre, exe string) error {
	info, err := os.Stat(exe)
	if err != nil {
		return fmt.Errorf("no se puede usar el ejecutable %q: %w", exe, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("la ruta del ejecutable no es un archivo: %q", exe)
	}
	exe = filepath.Clean(exe)
	script := `$ErrorActionPreference = 'Stop'
$s = New-Object -ComObject WScript.Shell
$n = ` + comillasPS(nombre+".lnk") + `
$e = ` + comillasPS(exe) + `
$d = [Environment]::GetFolderPath('Desktop')
$p = [Environment]::GetFolderPath('CommonDesktopDirectory')
if ([string]::IsNullOrWhiteSpace($d)) { throw 'Windows no devolvió la carpeta Escritorio del usuario' }
[IO.Directory]::CreateDirectory($d) | Out-Null
$personal = Join-Path $d $n
if (Test-Path -LiteralPath $personal) {
  $l = $s.CreateShortcut($personal)
  if ($l.TargetPath -ieq $e) { exit 0 }
} elseif ($p -and (Test-Path -LiteralPath (Join-Path $p $n))) {
  $publico = $s.CreateShortcut((Join-Path $p $n))
  if (($publico.TargetPath -ieq $e) -and (Test-Path -LiteralPath $e)) { exit 0 }
  $l = $s.CreateShortcut($personal)
} else {
  $l = $s.CreateShortcut($personal)
}
$l.TargetPath = $e
$l.WorkingDirectory = ` + comillasPS(filepath.Dir(exe)) + `
$l.IconLocation = $e
$l.Save()`
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: crearSinVentana}
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("acceso directo de %s: %v %s", nombre, err, out)
		return err
	}
	return nil
}

// ejecutableInstalado encuentra la ruta registrada del ejecutable y si hay indicios de que
// la app está instalada aunque falte su entrada normal en Agregar o quitar programas.
func ejecutableInstalado(app AppCatalogo) (string, bool) {
	inst := buscarInstalada(app.Registro())
	if inst != nil {
		if exe := ejecutableDe(app, inst); exe != "" {
			return exe, true
		}
	}
	instalacionRegistrada := carpetaInstalacionRegistrada(app)
	if instalacionRegistrada != "" {
		exe := filepath.Join(instalacionRegistrada, app.Exe())
		if info, err := os.Stat(exe); err == nil && info.Mode().IsRegular() {
			return exe, true
		}
	}
	// Codex instala tanto NSIS como MSI bajo esta carpeta; permite crear el enlace incluso si
	// Windows aún no terminó de reflejar la entrada de desinstalación en el registro.
	candidato := filepath.Join(carpetaSuite(), app.NombreCompleto(), app.Exe())
	if info, err := os.Stat(candidato); err == nil && info.Mode().IsRegular() {
		return candidato, true
	}
	return "", inst != nil || instalacionRegistrada != ""
}

// asegurarAccesoDeAppInstalada asegura el enlace sin depender del tipo de instalador.
func asegurarAccesoDeAppInstalada(app AppCatalogo) error {
	exe, instalada := ejecutableInstalado(app)
	if exe == "" {
		if !instalada {
			return fmt.Errorf("no se encontró una instalación de %s", app.NombreCompleto())
		}
		return fmt.Errorf("no se encontró el ejecutable de %s", app.NombreCompleto())
	}
	return crearAccesoEscritorio(app.NombreCompleto(), exe)
}

// asegurarAccesosDeAppsInstaladas repara los enlaces que falten para las instalaciones
// anteriores y para las apps que se instalaron fuera de la sesión actual de Codex.
func asegurarAccesosDeAppsInstaladas(cat Catalogo) {
	for _, app := range cat.Apps {
		if app.Ejemplo {
			continue
		}
		if _, instalada := ejecutableInstalado(app); !instalada {
			continue
		}
		if err := asegurarAccesoDeAppInstalada(app); err != nil {
			log.Printf("acceso directo de %s: %v", app.NombreCompleto(), err)
		}
	}
}
