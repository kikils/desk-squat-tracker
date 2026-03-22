#!/usr/bin/env bash
# Build helper/face_detect_mediapipe.py → helper/dist/face_detect (Nuitka onefile).
# Invoke: (cd helper && bash ../build/scripts/build_face_detect.sh) or task build:python:binary.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
HELPER="$PROJECT_ROOT/helper"
OUT_DIR="$HELPER/dist"

cd "$HELPER"
bash scripts/sync_uv_venv.sh --extra dev

MP_MOD="$(./.venv/bin/python -c "import mediapipe, pathlib; print(pathlib.Path(mediapipe.__file__).resolve().parent / 'modules')")"

mkdir -p "$OUT_DIR"
./.venv/bin/python -m nuitka --mode=onefile --follow-imports \
  --nofollow-import-to=tkinter \
  --nofollow-import-to=_tkinter \
  --include-data-dir="$MP_MOD=mediapipe/modules" \
  --output-dir="$OUT_DIR" \
  --output-filename=face_detect \
  --assume-yes-for-downloads \
  face_detect_mediapipe.py
