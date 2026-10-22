package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
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
	ID      int      `json:"id"`
	Mensaje *Mensaje `json:"mensaje"`
	Error   bool     `json:"error"`
}

type Ajustes struct {
	RevisarAuto    bool `json:"revisarAuto"`
	IntervaloHoras int  `json:"intervaloHoras"` // siempre 24: «cada 24 horas» o, sin RevisarAuto, manual
	// Bandeja: cerrar la ventana deja Codex abierto en la bandeja del sistema.
	Bandeja bool `json:"bandeja"`
	// IniciarSesion: Windows abre Codex (oculto en la bandeja) al iniciar sesión.
	IniciarSesion bool `json:"iniciarSesion"`
	// Idioma de la interfaz: «auto» (el del PC) o uno de idiomas.
	Idioma string `json:"idioma"`
}

type Launcher struct {
	mu        sync.Mutex
	cat       Catalogo
	datos     string
	ajustes   Ajustes
	releases  map[string]*Release
	errores   map[string]*Mensaje
	tareas    map[string]*Tarea
	revisado  time.Time
	revisando bool
	propia    *Release // última versión publicada del launcher
	aviso     Aviso
	salir     func() // cierra la ventana (para actualizar el launcher)
	// alCambiarAjustes avisa a la ventana (pone o quita el icono de la bandeja).
	alCambiarAjustes func()
}

func NuevoLauncher(cat Catalogo, datos string) *Launcher {
	l := &Launcher{cat: cat, datos: datos, releases: map[string]*Release{},
		errores: map[string]*Mensaje{}, tareas: map[string]*Tarea{},
		ajustes: Ajustes{RevisarAuto: true, IntervaloHoras: horasEntreRevisiones, Bandeja: true, Idioma: "auto"}}
	if b, err := os.ReadFile(filepath.Join(datos, "ajustes.json")); err == nil {
		_ = json.Unmarshal(b, &l.ajustes)
	}
	// Desde la 0.3.3 solo hay «cada 24 horas» o «manualmente» (las 1, 6 y 12 horas de antes
	// pasan a 24).
	l.ajustes.IntervaloHoras = horasEntreRevisiones
	// Lo último que se supo de cada app se muestra al instante, aun sin conexión.
	conocidas := cargarCache(datos)
	for _, app := range cat.Apps {
		if rel := conocidas[app.Repo]; rel != nil {
			l.releases[app.ID] = rel
		}
	}
	l.propia = conocidas[repoLauncher]
	return l
}

func (l *Launcher) avisar(m *Mensaje, esError bool) {
	l.mu.Lock()
	l.aviso = Aviso{ID: l.aviso.ID + 1, Mensaje: m, Error: esError}
	l.mu.Unlock()
	if esError {
		log.Println("aviso:", m.Clave, m.Datos)
	}
}

func (l *Launcher) AjustesActuales() Ajustes {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.ajustes
}

func (l *Launcher) GuardarAjustes(a Ajustes) error {
	a.IntervaloHoras = horasEntreRevisiones
	if !slices.Contains(idiomas, a.Idioma) {
		a.Idioma = "auto"
	}
	l.mu.Lock()
	antes := l.ajustes
	l.ajustes = a
	l.mu.Unlock()
	b, _ := json.MarshalIndent(a, "", "  ")
	_ = os.WriteFile(filepath.Join(l.datos, "ajustes.json"), b, 0o644)
	if l.alCambiarAjustes != nil {
		l.alCambiarAjustes()
	}
	if a.IniciarSesion != antes.IniciarSesion {
		if err := aplicarInicioSesion(a.IniciarSesion); err != nil {
			return msj("inicioSesionFallo", "detalle", err.Error())
		}
	}
	return nil
}

// ---------- Revisión de versiones ----------

// horasEntreRevisiones: con la revisión automática, Codex busca versiones nuevas cada 24 horas.
// Sin ella (manualmente), solo al abrirse y cuando se aprieta el botón de buscar actualizaciones.
const horasEntreRevisiones = 24

// Vigilar revisa al iniciar, también en modo manual, y luego cada IntervaloHoras si la
// revisión automática está activada.
func (l *Launcher) Vigilar() {
	l.Revisar()
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
			l.errores[app.ID] = comoMensaje(err)
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
	guardarCache()
}

// ---------- Estado para la interfaz ----------

type EstadoApp struct {
	AppCatalogo
	Instalada        bool     `json:"instalada"`
	VersionInstalada string   `json:"versionInstalada"`
	Ultima           *Release `json:"ultima"`
	Estado           string   `json:"estado"` // proximamente | sin-version | disponible | instalada | actualizacion
	Error            *Mensaje `json:"error"`
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
	Idioma      string      `json:"idioma"` // el que se muestra: «auto» ya resuelto
	Aviso       Aviso       `json:"aviso"`
	Carpeta     string      `json:"carpeta"`
	// LimiteHasta: GitHub pausó las consultas hasta esta hora (se muestran las últimas conocidas).
	LimiteHasta *time.Time `json:"limiteHasta,omitempty"`
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
		Ajustes: l.ajustes, Idioma: idiomaElegido(l.ajustes), Aviso: l.aviso, Carpeta: carpetaSuite()}
	if !l.revisado.IsZero() {
		t := l.revisado
		e.Revisado = &t
	}
	if t := limiteGitHub(); !t.IsZero() {
		e.LimiteHasta = &t
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
		return msj("appNoDisponible")
	}
	l.mu.Lock()
	rel := l.releases[id]
	l.mu.Unlock()
	if rel == nil || rel.Instalador == nil {
		return msj("sinInstalador")
	}
	if rel.Instalador.SHA256 == "" {
		return msj("sinSumaPublicada")
	}
	tipo := "instalar"
	if buscarInstalada(app.Registro()) != nil {
		tipo = "actualizar"
	}
	if !l.empezar(id, &Tarea{Tipo: tipo, Texto: "preparando", Progreso: 0}) {
		return msj("tareaEnCurso")
	}
	go func() {
		defer l.terminar(id)
		nombre := app.NombreCompleto()
		archivo, err := l.descargar(id, nombre, rel.Instalador)
		if err != nil {
			l.avisar(comoMensaje(err), true)
			return
		}
		defer os.Remove(archivo)
		l.avance(id, "instalando", -1)
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
			l.avisar(msj("instalarCancelado", "app", nombre), true)
		case err != nil:
			l.avisar(msj("instaladorNoEjecuto", "app", nombre, "detalle", err.Error()), true)
		case codigo == msiOcupado:
			l.avisar(msj("windowsOcupado", "app", nombre), true)
		case codigo != 0 && codigo != msiReiniciar:
			l.avisar(msj("instaladorCodigo", "app", nombre, "codigo", strconv.FormatUint(uint64(codigo), 10)), true)
		case tipo == "actualizar":
			l.avisar(msj("appActualizada", "app", nombre, "version", rel.Version), false)
		default:
			accesoDeAppInstalada(app)
			l.avisar(msj("appInstalada", "app", nombre), false)
		}
	}()
	return nil
}

func (l *Launcher) Desinstalar(id string) error {
	app, ok := l.app(id)
	if !ok {
		return msj("appDesconocida")
	}
	inst := buscarInstalada(app.Registro())
	if inst == nil || inst.Desinstalar == "" {
		return msj("appNoInstalada")
	}
	if !l.empezar(id, &Tarea{Tipo: "desinstalar", Texto: "desinstalando", Progreso: -1}) {
		return msj("tareaEnCurso")
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
		var codigo uint32
		var err error
		if inst.Maquina {
			codigo, err = ejecutarElevado(exe, args, true)
		} else {
			codigo, err = ejecutarNormal(exe, args, true)
		}
		if codigo == msiReiniciar {
			codigo = 0
		}
		switch {
		case errors.Is(err, errCancelado), codigo == msiCancelado:
			l.avisar(msj("permisoCancelado"), true)
		case err != nil || codigo != 0:
			l.avisar(msj("desinstalarFallo", "app", app.NombreCompleto()), true)
		default:
			// El desinstalador de NSIS se copia y sigue un momento después de salir.
			time.Sleep(1500 * time.Millisecond)
			l.avisar(msj("appDesinstalada", "app", app.NombreCompleto()), false)
		}
	}()
	return nil
}

func (l *Launcher) Abrir(id string) error {
	app, ok := l.app(id)
	if !ok {
		return msj("appDesconocida")
	}
	inst := buscarInstalada(app.Registro())
	if inst == nil {
		return msj("appNoInstalada")
	}
	exe := ejecutableDe(app, inst)
	if exe == "" {
		return msj("sinEjecutable", "app", app.NombreCompleto())
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
		return msj("launcherSinVersion")
	}
	if !l.empezar("_launcher", &Tarea{Tipo: "launcher", Texto: "descargando", Progreso: 0}) {
		return msj("launcherEnCurso")
	}
	go func() {
		defer l.terminar("_launcher")
		archivo, err := l.descargar("_launcher", tituloVentana, rel.Instalador)
		if err != nil {
			l.avisar(comoMensaje(err), true)
			return
		}
		// El instalador cierra este launcher, lo reemplaza y lo vuelve a abrir (/reiniciar).
		if _, err := ejecutarElevado(archivo, "/S /reiniciar", false); err != nil {
			if errors.Is(err, errCancelado) {
				l.avisar(msj("launcherCancelado"), true)
			} else {
				l.avisar(msj("instaladorNoEjecuto", "app", tituloVentana, "detalle", err.Error()), true)
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
// y lo rechaza si no coincide con la suma publicada. Los errores nombran a la app (nombre).
func (l *Launcher) descargar(id, nombre string, a *Archivo) (string, error) {
	fallo := func(err error) error { return msj("appErrorDetalle", "app", nombre, "detalle", err.Error()) }
	dir := filepath.Join(l.datos, "descargas")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fallo(err)
	}
	destino := filepath.Join(dir, filepath.Base(a.Nombre))
	parcial := destino + ".part"

	cli := &http.Client{Timeout: 60 * time.Minute}
	req, _ := http.NewRequest(http.MethodGet, a.URL, nil)
	req.Header.Set("User-Agent", "Codex/"+Version)
	resp, err := cli.Do(req)
	if err != nil {
		return "", msj("descargaSinConexion", "app", nombre)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", msj("descargaRespondio", "app", nombre, "codigo", strconv.Itoa(resp.StatusCode))
	}
	f, err := os.Create(parcial)
	if err != nil {
		return "", fallo(err)
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
				return "", fallo(werr)
			}
			h.Write(buf[:n])
			leido += int64(n)
			if total > 0 && time.Since(ultimo) > 150*time.Millisecond {
				l.avance(id, "descargando", float64(leido)/float64(total))
				ultimo = time.Now()
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			os.Remove(parcial)
			return "", msj("descargaCortada", "app", nombre)
		}
	}
	f.Close()
	l.avance(id, "verificando", 1)
	if hex.EncodeToString(h.Sum(nil)) != a.SHA256 {
		os.Remove(parcial)
		return "", msj("sumaNoCoincide", "app", nombre)
	}
	_ = os.Remove(destino)
	if err := os.Rename(parcial, destino); err != nil {
		return "", fallo(err)
	}
	return destino, nil
}
