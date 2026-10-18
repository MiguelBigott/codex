package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// cacheGitHub guarda la última respuesta por repositorio y su ETag: las consultas sin
// cambios (304) no gastan el límite de la API de GitHub.
type cacheGitHub struct {
	mu    sync.Mutex
	datos map[string]entradaCache
}

type entradaCache struct {
	ETag    string   `json:"etag"`
	Release *Release `json:"release"`
}

var cacheRel = &cacheGitHub{datos: map[string]entradaCache{}}

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
	return cliente.Do(req)
}

// ultimaRelease consulta /releases/latest del repositorio y elige el instalador según el
// formato: setup .exe (NSIS) o .msi.
func ultimaRelease(repo string, msi bool) (*Release, error) {
	cacheRel.mu.Lock()
	previa := cacheRel.datos[repo]
	cacheRel.mu.Unlock()

	resp, err := pedirGitHub("https://api.github.com/repos/"+repo+"/releases/latest", previa.ETag)
	if err != nil {
		return nil, errors.New("sin conexión con GitHub")
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotModified && previa.Release != nil:
		return previa.Release, nil
	case resp.StatusCode == http.StatusNotFound:
		return nil, errors.New("todavía no tiene versiones publicadas")
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests:
		return nil, errors.New("GitHub limitó las consultas; se reintentará más tarde")
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("GitHub respondió %d", resp.StatusCode)
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
		return nil, errors.New("respuesta de GitHub ilegible")
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
	return "", errors.New("el instalador no aparece en el archivo de sumas")
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
