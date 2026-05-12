# pgportable

> **Entorno PostgreSQL portable, zero-friction, para estudiantes.**
> Un solo ejecutable, sin admin, sin instalador, sin tocar el sistema.

![status](https://img.shields.io/badge/status-Windows%20validated-2bd49c) ![platform](https://img.shields.io/badge/platform-win%2Famd64-blue) ![scaffolding](https://img.shields.io/badge/Linux%2FmacOS-scaffolding-orange) ![license](https://img.shields.io/badge/license-MIT-lightgrey)

## ¿Qué hace?

- 🚀 Arranca / detiene PostgreSQL 17 desde una GUI mínima.
- 🔌 Auto-elige puerto si 5432 está ocupado.
- 🔒 Single-instance: si lo abres dos veces, te avisa.
- 🌳 Explorador de árbol: bases de datos → esquemas → tablas (con estimación de filas).
- 📁 Muestra dónde están los datos físicamente, con botón "Abrir carpeta".
- 📋 Copia con un clic: URI libpq, JDBC URL, comando psql, bloque de variables de entorno.
- 🧯 Resiliente: limpia `postmaster.pid` huérfanos tras apagones, se recupera si Postgres ya estaba corriendo.

## Stack

| Capa | Tecnología |
|---|---|
| Framework | Wails v2.12.0 (Go ↔ WebView2 IPC) |
| Backend | Go 1.22+ |
| Frontend | HTML / CSS / JS **vanilla** — sin npm, sin Vite, sin React |
| Cliente Postgres | `github.com/jackc/pgx/v5` (puro Go) |
| Binarios PG | EnterpriseDB v17 Win x86-64 (no incluidos en el repo) |
| Empaquetado | un solo `.exe` (~11 MB) + `pgsql/` (~132 MB) |

## Estructura

```
pgportable/
├── main.go                      ← entry + single-instance guard + wails.Run
├── app.go                       ← API expuesta al frontend (bind)
├── db_manager.go                ← initdb / pg_ctl / port-finder / pid cleanup
├── explorer.go                  ← queries pgx para databases/schemas/tables
├── single_instance.go           ← lock por PID + processAlive
├── platform_windows.go          ← Win32 helpers (hide cmd window, MessageBox)
├── platform_unix.go             ← stubs Mac/Linux
├── frontend/
│   ├── index.html
│   ├── style.css                ← tokens + componentes .c-*
│   └── app.js                   ← ICONS inline (Lucide style) + STR i18n
├── scripts/
│   ├── build-windows.ps1
│   ├── build-linux.sh
│   └── build-darwin.sh
├── wails.json
├── go.mod
└── AGENTS.md                    ← guía para completar Linux/macOS
```

## Compilar (Windows)

Requiere: Go 1.22+, MSYS2 gcc en `C:\msys64\mingw64\bin`, Wails CLI v2.

```powershell
.\scripts\build-windows.ps1
# salida: build\bin\PgPortable.exe (~11 MB)
```

Para armar el `deploy/` final con los binarios de Postgres:

1. Descarga PostgreSQL 17 Win x86-64 ZIP de https://www.enterprisedb.com/download-postgresql-binaries
2. Extrae y queda como `pgsql/` (debe contener `pgsql/bin/initdb.exe` etc.)
3. Crea `deploy/` con:
   - `deploy/PgPortable.exe` (copia de `build/bin/`)
   - `deploy/pgsql/` (solo `bin/`, `lib/`, `share/` — descarta `pgAdmin 4/`, `doc/`, `include/`, `StackBuilder/`)
   - `deploy/LEEME.txt` (instrucciones para el alumno)

Tamaño típico del deploy: **~143 MB**.

## Compilar (Linux / macOS)

Estos targets están **en scaffolding**. Lee [`AGENTS.md`](AGENTS.md) — tiene los comandos exactos para que un agente IA (o tú) los complete en pocos minutos en la máquina respectiva.

## Cómo lo usa el alumno

1. Recibe un ZIP con `PgPortable.exe`, `pgsql/`, `LEEME.txt`.
2. Descomprime donde quiera (escritorio, USB, OneDrive).
3. Doble clic en `PgPortable.exe`.
4. Clic en "Iniciar entorno". La primera vez tarda ~10 s creando la base.
5. Cuando aparece 🟢 Corriendo, abre DBeaver con los datos de la card "Conexión":
   - `localhost:5432`, usuario `alumno`, password `alumno`, base `postgres`.
6. El árbol abajo muestra las BD/esquemas/tablas que tiene dentro.
7. Al terminar: clic en "Detener entorno" o cierra la ventana.

## Flag útil para desarrolladores

```
PgPortable.exe --autostart
```
Arranca Postgres automáticamente al abrir la GUI (útil para tests o si el alumno ya sabe lo que hace).

## Diseño / decisiones técnicas

- **Por qué Wails y no Fyne**: stack consistente con otros proyectos del autor; WebView2 ya viene en Windows 10 22H2+ y todo Windows 11; HTML/CSS dan mucho más control visual con cero learning curve.
- **Por qué `pgx` y no usar `psql` por subproceso**: pgx es puro Go (sin CGO extra), pool de conexiones decente, parsing trivial. ~3MB al binario.
- **Por qué guardar el puerto en `postgresql.conf` y no en JSON aparte**: es la fuente canónica; un alumno avanzado puede editarlo a mano y la app lo respeta.
- **Por qué single-instance con PID-file y no named mutex**: portable, cross-platform sin syscalls específicas, y el archivo sirve también como "última huella conocida" si alguien quiere diagnosticar.

## Licencia

MIT. Ver `LICENSE`.

---

_Hecho con 💙 para alumnos universitarios que solo quieren conectar DBeaver y aprender SQL._
