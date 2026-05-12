# AGENTS.md — Guía prescriptiva para agentes IA que completen Linux y macOS

> Estado actual:
> - ✅ **Windows** — terminado y validado.
> - ✅ **Linux (Ubuntu 24.04 amd64)** — terminado y validado. Ver §1.
> - ⏳ **macOS** — pendiente. Ver §2 y aprovecha las lecciones aprendidas en §1.5.
>
> Lee este archivo de arriba a abajo y ejecuta cada bloque. No hay decisiones de diseño pendientes — todo está acotado.

---

## 0. Lo que ya está hecho (no tocar)

- ✅ Stack Go + Wails v2 + WebView2/WebKit + HTML/CSS/JS vanilla.
- ✅ Lógica `Manager` (initdb, start, stop, port-finder, cleanup pid huérfano) en `db_manager.go`.
- ✅ Build tags correctos: `platform_windows.go` (Win32 syscalls) y `platform_unix.go` (signal(0) + stderr fallback).
- ✅ Explorador con `pgx/v5` (puro Go, funciona igual en Linux/Mac sin cambios).
- ✅ Single-instance lock (`pgportable.lock` + PID + `processAlive`).
- ✅ Frontend totalmente cross-platform (HTML/CSS/JS).

**No modifiques** ningún archivo `.go`, HTML, CSS o JS para "preparar" Linux/Mac. Solo hace falta:
1. Compilar el binario en cada OS.
2. Conseguir los binarios portables de PostgreSQL para ese OS.
3. Armar el folder `deploy/` equivalente.

---

## 1. Linux (Ubuntu/Debian/Fedora/Arch — amd64) — ✅ COMPLETO

> Validado en Ubuntu 24.04 LTS. Pasos automatizados en `scripts/build-linux.sh`
> y `scripts/pgsql-portable-linux.sh`. Si tu distro es similar, todo funciona
> sin tocar código. Si vas a otra distro, lee también §1.5 (lecciones).

### 1.1 Toolchain
```bash
# Debian / Ubuntu (24.04 y derivadas)
sudo apt update
sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev

# Ubuntu 22.04 y anteriores: usa webkit2gtk-4.0
sudo apt install -y libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install -y gcc gtk3-devel webkit2gtk4.1-devel

# Arch
sudo pacman -S --needed base-devel gtk3 webkit2gtk
```

`scripts/build-linux.sh` autodetecta `webkit2gtk-4.0` vs `4.1` y aplica el
build tag correcto (`-tags webkit2_41` cuando solo está 4.1).

### 1.2 Go + Wails CLI
```bash
# Go 1.22+
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH="$HOME/go/bin:$PATH"
wails version    # v2.12.0
# Nota: `wails doctor` puede reportar "libwebkit Not Found" en Ubuntu 24.04;
# el chequeo busca el nombre 4.0. Es un falso negativo — el build funciona igual.
```

### 1.3 Compilar el binario
```bash
cd ~/pgportable
bash scripts/build-linux.sh
# → build/bin/PgPortable (≈13 MB)
```

### 1.4 Armar pgsql/ portable de PostgreSQL 17

Hay un script reproducible que descarga el `.deb` oficial, extrae los binarios
y los reorganiza al layout correcto:

```bash
bash scripts/pgsql-portable-linux.sh
# → ./pgsql/  (≈47 MB)
```

Por defecto baja PostgreSQL 17.9 para `pgdg24.04`. Para otra versión:
`bash scripts/pgsql-portable-linux.sh 17.9 22.04`.

### 1.5 Lecciones aprendidas (importante para macOS y otras distros)

Estos son los **gotchas** que se descubrieron al portar a Linux. La mayoría
también aplican a macOS si usas binarios estilo Homebrew/Postgres.app.

1. **Layout nativo, no plano.** Los binarios de Postgres usan `find_my_exec()`
   para resolver `share/`. Si están compilados con `--prefix=/usr` (Debian/
   Homebrew), buscan share en `<bindir>/../../share/postgresql/17`. Por eso
   el árbol final NO es plano (`pgsql/{bin,lib,share}`) sino:
   ```
   pgsql/
     bin/                     ← wrappers shell que el código Go invoca
     lib/postgresql/17/bin/   ← binarios reales
     lib/postgresql/17/lib/   ← extensiones .so + libpq
     share/postgresql/17/     ← share data
   ```
   Los wrappers en `pgsql/bin/` hacen `exec` al binario real → reemplazan el
   proceso → `find_my_exec` ve la ubicación real y resuelve share/ correcto.
   Sin esto, `initdb` falla con `"/usr/share/postgresql/17/postgres.bki does
   not exist"`.

2. **`PGSHAREDIR` NO funciona como esperarías.** `initdb` regenera el
   `postgresql.conf` con los defaults compilados en el binario, ignorando la
   env var. La única salida es el layout nativo del punto 1, o pasar `-L
   sharedir` manualmente (no escalable).

3. **`unix_socket_directories` por defecto = `/var/run/postgresql`.** Los
   `.deb` Debian/Ubuntu compilan ese path como default. La carpeta solo
   existe instalando el paquete oficial postgresql y no es escribible sin
   sudo. **Fix ya aplicado**: `db_manager.go::applyConfig()` añade
   `unix_socket_directories = '/tmp'` cuando `runtime.GOOS == "linux"`.
   En macOS los binarios de Postgres.app/Homebrew compilan default `/tmp`,
   así que el fix Linux-only no aplica — pero verifica con `postgres -C
   unix_socket_directories` antes de descartar.

4. **`libpq.so.5` del sistema es vieja.** Ubuntu 24.04 trae libpq 16.x, sin
   `PQchangePassword`. Si los binarios de `psql` 17 cargan esa libpq, fallan
   con `undefined symbol`. Solución: incluir `libpq.so.5` 17+ en
   `pgsql/lib/postgresql/17/lib/` y aplicar `rpath = '$ORIGIN/../lib'` en los
   binarios reales con `patchelf`. El script ya lo hace.

5. **No `LD_LIBRARY_PATH` ni `DYLD_LIBRARY_PATH` en Go.** El código
   `db_manager.go` no inyecta esas vars. Toda la portabilidad depende del
   `rpath` correcto en los binarios. Si en macOS necesitas ajustar rutas
   dylib, usa `install_name_tool -change` o `install_name_tool -add_rpath`.

6. **Cambio chiquito al código Go (justificado).** Solo se añadió el `if
   runtime.GOOS == "linux"` mencionado en el punto 3. Cero impacto en
   Windows o macOS. Si necesitas algo análogo en Darwin, agrega otro `if`
   bajo el mismo patrón.

### 1.6 Armar deploy/
```bash
mkdir deploy
cp build/bin/PgPortable deploy/PgPortable
cp -r pgsql              deploy/pgsql
cp LEEME.txt             deploy/LEEME.txt   # versión Linux (ver §3)
chmod +x deploy/PgPortable
```

### 1.7 Validación
```bash
cd deploy
./PgPortable --autostart &
sleep 5
PGPASSWORD=alumno psql -h localhost -p 5432 -U alumno -d postgres -c "SELECT version();"
# debe responder PostgreSQL 17.x on x86_64-pc-linux-gnu
pkill -TERM -f PgPortable    # cierra GUI → manager detiene postgres limpio
```

Single-instance: ejecutar dos veces seguidas; el segundo debe imprimir
`"Ya está corriendo otra instancia... PID NNN"` y salir.

---

## 2. macOS (arm64 Apple Silicon / amd64 Intel) — ⏳ pendiente

> Antes de empezar lee §1.5 (Lecciones aprendidas en Linux). Los puntos 1, 2,
> 4 y 5 aplican casi 1:1 a macOS. Punto 3 (`unix_socket_directories`) NO
> aplica si usas Postgres.app pero SÍ podría aplicar con Homebrew.

### 2.1 Toolchain
```bash
xcode-select --install   # si no está
brew install go          # o descarga de go.dev
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH="$HOME/go/bin:$PATH"
wails doctor
```

### 2.2 Compilar
```bash
cd ~/pgportable
chmod +x scripts/build-darwin.sh
./scripts/build-darwin.sh arm64    # Apple Silicon
# ó:
./scripts/build-darwin.sh amd64    # Intel
# → build/bin/PgPortable.app
```

### 2.3 Binarios portables de PostgreSQL 17 para macOS

EDB tampoco da tarballs portables para macOS. Las rutas reales:

**Ruta A — Homebrew (más fácil)**
```bash
brew install postgresql@17
# brew lo instala en /opt/homebrew/opt/postgresql@17/ (arm64) o /usr/local/opt/...
# Copia bin/, lib/, share/ a tu pgsql/:
PG=/opt/homebrew/opt/postgresql@17
mkdir -p pgsql/{bin,lib,share}
cp -R "$PG/bin/"   pgsql/bin/
cp -R "$PG/lib/"   pgsql/lib/
cp -R "$PG/share/postgresql@17/" pgsql/share/

# Las dylibs de Homebrew tienen rpaths absolutos a /opt/homebrew. Para portabilidad:
install_name_tool -change ...   # complicado
# Más sencillo: documentar que el estudiante necesite el mismo prefix /opt/homebrew/opt/postgresql@17
# O usar postgres.app (universal binary, ver Ruta B).
```

**Ruta B — Postgres.app (universal, mejor opción)**
- Descargar Postgres.app desde https://postgresapp.com/
- Dentro del .app: `Contents/Versions/17/` contiene bin/ lib/ share/
- Es universal binary (arm64 + amd64), portable, con `@loader_path` en los
  rpaths internos → no necesitas `install_name_tool` ni patchelf.
- **Layout sugerido (mismo principio que Linux, ver §1.5 punto 1):**
  ```
  pgsql/
    bin/                       ← wrappers shell exec a binarios reales
    lib/postgresql/17/bin/
    lib/postgresql/17/lib/
    share/postgresql/17/
  ```
  Si Postgres.app ya trae `share/postgresql/17/postgres.bki` en una ruta
  consistente con sus binarios, puedes simplificar el layout — verifica con:
  ```bash
  /Applications/Postgres.app/Contents/Versions/17/bin/initdb -D /tmp/test
  ```
  Si funciona desde su ubicación original, copiar el árbol tal cual a `pgsql/`
  debería bastar. Si falla, replica el layout nativo Debian/Linux.

- Verifica también `unix_socket_directories`:
  ```bash
  /Applications/Postgres.app/Contents/Versions/17/bin/postgres -C unix_socket_directories
  ```
  Si responde `/tmp` no necesitas tocar nada. Si responde una ruta absoluta
  que no es escribible (poco probable en Postgres.app), añade un `if
  runtime.GOOS == "darwin"` en `db_manager.go::applyConfig()` siguiendo el
  patrón Linux.

### 2.4 Armar deploy/
```bash
mkdir deploy
cp -r build/bin/PgPortable.app deploy/PgPortable.app
cp -r pgsql                    deploy/pgsql
# IMPORTANT: para macOS conviene meter pgsql/ DENTRO del .app:
mv deploy/pgsql deploy/PgPortable.app/Contents/Resources/pgsql
```

⚠️ Si pones `pgsql/` dentro del `.app/Contents/Resources/`, ajusta `app.go` para detectar el resource path en macOS. Como esto es una decisión de empaquetado que afecta al código Go, **antes de moverlo**, lee el comentario en `app.go` `func NewApp` y considera dejar `pgsql/` al lado del `.app` en su lugar (más simple, el alumno descomprime y obtiene `PgPortable.app` + `pgsql/` hermanos).

### 2.5 Code signing & notarization
Sin firmar, macOS Gatekeeper bloquea apps descargadas. Para estudiantes:
- Si es para uso local, instruir: clic derecho → Abrir → confirmar.
- Para distribución amplia: necesitas Apple Developer Account ($99/año) y firmar:
  ```bash
  codesign --deep --force --sign "Developer ID Application: <tu nombre>" deploy/PgPortable.app
  ```

---

## 3. LEEME.txt por plataforma

`deploy/LEEME.txt` (la versión Windows) está en español y se adapta así:

- **Linux**: cambia "Doble clic en PgPortable.exe" por "Ejecuta `./PgPortable` desde la terminal o crea un .desktop". Requisitos: glibc 2.31+, libwebkit2gtk.
- **macOS**: "Doble clic en PgPortable.app". Requisitos: macOS 11+ (Big Sur).

---

## 4. Checklist final por plataforma

Antes de marcar el plataforma como listo, valida:

- [ ] `./PgPortable[.exe|.app]` arranca y muestra GUI.
- [ ] Botón "Iniciar entorno" ejecuta initdb la primera vez y arranca postgres.
- [ ] Conexión con psql / DBeaver responde en el puerto que muestra la app.
- [ ] Si el puerto 5432 está ocupado, la app elige 5433 automáticamente.
- [ ] Cerrar la ventana detiene postgres limpiamente.
- [ ] Re-abrir la app mientras corre detecta postgres y no levanta otro.
- [ ] Ejecutar el .exe/binario DOS veces muestra el mensaje de single-instance.
- [ ] El árbol explorador lista la BD `postgres` y schema `public`.

---

## 5. PRs

- Branch: `feat/linux-deploy` o `feat/darwin-deploy`.
- Incluye: scripts ajustados, `deploy/` adjunto NO (es muy grande para git), pero documenta cómo armarlo.
- Si tocas código Go, justifica por qué — el código Windows está validado y debería ser cross-platform sin cambios.
