package main

import (
	_ "embed"
	"encoding/json"
	"sync"
)

// Panel de detalles: lo que se ve al hacer clic en una app. La parte que no cambia (titular,
// descripción, funciones y plataformas) viene de detalles.json; la que cambia con cada versión
// (número, fecha, tamaño y notas) sale de la release de GitHub que Codex ya consulta.
//
// Los textos de las apps de ArtCraft (y de SolidCraft y MailCraft) son los de su propio launcher
// (storytold/craft-launcher, crates/core/catalog.toml, licencia Apache 2.0 o MIT), copiados
// tal cual y en inglés, como allí. Los de ClipVault, PngWebp y PdfEpub (en español, con inglés)
// y los de Orchard y OBS Studio (en inglés, con español, sacados de sus README) están escritos a
// mano. Se sirven aparte para que /api/estado siga siendo liviano.

//go:embed detalles.json
var detallesJSON []byte

type Funcion struct {
	Titulo string `json:"titulo"`
	Texto  string `json:"texto"`
}

type Detalle struct {
	// Idioma del texto original: «en» (ArtCraft, Orchard, OBS) o «es» (las apps de Codex).
	Idioma      string    `json:"idioma,omitempty"`
	Titular     string    `json:"titular"`
	Acerca      string    `json:"acerca"`
	Plataformas string    `json:"plataformas,omitempty"`
	Funciones   []Funcion `json:"funciones"`
	// Traducciones: el mismo texto en el otro idioma. El panel se abre en español si Codex está
	// en español y en inglés en los demás idiomas, y un botón cambia entre los dos.
	Traducciones map[string]Detalle `json:"traducciones,omitempty"`
}

var (
	detallesUnaVez sync.Once
	detallesApps   map[string]Detalle
	errDetalles    error
)

// cargarDetalles lee detalles.json una sola vez, la primera vez que se abre un panel.
func cargarDetalles() (map[string]Detalle, error) {
	detallesUnaVez.Do(func() { errDetalles = json.Unmarshal(detallesJSON, &detallesApps) })
	return detallesApps, errDetalles
}

type RespuestaDetalles struct {
	Detalle *Detalle `json:"detalle"`
	Cambios string   `json:"cambios"`
}

// Detalles arma el panel de una app: su página (si la tiene) y las notas de su última versión.
func (l *Launcher) Detalles(id string) (*RespuestaDetalles, error) {
	var app *AppCatalogo
	for i := range l.cat.Apps { // el catálogo no cambia después de cargarse
		if l.cat.Apps[i].ID == id {
			app = &l.cat.Apps[i]
			break
		}
	}
	if app == nil {
		return nil, msj("appDesconocida")
	}
	r := &RespuestaDetalles{}
	if d, err := cargarDetalles(); err == nil {
		if x, ok := d[id]; ok {
			r.Detalle = &x
		}
	}
	if app.Repo != "" {
		r.Cambios = cambiosDe(app.Repo)
	}
	return r, nil
}
