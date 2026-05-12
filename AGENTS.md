# AGENTS.md — Guía prescriptiva para agentes IA que completen Linux y macOS

> Este proyecto está **terminado y validado en Windows**. Linux y macOS quedan como scaffolding listo para que un agente (Claude / Codex / Cursor / etc.) los complete en pocos minutos en sus respectivas máquinas.
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

## 1. Linux (Ubuntu/Debian/Fedora/Arch — amd64)

### 1.1 Toolchain
```bash
# Debian / Ubuntu
sudo apt update
sudo apt install -y build-essential libgtk-3-dev libwebkit2gtk-4.1-dev pkg-config

# Si tu distro solo trae webkit2gtk-4.0 (Ubuntu 22.04 y anteriores):
sudo apt install -y libwebkit2gtk-4.0-dev
# y agrega `-tags webkit2_40` al wails build de abajo.

# Fedora
sudo dnf install -y gcc gtk3-devel webkit2gtk4.1-devel

# Arch
sudo pacman -S --needed base-devel gtk3 webkit2gtk
```

### 1.2 Go + Wails CLI
```bash
# Go 1.22+ (descarga de go.dev si tu distro tiene una versión vieja)
go version   # debe decir 1.22 o superior

go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH="$HOME/go/bin:$PATH"
wails doctor    # debe pasar todos los checks
```

### 1.3 Compilar el binario
```bash
cd ~/pgportable    # clon del repo
chmod +x scripts/build-linux.sh
./scripts/build-linux.sh
# → build/bin/PgPortable
```

### 1.4 Conseguir los binarios portables de PostgreSQL 17 para Linux

EnterpriseDB **no ofrece** tarballs portables oficiales para Linux. Hay dos rutas:

**Ruta A — Construir desde el .deb/.rpm oficial (recomendada)**
```bash
# Descargar el .deb de PostgreSQL 17 (Ubuntu 22.04 ejemplo)
wget https://apt.postgresql.org/pub/repos/apt/pool/main/p/postgresql-17/postgresql-17_17.9-1.pgdg22.04+1_amd64.deb

# Extraer SIN instalar
dpkg-deb -x postgresql-17_17.9-1.pgdg22.04+1_amd64.deb extracted/

# El árbol relevante queda en: extracted/usr/lib/postgresql/17/
# Reorganizar a la estructura que espera nuestro código:
mkdir -p pgsql/{bin,lib,share}
cp -r extracted/usr/lib/postgresql/17/bin/*       pgsql/bin/
cp -r extracted/usr/lib/x86_64-linux-gnu/*        pgsql/lib/  # solo lo necesario; ver abajo
cp -r extracted/usr/share/postgresql/17/*         pgsql/share/

# Test
./pgsql/bin/initdb --version    # debe decir initdb (PostgreSQL) 17.x
```

⚠️ Las dependencias de glibc/libicu/libssl pueden ser un dolor en Linux. Si el target es **Ubuntu 22.04 LTS** (común en estudiantes), construye el deploy ahí mismo y prueba que los `*.so` necesarios estén dentro de `pgsql/lib/` o sean parte de la base de Ubuntu 22.04. Para detectarlos:
```bash
ldd pgsql/bin/postgres | grep "not found"
```

**Ruta B — Usar `pgsql-tools` portable de un tercero (Bitnami / Postgres.app-style)**
Si la ruta A es muy frágil, busca un tarball portable mantenido por terceros o construye con Docker el binario estático.

### 1.5 Armar deploy/
```bash
mkdir deploy
cp build/bin/PgPortable deploy/PgPortable
cp -r pgsql              deploy/pgsql
cp LEEME.txt             deploy/LEEME.txt   # adapta para Linux (ver §3)
chmod +x deploy/PgPortable
```

### 1.6 Validación
```bash
cd deploy
./PgPortable --autostart &
sleep 15
psql -h localhost -p 5432 -U alumno -d postgres -c "SELECT version();"
# debe responder PostgreSQL 17.x on x86_64-linux
pkill PgPortable
```

---

## 2. macOS (arm64 Apple Silicon / amd64 Intel)

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
- Copiar a `pgsql/`
- Es universal binary (arm64 + amd64), portable, sin rpaths externos.

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
