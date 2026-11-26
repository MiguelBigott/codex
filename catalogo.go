package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

// El catálogo es la lista de apps que conoce el launcher. Cada app publica sus versiones
// como releases de GitHub (`repo`, público) con su instalador NSIS y su suma SHA-256.

//go:embed catalogo.json
var catalogoJSON []byte

type Categoria struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

type AppCatalogo struct {
	ID          string `json:"id"`
	Nombre      string `json:"nombre"` // "Video"
	Sufijo      string `json:"sufijo"` // "vault", "Craft": se muestra con el color de la app
	Descripcion string `json:"descripcion"`
	Similar     string `json:"similar"`
	Color       string `json:"color"`
	Categoria   string `json:"categoria"`
	Repo        string `json:"repo"` // "usuario/repositorio" en GitHub
	Web         string `json:"web"`
	// Nombre con el que el instalador se registra en Windows (por defecto Nombre+Sufijo).
	NombreRegistro string `json:"nombreRegistro"`
	// Ejecutable dentro de la carpeta de la app (por defecto Nombre+Sufijo+".exe").
	Ejecutable string `json:"ejecutable"`
	// Ejemplo: app que aún no existe; se muestra como «Próximamente».
	Ejemplo bool `json:"ejemplo"`
	// Formato del instalador: "nsis" (setup .exe, por defecto) o "msi".
	Formato string `json:"formato"`
	// Propiedad del MSI que fija la carpeta de instalación (p. ej. INSTALLFOLDER).
	PropiedadCarpeta string `json:"propiedadCarpeta"`
	// Autor, para apps de terceros (se muestra como crédito).
	Autor string `json:"autor"`
	// Icono: URL https de una imagen; si falta, se dibuja la inicial.
	Icono string `json:"icono"`
	// Número de la app en su colección: ArtCraft numera las suyas (PhotoCraft 01... GridCraft 13).
	Numero int `json:"numero,omitempty"`
	// Color del sufijo («Craft», «vault») sobre fondo oscuro; si falta, se usa Color.
	ColorSufijo string `json:"colorSufijo,omitempty"`
	// Descripción y lema en otros idiomas («en», «fr»...); lo que falte se muestra en español.
	Traducciones map[string]Traduccion `json:"traducciones,omitempty"`
	// PorUsuario: el instalador se instala solo para el usuario, sin permiso de administrador,
	// y elige su propia carpeta. Codex lo ejecuta sin elevar y sin /D=.
	PorUsuario bool `json:"porUsuario,omitempty"`
	// SeActualizaSola: la app trae su propio actualizador. Codex la instala, la abre y la
	// desinstala, pero no ofrece actualizaciones (no pelea con el de la app).
	SeActualizaSola bool `json:"seActualizaSola,omitempty"`
	// Betas: si la última versión estable no trae instalador para Windows, se usa la versión
	// más reciente que sí lo traiga, aunque sea beta.
	Betas bool `json:"betas,omitempty"`
	// InstalarDesdeWeb: la app no publica su instalador con una suma SHA-256 que Codex pueda
	// verificar, así que Codex no lo descarga: «Instalar» abre su página oficial (`web`). La
	// detecta, la abre y la desinstala igual que a las demás.
	InstalarDesdeWeb bool `json:"instalarDesdeWeb,omitempty"`
	// ManifiestoVersion: manifiesto de versiones de la app fuera de GitHub (formato
	// orchard-signed-v1: un «payload» en base64 con su «version»). Solo sirve para mostrar la
	// última versión; nunca para descargar nada.
	ManifiestoVersion string `json:"manifiestoVersion,omitempty"`
	// NotasVersion: las notas de esas versiones (lista de {version, date, title, sections}),
	// para el panel de detalles.
	NotasVersion string `json:"notasVersion,omitempty"`
}

type Traduccion struct {
	Descripcion string `json:"descripcion,omitempty"`
	Similar     string `json:"similar,omitempty"`
}

func (a AppCatalogo) EsMSI() bool { return strings.EqualFold(a.Formato, "msi") }

func (a AppCatalogo) NombreCompleto() string { return a.Nombre + a.Sufijo }

func (a AppCatalogo) Registro() string {
	if a.NombreRegistro != "" {
		return a.NombreRegistro
	}
	return a.NombreCompleto()
}

func (a AppCatalogo) Exe() string {
	if a.Ejecutable != "" {
		return a.Ejecutable
	}
	return a.NombreCompleto() + ".exe"
}

func (a AppCatalogo) URLRepo() string {
	if a.Repo == "" {
		return ""
	}
	return "https://github.com/" + a.Repo
}

type Catalogo struct {
	Categorias []Categoria   `json:"categorias"`
	Apps       []AppCatalogo `json:"apps"`
}

func cargarCatalogo() (Catalogo, error) {
	var c Catalogo
	if err := json.Unmarshal(catalogoJSON, &c); err != nil {
		return c, fmt.Errorf("catálogo: %w", err)
	}
	ids := map[string]bool{}
	for _, a := range c.Apps {
		if a.ID == "" || ids[a.ID] {
			return c, fmt.Errorf("catálogo: id vacío o repetido (%q)", a.ID)
		}
		if a.Repo != "" && strings.Count(a.Repo, "/") != 1 {
			return c, fmt.Errorf("catálogo: repo inválido en %s", a.ID)
		}
		if f := strings.ToLower(a.Formato); f != "" && f != "nsis" && f != "msi" {
			return c, fmt.Errorf("catálogo: formato %q desconocido en %s", a.Formato, a.ID)
		}
		if a.Icono != "" && !strings.HasPrefix(a.Icono, "https://") {
			return c, fmt.Errorf("catálogo: el icono de %s debe ser una URL https", a.ID)
		}
		for _, u := range []string{a.ManifiestoVersion, a.NotasVersion} {
			if u != "" && !strings.HasPrefix(u, "https://") {
				return c, fmt.Errorf("catálogo: las direcciones de versiones de %s deben ser https", a.ID)
			}
		}
		if a.InstalarDesdeWeb && !strings.HasPrefix(a.Web, "https://") {
			return c, fmt.Errorf("catálogo: %s se instala desde su web, pero no tiene una web https", a.ID)
		}
		ids[a.ID] = true
	}
	return c, nil
}
