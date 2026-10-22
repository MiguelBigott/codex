"""Genera el icono de PdfEpub para Windows (Pillow)."""
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "build" / "app.ico"
S = 512
im = Image.new("RGBA", (S, S), (0, 0, 0, 0))
d = ImageDraw.Draw(im)
d.rounded_rectangle((16, 16, 496, 496), radius=104, fill="#24382d", outline="#405646", width=7)
# Tapa y lomo del códice.
d.polygon([(112, 78), (322, 78), (402, 150), (402, 438), (112, 438)], fill="#f4efe4")
d.polygon([(322, 78), (402, 150), (322, 150)], fill="#d8ccb2")
d.rectangle((112, 78, 143, 438), fill="#c9a45c")
# Líneas de página y título.
for y, x2 in [(218, 346), (264, 346), (310, 346), (356, 306)]:
    d.rounded_rectangle((183, y, x2, y + 13), radius=6, fill="#405646")
# Cinta de lectura.
d.polygon([(350, 344), (393, 344), (393, 463), (371, 444), (350, 463)], fill="#c9a45c")
im.save(OUT, format="ICO", sizes=[(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)])
print(OUT)
