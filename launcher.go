package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Tarea en curso sobre una app (o el propio launcher).
type Tarea struct {
	Tipo     string  `json:"tipo"` // instalar | actualizar | desinstalar | launcher
	Texto    string  `json:"texto"`
	Progreso float64 `json:"progreso"` // 0..1; -1 = sin porcentaje
}

type Aviso struct {
	ID    int    `json:"id"`
	Texto string `json:"texto"`
	Error bool   `json:"error"`
}

type Ajustes struct {
	RevisarAuto    bool `json:"revisarAuto"`
	IntervaloHoras int  `json:"intervaloHoras"`
}

type Launcher struct {
	mu        sync.Mutex
	cat       Catalogo
	datos     string
	ajustes   Ajustes
	releases  map[string]*Release
	errores   map[string]string
	tareas    map[string]*Tarea
	revisado  time.Time
	revisando bool
	propia    *Release // última versión publicada del launcher
	aviso     Aviso
	salir     func() // cierra la ventana (para actualizar el launcher)
}

func NuevoLauncher(cat Catalogo, datos string) *Launcher {
	l := &Launcher{cat: cat, datos: datos, releases: map[string]*Release{},
		errores: map[string]string{}, tareas: map[string]*Tarea{},
		ajustes: Ajustes{RevisarAuto: true, IntervaloHoras: 6}}
	if b, err := os.ReadFile(filepath.Join(datos, "ajustes.json")); err == nil {
		_ = json.Unmarshal(b, &l.ajustes)
	}
	if l.ajustes.IntervaloHoras <= 0 {
		l.ajustes.IntervaloHoras = 6
	}
	return l
}

func (l *Launcher) avisar(texto string, esError bool) {
	l.mu.Lock()
	l.aviso = Aviso{ID: l.aviso.ID + 1, Texto: texto, Error: esError}
	l.mu.Unlock()
	if esError {
		log.Println("aviso:", texto)
	}
}

func (l *Launcher) GuardarAjustes(a Ajustes) {
	if a.IntervaloHoras <= 0 {
		a.IntervaloHoras = 6
	}
	l.mu.Lock()
	l.ajustes = a
	l.mu.Unlock()
	b, _ := json.MarshalIndent(a, "", "  ")
	_ = os.WriteFile(filepath.Join(l.datos, "ajustes.json"), b, 0o644)
}

// ---------- Revisión de versiones ----------

// Vigilar revisa al iniciar y luego cada IntervaloHoras, si está activado.
func (l *Launcher) Vigilar() {
	for {
		l.mu.Lock()
		a, ultima := l.ajustes, l.revisado
		l.mu.Unlock()
		if a.RevisarAuto && time.Since(ultima) >= time.Duration(a.IntervaloHoras)*time.Hour {
			l.Revisar()
		}
		time.Sleep(time.Minute)
	}
}

func (l *Launcher) Revisar() {
	l.mu.Lock()
	if l.revisando {
		l.mu.Unlock()
		return
	}
	l.revisando = true
	l.mu.Unlock()
	defer func() {
		l.mu.Lock()
		l.revisando, l.revisado = false, time.Now()
		l.mu.Unlock()
	}()

	for _, app := range l.cat.Apps {
		if app.Ejemplo || app.Repo == "" {
			continue
		}
		rel, err := ultimaRelease(app.Repo, app.EsMSI())
		l.mu.Lock()
		if err != nil {
			l.errores[app.ID] = err.Error()
		} else {
			delete(l.errores, app.ID)
			l.releases[app.ID] = rel
		}
		l.mu.Unlock()
	}
	if repoLauncher != "" {
		if rel, err := ultimaRelease(repoLauncher, false); err == nil {
			l.mu.Lock()
			l.propia = rel
			l.mu.Unlock()
		}
	}
}

// ---------- Estado para la interfaz ----------

type EstadoApp struct {
	AppCatalogo
	Instalada        bool     `json:"instalada"`
	VersionInstalada string   `json:"versionInstalada"`
	Ultima           *Release `json:"ultima"`
	Estado           string   `json:"estado"` // proximamente | sin-version | disponible | instalada | actualizacion
	Error            string   `json:"error"`
	Tarea            *Tarea   `json:"tarea"`
	URLRepoGitHub    string   `json:"urlRepo"`
}

type EstadoLauncher struct {
	Version     string      `json:"version"`
	Plataforma  string      `json:"plataforma"`
	Revisado    *time.Time  `json:"revisado"`
	Revisando   bool        `json:"revisando"`
	Categorias  []Categoria `json:"categorias"`
	Apps        []EstadoApp `json:"apps"`
	Propia      *Release    `json:"propia"`
	HayNueva    bool        `json:"hayNueva"`
	TareaPropia *Tarea      `json:"tareaPropia"`
	Ajustes     Ajustes     `json:"ajustes"`
	Aviso       Aviso       `json:"aviso"`
	Carpeta     string      `json:"carpeta"`
}

func (l *Launcher) Estado() EstadoLauncher {
	// El registro se lee fuera del candado: puede tardar unos milisegundos.
	inst := map[string]*Instalada{}
	for _, app := range l.cat.Apps {
		if !app.Ejemplo {
			inst[app.ID] = buscarInstalada(app.Registro())
		}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	e := EstadoLauncher{Version: Version, Plataforma: "Windows " + map[string]string{"amd64": "x64", "arm64": "Arm64", "386": "x86"}[runtime.GOARCH],
		Revisando: l.revisando, Categorias: l.cat.Categorias, Propia: l.propia, TareaPropia: l.tareas["_launcher"],
		Ajustes: l.ajustes, Aviso: l.aviso, Carpeta: carpetaSuite()}
	if !l.revisado.IsZero() {
		t := l.revisado
		e.Revisado = &t
	}
	e.HayNueva = l.propia != nil && compararVersiones(l.propia.Version, Version) > 0
	for _, app := range l.cat.Apps {
		ea := EstadoApp{AppCatalogo: app, Ultima: l.releases[app.ID], Error: l.errores[app.ID], Tarea: l.tareas[app.ID], URLRepoGitHub: app.URLRepo()}
		if i := inst[app.ID]; i != nil {
			ea.Instalada, ea.VersionInstalada = true, i.Version
		}
		switch {
		case app.Ejemplo:
			ea.Estado = "proximamente"
		case ea.Instalada && ea.Ultima != nil && compararVersiones(ea.Ultima.Version, ea.VersionInstalada) > 0:
			ea.Estado = "actualizacion"
		case ea.Instalada:
			ea.Estado = "instalada"
		case ea.Ultima != nil && ea.Ultima.Instalador != nil:
			ea.Estado = "disponible"
		default:
			ea.Estado = "sin-version"
		}
		e.Apps = append(e.Apps, ea)
	}
	return e
}

// ---------- Acciones ----------

func (l *Launcher) app(id string) (AppCatalogo, bool) {
	for _, a := range l.cat.Apps {
		if a.ID == id {
			return a, true
		}
	}
	return AppCatalogo{}, false
}

func (l *Launcher) empezar(id string, t *Tarea) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.tareas[id] != nil {
		return false
	}
	l.tareas[id] = t
	return true
}

func (l *Launcher) avance(id, texto string, p float64) {
	l.mu.Lock()
	if t := l.tareas[id]; t != nil {
		t.Texto, t.Progreso = texto, p
	}
	l.mu.Unlock()
}

func (l *Launcher) terminar(id string) {
	l.mu.Lock()
	delete(l.tareas, id)
	l.mu.Unlock()
}

// Códigos de salida de Windows Installer (msiexec).
const (
	msiCancelado = 1602 // el usuario canceló
	msiOcupado   = 1618 // hay otra instalación en curso
	msiReiniciar = 3010 // terminó bien; pide reiniciar
)

var reGUID = regexp.MustCompile(`\{[0-9A-Fa-f-]{36}\}`)

// codigoProductoMSI devuelve el {GUID} si el desinstalador registrado es de Windows Installer.
func codigoProductoMSI(cmd string) string {
	if !strings.Contains(strings.ToLower(cmd), "msiexec") {
		return ""
	}
	return reGUID.FindString(cmd)
}

// carpetaSuite es C:\Program Files\Codex: el launcher y cada app en su subcarpeta.
func carpetaSuite() string {
	pf := os.Getenv("ProgramW6432")
	if pf == "" {
		pf = os.Getenv("ProgramFiles")
	}
	return filepath.Join(pf, "Codex")
}

// Instalar descarga la última versión, la verifica y ejecuta su instalador. Sirve también
// para actualizar: el instalador NSIS reemplaza la versión anterior.
func (l *Launcher) Instalar(id string) error {
	app, ok := l.app(id)
	if !ok || app.Ejemplo {
		return errors.New("esa app no está disponible")
	}
	l.mu.Lock()
	rel := l.releases[id]
	l.mu.Unlock()
	if rel == nil || rel.Instalador == nil {
		return errors.New("no hay un instalador publicado para Windows")
	}
	if rel.Instalador.SHA256 == "" {
		return errors.New("la versión publicada no trae su suma SHA-256; por seguridad no se instala")
	}
	tipo := "instalar"
	if buscarInstalada(app.Registro()) != nil {
		tipo = "actualizar"
	}
	if !l.empezar(id, &Tarea{Tipo: tipo, Texto: "Preparando", Progreso: 0}) {
		return errors.New("ya hay una tarea en curso para esta app")
	}
	go func() {
		defer l.terminar(id)
		nombre := app.NombreCompleto()
		archivo, err := l.descargar(id, rel.Instalador)
		if err != nil {
			l.avisar(fmt.Sprintf("%s: %v", nombre, err), true)
			return
		}
		defer os.Remove(archivo)
		l.avance(id, "Instalando", -1)
		carpeta := filepath.Join(carpetaSuite(), nombre)
		exe, args := archivo, "/S /D="+carpeta // NSIS: /D= siempre al final y sin comillas
		if app.EsMSI() {
			exe, args = "msiexec.exe", `/i "`+archivo+`" /qn /norestart`
			if app.PropiedadCarpeta != "" {
				args += ` ` + app.PropiedadCarpeta + `="` + carpeta + `"`
			}
		}
		codigo, err := ejecutarElevado(exe, args, true)
		switch {
		case errors.Is(err, errCancelado), codigo == msiCancelado:
			l.avisar("Se canceló el permiso de administrador; "+nombre+" no se instaló.", true)
		case err != nil:
			l.avisar(fmt.Sprintf("%s: no se pudo ejecutar el instalador (%v)", nombre, err), true)
		case codigo == msiOcupado:
			l.avisar(nombre+": Windows está terminando otra instalación; vuelve a intentarlo en un momento.", true)
		case codigo != 0 && codigo != msiReiniciar:
			l.avisar(fmt.Sprintf("%s: el instalador terminó con error (código %d)", nombre, codigo), true)
		case tipo == "actualizar":
			l.avisar(nombre+" se actualizó a la versión "+rel.Version+".", false)
		default:
			l.avisar(nombre+" se instaló.", false)
		}
	}()
	return nil
}

func (l *Launcher) Desinstalar(id string) error {
	app, ok := l.app(id)
	if !ok {
		return errors.New("app desconocida")
	}
	inst := buscarInstalada(app.Registro())
	if inst == nil || inst.Desinstalar == "" {
		return errors.New("la app no está instalada")
	}
	if !l.empezar(id, &Tarea{Tipo: "desinstalar", Texto: "Desinstalando", Progreso: -1}) {
		return errors.New("ya hay una tarea en curso para esta app")
	}
	go func() {
		defer l.terminar(id)
		exe, args := separarComando(inst.Desinstalar)
		if guid := codigoProductoMSI(inst.Desinstalar); guid != "" {
			// MSI: «MsiExec.exe /I{...}» abre el asistente; /x desinstala en silencio.
			exe, args = "msiexec.exe", "/x "+guid+" /qn /norestart"
		} else if !inst.Silencioso {
			args += " /S"
		}
		codigo, err := ejecutarElevado(exe, args, true)
		if codigo == msiReiniciar {
			codigo = 0
		}
		switch {
		case errors.Is(err, errCancelado), codigo == msiCancelado:
			l.avisar("Se canceló el permiso de administrador.", true)
		case err != nil || codigo != 0:
			l.avisar(app.NombreCompleto()+": no se pudo desinstalar.", true)
		default:
			// El desinstalador de NSIS se copia y sigue un momento después de salir.
			time.Sleep(1500 * time.Millisecond)
			l.avisar(app.NombreCompleto()+" se desinstaló.", false)
		}
	}()
	return nil
}

func (l *Launcher) Abrir(id string) error {
	app, ok := l.app(id)
	if !ok {
		return errors.New("app desconocida")
	}
	inst := buscarInstalada(app.Registro())
	if inst == nil {
		return errors.New("la app no está instalada")
	}
	exe := ejecutableDe(app, inst)
	if exe == "" {
		return errors.New("no se encontró el ejecutable de " + app.NombreCompleto())
	}
	cmd := exec.Command(exe)
	cmd.Dir = filepath.Dir(exe)
	return cmd.Start()
}

// ActualizarLauncher descarga la nueva versión del launcher, la instala y lo reinicia.
func (l *Launcher) ActualizarLauncher() error {
	l.mu.Lock()
	rel := l.propia
	l.mu.Unlock()
	if rel == nil || rel.Instalador == nil || rel.Instalador.SHA256 == "" {
		return errors.New("no hay una versión nueva del launcher lista para instalar")
	}
	if !l.empezar("_launcher", &Tarea{Tipo: "launcher", Texto: "Descargando", Progreso: 0}) {
		return errors.New("la actualización ya está en curso")
	}
	go func() {
		defer l.terminar("_launcher")
		archivo, err := l.descargar("_launcher", rel.Instalador)
		if err != nil {
			l.avisar("Launcher: "+err.Error(), true)
			return
		}
		// El instalador cierra este launcher, lo reemplaza y lo vuelve a abrir (/reiniciar).
		if _, err := ejecutarElevado(archivo, "/S /reiniciar", false); err != nil {
			if errors.Is(err, errCancelado) {
				l.avisar("Se canceló el permiso de administrador; el launcher no se actualizó.", true)
			} else {
				l.avisar("Launcher: no se pudo ejecutar el instalador.", true)
			}
			return
		}
		if l.salir != nil {
			l.salir()
		}
	}()
	return nil
}

// descargar baja el archivo a la carpeta de datos, calcula su SHA-256 mientras lo escribe
// y lo rechaza si no coincide con la suma publicada.
func (l *Launcher) descargar(id string, a *Archivo) (string, error) {
	dir := filepath.Join(l.datos, "descargas")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	destino := filepath.Join(dir, filepath.Base(a.Nombre))
	parcial := destino + ".part"

	cli := &http.Client{Timeout: 60 * time.Minute}
	req, _ := http.NewRequest(http.MethodGet, a.URL, nil)
	req.Header.Set("User-Agent", "Codex/"+Version)
	resp, err := cli.Do(req)
	if err != nil {
		return "", errors.New("sin conexión para descargar")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("la descarga respondió %d", resp.StatusCode)
	}
	f, err := os.Create(parcial)
	if err != nil {
		return "", err
	}
	total := resp.ContentLength
	if total <= 0 {
		total = a.Tamano
	}
	h := sha256.New()
	buf := make([]byte, 256*1024)
	var leido int64
	ultimo := time.Now()
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				os.Remove(parcial)
				return "", werr
			}
			h.Write(buf[:n])
			leido += int64(n)
			if total > 0 && time.Since(ultimo) > 150*time.Millisecond {
				l.avance(id, "Descargando", float64(leido)/float64(total))
				ultimo = time.Now()
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			os.Remove(parcial)
			return "", errors.New("la descarga se cortó")
		}
	}
	f.Close()
	l.avance(id, "Verificando", 1)
	if hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
		os.Remove(parcial)
		return "", errors.New("el archivo descargado no coincide con su suma SHA-256; se descartó")
	}
	_ = os.Remove(destino)
	if err := os.Rename(parcial, destino); err != nil {
		return "", err
	}
	return destino, nil
}
