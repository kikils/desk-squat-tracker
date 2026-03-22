#!/usr/bin/env bash
# Create/update helper/.venv with uv-managed CPython 3.12 (ignores asdf/.tool-versions quirks).
# Usage: bash scripts/sync_uv_venv.sh [--extra dev]
# Run from helper/ or via task install:python:deps / build_face_detect.sh.

set -euo pipefail

HELPER="$(cd "$(dirname "$0")/.." && pwd)"
cd "$HELPER"

uv python install 3.12
UV_PYTHON="$(uv python find 3.12)"
export UV_PROJECT_ENVIRONMENT="$HELPER/.venv"

if [ ! -x .venv/bin/python ]; then
  uv venv --python "$UV_PYTHON"
fi

uv sync --python "$UV_PYTHON" "$@"
