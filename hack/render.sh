#!/bin/sh
# Render hack/demo.cast -> hack/demo.gif
#
# Requirements:
#   brew install asciinema agg
#
# Re-record the demo (run from repo root):
#   asciinema rec --command "bash hack/demo_script.sh" --overwrite hack/demo.cast
#
# Then render:
#   hack/render.sh

set -e

cd "$(dirname "$0")/.."

agg \
  --font-size 14 \
  --cols 100 \
  --rows 28 \
  --speed 1.5 \
  hack/demo.cast \
  hack/demo.gif

echo "Rendered hack/demo.gif"
