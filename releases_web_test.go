package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// La página pública dice la etiqueta gratis: con la misma etiqueta no se toca la API;
// con una etiqueta nueva, una sola consulta a la API.
func TestEtiquetaPublicaEvitaLaAPI(t *testing.T) {
	var api atomic.Int32
	var etiqueta atomic.Value
	etiqueta.Store("v1.0.0")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/") {
			api.Add(1)
			e := etiqueta.Load().(string)
			_, _ = w.Write([]byte(`{"tag_name":"` + e + `","published_at":"2026-10-18T00:00:00Z","assets":[]}`))
			return
		}
		if r.URL.Path == "/yo/sinversion/releases/latest" {
			http.Redirect(w, r, "/yo/sinversion/releases", http.StatusFound)
			return
		}
		// Repositorio mudado a una organización: GitHub redirige a la dirección nueva.
		if r.URL.Path == "/yo/mudada/releases/latest" {
			http.Redirect(w, r, "/organizacion/mudada/releases/latest", http.StatusMovedPermanently)
			return
		}
		http.Redirect(w, r, "/yo/app/releases/tag/"+etiqueta.Load().(string), http.StatusFound)
	}))
	defer srv.Close()

	antes, antesWeb := apiGitHub, webGitHub
	apiGitHub, webGitHub = srv.URL, srv.URL
	cacheRel.mu.Lock()
	cacheRel.datos, cacheRel.limiteHasta, cacheRel.ruta = map[string]entradaCache{}, time.Time{}, ""
	cacheRel.mu.Unlock()
	defer func() {
		apiGitHub, webGitHub = antes, antesWeb
		cacheRel.mu.Lock()
		cacheRel.datos = map[string]entradaCache{}
		cacheRel.mu.Unlock()
	}()

	if e, err := etiquetaPublica("yo/app"); err != nil || e != "v1.0.0" {
		t.Fatalf("etiquetaPublica = %q, %v", e, err)
	}
	if e, err := etiquetaPublica("yo/mudada"); err != nil || e != "v1.0.0" {
		t.Fatalf("repositorio mudado: etiquetaPublica = %q, %v", e, err)
	}
	// Primera vez: no hay nada guardado, se usa la API una vez.
	if rel, err := ultimaRelease("yo/app", false); err != nil || rel.Version != "1.0.0" {
		t.Fatalf("primera consulta: %v %+v", err, rel)
	}
	if api.Load() != 1 {
		t.Fatalf("la primera consulta debía usar la API una vez, usó %d", api.Load())
	}
	// Misma etiqueta: diez revisiones y ninguna consulta a la API.
	for i := 0; i < 10; i++ {
		if _, err := ultimaRelease("yo/app", false); err != nil {
			t.Fatal(err)
		}
	}
	if api.Load() != 1 {
		t.Fatalf("con la misma etiqueta no debía usar la API; usó %d consultas", api.Load()-1)
	}
	// Sale una versión nueva: una sola consulta a la API para traerla.
	etiqueta.Store("v1.1.0")
	if rel, err := ultimaRelease("yo/app", false); err != nil || rel.Version != "1.1.0" {
		t.Fatalf("versión nueva: %v %+v", err, rel)
	}
	if api.Load() != 2 {
		t.Fatalf("la versión nueva debía usar la API una vez más, total %d", api.Load())
	}
	// Un repositorio sin versiones se reconoce sin la API.
	if _, err := ultimaRelease("yo/sinversion", false); err != errSinVersiones {
		t.Fatalf("sin versiones: %v", err)
	}
	if api.Load() != 2 {
		t.Fatal("un repositorio sin versiones no debía usar la API")
	}
}
