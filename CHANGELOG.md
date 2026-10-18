# Historial de versiones

Todas las versiones siguen [Versionado Semántico](https://semver.org/lang/es/). Cada una tiene su
etiqueta `vX.Y.Z` y su página en [Releases](../../releases).

## [0.1.3] - 2026-10-18

### Cambiado
- El título dice `~/codex app suite`.
- En la vista de tabla, el estado de cada app se lee completo («disponible para instalar» ya no se corta).

## [0.1.2] - 2026-10-18

### Quitado
- Las apps de ejemplo que quedaban (PantallaAngel, PdfAngel, NotaAngel y ArchivoAngel) y la categoría Productividad, que quedó vacía. El catálogo muestra solo apps reales: Clipvault y las siete de ArtCraft.

## [0.1.1] - 2026-10-18

### Añadido
- En el catálogo, de ArtCraft: [PdfCraft](https://github.com/storytold/pdfcraft) (PDF), [EffectCraft](https://github.com/storytold/effectcraft) (animación gráfica) y [SoundCraft](https://github.com/storytold/soundcraft) (audio).

### Quitado
- Las apps de ejemplo FotoAngel y AudioAngel.

## [0.1.0] - 2026-10-18

Primera versión de **Codex**, el launcher de mis apps de Windows.

### Añadido
- La biblioteca de apps en cuadrícula o en tabla, con filtros, orden y búsqueda.
- Detección de las apps instaladas y su versión desde el registro de Windows.
- Instalación, actualización y desinstalación con un clic, desde los releases de GitHub de cada app, con verificación SHA-256. Acepta setups NSIS y paquetes MSI.
- Aviso y actualización del propio launcher.
- Skin «Terminal» y logo «Códice cerrado».
- Ventana nativa de Windows (WebView2) con icono nítido en la barra de tareas.
- En el catálogo: [Clipvault](https://github.com/MiguelBigott/clipvault) y, de ArtCraft, [PhotoCraft](https://github.com/storytold/photocraft), [FilmCraft](https://github.com/storytold/filmcraft), [WordCraft](https://github.com/storytold/wordcraft) y [GridCraft](https://github.com/storytold/gridcraft).

[0.1.3]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.3
[0.1.2]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.2
[0.1.1]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.1
[0.1.0]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.0
