package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Apps que publican sus versiones de Windows fuera de GitHub (Orchard, desde su generación 2.0):
// Codex lee la última versión y sus notas de la web oficial de la app, solo para mostrarlas.
// Nunca descarga nada de ahí: esas apps se instalan desde su página oficial (InstalarDesdeWeb).

var reVersionWeb = regexp.MustCompile(`^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)

// leerJSON pide una dirección https y decodifica su JSON, sin pasar de max bytes.
func leerJSON(direccion string, max int64, v any) error {
	req, err := http.NewRequest(http.MethodGet, direccion, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Codex/"+Version)
	resp, err := cliente.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return io.ErrUnexpectedEOF
	}
	return json.NewDecoder(io.LimitReader(resp.Body, max)).Decode(v)
}

// releaseDeManifiesto arma la última versión de la app desde su manifiesto oficial y, si las
// hay, sus notas. Guarda el resultado como cualquier versión, para el panel de detalles.
func releaseDeManifiesto(app AppCatalogo) (*Release, error) {
	fallo := msj("versionWebError", "app", app.NombreCompleto())
	var firmado struct {
		Payload string `json:"payload"`
	}
	if err := leerJSON(app.ManifiestoVersion, 1<<20, &firmado); err != nil {
		return nil, fallo
	}
	crudo, err := base64.StdEncoding.DecodeString(firmado.Payload)
	if err != nil {
		return nil, fallo
	}
	var m struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(crudo, &m); err != nil || !reVersionWeb.MatchString(m.Version) {
		return nil, fallo
	}

	clave := app.Repo
	if clave == "" {
		clave = "web:" + app.ID
	}
	cacheRel.mu.Lock()
	previa := cacheRel.datos[clave]
	cacheRel.mu.Unlock()
	if previa.Release != nil && previa.Release.Version == m.Version && previa.ConCambios {
		return previa.Release, nil
	}

	rel := &Release{Version: m.Version, Tag: "v" + m.Version}
	cambios := ""
	if app.NotasVersion != "" {
		var notas []struct {
			Version   string `json:"version"`
			Fecha     string `json:"date"`
			Titulo    string `json:"title"`
			Secciones []struct {
				Titulo string   `json:"title"`
				Items  []string `json:"items"`
			} `json:"sections"`
		}
		if leerJSON(app.NotasVersion, 4<<20, &notas) == nil {
			for _, n := range notas {
				if n.Version != m.Version {
					continue
				}
				// A mediodía UTC: así el día no cambia al mostrarlo en otra zona horaria.
				if f, err := time.Parse("2006-01-02", n.Fecha); err == nil {
					rel.Fecha = time.Date(f.Year(), f.Month(), f.Day(), 12, 0, 0, 0, time.UTC)
				}
				var b strings.Builder
				if n.Titulo != "" {
					b.WriteString("## " + n.Titulo + "\n")
				}
				for _, s := range n.Secciones {
					if s.Titulo != "" {
						b.WriteString("### " + s.Titulo + "\n")
					}
					for _, it := range s.Items {
						b.WriteString("- " + it + "\n")
					}
				}
				cambios = recortarCambios(b.String())
				break
			}
		}
	}
	cacheRel.mu.Lock()
	cacheRel.datos[clave] = entradaCache{Release: rel, Cambios: cambios, ConCambios: true}
	cacheRel.mu.Unlock()
	return rel, nil
}
