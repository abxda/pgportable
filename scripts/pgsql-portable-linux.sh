#!/usr/bin/env bash
# Arma un árbol pgsql/ portable de PostgreSQL 17 para Linux amd64.
#
# El layout final es nativo de Debian/Ubuntu (find_my_exec lo entiende):
#   pgsql/
#     bin/                          ← wrappers shell que el código Go invoca
#     lib/postgresql/17/bin/        ← binarios reales (postgres, initdb, ...)
#     lib/postgresql/17/lib/        ← extensiones .so + libpq.so.5
#     share/postgresql/17/          ← postgres.bki, schemas, samples, ...
#
# Por qué este layout:
#  - Los binarios del .deb tienen find_my_exec compilado para encontrar share/
#    en `<bindir>/../../share/postgresql/17`. Si los movemos a `pgsql/bin/`
#    pelado, no encuentran share/ y `initdb` falla con
#    "/usr/share/postgresql/17/postgres.bki does not exist".
#  - Los wrappers en pgsql/bin/ exec'an al binario real, lo que reemplaza el
#    proceso → find_my_exec ve la ubicación real y resuelve share/ relativo.
#
# Otras decisiones importantes:
#  - rpath '$ORIGIN/../lib' en los binarios reales para que carguen libpq.so.5
#    (la 18.x con PQchangePassword) sin LD_LIBRARY_PATH. La libpq vieja del
#    sistema (Ubuntu 24.04 trae 16.x) hace fallar `psql` con undefined symbol.
#  - unix_socket_directories: los .deb compilan default '/var/run/postgresql'
#    que no existe sin instalar el paquete oficial. db_manager.go ya lo
#    sobrescribe a '/tmp' en runtime.GOOS=="linux".
#
# Requisitos:
#  - dpkg-deb (presente en Debian/Ubuntu/derivadas)
#  - wget
#  - patchelf (se descarga binario portable si no está)
#
# Uso:
#  cd ~/pgportable
#  bash scripts/pgsql-portable-linux.sh [PG_VERSION] [UBUNTU_CODENAME]
#  # default: PG_VERSION=17.9, UBUNTU_CODENAME=24.04
#
# El script deja el árbol en ./pgsql/ del directorio actual.

set -euo pipefail

PG_VERSION="${1:-17.9}"
UBUNTU_VER="${2:-24.04}"
PG_MAJOR="${PG_VERSION%%.*}"
LIBPQ_VERSION="18.3"   # libpq con PQchangePassword (cualquier ≥17 sirve)

WORK="$(mktemp -d -t pgportable-build-XXXXXX)"
trap 'rm -rf "$WORK"' EXIT

echo "==> Trabajando en $WORK"
cd "$WORK"

# ── 1. patchelf portable ──────────────────────────────────────────────
if command -v patchelf >/dev/null 2>&1; then
  PATCHELF="$(command -v patchelf)"
else
  echo "==> Descargando patchelf portable"
  wget -q https://github.com/NixOS/patchelf/releases/download/0.18.0/patchelf-0.18.0-x86_64.tar.gz
  mkdir patchelf_dl && tar xzf patchelf-0.18.0-x86_64.tar.gz -C patchelf_dl
  PATCHELF="$WORK/patchelf_dl/bin/patchelf"
fi
$PATCHELF --version

# ── 2. Descargar .deb oficiales ───────────────────────────────────────
PGDG=https://apt.postgresql.org/pub/repos/apt/pool/main
SERVER_DEB="postgresql-${PG_MAJOR}_${PG_VERSION}-1.pgdg${UBUNTU_VER}+1_amd64.deb"
CLIENT_DEB="postgresql-client-${PG_MAJOR}_${PG_VERSION}-1.pgdg${UBUNTU_VER}+1_amd64.deb"
LIBPQ_DEB="libpq5_${LIBPQ_VERSION}-1.pgdg${UBUNTU_VER}+1_amd64.deb"

echo "==> Descargando $SERVER_DEB"
wget -q --show-progress "$PGDG/p/postgresql-${PG_MAJOR}/$SERVER_DEB"
echo "==> Descargando $CLIENT_DEB"
wget -q --show-progress "$PGDG/p/postgresql-${PG_MAJOR}/$CLIENT_DEB"
echo "==> Descargando $LIBPQ_DEB"
wget -q --show-progress "$PGDG/p/postgresql-18/$LIBPQ_DEB"

# ── 3. Extraer ────────────────────────────────────────────────────────
mkdir ex_server ex_client ex_libpq
dpkg-deb -x "$SERVER_DEB" ex_server
dpkg-deb -x "$CLIENT_DEB" ex_client
dpkg-deb -x "$LIBPQ_DEB"  ex_libpq

# ── 4. Layout final ───────────────────────────────────────────────────
OUT="$WORK/pgsql"
mkdir -p "$OUT"/{bin,lib/postgresql/${PG_MAJOR}/bin,lib/postgresql/${PG_MAJOR}/lib,share/postgresql/${PG_MAJOR}}

cp ex_server/usr/lib/postgresql/${PG_MAJOR}/bin/* "$OUT/lib/postgresql/${PG_MAJOR}/bin/"
cp ex_client/usr/lib/postgresql/${PG_MAJOR}/bin/* "$OUT/lib/postgresql/${PG_MAJOR}/bin/"
cp -r ex_server/usr/lib/postgresql/${PG_MAJOR}/lib/* "$OUT/lib/postgresql/${PG_MAJOR}/lib/"
cp -r ex_server/usr/share/postgresql/${PG_MAJOR}/* "$OUT/share/postgresql/${PG_MAJOR}/"

# libpq al lado de los binarios reales (resolverá vía rpath $ORIGIN/../lib)
cp ex_libpq/usr/lib/x86_64-linux-gnu/libpq.so.5.* "$OUT/lib/postgresql/${PG_MAJOR}/lib/"
ln -sf libpq.so.5.${LIBPQ_VERSION%%.*} "$OUT/lib/postgresql/${PG_MAJOR}/lib/libpq.so.5"

# ── 5. rpath en binarios reales ───────────────────────────────────────
echo "==> Aplicando rpath \$ORIGIN/../lib"
for b in "$OUT/lib/postgresql/${PG_MAJOR}/bin/"*; do
  if file "$b" | grep -q "ELF.*executable"; then
    $PATCHELF --set-rpath '$ORIGIN/../lib' "$b" || echo "WARN: rpath fallo en $b"
  fi
done

# ── 6. Wrappers en pgsql/bin/ ─────────────────────────────────────────
for name in initdb pg_ctl postgres psql; do
  cat > "$OUT/bin/$name" <<EOF
#!/bin/sh
SELF="\$(readlink -f "\$0")"
DIR="\$(dirname "\$SELF")"
exec "\$DIR/../lib/postgresql/${PG_MAJOR}/bin/$name" "\$@"
EOF
  chmod +x "$OUT/bin/$name"
done

# ── 7. Smoke test ─────────────────────────────────────────────────────
echo "==> Smoke test"
"$OUT/bin/initdb" --version
"$OUT/bin/postgres" --version
"$OUT/bin/psql" --version

# ── 8. Mover al cwd del usuario ───────────────────────────────────────
TARGET="$OLDPWD/pgsql"
if [ -e "$TARGET" ]; then
  echo "ERROR: $TARGET ya existe. Bórralo si quieres rehacerlo." >&2
  exit 1
fi
mv "$OUT" "$TARGET"
echo "✓ pgsql/ portable listo en $TARGET ($(du -sh "$TARGET" | cut -f1))"
