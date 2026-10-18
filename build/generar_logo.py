"""Logo de Codex: «Códice abierto» (un libro abierto con < y > en sus páginas).

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

FONDO, BORDE, AMBAR, AMBAR_OSC = "#11151b", "#262d37", "#ffb454", "#d98a2b"


def svg(chico=False):
    """Logo en viewBox 512. `chico`: libro más grande y trazos más gruesos para 16-32 px."""
    if chico:
        izq = "M250 160 C210 132 150 124 76 136 L76 384 C150 372 210 380 250 410 Z"
        der = "M262 160 C302 132 362 124 436 136 L436 384 C362 372 302 380 262 410 Z"
        signos = '<path d="M196 214 L136 270 L196 326"/><path d="M316 214 L376 270 L316 326"/>'
        trazo, borde = 40, 0
    else:
        izq = "M248 176 C212 150 160 142 100 152 L100 370 C160 360 212 368 248 394 Z"
        der = "M264 176 C300 150 352 142 412 152 L412 370 C352 360 300 368 264 394 Z"
        signos = '<path d="M200 222 L152 266 L200 310"/><path d="M312 222 L360 266 L312 310"/>'
        trazo, borde = 24, 10
    return f'''<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="512" height="512">
  <rect x="8" y="8" width="496" height="496" rx="112" fill="{FONDO}" stroke="{BORDE}" stroke-width="{borde}"/>
  <path d="{izq}" fill="{AMBAR}"/>
  <path d="{der}" fill="{AMBAR_OSC}"/>
  <g fill="none" stroke="{FONDO}" stroke-width="{trazo}" stroke-linecap="round" stroke-linejoin="round">{signos}</g>
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
