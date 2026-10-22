# Historial de versiones

Todas las versiones siguen [Versionado Semántico](https://semver.org/lang/es/). Cada una tiene su
etiqueta `vX.Y.Z` y su página en [Releases](../../releases).

## [0.3.4] - 2026-10-22

### Añadido
- PdfEpub en el catálogo: convierte libros PDF a EPUB, con portada, datos y capítulos editables.

### Seguridad
- Los desinstaladores registrados para el usuario actual se ejecutan sin elevar permisos; solo se eleva una entrada registrada para toda la máquina.
- Actualiza el runtime de Go a 1.27.3.

## [0.3.3] - 2026-10-20

### Cambiado
- Buscar actualizaciones tiene dos opciones: cada 24 horas o manualmente. Van unidas al interruptor: activado es cada 24 horas; apagado es manual, con el botón de buscar actualizaciones junto al engranaje; en modo manual, Codex igual revisa una vez al abrirse. Las revisiones cada 1, 6 o 12 horas de antes pasan a 24.

## [0.3.2] - 2026-10-19

### Corregido
- Al instalar ClipVault o PngWebp, Codex crea su acceso directo en el escritorio si falta. Con una cuenta de Windows que no es administradora, el instalador lo dejaba en el escritorio de la cuenta que dio el permiso.

## [0.3.1] - 2026-10-19

### Corregido
- Al actualizar, el escritorio y la barra de tareas seguían mostrando el icono anterior: el instalador ahora le pide a Windows que refresque sus iconos.
- Si un repositorio del catálogo se muda a otra cuenta u organización de GitHub, Codex sigue la redirección y no lo da por «sin versiones».

## [0.3.0] - 2026-10-19

### Añadido
- Codex puede seguir abierto en la bandeja del sistema al cerrar la ventana (activado por defecto): clic en el icono para volver y clic derecho para «Abrir Codex» o «Salir». Mientras tanto sigue buscando actualizaciones cada 6 horas, sin gastar el límite de GitHub.
- Opción para abrir Codex al iniciar sesión, directo en la bandeja (desactivada por defecto). El desinstalador la quita.
- Idiomas: español, inglés, francés, alemán, ruso, portugués, chino y japonés, con la descripción de cada app traducida. Por defecto, el idioma del PC.

### Cambiado
- El icono de la app y el logo del encabezado son el libro solo, sin el cuadro café, y más grandes.
- Cuadrícula tenue de fondo detrás de las tarjetas.
- La imagen del instalador muestra el libro sin marco y «~/codex_» más grande, sobre el café del logo.
- Licencia MIT a nombre de Miguel Bigott: el código es de uso libre manteniendo el aviso de autor. El instalador deja la licencia y las de terceros (`LICENCIAS-TERCEROS.txt`) junto a la app, y el instalador y el .exe dicen «© 2026 Miguel Bigott».

### Corregido
- Las tildes del instalador se veían rotas (por ejemplo «tambiÃ©n»).

## [0.2.2] - 2026-10-19

### Cambiado
- Al pasar el mouse por una app de ArtCraft, la barra de arriba de su tarjeta toma el color de esa app, como en el launcher de ArtCraft. Las apps propias la mantienen en el dorado del skin.
- ClipVault se escribe con V mayúscula.

### Corregido
- Después de abrir una app, el botón «abrir» quedaba gris y no se podía volver a usar hasta reiniciar Codex.

## [0.2.1] - 2026-10-18

### Corregido
- Codex ya no agota las 60 consultas por hora que GitHub permite sin cuenta. Sin cuenta, la API gasta una consulta incluso cuando responde «sin cambios», así que cada revisión costaba 16. Ahora Codex mira primero la página pública de releases de cada app, que no cuenta para el límite, y usa la API solo cuando hay una versión nueva. Una revisión normal no gasta ninguna consulta.

## [0.2.0] - 2026-10-18

### Cambiado
- Skin «Nítido»: Codex se ve más limpio y claro. El fondo es gris neutro, las líneas son continuas y bien visibles (antes eran casi invisibles y algunas punteadas), los textos son más blancos, los nombres de las apps van en Inter, los estados llevan borde y las esquinas son más finas. El ámbar sigue en los botones y en el cursor.

## [0.1.9] - 2026-10-18

### Cambiado
- Las apps de ArtCraft muestran solo su nombre, su lema y su autor, como en su propio launcher.
- PNGWebp se muestra como «PngWebp», para que su nombre se vea parejo con el resto en la letra del launcher.

## [0.1.8] - 2026-10-18

### Cambiado
- Codex recuerda lo último que supo de cada app aunque lo cierres (en `%LOCALAPPDATA%\Codex\releases.json`): al abrirlo muestra las versiones al instante, y al revisar, GitHub solo responde «sin cambios», que no gasta su límite de 60 consultas por hora.
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

[0.2.2]: https://github.com/MiguelBigott/codex/releases/tag/v0.2.2
[0.2.1]: https://github.com/MiguelBigott/codex/releases/tag/v0.2.1
[0.2.0]: https://github.com/MiguelBigott/codex/releases/tag/v0.2.0
[0.1.9]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.9
[0.1.8]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.8
[0.1.7]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.7
[0.1.6]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.6
[0.1.5]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.5
[0.1.4]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.4
[0.1.3]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.3
[0.1.2]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.2
[0.1.1]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.1
[0.1.0]: https://github.com/MiguelBigott/codex/releases/tag/v0.1.0
