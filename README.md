<div align="center">

<img src="build/logo.png" alt="Logo de Codex" width="96">

# Codex

Un solo lugar para instalar, actualizar y abrir mis apps de Windows.

[![CI](https://github.com/MiguelBigott/codex/actions/workflows/ci.yml/badge.svg)](https://github.com/MiguelBigott/codex/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Windows](https://img.shields.io/badge/Windows-10%20·%2011-0078D4)](#dónde-se-instala)

</div>

## Qué hace

- Muestra las apps del catálogo en cuadrícula o tabla, con filtros (todas, instaladas,
  actualizaciones, no instaladas), orden y búsqueda.
- Detecta qué apps están instaladas y en qué versión leyendo el registro de Windows.
- Consulta la última versión de cada app en sus releases de GitHub, descarga el instalador,
  **verifica su SHA-256** y lo ejecuta como administrador (Windows muestra el aviso de UAC).
- Instala setups `.exe` (NSIS) y paquetes `.msi` (Windows Installer).
- Avisa cuando hay una versión nueva del propio launcher y se actualiza con un clic.

## Dónde se instala

Todo queda en `C:\Program Files\Codex`:

- `Codex.exe`: el launcher.
- Una subcarpeta por app: `C:\Program Files\Codex\<App>`.

La configuración vive en `%LOCALAPPDATA%\Codex`.

## Catálogo

Las apps se definen en [`catalogo.json`](catalogo.json):

| Campo | Para qué |
|---|---|
| `id` | Identificador único |
| `nombre` + `sufijo` | Se muestran juntos (`Photo` + `Craft`) |
| `descripcion`, `similar` | Las líneas bajo el nombre |
| `autor` | Crédito para apps de terceros («de ArtCraft») |
| `color` | Color de la app (icono de respaldo) |
| `icono` | URL `https` de su icono oficial (opcional) |
| `categoria` | Sección donde aparece |
| `repo` | Repositorio público de GitHub (`usuario/repo`) donde publica sus versiones |
| `web` | Sitio web (opcional) |
| `formato` | `nsis` (setup `.exe`, por defecto) o `msi` |
| `propiedadCarpeta` | Solo MSI: propiedad que fija la carpeta (p. ej. `INSTALLFOLDER`) |
| `nombreRegistro` | Nombre con el que se registra en Windows (por defecto `nombre`+`sufijo`) |
| `ejecutable` | `.exe` dentro de su carpeta (por defecto `nombre`+`sufijo`.exe) |
| `ejemplo` | `true` para mostrarla como «Próximamente» |

Cada release debe incluir el instalador para Windows x64 (un `.exe` con `setup` en el nombre, o
un `.msi` con `x64`) y su suma SHA-256: GitHub ya la publica para cada archivo, o puede ir en un
`SHA256SUMS.txt`. Sin suma, el launcher no instala.

### Apps de terceros

[PhotoCraft](https://github.com/storytold/photocraft) es de [ArtCraft](https://getartcraft.com),
con licencia Apache 2.0. Codex lo muestra con su nombre, su icono oficial y su crédito, y lo
descarga de sus releases oficiales sin modificarlo.

## Compilar

Requisitos: [Go](https://go.dev/dl), [NSIS](https://nsis.sourceforge.io) y go-winres
(`go install github.com/tc-hib/go-winres@latest`).

```powershell
.\build\compilar.ps1
```

Genera `dist\Codex.exe` y `dist\Codex-Setup-<versión>.exe`. La versión se toma de
`const Version` en `main.go`.

El logo («Códice abierto») se define en `build/generar_logo.py`, que escribe `web/logo.svg`,
`build/logo.png`, `build/app.ico` (con una versión simplificada para 16-32 px) y
`build/lateral.bmp` (requiere Python con Pillow y Microsoft Edge):

```powershell
python build\generar_logo.py
```

Para revisar la interfaz en un navegador, sin ventana, ejecuta el launcher con la variable
`CODEX_NAVEGADOR=1`: escribe la dirección en `%LOCALAPPDATA%\Codex\url-desarrollo.txt`.

## Cómo está hecho

| Archivo | Contenido |
|---|---|
| `main.go` | Arranque, instancia única y ventana |
| `ventana.go` | Ventana nativa sin barra de título (el encabezado de la interfaz hace de barra) |
| `servidor.go` | Servidor local (solo `127.0.0.1`, con clave por sesión) que usa la interfaz |
| `launcher.go` | Estado, revisión de versiones, descargas verificadas e instalaciones (NSIS y MSI) |
| `releases.go` | Consulta a la API de GitHub y sumas SHA-256 |
| `instalados.go` | Apps instaladas según el registro de Windows |
| `elevar.go` | Ejecución como administrador (UAC) |
| `web/` | Interfaz (HTML, CSS y JS) con las tipografías Inter y Space Mono (OFL) |
| `build/` | Logo, instalador NSIS y script de compilación |
