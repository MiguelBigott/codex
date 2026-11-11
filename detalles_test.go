package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// Cada página de detalles corresponde a una app del catálogo y trae lo que el panel muestra.
func TestDetalles(t *testing.T) {
	d, err := cargarDetalles()
	if err != nil {
		t.Fatal(err)
	}
	cat, err := cargarCatalogo()
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, a := range cat.Apps {
		ids[a.ID] = true
	}
	if len(d) == 0 {
		t.Fatal("detalles.json está vacío")
	}
	for id, x := range d {
		if !ids[id] {
			t.Errorf("detalles de %q, que no está en el catálogo", id)
		}
		if x.Titular == "" || x.Acerca == "" || x.Plataformas == "" {
			t.Errorf("%s: faltan titular, descripción o plataformas", id)
		}
		for _, f := range x.Funciones {
			if f.Titulo == "" || f.Texto == "" {
				t.Errorf("%s: función sin título o sin texto", id)
			}
		}
		// Toda página está en español y en inglés: la original y su traducción.
		otro := map[string]string{"es": "en", "en": "es"}[x.Idioma]
		if otro == "" {
			t.Errorf("%s: idioma original %q (debe ser «es» o «en»)", id, x.Idioma)
		} else if _, ok := x.Traducciones[otro]; !ok {
			t.Errorf("%s: falta la traducción %q", id, otro)
		}
		for idioma, tr := range x.Traducciones {
			if tr.Titular == "" || tr.Acerca == "" || len(tr.Funciones) != len(x.Funciones) {
				t.Errorf("%s: la traducción %q está incompleta", id, idioma)
			}
		}
		// Ningún texto visible nombra al sitio de videos ni a su motor de descarga.
		if b, _ := json.Marshal(x); strings.Contains(strings.ToLower(string(b)), "youtube") || strings.Contains(strings.ToLower(string(b)), "yt-dlp") {
			t.Errorf("%s: el texto nombra al sitio de videos o a su motor", id)
		}
	}

	l := &Launcher{cat: cat}
	if _, err := l.Detalles("no-existe"); err == nil {
		t.Error("una app desconocida debería dar error")
	}
	if r, err := l.Detalles("figcraft"); err != nil || r.Detalle == nil || len(r.Detalle.Funciones) == 0 {
		t.Errorf("FigCraft sin detalles: %v %+v", err, r)
	}
}

// Las notas largas se recortan en un salto de línea y sin pasar del máximo.
func TestRecortarCambios(t *testing.T) {
	if got := recortarCambios("a\r\nb"); got != "a\nb" {
		t.Errorf("saltos de línea: %q", got)
	}
	largo := strings.Repeat("* un cambio\n", maxCambios/6)
	got := recortarCambios(largo)
	if len(got) > maxCambios || strings.HasSuffix(got, "* un cam") {
		t.Errorf("recorte: %d bytes, termina en %q", len(got), got[len(got)-10:])
	}
}
