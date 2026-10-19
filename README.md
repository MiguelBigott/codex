<div align="center">

<img src="build/logo.png" alt="Logo de Codex" width="96">

# Codex

> Un solo lugar para **instalar, actualizar y abrir** mis apps de Windows: eliges una app, Codex baja su instalador oficial, verifica su firma SHA-256 y la deja lista.

[![Versión](https://img.shields.io/badge/versión-v0.3.0-ffb454)](CHANGELOG.md)
[![CI](https://github.com/MiguelBigott/codex/actions/workflows/ci.yml/badge.svg)](https://github.com/MiguelBigott/codex/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Windows](https://img.shields.io/badge/Windows-10%20·%2011-0078D4)](#-instalación)
[![Apps](https://img.shields.io/badge/apps-15-c9a45c)](#-catálogo)
[![Licencia](https://img.shields.io/badge/licencia-MIT-3f9d6e)](LICENSE)

[Historial de versiones](CHANGELOG.md) · [Releases](../../releases) · [Descargar el instalador](../../releases/latest)

<img src="https://github.com/MiguelBigott/codex/raw/estadisticas/estadisticas.svg" alt="Actividad del repositorio en los últimos 30 días" width="100%">

<sub>Panel de actividad generado a diario por <a href=".github/workflows/estadisticas.yml">GitHub Actions</a>, sin servicios de terceros.</sub>

</div>

---

## 📖 Índice

- [Funciones](#-funciones)
- [Catálogo](#-catálogo)
- [Stack](#-stack)
- [Instalación](#-instalación)
- [Compilar](#️-compilar)
- [Añadir una app](#-añadir-una-app)
- [Seguridad](#-seguridad)
- [Licencia](#️-licencia)
- [Créditos](#️-créditos)

## ✨ Funciones

- **La biblioteca**: las apps del catálogo en cuadrícula o en tabla, con filtros (todas, instaladas, con actualización, sin instalar), orden y búsqueda.
- **Sabe qué tienes**: detecta qué apps están instaladas y en qué versión leyendo el registro de Windows, aunque no las hayas instalado desde Codex.
- **Siempre al día**: consulta la última versión de cada app en sus releases de GitHub y marca las que tienen actualización. Mira primero la página pública de releases, que no cuenta para el límite de consultas de GitHub, y usa la API solo cuando hay una versión nueva; si GitHub pausa las consultas, sigue mostrando las últimas versiones conocidas.
- **Instala con un clic**: baja el instalador, **verifica su SHA-256** y lo ejecuta como administrador (Windows muestra el aviso de UAC). Acepta setups `.exe` (NSIS) y paquetes `.msi`, y también desinstala.
- **Se actualiza solo**: avisa cuando hay una versión nueva del propio launcher y se pone al día con un clic.
- **Skin «Terminal», versión «Nítido»**: fondo gris neutro con líneas continuas y bien visibles, nombres de las apps en Inter, etiquetas en minúsculas monoespaciadas, estados con borde y un solo acento ámbar, con el cursor de `~/codex_`.
- **Ventana propia**: ventana nativa de Windows (WebView2) con su icono nítido en la barra de tareas, sin barra de título del sistema.
- **En segundo plano**: al cerrar la ventana, Codex sigue en la bandeja del sistema (clic en el icono para volver, clic derecho para salir), y puede abrirse solo al iniciar sesión. Las dos opciones están en Ajustes.
- **8 idiomas**: español, inglés, francés, alemán, ruso, portugués, chino y japonés. Por defecto usa el idioma del PC (si no es ninguno de esos, inglés), y se puede elegir otro en Ajustes.

## 📚 Catálogo

Las apps de ArtCraft llevan el mismo número y los mismos colores que en su propio launcher (PhotoCraft 01 … GridCraft 13).

| # | App | Qué hace | Autor | Repositorio |
|---|---|---|---|---|
|  | <img src="https://raw.githubusercontent.com/MiguelBigott/clipvault/main/build/logo.png" width="20" alt=""> **ClipVault** | Descarga de video y audio · MP4 o MP3 con portada y metadatos | Miguel Bigott | [MiguelBigott/clipvault](https://github.com/MiguelBigott/clipvault) |
|  | <img src="https://raw.githubusercontent.com/MiguelBigott/pngwebp/main/build/logo.png" width="20" alt=""> **PngWebp** | Conversión de imágenes a WebP · PNG o JPG con la máxima calidad y el menor peso | Miguel Bigott | [MiguelBigott/pngwebp](https://github.com/MiguelBigott/pngwebp) |
| 01 | <img src="https://raw.githubusercontent.com/storytold/photocraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.photocraft.png" width="20" alt=""> **PhotoCraft** | Edición de imágenes | [ArtCraft](https://getartcraft.com) | [storytold/photocraft](https://github.com/storytold/photocraft) |
| 02 | <img src="https://raw.githubusercontent.com/storytold/vectorcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.vectorcraft.png" width="20" alt=""> **VectorCraft** | Ilustración vectorial | [ArtCraft](https://getartcraft.com) | [storytold/vectorcraft](https://github.com/storytold/vectorcraft) |
| 03 | <img src="https://raw.githubusercontent.com/storytold/filmcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.filmcraft.png" width="20" alt=""> **FilmCraft** | Edición de video, color y sonido | [ArtCraft](https://getartcraft.com) | [storytold/filmcraft](https://github.com/storytold/filmcraft) |
| 04 | <img src="https://raw.githubusercontent.com/storytold/lightcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.lightcraft.png" width="20" alt=""> **LightCraft** | Revelado y edición de fotos | [ArtCraft](https://getartcraft.com) | [storytold/lightcraft](https://github.com/storytold/lightcraft) |
| 05 | <img src="https://raw.githubusercontent.com/storytold/pdfcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.pdfcraft.png" width="20" alt=""> **PdfCraft** | Leer, organizar y proteger PDF | [ArtCraft](https://getartcraft.com) | [storytold/pdfcraft](https://github.com/storytold/pdfcraft) |
| 06 | <img src="https://raw.githubusercontent.com/storytold/effectcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.effectcraft.png" width="20" alt=""> **EffectCraft** | Gráficos en movimiento y efectos visuales | [ArtCraft](https://getartcraft.com) | [storytold/effectcraft](https://github.com/storytold/effectcraft) |
| 07 | <img src="https://raw.githubusercontent.com/storytold/designcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.designcraft.png" width="20" alt=""> **DesignCraft** | Maquetación de páginas | [ArtCraft](https://getartcraft.com) | [storytold/designcraft](https://github.com/storytold/designcraft) |
| 08 | <img src="https://raw.githubusercontent.com/storytold/figcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.figcraft.png" width="20" alt=""> **FigCraft** | Diseño de interfaces y prototipos | [ArtCraft](https://getartcraft.com) | [storytold/figcraft](https://github.com/storytold/figcraft) |
| 09 | <img src="https://raw.githubusercontent.com/storytold/wordcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.wordcraft.png" width="20" alt=""> **WordCraft** | Escritura y diseño de documentos | [ArtCraft](https://getartcraft.com) | [storytold/wordcraft](https://github.com/storytold/wordcraft) |
| 10 | <img src="https://raw.githubusercontent.com/storytold/soundcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.soundcraft.png" width="20" alt=""> **SoundCraft** | Grabación, edición y mezcla de audio | [ArtCraft](https://getartcraft.com) | [storytold/soundcraft](https://github.com/storytold/soundcraft) |
| 11 | <img src="https://raw.githubusercontent.com/storytold/deckcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.deckcraft.png" width="20" alt=""> **DeckCraft** | Presentaciones y diapositivas | [ArtCraft](https://getartcraft.com) | [storytold/deckcraft](https://github.com/storytold/deckcraft) |
| 12 | <img src="https://raw.githubusercontent.com/storytold/cadcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.cadcraft.png" width="20" alt=""> **CADCraft** | Diseño y dibujo técnico | [ArtCraft](https://getartcraft.com) | [storytold/cadcraft](https://github.com/storytold/cadcraft) |
| 13 | <img src="https://raw.githubusercontent.com/storytold/gridcraft/HEAD/assets/app-icon/hicolor/256x256/apps/ai.storyteller.gridcraft.png" width="20" alt=""> **GridCraft** | Hojas de cálculo y datos | [ArtCraft](https://getartcraft.com) | [storytold/gridcraft](https://github.com/storytold/gridcraft) |

Cada app nueva entra al catálogo cuando tiene su primera versión publicada.

## 🧱 Stack

| Capa | Tecnología |
|---|---|
| Aplicación | [Go](https://go.dev): servidor local con la biblioteca estándar |
| Ventana | [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/) con [go-webview2](https://github.com/jchv/go-webview2) |
| Interfaz | HTML, CSS y JavaScript embebidos (`go:embed`), tipografías Inter y Space Mono |
| Versiones | API de releases de GitHub, con caché por ETag |
| Instalador | [NSIS](https://nsis.sourceforge.io) |
| CI | GitHub Actions |

## 📦 Instalación

Descarga `Codex-Setup-<versión>.exe` desde [Releases](../../releases/latest) y ejecútalo.

- Todo queda en `C:\Program Files\Codex`: el launcher (`Codex.exe`) y una subcarpeta por app (`C:\Program Files\Codex\<App>`).
- La configuración vive en `%LOCALAPPDATA%\Codex`.
- El instalador ofrece crear un acceso directo en el escritorio.

## 🛠️ Compilar

Requisitos: [Go](https://go.dev/dl) 1.27 o superior, [NSIS](https://nsis.sourceforge.io) y go-winres (`go install github.com/tc-hib/go-winres@latest`).

```powershell
.\build\compilar.ps1
```

Genera `dist\Codex.exe` y `dist\Codex-Setup-<versión>.exe`. La versión se toma de `const Version` en `main.go`.

El logo «Códice cerrado» se define en `build/generar_logo.py`, que escribe `web/logo.svg`, `build/logo.png`, `build/app.ico` (con una versión simplificada para 16-32 px) y `build/lateral.bmp` (requiere Python con Pillow y Microsoft Edge):

```powershell
python build\generar_logo.py
```

Para revisar la interfaz en un navegador, sin ventana, ejecuta el launcher con `CODEX_NAVEGADOR=1`: la dirección queda en `%LOCALAPPDATA%\Codex\url-desarrollo.txt`.

<details>
<summary>Cómo está hecho</summary>

| Archivo | Contenido |
|---|---|
| `main.go` | Arranque, instancia única y ventana |
| `ventana.go` | Ventana nativa sin barra de título (el encabezado de la interfaz hace de barra) |
| `servidor.go` | Servidor local (solo `127.0.0.1`, con clave por sesión) que usa la interfaz |
| `launcher.go` | Estado, revisión de versiones, descargas verificadas e instalaciones (NSIS y MSI) |
| `releases.go` | Consulta a la API de GitHub y sumas SHA-256 |
| `instalados.go` | Apps instaladas según el registro de Windows |
| `elevar.go` | Ejecución como administrador (UAC) |
| `catalogo.go`, `catalogo.json` | El catálogo de apps |
| `web/` | Interfaz (HTML, CSS y JS) |
| `build/` | Logo, instalador NSIS y script de compilación |

</details>

## ➕ Añadir una app

Cada app vive en su propio repositorio público y se define en [`catalogo.json`](catalogo.json):

| Campo | Para qué |
|---|---|
| `id` | Identificador único |
| `nombre` + `sufijo` | Se muestran juntos (`Clip` + `vault`) |
| `descripcion`, `similar` | Las líneas bajo el nombre, en español |
| `traducciones` | Las mismas líneas en otros idiomas: `{"en": {"descripcion": "…", "similar": "…"}}`; lo que falte se muestra en español |
| `autor` | Crédito para apps de terceros («de ArtCraft») |
| `color` | Color de la app (icono de respaldo) |
| `icono` | URL `https` de su icono oficial (opcional) |
| `categoria` | Sección donde aparece |
| `repo` | Repositorio de GitHub (`usuario/repo`) donde publica sus versiones |
| `web` | Sitio web (opcional) |
| `formato` | `nsis` (setup `.exe`, por defecto) o `msi` |
| `propiedadCarpeta` | Solo MSI: propiedad que fija la carpeta (p. ej. `INSTALLFOLDER`) |
| `nombreRegistro` | Nombre con el que se registra en Windows (por defecto `nombre`+`sufijo`) |
| `ejecutable` | `.exe` dentro de su carpeta (por defecto `nombre`+`sufijo`.exe) |
| `ejemplo` | `true` para mostrarla como «Próximamente» |

Cada release debe incluir el instalador para Windows x64 (un `.exe` con `setup` en el nombre, o un `.msi` con `x64`) y su suma SHA-256: GitHub ya la publica para cada archivo, o puede ir en un `SHA256SUMS.txt`.

## 🔐 Seguridad

- **Sin suma, no instala**: cada instalador se compara con su SHA-256 antes de ejecutarse; si no coincide, se borra.
- **Solo desde GitHub**: los instaladores se bajan de los releases oficiales de cada repositorio del catálogo, por HTTPS.
- **Servidor cerrado**: la interfaz habla con un servidor en `127.0.0.1` que exige una clave nueva en cada sesión.
- **Administrador solo para instalar**: el launcher corre como usuario normal y pide permiso (UAC) únicamente al instalar o desinstalar.
- **CI con acciones fijadas por SHA** y permisos mínimos en cada workflow.

## ⚖️ Licencia

**Creado por [Miguel Bigott](https://github.com/MiguelBigott)** y publicado con licencia [MIT](LICENSE): cualquiera puede usarlo, copiarlo, modificarlo y compartirlo libremente, también en sus propios proyectos, siempre que mantenga el aviso de autor. Se entrega tal cual, sin garantías.

Las apps del catálogo tienen sus propias licencias y Codex las descarga de sus releases oficiales sin modificarlas. Lo de terceros que va dentro de Codex (Go, go-webview2 y las tipografías) conserva su licencia; el instalador las deja en `LICENCIAS-TERCEROS.txt`, junto a la app.

## 🖋️ Créditos

- [PhotoCraft](https://github.com/storytold/photocraft), [LightCraft](https://github.com/storytold/lightcraft), [VectorCraft](https://github.com/storytold/vectorcraft), [DesignCraft](https://github.com/storytold/designcraft), [FigCraft](https://github.com/storytold/figcraft), [FilmCraft](https://github.com/storytold/filmcraft), [EffectCraft](https://github.com/storytold/effectcraft), [SoundCraft](https://github.com/storytold/soundcraft), [WordCraft](https://github.com/storytold/wordcraft), [GridCraft](https://github.com/storytold/gridcraft), [DeckCraft](https://github.com/storytold/deckcraft), [PdfCraft](https://github.com/storytold/pdfcraft) y [CADCraft](https://github.com/storytold/cadcraft) son de [ArtCraft](https://getartcraft.com), con licencia Apache 2.0. Codex las muestra con su nombre, su icono oficial y su crédito, y las descarga de sus releases oficiales sin modificarlas.
- Tipografías [Inter](https://rsms.me/inter/) y [Space Mono](https://fonts.google.com/specimen/Space+Mono), con licencia SIL Open Font License.
- Ventana con [go-webview2](https://github.com/jchv/go-webview2) (MIT).

<div align="center">
<sub>Hecho por <a href="https://github.com/MiguelBigott">Miguel Bigott</a></sub>
</div>
