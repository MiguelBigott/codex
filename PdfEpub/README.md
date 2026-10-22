# PdfEpub

Aplicación de escritorio libre para convertir archivos PDF en libros EPUB. Está escrita en Go y usa pdf.js para leer los documentos. La interfaz permite revisar los datos del libro, diseñar la portada y ajustar los capítulos antes de exportar.

## Funciones

- **Texto fluido:** reconstruye párrafos y títulos para que el lector pueda adaptar la letra y el tamaño de página.
- **Páginas fijas:** conserva cada página como imagen, útil para cómics, revistas, libros ilustrados y documentos escaneados.
- **Portada:** usa una página del PDF, una imagen propia o un fondo de color; incluye plantillas de título, tipografías, encuadre, zoom y oscurecimiento.
- **Datos del libro:** título, autores, idioma, fecha, editorial, ISBN, serie, etiquetas y descripción.
- **Capítulos:** detecta el índice del PDF y los títulos; permite cambiar nombres, excluir secciones, unirlas y dividir desde la vista previa.
- **Contraseñas:** abre PDFs protegidos si se proporciona la clave.
- **Privacidad:** los archivos se procesan localmente. La aplicación no envía documentos a un servicio en línea.

Los PDF escaneados no incluyen texto que se pueda extraer. PdfEpub los ofrece como páginas fijas; no incorpora OCR. En el modo de texto fluido, el diseño original del PDF se reconstruye y puede requerir correcciones en la separación de párrafos o capítulos.

## Ejecutar

La aplicación está dirigida a Windows de 64 bits y requiere Microsoft Edge WebView2 Runtime. Windows 11 normalmente ya lo incluye.

Puedes usar `dist\PdfEpub.exe` de una compilación local o compilarlo desde el código fuente. El ejecutable se puede usar de forma portable. La compilación también genera un instalador para el usuario actual, con accesos directos y desinstalación desde Windows.

## Compilar en Windows

Requisitos: Go 1.27.3 o posterior, Python 3 con Pillow y `go-winres`.

```powershell
go install github.com/tc-hib/go-winres@latest
.\build\compilar.ps1
```

El ejecutable queda en `dist\PdfEpub.exe` y el instalador en `dist\PdfEpub-Setup-1.0.1.exe`. Para compilar solo el ejecutable portable, sin crear el instalador:

```powershell
.\build\compilar.ps1 -SoloExe
```

## Licencia

PdfEpub se distribuye bajo la licencia MIT. Las bibliotecas, datos de fuentes y recursos de terceros mantienen sus propias licencias; consulta [THIRD-PARTY-NOTICES.md](THIRD-PARTY-NOTICES.md) y el texto completo en [THIRD-PARTY-LICENSES.txt](THIRD-PARTY-LICENSES.txt).
