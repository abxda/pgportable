#!/usr/bin/env bash
# Build de PgPortable para macOS (arm64 / Apple Silicon o amd64 / Intel).
# Requiere:
#   - Go 1.22+
#   - Xcode Command Line Tools (xcode-select --install)
#   - Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
#
# IMPORTANTE: este script genera SOLO el .app de macOS.
# Los binarios portables de PostgreSQL para Darwin NO se incluyen aquí —
# ver AGENTS.md sección "macOS" para cómo armarlos.

set -euo pipefail

cd "$(dirname "$0")/.."

ARCH="${1:-arm64}"   # uso: scripts/build-darwin.sh [arm64|amd64]

export CGO_ENABLED=1

echo "→ go mod tidy"
go mod tidy

echo "→ wails build (darwin/${ARCH})"
wails build -platform "darwin/${ARCH}" -trimpath

app="build/bin/PgPortable.app"
[ -d "$app" ] || { echo "Build failed: $app missing"; exit 1; }
echo "✓ Built $app"
