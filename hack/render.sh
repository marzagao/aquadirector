#!/bin/sh
# Render hack/demo.cast -> hack/demo.svg
#
# Requirements:
#   brew install asciinema
#   npm install -g svg-term-cli
#
# Re-record the demo (run from repo root):
#   asciinema rec --window-size 100x28 --command "bash hack/demo_script.sh" \
#     --overwrite hack/demo.cast
#
# The --window-size must match the svg-term --width/--height below or the
# recording will wrap/truncate before svg-term sees it.
#
# Then render:
#   hack/render.sh

set -e

cd "$(dirname "$0")/.."

# Convert v3 -> v2 (svg-term requires v2) and slow down output events by
# inserting a small delay between each character burst so the animation
# is readable. Pauses (gaps > 0.3s) are preserved as-is.
python3 -c "
import json, sys

lines = open('hack/demo.cast').readlines()
h = json.loads(lines[0])
v2 = {'version': 2, 'width': h['term']['cols'], 'height': h['term']['rows'],
      'timestamp': h.get('timestamp'), 'command': h.get('command', '')}

events = [json.loads(l) for l in lines[1:] if l.strip()]

# v3 timestamps are relative (delta from previous event); convert to absolute.
abs_events = []
t = 0.0
for delta, kind, data in events:
    t += delta
    abs_events.append((t, kind, data))

# Shift timeline so first event starts at t=0 (eliminates blank frame
# at the start of the loop).
if abs_events:
    offset = abs_events[0][0]
    abs_events = [(t - offset, kind, data) for t, kind, data in abs_events]

sys.stdout.write(json.dumps(v2) + '\n')
for abs_t, kind, data in abs_events:
    sys.stdout.write(json.dumps([round(abs_t, 4), kind, data]) + '\n')
" | svg-term --window --width 100 --height 28 --padding-x 20 --padding-y 0 | python3 -c "
import sys, re

svg = sys.stdin.read()

# svg-term quirks we fix here:
#   1. The last row's text baseline (line.y + fontSize) overshoots the
#      Document viewBox by ~fontSize, clipping the last line. Extend the
#      Document viewBox and pixel height by one row to include it.
#   2. --padding-y applies to top and bottom equally; we want zero padding
#      at the top (cursor snug under the window chrome) but generous
#      breathing room at the bottom. Grow only the outer Window height.

EXTRA_BOTTOM_PX = 30   # extra bottom padding in pixels

m = re.search(r'<svg ([^>]*)viewBox=\"0 0 (\d+(?:\.\d+)?) (\d+(?:\.\d+)?)\"', svg)
if m:
    vb_w = float(m.group(2))
    vb_h = float(m.group(3))
    rows = 28
    row_vb = vb_h / rows

    hm = re.search(r'<svg [^>]*height=\"(\d+(?:\.\d+)?)\"[^>]*viewBox=', svg)
    doc_px_h = float(hm.group(1))
    row_px = doc_px_h / rows

    new_vb_h = vb_h + row_vb
    new_px_h = doc_px_h + row_px

    svg = svg.replace(
        f'viewBox=\"0 0 {m.group(2)} {m.group(3)}\"',
        f'viewBox=\"0 0 {m.group(2)} {new_vb_h:.3f}\"',
        1,
    )
    svg = svg.replace(
        f'height=\"{hm.group(1)}\"',
        f'height=\"{new_px_h:.3f}\"',
        1,
    )
    om = re.match(r'(<svg [^>]*height=\")(\d+(?:\.\d+)?)(\")', svg)
    if om:
        old_win_h = float(om.group(2))
        new_win_h = old_win_h + row_px + EXTRA_BOTTOM_PX
        svg = svg.replace(
            f'{om.group(1)}{om.group(2)}{om.group(3)}',
            f'{om.group(1)}{new_win_h:.3f}{om.group(3)}',
            1,
        )
        # Expand the background rect to match the new Window height,
        # otherwise the dark terminal background stops at the old size.
        svg = re.sub(
            r'(<rect[^>]*key=\"bg\"[^>]*height=\")(\d+(?:\.\d+)?)(\")',
            lambda mm: f'{mm.group(1)}{new_win_h:.3f}{mm.group(3)}',
            svg, count=1,
        )
        svg = re.sub(
            r'(<rect width=\"\d+(?:\.\d+)?\" height=\")(\d+(?:\.\d+)?)(\"[^>]*rx=)',
            lambda mm: f'{mm.group(1)}{new_win_h:.3f}{mm.group(3)}',
            svg, count=1,
        )

sys.stdout.write(svg)
" > hack/demo.svg

echo "Rendered hack/demo.svg"
