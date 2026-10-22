// Codex: instala, actualiza y abre las apps de Codex.
package main

import (
	"embed"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	webview2 "github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

const Version = "0.3.4"

// repoLauncher es el repositorio público de GitHub donde se publican las versiones del
// launcher ("usuario/repositorio"). Vacío: el launcher no busca actualizaciones propias.
const repoLauncher = "MiguelBigott/codex"

//go:embed web
var webFS embed.FS

func carpetaDatos() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base, _ = os.UserConfigDir()
	}
	return filepath.Join(base, "Codex")
}

func main() {
	datos := carpetaDatos()
	_ = os.MkdirAll(datos, 0o755)
	if f, err := os.OpenFile(filepath.Join(datos, "launcher.log"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644); err == nil {
		log.SetOutput(f)
	}

	// Al iniciar sesión, Windows abre Codex con --bandeja: arranca oculto, junto al reloj.
	enBandeja := slices.Contains(os.Args[1:], argBandeja)

	// Una sola instancia: si ya está abierto, lo traemos al frente (salvo al iniciar sesión).
	nombre, _ := windows.UTF16PtrFromString("Local\\Codex")
	if _, err := windows.CreateMutex(nil, false, nombre); errors.Is(err, windows.ERROR_ALREADY_EXISTS) {
		if !enBandeja {
			traerAlFrente()
		}
		return
	}

	cat, err := cargarCatalogo()
	if err != nil {
		log.Fatal(err)
	}
	l := NuevoLauncher(cat, datos)
	go l.Vigilar()

	srv, ln, err := iniciarServidor(l)
	if err != nil {
		log.Fatal("servidor: ", err)
	}
	go func() {
		if err := http.Serve(ln, srv.handler()); err != nil {
			log.Println("servidor:", err)
		}
	}()
	url := "http://" + srv.host + "/#" + srv.clave

	// Modo de desarrollo: sin ventana, para abrir la interfaz en un navegador.
	if os.Getenv("CODEX_NAVEGADOR") != "" {
		_ = os.WriteFile(filepath.Join(datos, "url-desarrollo.txt"), []byte(url), 0o600)
		log.Println("modo navegador:", srv.host)
		select {}
	}

	ancho, alto := tamanoInicial()
	w := webview2.NewWithOptions(webview2.WebViewOptions{
		AutoFocus: true,
		DataPath:  filepath.Join(datos, "webview"),
		WindowOptions: webview2.WindowOptions{
			Title: tituloVentana, Width: ancho, Height: alto, Center: true, IconId: idIcono,
		},
	})
	if w == nil {
		mensaje(textosDe(idiomaElegido(l.AjustesActuales())).FaltaWebView)
		return
	}
	defer w.Destroy()
	hwnd := uintptr(w.Window())
	prepararVentana(hwnd)
	w.SetSize(1040, 660, webview2.HintMin)
	prepararBandeja(hwnd, l.AjustesActuales)
	defer quitarIconoAlSalir()
	if enBandeja && cerrarAlaBandeja() {
		ocultarVentana()
	}
	if l.AjustesActuales().IniciarSesion {
		_ = aplicarInicioSesion(true) // por si Codex cambió de carpeta
	}
	l.salir = func() { w.Dispatch(salirDeVerdad) }
	l.alCambiarAjustes = func() { w.Dispatch(actualizarBandeja) }

	_ = w.Bind("caArrastrar", ventanaArrastrar)
	_ = w.Bind("caRedimensionar", ventanaRedimensionar)
	_ = w.Bind("caMinimizar", ventanaMinimizar)
	_ = w.Bind("caMaximizar", ventanaMaximizar)
	_ = w.Bind("caMaximizada", ventanaMaximizada)
	_ = w.Bind("caCerrar", ventanaCerrar)

	w.Navigate(url)
	w.Run()
}

func mensaje(texto string) {
	t, _ := windows.UTF16PtrFromString(texto)
	c, _ := windows.UTF16PtrFromString(tituloVentana)
	windows.MessageBox(0, t, c, windows.MB_OK|windows.MB_ICONWARNING)
}
