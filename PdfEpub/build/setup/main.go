//go:build windows

package main

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

//go:embed payload.zip
var payload []byte

var version = "1.0.1"

const (
	mbYesNo        = 0x00000004
	mbIconInfo     = 0x00000040
	mbIconQuestion = 0x00000020
	mbIconError    = 0x00000010
	idYes          = 6
	createNoWindow = 0x08000000
)

func message(title, body string, flags uintptr) int {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	bodyPtr, _ := syscall.UTF16PtrFromString(body)
	result, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		0, uintptr(unsafe.Pointer(bodyPtr)), uintptr(unsafe.Pointer(titlePtr)), flags)
	return int(result)
}

func main() {
	if message("Instalar PdfEpub", "PdfEpub se instalará para tu usuario, con accesos directos y opción para quitarlo desde Windows. ¿Continuar?", mbYesNo|mbIconQuestion) != idYes {
		return
	}

	tempDir, err := os.MkdirTemp("", "PdfEpub-setup-")
	if err != nil {
		message("Instalar PdfEpub", "No se pudieron preparar los archivos de instalación: "+err.Error(), mbIconError)
		return
	}
	defer os.RemoveAll(tempDir)

	if err := extraer(tempDir); err != nil {
		message("Instalar PdfEpub", "No se pudieron extraer los archivos: "+err.Error(), mbIconError)
		return
	}

	powershell := filepath.Join(os.Getenv("WINDIR"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if _, err := os.Stat(powershell); err != nil {
		powershell = "powershell.exe"
	}
	cmd := exec.Command(powershell, "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(tempDir, "instalar.ps1"), "-Version", version)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	if out, err := cmd.CombinedOutput(); err != nil {
		detalle := strings.TrimSpace(string(out))
		if detalle == "" {
			detalle = err.Error()
		}
		message("Instalar PdfEpub", "La instalación no pudo completarse.\n\n"+detalle, mbIconError)
	}
}

func extraer(destino string) error {
	zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
	if err != nil {
		return err
	}
	for _, archivo := range zr.File {
		nombre := filepath.Clean(filepath.FromSlash(archivo.Name))
		if filepath.IsAbs(nombre) || nombre == "." || nombre == ".." || strings.HasPrefix(nombre, ".."+string(filepath.Separator)) {
			return syscall.EINVAL
		}
		r, err := archivo.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(filepath.Join(destino, nombre), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			r.Close()
			return err
		}
		_, copyErr := io.Copy(out, r)
		closeOutErr := out.Close()
		closeInErr := r.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}
		if closeInErr != nil {
			return closeInErr
		}
	}
	return nil
}
