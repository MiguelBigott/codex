package main

import "testing"

func TestVersionDeEtiqueta(t *testing.T) {
	casos := map[string]string{
		"v1.0.0":           "1.0.0",
		"1.6.0":            "1.6.0",
		"artcraft-v0.41.0": "0.41.0",
		"v0.10.0":          "0.10.0",
	}
	for tag, quiere := range casos {
		if got := versionDeEtiqueta(tag); got != quiere {
			t.Errorf("versionDeEtiqueta(%q) = %q, quiere %q", tag, got, quiere)
		}
	}
	if compararVersiones(versionDeEtiqueta("artcraft-v0.41.0"), "0.41.0") != 0 {
		t.Error("ArtCraft 0.41.0 instalado no debe pedir actualización")
	}
	if compararVersiones(versionDeEtiqueta("artcraft-v0.42.0"), "0.41.0") <= 0 {
		t.Error("ArtCraft 0.42.0 debe aparecer como actualización de la 0.41.0")
	}
}
