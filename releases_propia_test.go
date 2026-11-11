package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Codex encuentra su versión nueva sin tocar la API: etiqueta de la página pública e
// instalador y suma del SHA256SUMS.txt de la release.
func TestReleasePropiaSinAPI(t *testing.T) {
	sha := "b826dfe8aca362a4f0a6480bbbf522674ea8dd7049bba16c872e9643d4ed9040"
	usoAPI := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/yo/codex/releases/latest":
			http.Redirect(w, r, "/yo/codex/releases/tag/v9.9.9", http.StatusFound)
		case r.URL.Path == "/yo/codex/releases/download/v9.9.9/SHA256SUMS.txt":
			_, _ = w.Write([]byte(sha + " *Codex-Setup-9.9.9.exe\n"))
		case len(r.URL.Path) > 6 && r.URL.Path[:7] == "/repos/":
			usoAPI = true
			http.Error(w, "límite", http.StatusForbidden)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	antesWeb, antesAPI := webGitHub, apiGitHub
	webGitHub, apiGitHub = srv.URL, srv.URL
	cacheRel.mu.Lock()
	cacheRel.datos, cacheRel.limiteHasta, cacheRel.ruta = map[string]entradaCache{}, time.Time{}, ""
	cacheRel.mu.Unlock()
	defer func() {
		webGitHub, apiGitHub = antesWeb, antesAPI
		cacheRel.mu.Lock()
		cacheRel.datos = map[string]entradaCache{}
		cacheRel.mu.Unlock()
	}()

	rel, err := releasePropia("yo/codex")
	if err != nil {
		t.Fatal(err)
	}
	if usoAPI {
		t.Error("usó la API de GitHub")
	}
	if rel.Version != "9.9.9" || rel.Instalador == nil || rel.Instalador.SHA256 != sha ||
		rel.Instalador.URL != srv.URL+"/yo/codex/releases/download/v9.9.9/Codex-Setup-9.9.9.exe" {
		t.Fatalf("versión mal armada: %+v %+v", rel, rel.Instalador)
	}
	// La segunda revisión, con la misma etiqueta, reutiliza lo guardado.
	if otra, err := releasePropia("yo/codex"); err != nil || otra != rel {
		t.Errorf("no reutilizó la versión guardada: %v", err)
	}
}
