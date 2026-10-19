package main

// Icono en la bandeja del sistema. Con la opción activada, cerrar la ventana solo la oculta y
// Codex sigue abierto junto al reloj: clic en el icono para volver, clic derecho para el menú
// (abrir o salir). También aquí: abrir Codex al iniciar sesión, con la clave Run del usuario.

import (
	"log"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	shell32 = windows.NewLazySystemDLL("shell32.dll")

	procShellNotifyIconW       = shell32.NewProc("Shell_NotifyIconW")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procRegisterWindowMessageW = user32.NewProc("RegisterWindowMessageW")
	procIsIconic               = user32.NewProc("IsIconic")
	procGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
)

const (
	wmBandeja   = 0x8000 + 1 // WM_APP + 1: lo que pasa sobre el icono de la bandeja
	wmLButtonUp = 0x0202
	wmRButtonUp = 0x0205
	wmNull      = 0x0000

	nimAdd, nimDelete           = 0, 2
	nifMessage, nifIcon, nifTip = 0x1, 0x2, 0x4

	mfString, mfSeparator      = 0x0, 0x800
	tpmRightButton, tpmRetorno = 0x0002, 0x0100 // TPM_RIGHTBUTTON, TPM_RETURNCMD

	swHide = 0
	swShow = 5

	menuAbrir, menuSalir = 1, 2

	// Argumento con el que Windows abre Codex al iniciar sesión: arranca oculto en la bandeja.
	argBandeja = "--bandeja"
	claveRun   = `Software\Microsoft\Windows\CurrentVersion\Run`
)

// notifyIconData es NOTIFYICONDATAW (shellapi.h); los campos van en el mismo orden.
type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GUIDItem         windows.GUID
	HBalloonIcon     uintptr
}

var bandeja struct {
	mu       sync.Mutex
	hwnd     uintptr
	puesta   bool           // el icono está en la bandeja
	saliendo bool           // «Salir» o una actualización: cerrar de verdad
	ajustes  func() Ajustes // los ajustes actuales del launcher
	// taskbarCreated llega cuando el Explorador se reinicia: hay que volver a poner el icono.
	taskbarCreated uintptr
}

func prepararBandeja(hwnd uintptr, ajustes func() Ajustes) {
	nombre, _ := windows.UTF16PtrFromString("TaskbarCreated")
	tc, _, _ := procRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(nombre)))
	bandeja.mu.Lock()
	bandeja.hwnd, bandeja.ajustes, bandeja.taskbarCreated = hwnd, ajustes, tc
	bandeja.mu.Unlock()
	actualizarBandeja()
}

func datosIcono() notifyIconData {
	d := notifyIconData{HWnd: bandeja.hwnd, UID: 1}
	d.CbSize = uint32(unsafe.Sizeof(d))
	return d
}

// actualizarBandeja pone o quita el icono según el ajuste «seguir en la bandeja».
func actualizarBandeja() {
	bandeja.mu.Lock()
	defer bandeja.mu.Unlock()
	if bandeja.hwnd == 0 {
		return
	}
	quiere := bandeja.ajustes().Bandeja && !bandeja.saliendo
	switch {
	case quiere && !bandeja.puesta:
		d := datosIcono()
		d.UFlags = nifMessage | nifIcon | nifTip
		d.UCallbackMessage = wmBandeja
		d.HIcon = iconoBandeja()
		copy(d.SzTip[:], windows.StringToUTF16(tituloVentana))
		r, _, err := procShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&d)))
		bandeja.puesta = r != 0
		if r == 0 {
			log.Println("bandeja: no se pudo poner el icono:", err)
		}
	case !quiere && bandeja.puesta:
		d := datosIcono()
		procShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&d)))
		bandeja.puesta = false
	}
}

// iconoBandeja carga el icono del .exe al tamaño de la bandeja (16 px a 100 %, 20 a 125 %...).
func iconoBandeja() uintptr {
	inst, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW").Call(0)
	dpi, _, _ := procGetDpiForWindow.Call(bandeja.hwnd)
	lado, _, _ := procGetSystemMetricsDpi.Call(49, dpi) // SM_CXSMICON
	if lado == 0 {
		lado, _, _ = procGetSystemMetrics.Call(49)
	}
	ico, _, _ := procLoadImageW.Call(inst, idIcono, 1, lado, lado, 0) // IMAGE_ICON
	return ico
}

// cerrarAlaBandeja decide qué hace la X: true si la ventana se oculta en vez de cerrarse.
func cerrarAlaBandeja() bool {
	bandeja.mu.Lock()
	defer bandeja.mu.Unlock()
	return bandeja.puesta && !bandeja.saliendo
}

func mostrarVentana() {
	h := bandeja.hwnd
	if ic, _, _ := procIsIconic.Call(h); ic != 0 {
		procShowWindow.Call(h, swRestore)
	} else {
		procShowWindow.Call(h, swShow)
	}
	procSetForegroundWindow.Call(h)
}

func ocultarVentana() { procShowWindow.Call(bandeja.hwnd, swHide) }

// salirDeVerdad quita el icono y cierra Codex (menú «Salir» o una actualización).
func salirDeVerdad() {
	bandeja.mu.Lock()
	bandeja.saliendo = true
	bandeja.mu.Unlock()
	actualizarBandeja()
	procPostMessageW.Call(bandeja.hwnd, wmClose, 0, 0)
}

// quitarIconoAlSalir evita que quede un icono fantasma en la bandeja.
func quitarIconoAlSalir() {
	bandeja.mu.Lock()
	bandeja.saliendo = true
	bandeja.mu.Unlock()
	actualizarBandeja()
}

// mensajeBandeja atiende los clics sobre el icono; devuelve true si el mensaje era suyo.
func mensajeBandeja(hwnd, msg, lp uintptr) bool {
	switch {
	case msg == wmBandeja:
		switch lp & 0xFFFF {
		case wmLButtonUp:
			mostrarVentana()
		case wmRButtonUp:
			menuBandeja(hwnd)
		}
		return true
	case msg != 0 && msg == bandeja.taskbarCreated:
		bandeja.mu.Lock()
		bandeja.puesta = false
		bandeja.mu.Unlock()
		actualizarBandeja()
		return true
	}
	return false
}

func menuBandeja(hwnd uintptr) {
	t := textosDe(idiomaElegido(bandeja.ajustes()))
	menu, _, _ := procCreatePopupMenu.Call()
	if menu == 0 {
		return
	}
	defer procDestroyMenu.Call(menu)
	abrir, _ := windows.UTF16PtrFromString(t.Abrir)
	salir, _ := windows.UTF16PtrFromString(t.Salir)
	procAppendMenuW.Call(menu, mfString, menuAbrir, uintptr(unsafe.Pointer(abrir)))
	procAppendMenuW.Call(menu, mfSeparator, 0, 0)
	procAppendMenuW.Call(menu, mfString, menuSalir, uintptr(unsafe.Pointer(salir)))
	var p struct{ X, Y int32 }
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&p)))
	// Sin esto el menú no se cierra al hacer clic fuera (comportamiento documentado de Windows).
	procSetForegroundWindow.Call(hwnd)
	elegido, _, _ := procTrackPopupMenu.Call(menu, tpmRightButton|tpmRetorno, uintptr(p.X), uintptr(p.Y), 0, hwnd, 0)
	procPostMessageW.Call(hwnd, wmNull, 0, 0)
	switch elegido {
	case menuAbrir:
		mostrarVentana()
	case menuSalir:
		salirDeVerdad()
	}
}

// ---------- Abrir al iniciar sesión ----------

// aplicarInicioSesion agrega o quita Codex de los programas que abre Windows al iniciar sesión.
func aplicarInicioSesion(activo bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, claveRun, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !activo {
		if err := k.DeleteValue(tituloVentana); err != nil && err != registry.ErrNotExist {
			return err
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(tituloVentana, `"`+exe+`" `+argBandeja)
}
