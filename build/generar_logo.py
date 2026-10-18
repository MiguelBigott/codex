"""Logo de Codex: «Códice cerrado», en el estilo del Copista (pergamino sobre tinta, oro y rojo de rúbrica).

    python build/generar_logo.py

Escribe:
  web/logo.svg       logo para el encabezado del launcher
  build/logo.png     logo para el README (256 px)
  build/app.ico      icono del .exe: versión simplificada en 16-32 px, normal desde 36 px
  build/lateral.bmp  imagen lateral del instalador (al doble de tamaño, para pantallas con escala)
Requiere Pillow y Microsoft Edge (renderiza el SVG).
"""
import os
import subprocess
import tempfile

from PIL import Image

BUILD = os.path.dirname(os.path.abspath(__file__))
RAIZ = os.path.dirname(BUILD)
EDGE = r"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"

TINTA, BORDE, PERGAMINO, ORO, RUBRICA = "#16120e", "#3a3128", "#ece2cf", "#c9a45c", "#c8483a"


def svg(chico=False):
    """Logo en viewBox 512: el códice cerrado (tapa pergamino, lomo recortado, banda y broche
    de oro, marcapáginas rojo). `chico`: libro más grande y detalles más gruesos para 16-32 px."""
    if chico:
        tapa = '<rect x="118" y="70" width="252" height="342" rx="18" fill="{p}"/>'
        lomo = '<path d="M164 70 V412" stroke="{t}" stroke-width="16"/>'
        banda = '<path d="M164 138 H370" stroke="{o}" stroke-width="24"/>'
        broche = '<rect x="330" y="200" width="82" height="76" rx="12" fill="{o}"/><circle cx="371" cy="238" r="13" fill="{t}"/>'
        marca = '<path d="M290 412 V472 L304 458 L318 472 V412 Z" fill="{r}"/>'
        borde = 0
    else:
        tapa = '<rect x="146" y="92" width="220" height="300" rx="16" fill="{p}"/>'
        lomo = '<path d="M184 92 V392" stroke="{t}" stroke-width="10"/>'
        banda = '<path d="M184 150 H366" stroke="{o}" stroke-width="14"/>'
        broche = '<rect x="330" y="214" width="62" height="56" rx="10" fill="{o}"/><circle cx="361" cy="242" r="9" fill="{t}"/>'
        marca = '<path d="M296 392 V448 L306 438 L316 448 V392 Z" fill="{r}"/>'
        borde = 10
    partes = "\n  ".join(x.format(p=PERGAMINO, t=TINTA, o=ORO, r=RUBRICA) for x in (tapa, lomo, banda, broche, marca))
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="512" height="512">
  <rect x="8" y="8" width="496" height="496" rx="112" fill="{TINTA}" stroke="{BORDE}" stroke-width="{borde}"/>
  {partes}
</svg>
'''


def renderizar(contenido_html, lado_x, lado_y, destino):
    with tempfile.TemporaryDirectory() as perfil:
        html = os.path.join(perfil, "p.html")
        with open(html, "w", encoding="utf-8") as f:
            f.write(contenido_html)
        subprocess.run([EDGE, "--headless", "--disable-gpu", "--hide-scrollbars", "--allow-file-access-from-files",
                        f"--user-data-dir={perfil}", "--default-background-color=00000000",
                        f"--window-size={lado_x},{lado_y}", "--virtual-time-budget=3000",
                        f"--screenshot={destino}", "file:///" + html.replace(os.sep, "/")],
                       check=True, capture_output=True, timeout=60)


def png_de_svg(texto_svg, lado, destino):
    with tempfile.TemporaryDirectory() as tmp:
        ruta = os.path.join(tmp, "logo.svg")
        with open(ruta, "w", encoding="utf-8") as f:
            f.write(texto_svg)
        renderizar(f'<html><body style="margin:0;background:transparent"><img src="file:///{ruta.replace(os.sep, "/")}" '
                   f'width="{lado}" height="{lado}" style="display:block"></body></html>', lado, lado, destino)
    im = Image.open(destino).convert("RGBA")
    im.load()
    return im


def lateral(logo_png, destino):
    fuente = os.path.join(RAIZ, "web", "fonts", "space-mono-latin-400-normal.woff2").replace(os.sep, "/")
    with tempfile.TemporaryDirectory() as tmp:
        png = os.path.join(tmp, "lateral.png")
        renderizar(f'''<html><head><style>
@font-face {{ font-family: Mono; src: url("file:///{fuente}"); }}
html, body {{ margin: 0; width: 328px; height: 628px; overflow: hidden; background: #0d0f12; }}
body {{ display: flex; flex-direction: column; align-items: center; font-family: Mono, monospace; color: #d9dfe6; }}
img {{ width: 200px; height: 200px; margin-top: 150px; }}
h1 {{ margin: 34px 0 0; font-weight: 400; font-size: 34px; }}
h1 b {{ color: #ffb454; font-weight: 400; }}
</style></head><body><img src="file:///{logo_png.replace(os.sep, "/")}"><h1><b>~/</b>codex<b>_</b></h1></body></html>''', 328, 628, png)
        Image.open(png).convert("RGB").save(destino, format="BMP")


def main():
    with open(os.path.join(RAIZ, "web", "logo.svg"), "w", encoding="utf-8", newline="\n") as f:
        f.write(svg())
    with tempfile.TemporaryDirectory() as tmp:
        normal = png_de_svg(svg(), 1024, os.path.join(tmp, "normal.png"))
        chico = png_de_svg(svg(chico=True), 1024, os.path.join(tmp, "chico.png"))
        normal.resize((256, 256), Image.LANCZOS).save(os.path.join(BUILD, "logo.png"))
        tam = [16, 20, 24, 30, 32, 36, 40, 48, 64, 128, 256]
        cuadros = [(chico if s <= 32 else normal).resize((s, s), Image.LANCZOS) for s in tam]
        cuadros[-1].save(os.path.join(BUILD, "app.ico"), format="ICO",
                         sizes=[(s, s) for s in tam], append_images=cuadros[:-1])
        logo_lat = os.path.join(tmp, "lat.png")
        normal.resize((400, 400), Image.LANCZOS).save(logo_lat)
        lateral(logo_lat, os.path.join(BUILD, "lateral.bmp"))
    print("logo generado")


if __name__ == "__main__":
    main()
