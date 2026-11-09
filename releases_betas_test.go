package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Con betas, se usa la versión más reciente que trae instalador para Windows x64: se salta
// la estable sin archivos, la que solo es de Android y el instalador para Arm64.
func TestReleaseConBetas(t *testing.T) {
	sha := "e2abc552b12ec85ce1bfbe5330bad919bb7bed571f88fd4e4874a10b3d842211"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/yo/betas/releases" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`[
		  {"tag_name":"v2.0.0-beta.10-mobile","draft":false,"assets":[{"name":"Orchard-2.0.0.apk","browser_download_url":"https://x/a.apk","size":1,"digest":"sha256:` + sha + `"}]},
		  {"tag_name":"v1.0.0-canary.9","draft":false,"assets":[]},
		  {"tag_name":"v6.0.0-beta.1","draft":true,"assets":[{"name":"Orchard-Setup-6.0.0.exe","browser_download_url":"https://x/borrador.exe","size":1,"digest":"sha256:` + sha + `"}]},
		  {"tag_name":"v5.0.0-beta.9","published_at":"2026-09-10T18:25:10Z","draft":false,"assets":[
		    {"name":"Orchard-Setup-5.0.0-arm64.exe","browser_download_url":"https://x/arm64.exe","size":1,"digest":"sha256:` + sha + `"},
		    {"name":"Orchard-Setup-5.0.0.exe","browser_download_url":"https://x/setup.exe","size":2,"digest":"sha256:` + sha + `"}]},
		  {"tag_name":"v5.0.0-beta.8","draft":false,"assets":[{"name":"Orchard-Setup-5.0.0.exe","browser_download_url":"https://x/vieja.exe","size":1,"digest":"sha256:` + sha + `"}]}
		]`))
	}))
	defer srv.Close()

	antes := apiGitHub
	apiGitHub = srv.URL
	cacheRel.mu.Lock()
	cacheRel.datos, cacheRel.limiteHasta, cacheRel.ruta = map[string]entradaCache{}, time.Time{}, ""
	cacheRel.mu.Unlock()
	defer func() {
		apiGitHub = antes
		cacheRel.mu.Lock()
		cacheRel.datos = map[string]entradaCache{}
		cacheRel.mu.Unlock()
	}()

	rel, err := ultimaRelease("yo/betas", false, true)
	if err != nil {
		t.Fatal(err)
	}
	if rel.Tag != "v5.0.0-beta.9" || rel.Instalador == nil || rel.Instalador.URL != "https://x/setup.exe" || rel.Instalador.SHA256 != sha {
		t.Fatalf("eligió %s %+v", rel.Tag, rel.Instalador)
	}
}
