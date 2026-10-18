# Historial de versiones

Todas las versiones siguen [Versionado Semántico](https://semver.org/lang/es/). Cada una tiene su
etiqueta `vX.Y.Z` y su página en [Releases](../../releases).

## [0.1.8] - 2026-10-18

### Cambiado
- Codex recuerda lo último que supo de cada app aunque lo cierres (en `%LOCALAPPDATA%\Codexeleases.json`): al abrirlo muestra las versiones al instante, y al revisar, GitHub solo responde «sin cambios», que no gasta su límite de 60 consultas por hora.
- Si GitHub pausa las consultas, Codex deja de preguntar hasta la hora que indica, sigue mostrando las últimas versiones conocidas y lo avisa en el pie, en lugar de un error rojo en cada app.

## [0.1.7] - 2026-10-18

### Cambiado
- «vault» y «Webp» se ven en el dorado del skin de Codex, como el resto de sus acentos.

## [0.1.6] - 2026-10-18

### Cambiado
- Las apps de ArtCraft llevan su numeración oficial (PhotoCraft 01, VectorCraft 02 … GridCraft 13), en un recuadro del color de cada app, y aparecen en ese mismo orden. Mis apps ya no llevan número.
- «Craft» (y «vault» o «Webp» en mis apps) se ve en el color de cada app, con los colores oficiales de ArtCraft.

### Quitado
- La app principal ArtCraft: el catálogo lleva las apps de ArtCraft, no su app principal.

## [0.1.5] - 2026-10-18

### Añadido
- Todo ArtCraft en el catálogo: [ArtCraft](https://github.com/storytold/artcraft) (la app principal), [LightCraft](https://github.com/storytold/lightcraft), [VectorCraft](https://github.com/storytold/vectorcraft), [DesignCraft](https://github.com/storytold/designcraft), [FigCraft](https://github.com/storytold/figcraft), [DeckCraft](https://github.com/storytold/deckcraft) y [CADCraft](https://github.com/storytold/cadcraft).
- Cada app lleva su número (01, 02...) en el color de la app, en la cuadrícula y en la tabla.

### Cambiado
- La sección de mis apps se llama «codex» en lugar de «multimedia».
- Los nombres de las apps se ven en blanco puro.
- Leyendas nuevas para FilmCraft, PdfCraft, EffectCraft, WordCraft, GridCraft y SoundCraft.

### Corregido
- Las versiones de apps cuyas etiquetas llevan el nombre delante (ArtCraft publica «artcraft-v0.41.0») ahora se leen bien («0.41.0»), así que los avisos de actualización funcionan con ellas.
- En la vista de tabla, las fechas largas («26 sept 2026») ya no se cortan.

## [0.1.4] - 2026-10-18

### Añadido
- En el catálogo: [PNGWebp](https://github.com/MiguelBigott/pngwebp), que convierte imágenes PNG y JPG a WebP con la máxima calidad y el menor peso.

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

[0.1.8]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.8
[0.1.7]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.7
[0.1.6]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.6
[0.1.5]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.5
[0.1.4]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.4
[0.1.3]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.3
[0.1.2]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.2
[0.1.1]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.1
[0.1.0]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.0
