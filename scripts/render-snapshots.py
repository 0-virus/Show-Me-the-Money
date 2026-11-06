"""Render captured tcell cells for visual review (requires Pillow on review host)."""
import json
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
mono = ImageFont.truetype('C:/Windows/Fonts/consola.ttf', 18)
korean = ImageFont.truetype('C:/Windows/Fonts/malgun.ttf', 17)
cw, ch, margin = 11, 22, 24
for path in (ROOT / 'artifacts/screens').glob('*.json'):
    data = json.loads(path.read_text(encoding='utf-8'))
    image = Image.new('RGB', (data['Width']*cw+margin*2, data['Height']*ch+margin*2), '#08120f')
    draw = ImageDraw.Draw(image)
    for c in data['Cells']:
        x,y = margin+c['X']*cw, margin+c['Y']*ch
        fg,bg = c['FG'] & 0xffffff, c['BG'] & 0xffffff
        fg,bg = f'#{fg:06x}',f'#{bg:06x}'
        text = c['Text']
        wide = ord(text[0]) >= 0x2e80
        draw.rectangle((x,y,x+cw*(2 if wide else 1)-1,y+ch-1), fill=bg)
        if text == '▀':
            draw.rectangle((x,y,x+cw-1,y+ch//2-1),fill=fg)
        else:
            draw.text((x,y-1), text, font=korean if wide else mono, fill=fg)
    image.save(path.with_suffix('.png'))
    print(path.with_suffix('.png'))
