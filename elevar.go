package main

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ShellExecuteExW con el verbo "runas": Windows muestra el aviso de UAC y ejecuta el
// instalador como administrador (escribe en C:\Program Files). x/sys no lo incluye.

type shellExecuteInfo struct {
	cbSize       uint32
	fMask        uint32
	hwnd         uintptr
	lpVerb       *uint16
	lpFile       *uint16
	lpParameters *uint16
	lpDirectory  *uint16
	nShow        int32
	hInstApp     uintptr
	lpIDList     uintptr
	lpClass      *uint16
	hkeyClass    uintptr
	dwHotKey     uint32
	hIcon        uintptr
	hProcess     windows.Handle
}

const (
	seeMaskNoCloseProcess = 0x00000040
	seeMaskNoAsync        = 0x00000100
	errorCancelado        = 1223 // ERROR_CANCELLED: el usuario dijo que no en UAC
)

var (
	procShellExecuteExW = windows.NewLazySystemDLL("shell32.dll").NewProc("ShellExecuteExW")
	errCancelado        = errors.New("permiso de administrador cancelado")
)

// ejecutarElevado ejecuta exe como administrador. Si esperar es true, espera a que termine
// y devuelve su código de salida.
func ejecutarElevado(exe, args string, esperar bool) (uint32, error) {
	return ejecutarConVerbo("runas", exe, args, esperar)
}

// ejecutarNormal inicia un comando con los permisos actuales, para operaciones registradas
// en HKCU que pertenecen al usuario y no requieren acceso administrativo.
func ejecutarNormal(exe, args string, esperar bool) (uint32, error) {
	return ejecutarConVerbo("open", exe, args, esperar)
}

func ejecutarConVerbo(nombreVerbo, exe, args string, esperar bool) (uint32, error) {
	verbo, err := windows.UTF16PtrFromString(nombreVerbo)
	if err != nil {
		return 0, err
	}
	archivo, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return 0, err
	}
	parametros, _ := windows.UTF16PtrFromString(args)
	info := shellExecuteInfo{fMask: seeMaskNoCloseProcess | seeMaskNoAsync, hwnd: vent.hwnd,
		lpVerb: verbo, lpFile: archivo, lpParameters: parametros, nShow: 1}
	info.cbSize = uint32(unsafe.Sizeof(info))
	r, _, e := procShellExecuteExW.Call(uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		if errno, ok := e.(windows.Errno); ok && errno == errorCancelado {
			return 0, errCancelado
		}
		return 0, e
	}
	if info.hProcess == 0 {
		return 0, nil
	}
	defer windows.CloseHandle(info.hProcess)
	if !esperar {
		return 0, nil
	}
	if _, err := windows.WaitForSingleObject(info.hProcess, windows.INFINITE); err != nil {
		return 0, err
	}
	var codigo uint32
	if err := windows.GetExitCodeProcess(info.hProcess, &codigo); err != nil {
		return 0, err
	}
	return codigo, nil
}
