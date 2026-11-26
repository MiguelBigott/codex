package main

import (
	"encoding/json"
	"testing"
)

// El DisplayName del registro cuenta si es el nombre de la app, o el nombre seguido de su versión
// (electron-builder registra «Recordly 1.4.0»).
func TestCoincideNombre(t *testing.T) {
	casos := []struct {
		registrado, nombre string
		ok                 bool
	}{
		{"Recordly", "Recordly", true},
		{"  recordly ", "Recordly", true},
		{"Recordly 1.4.0", "Recordly", true},
		{"Recordly v2.0.0-beta.1", "Recordly", true},
		{"Recordly Helper", "Recordly", false},
		{"Recordly", "Record", false},
		{"Recordly 1.4.0", "Record", false},
		{"OBS Studio", "OBS Studio", true},
		{"OBS Studio Plugins", "OBS Studio", false},
		{"Recordly ", "Recordly", true},
		{"", "Recordly", false},
	}
	for _, c := range casos {
		if got := coincideNombre(c.registrado, c.nombre); got != c.ok {
			t.Errorf("coincideNombre(%q, %q) = %v", c.registrado, c.nombre, got)
		}
	}
}

// Con archivoInstalador, se elige ese archivo aunque no se llame «setup» ni «installer».
func TestReleaseDeArchivoExacto(t *testing.T) {
	sha := "c65270948c70727ff461cd69eeaf4ebaf5ec7284852fe6a94dfe62893504eaa"
	var gh releaseGitHub
	if err := json.Unmarshal([]byte(`{"tag_name":"v1.4.0","assets":[
	  {"name":"Recordly-windows-x64.exe.blockmap","browser_download_url":"https://x/b","size":1,"digest":"sha256:`+sha+`1"},
	  {"name":"Recordly-x64.zip","browser_download_url":"https://x/z","size":1,"digest":"sha256:`+sha+`1"},
	  {"name":"Recordly-windows-x64.exe","browser_download_url":"https://x/setup.exe","size":185307721,"digest":"sha256:`+sha+`1"}]}`), &gh); err != nil {
		t.Fatal(err)
	}
	rel := releaseDe(gh, false, "Recordly-windows-x64.exe")
	if rel.Instalador == nil || rel.Instalador.URL != "https://x/setup.exe" || rel.Instalador.SHA256 != sha+"1" {
		t.Fatalf("eligió %+v", rel.Instalador)
	}
	if releaseDe(gh, false, "").Instalador != nil {
		t.Error("sin archivoInstalador no debería reconocer ese nombre")
	}
}
