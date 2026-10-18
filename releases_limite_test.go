package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Simula a GitHub: responde una release, luego pausa las consultas (límite agotado).
func TestLimiteDeGitHubUsaLoConocido(t *testing.T) {
	var pedidos, pausado atomic.Int32
	reinicio := time.Now().Add(30 * time.Minute).Truncate(time.Second)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Esta prueba simula que la página pública no responde, para probar la pausa de la API.
		if !strings.HasPrefix(r.URL.Path, "/repos/") {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		pedidos.Add(1)
		if pausado.Load() == 1 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reinicio.Unix(), 10))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("ETag", `"v1"`)
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.1","published_at":"2026-10-18T00:00:00Z","assets":[
			{"name":"App-Setup-1.0.1.exe","browser_download_url":"https://example.com/a.exe","size":10,
			 "digest":"sha256:` + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + `"}]}`))
	}))
	defer srv.Close()

	antes, antesWeb := apiGitHub, webGitHub
	apiGitHub, webGitHub = srv.URL, srv.URL
	cacheRel.mu.Lock()
	cacheRel.datos, cacheRel.limiteHasta = map[string]entradaCache{}, time.Time{}
	cacheRel.mu.Unlock()
	defer func() {
		apiGitHub, webGitHub = antes, antesWeb
		cacheRel.mu.Lock()
		cacheRel.datos, cacheRel.limiteHasta, cacheRel.ruta = map[string]entradaCache{}, time.Time{}, ""
		cacheRel.mu.Unlock()
	}()
	carpeta := t.TempDir()
	cargarCache(carpeta)

	rel, err := ultimaRelease("yo/app", false)
	if err != nil || rel.Version != "1.0.1" {
		t.Fatalf("primera consulta: %v %+v", err, rel)
	}
	guardarCache()

	// GitHub pausa: se usa lo conocido, sin error, y se anota hasta cuándo.
	pausado.Store(1)
	rel, err = ultimaRelease("yo/app", false)
	if err != nil || rel == nil || rel.Version != "1.0.1" {
		t.Fatalf("durante el límite debía usar lo conocido: %v %+v", err, rel)
	}
	if got := limiteGitHub(); !got.Equal(reinicio) {
		t.Fatalf("limiteGitHub = %v, quiere %v", got, reinicio)
	}

	// Durante la pausa no se vuelve a consultar.
	n := pedidos.Load()
	if _, err := ultimaRelease("yo/app", false); err != nil {
		t.Fatal(err)
	}
	if pedidos.Load() != n {
		t.Fatal("no debía consultar a GitHub durante la pausa")
	}
	// Una app nunca consultada recibe un error claro con la hora.
	if _, err := ultimaRelease("yo/otra", false); err == nil {
		t.Fatal("una app sin datos previos debía dar error durante la pausa")
	}

	// Al reabrir Codex, lo conocido sale del disco.
	cacheRel.mu.Lock()
	cacheRel.datos = map[string]entradaCache{}
	cacheRel.mu.Unlock()
	if conocidas := cargarCache(carpeta); conocidas["yo/app"] == nil || conocidas["yo/app"].Version != "1.0.1" {
		t.Fatalf("la caché en disco no trajo la release: %+v", conocidas)
	}
}
