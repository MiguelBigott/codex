package main

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Orchard publica sus versiones de Windows fuera de GitHub: Codex lee la versión de su
// manifiesto y las notas de su web, sin instalador (se instala desde su página oficial).
func TestReleaseDeManifiesto(t *testing.T) {
	payload := base64.StdEncoding.EncodeToString([]byte(`{"schema":1,"channel":"canary","version":"2.0.13-canary.1"}`))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/canary.json":
			_, _ = w.Write([]byte(`{"format":"orchard-signed-v1","payload":"` + payload + `"}`))
		case "/canary-notes.json":
			_, _ = w.Write([]byte(`[
			  {"version":"2.0.13-canary.1","date":"2026-11-25","title":"Orchard on Discord",
			   "sections":[{"title":"Fixes","items":["Lyrics stay in sync","Faster startup"]}]},
			  {"version":"2.0.12-canary.1","date":"2026-11-20","title":"Older","sections":[]}]`))
		case "/roto.json":
			_, _ = w.Write([]byte(`{"payload":"no es base64!"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	cacheRel.mu.Lock()
	cacheRel.datos, cacheRel.limiteHasta, cacheRel.ruta = map[string]entradaCache{}, time.Time{}, ""
	cacheRel.mu.Unlock()
	defer func() {
		cacheRel.mu.Lock()
		cacheRel.datos = map[string]entradaCache{}
		cacheRel.mu.Unlock()
	}()

	app := AppCatalogo{ID: "orchard", Nombre: "Orchard", Repo: "yo/orchard", Web: "https://ejemplo.dev/orchard",
		InstalarDesdeWeb: true, ManifiestoVersion: srv.URL + "/canary.json", NotasVersion: srv.URL + "/canary-notes.json"}
	rel, err := releaseDeManifiesto(app)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "2.0.13-canary.1" || rel.Instalador != nil || rel.Fecha.Format("2006-01-02") != "2026-11-25" {
		t.Fatalf("versión mal leída: %+v", rel)
	}
	if c := cambiosDe("yo/orchard"); !strings.Contains(c, "## Orchard on Discord") || !strings.Contains(c, "### Fixes") || !strings.Contains(c, "- Faster startup") || strings.Contains(c, "Older") {
		t.Errorf("notas: %q", c)
	}
	if otra, err := releaseDeManifiesto(app); err != nil || otra != rel {
		t.Errorf("no reutilizó la versión guardada: %v", err)
	}

	app.ManifiestoVersion = srv.URL + "/roto.json"
	if _, err := releaseDeManifiesto(app); err == nil {
		t.Error("un manifiesto ilegible debería dar error")
	}
}

// Una app que se instala desde su web nunca se descarga, aunque quede guardada una versión
// vieja con instalador (la 5.0.0-beta.9 de Orchard, de cuando lo publicaba en GitHub).
func TestInstalarDesdeWebNoDescarga(t *testing.T) {
	app := AppCatalogo{ID: "orchard", Nombre: "Orchard", Web: "https://ejemplo.dev/orchard", InstalarDesdeWeb: true}
	l := &Launcher{cat: Catalogo{Apps: []AppCatalogo{app}}, releases: map[string]*Release{
		"orchard": {Version: "5.0.0-beta.9", Instalador: &Archivo{Nombre: "Orchard-Setup-5.0.0.exe", URL: "https://x/vieja.exe",
			SHA256: "e2abc552b12ec85ce1bfbe5330bad919bb7bed571f88fd4e4874a10b3d842211"}},
	}, tareas: map[string]*Tarea{}, errores: map[string]*Mensaje{}}
	if err := l.Instalar("orchard"); err == nil {
		t.Fatal("instaló una app que se instala desde su web")
	}
	if len(l.tareas) != 0 {
		t.Errorf("empezó una tarea: %v", l.tareas)
	}
}
