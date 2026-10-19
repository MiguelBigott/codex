package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Release es la última versión publicada de una app (o del propio launcher) en GitHub.
type Release struct {
	Version    string    `json:"version"`
	Tag        string    `json:"tag"`
	Fecha      time.Time `json:"fecha"`
	Notas      string    `json:"notas"` // página de la release
	Instalador *Archivo  `json:"instalador,omitempty"`
}

type Archivo struct {
	Nombre string `json:"nombre"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
	Tamano int64  `json:"tamano"`
}

var cliente = &http.Client{Timeout: 30 * time.Second}

// apiGitHub y webGitHub son las direcciones de GitHub (las pruebas las reemplazan por un
// servidor falso).
var (
	apiGitHub = "https://api.github.com"
	webGitHub = "https://github.com"
)

// clienteWeb no sigue redirecciones: la página …/releases/latest redirige a
// …/releases/tag/<etiqueta>, y esa dirección es justo lo que se necesita.
var clienteWeb = &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
	return http.ErrUseLastResponse
}}

var errSinVersiones = msj("sinVersiones")

// etiquetaPublica lee la etiqueta de la última versión desde la página pública de GitHub, que
// no cuenta para el límite de la API. Sin cuenta, la API gasta una consulta incluso cuando
// responde «sin cambios» (304), así que se usa solo cuando la etiqueta cambió.
func etiquetaPublica(repo string) (string, error) {
	req, err := http.NewRequest(http.MethodHead, webGitHub+"/"+repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Codex/"+Version)
	resp, err := clienteWeb.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther, http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		destino := resp.Header.Get("Location")
		i := strings.Index(destino, "/releases/tag/")
		if i < 0 {
			return "", errSinVersiones // redirige a …/releases: no hay ninguna versión
		}
		etiqueta, err := url.PathUnescape(destino[i+len("/releases/tag/"):])
		if err != nil || etiqueta == "" {
			return "", fmt.Errorf("etiqueta ilegible en %q", destino)
		}
		return etiqueta, nil
	case http.StatusNotFound:
		return "", msj("repoNoExiste")
	}
	return "", msj("githubRespondio", "codigo", strconv.Itoa(resp.StatusCode))
}

// cacheGitHub guarda la última respuesta por repositorio y su ETag: las consultas sin
// cambios (304) no gastan el límite de la API de GitHub (60 por hora sin cuenta, por
// conexión). Se guarda en disco para que al reabrir Codex tampoco se gaste, y recuerda
// hasta cuándo GitHub pausó las consultas para no insistir antes de tiempo.
type cacheGitHub struct {
	mu          sync.Mutex
	datos       map[string]entradaCache
	ruta        string
	limiteHasta time.Time
}

type entradaCache struct {
	ETag    string   `json:"etag"`
	Release *Release `json:"release"`
}

var cacheRel = &cacheGitHub{datos: map[string]entradaCache{}}

// cargarCache lee releases.json de la carpeta de datos y devuelve lo que se conocía.
func cargarCache(datos string) map[string]*Release {
	cacheRel.mu.Lock()
	defer cacheRel.mu.Unlock()
	cacheRel.ruta = filepath.Join(datos, "releases.json")
	if b, err := os.ReadFile(cacheRel.ruta); err == nil {
		_ = json.Unmarshal(b, &cacheRel.datos)
	}
	conocidas := map[string]*Release{}
	for repo, e := range cacheRel.datos {
		if e.Release != nil {
			conocidas[repo] = e.Release
		}
	}
	return conocidas
}

func guardarCache() {
	cacheRel.mu.Lock()
	defer cacheRel.mu.Unlock()
	if cacheRel.ruta == "" {
		return
	}
	if b, err := json.MarshalIndent(cacheRel.datos, "", "  "); err == nil {
		_ = os.WriteFile(cacheRel.ruta, b, 0o644)
	}
}

// limiteGitHub devuelve hasta cuándo GitHub pausó las consultas (cero si no hay pausa).
func limiteGitHub() time.Time {
	cacheRel.mu.Lock()
	defer cacheRel.mu.Unlock()
	if time.Now().After(cacheRel.limiteHasta) {
		return time.Time{}
	}
	return cacheRel.limiteHasta
}

// finDelLimite lee de la respuesta hasta cuándo dura la pausa; si no lo dice, una hora.
func finDelLimite(resp *http.Response) time.Time {
	if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
		return time.Now().Add(time.Duration(s) * time.Second)
	}
	if t, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil && t > 0 {
		return time.Unix(t, 0)
	}
	return time.Now().Add(time.Hour)
}

func esLimite(resp *http.Response) bool {
	return resp.StatusCode == http.StatusTooManyRequests ||
		(resp.StatusCode == http.StatusForbidden && (resp.Header.Get("X-RateLimit-Remaining") == "0" || resp.Header.Get("Retry-After") != ""))
}

func errorLimite(hasta time.Time) error {
	return msj("githubPausa", "hora", hasta.Local().Format("15:04"))
}

func pedirGitHub(url, etag string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "Codex/"+Version)
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}
	// Solo para pruebas: con un token, GitHub permite 5.000 consultas por hora en vez de 60.
	if t := os.Getenv("CODEX_GITHUB_TOKEN"); t != "" {
		req.Header.Set("Authorization", "Bearer "+t)
	}
	return cliente.Do(req)
}

// ultimaRelease consulta /releases/latest del repositorio y elige el instalador según el
// formato: setup .exe (NSIS) o .msi.
func ultimaRelease(repo string, msi bool) (*Release, error) {
	cacheRel.mu.Lock()
	previa := cacheRel.datos[repo]
	pausa := cacheRel.limiteHasta
	cacheRel.mu.Unlock()

	// Primero la página pública (gratis): si la etiqueta es la misma, no hace falta la API.
	if etiqueta, err := etiquetaPublica(repo); err == nil {
		if previa.Release != nil && previa.Release.Tag == etiqueta {
			return previa.Release, nil
		}
	} else if errors.Is(err, errSinVersiones) {
		return nil, err
	}

	// Durante la pausa no se consulta: se usa lo último conocido.
	if time.Now().Before(pausa) {
		if previa.Release != nil {
			return previa.Release, nil
		}
		return nil, errorLimite(pausa)
	}

	resp, err := pedirGitHub(apiGitHub+"/repos/"+repo+"/releases/latest", previa.ETag)
	if err != nil {
		return nil, msj("sinConexionGitHub")
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotModified && previa.Release != nil:
		return previa.Release, nil
	case resp.StatusCode == http.StatusNotFound:
		return nil, errSinVersiones
	case esLimite(resp):
		hasta := finDelLimite(resp)
		cacheRel.mu.Lock()
		cacheRel.limiteHasta = hasta
		cacheRel.mu.Unlock()
		if previa.Release != nil {
			return previa.Release, nil
		}
		return nil, errorLimite(hasta)
	case resp.StatusCode != http.StatusOK:
		return nil, msj("githubRespondio", "codigo", strconv.Itoa(resp.StatusCode))
	}

	var gh struct {
		Tag       string    `json:"tag_name"`
		Publicada time.Time `json:"published_at"`
		URL       string    `json:"html_url"`
		Assets    []struct {
			Nombre string `json:"name"`
			URL    string `json:"browser_download_url"`
			Tamano int64  `json:"size"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&gh); err != nil {
		return nil, msj("respuestaIlegible")
	}
	rel := &Release{Version: versionDeEtiqueta(gh.Tag), Tag: gh.Tag, Fecha: gh.Publicada, Notas: gh.URL}

	var sumas string
	for _, a := range gh.Assets {
		n := strings.ToLower(a.Nombre)
		if esSumas(n) {
			sumas = a.URL
			continue
		}
		elegible := esInstaladorWin64(n)
		if msi {
			elegible = esMSIWin64(n)
		}
		if rel.Instalador == nil && elegible {
			rel.Instalador = &Archivo{Nombre: a.Nombre, URL: a.URL, Tamano: a.Tamano,
				SHA256: strings.TrimPrefix(strings.ToLower(a.Digest), "sha256:")}
		}
	}
	if rel.Instalador != nil && !esHex64(rel.Instalador.SHA256) {
		rel.Instalador.SHA256 = ""
		if sumas != "" {
			if s, err := sumaDesdeArchivo(sumas, rel.Instalador.Nombre); err == nil {
				rel.Instalador.SHA256 = s
			}
		}
	}

	cacheRel.mu.Lock()
	cacheRel.datos[repo] = entradaCache{ETag: resp.Header.Get("ETag"), Release: rel}
	cacheRel.mu.Unlock()
	return rel, nil
}

func esSumas(n string) bool {
	return n == "sha256sums.txt" || n == "sha256sums" || n == "checksums.sha256" || n == "sha2-256sums"
}

// esInstaladorWin64 elige el instalador .exe para Windows x64 (el setup NSIS de cada app).
func esInstaladorWin64(n string) bool {
	if !strings.HasSuffix(n, ".exe") || !strings.Contains(n, "setup") {
		return false
	}
	return !strings.Contains(n, "arm64") && !strings.Contains(n, "x86") && !strings.Contains(n, "win32")
}

// esMSIWin64 elige el paquete .msi para Windows x64 (por ejemplo photocraft-0.10.0-windows-x64.msi).
func esMSIWin64(n string) bool {
	if !strings.HasSuffix(n, ".msi") || strings.Contains(n, "arm64") {
		return false
	}
	return strings.Contains(n, "x64") || strings.Contains(n, "amd64") || strings.Contains(n, "x86_64")
}

var reHex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

func esHex64(s string) bool { return reHex64.MatchString(s) }

// sumaDesdeArchivo lee un archivo de sumas con el formato de sha256sum: "<hash>  <nombre>".
func sumaDesdeArchivo(url, nombre string) (string, error) {
	resp, err := pedirGitHub(url, "")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sumas: %d", resp.StatusCode)
	}
	sc := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20))
	for sc.Scan() {
		campos := strings.Fields(sc.Text())
		if len(campos) >= 2 && strings.TrimPrefix(campos[len(campos)-1], "*") == nombre {
			if h := strings.ToLower(campos[0]); esHex64(h) {
				return h, nil
			}
		}
	}
	return "", msj("sinSuma")
}

// versionDeEtiqueta saca la versión de la etiqueta del release: «v1.2.0» → «1.2.0». Algunos
// repos anteponen el nombre de la app (ArtCraft: «artcraft-v0.41.0» → «0.41.0»).
func versionDeEtiqueta(tag string) string {
	if i := strings.IndexAny(tag, "0123456789"); i >= 0 {
		return tag[i:]
	}
	return tag
}

// compararVersiones devuelve -1, 0 o 1 comparando "1.2.10" con "1.3".
func compararVersiones(a, b string) int {
	pa, pb := partes(a), partes(b)
	for i := 0; i < len(pa) || i < len(pb); i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func partes(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	var out []int
	for _, p := range strings.Split(v, ".") {
		n, _ := strconv.Atoi(p)
		out = append(out, n)
	}
	return out
}
