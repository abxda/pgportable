#!/usr/bin/env bash
# Build de PgPortable para Linux amd64.
# Requiere:
#   - Go 1.22+
#   - gcc (build-essential en Debian/Ubuntu, base-devel en Arch)
#   - Wails dependencies: libgtk-3-dev libwebkit2gtk-4.1-dev (o 4.0 en distros antiguas)
#   - Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`
#
# Si tu distro solo tiene webkit2gtk-4.0 (Ubuntu 22.04), agrega:
#   `-tags webkit2_40` al wails build (algunas versiones de Wails lo necesitan).
#
# IMPORTANTE: este script genera SOLO el ejecutable Linux.
# Los binarios portables de PostgreSQL para Linux NO se incluyen aquí —
# ver AGENTS.md sección "Linux" para cómo armarlos.

set -euo pipefail

cd "$(dirname "$0")/.."

export CGO_ENABLED=1

echo "→ go mod tidy"
go mod tidy

echo "→ wails build (linux/amd64)"
wails build -platform linux/amd64 -trimpath

bin="build/bin/PgPortable"
[ -f "$bin" ] || { echo "Build failed: $bin missing"; exit 1; }
size=$(du -h "$bin" | cut -f1)
echo "✓ Built $bin · $size"
