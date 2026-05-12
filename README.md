# pgportable

> **PostgreSQL completo en un solo `.exe` — descomprime, doble clic, y a programar.**
> Sin instalador, sin permisos de administrador, sin tocar tu computadora.

![status](https://img.shields.io/badge/status-Windows%20%2B%20Linux%20validated-2bd49c)
![platform](https://img.shields.io/badge/platform-win%2Famd64%20%7C%20linux%2Famd64-blue)
![scaffolding](https://img.shields.io/badge/macOS-scaffolding-orange)
![license](https://img.shields.io/badge/license-MIT-lightgrey)

---

## 🎯 ¿Para qué sirve?

Si eres **alumno** de una clase de bases de datos y tu profesor te dice
*"instala PostgreSQL en tu laptop"*, normalmente pasa esto:

- 😩 El instalador oficial pide **contraseña de administrador**.
- 😩 Te pregunta por **locale**, **puerto**, **superusuario**, **password de postgres**, y tú no sabes qué poner.
- 😩 Si ya tenías otra versión instalada, **chocan**.
- 😩 Cuando quieres desinstalarlo, **deja basura en el sistema**.

**pgportable** te da PostgreSQL 17 en un único archivo `.exe` que:

- 🚀 **Funciona con doble clic** — no hay instalador.
- 🔒 **No requiere admin** — corre como tu usuario normal.
- 💾 **No toca el sistema** — todo vive en la carpeta donde lo pones.
- 🗑️ **Se desinstala borrando la carpeta** — cero residuo.
- 🧳 **Es portable** — cópialo a un USB y se va contigo.

> Pensado para profesores que solo quieren que sus alumnos abran DBeaver y aprendan SQL,
> sin perder 3 clases peleando con instaladores.

---

## 📥 Descarga rápida (alumnos)

> **Aún no hay binarios pre-armados en Releases.**
> Por ahora, pídele a tu profesor el `.zip` con `PgPortable.exe` + carpeta `pgsql/`.
> Cuando haya release oficial, este enlace apuntará al ZIP correcto.

Requisitos:

- Windows 10 22H2 / Windows 11 (cualquier edición).
- ~200 MB libres en disco.
- **No** necesitas .NET, ni Java, ni Visual C++ Redistributable, ni admin.

---

## 🖥️ ¿Qué ves cuando lo abres?

```
┌────────────────────────────────────────────────────────────┐
│  🟦 PostgreSQL Portable                  ● Detenido        │
│     Entorno de alumno · sin instalación                    │
├────────────────────────────────────────────────────────────┤
│  [ ▶ Iniciar entorno ]     [ ⏹ Detener entorno ]           │
├──────────────────────────┬─────────────────────────────────┤
│  CONEXIÓN                │  ALMACENAMIENTO FÍSICO          │
│  Host       localhost    │  Carpeta datos                  │
│  Puerto     5432         │    C:\...\pgportable\data       │
│  Usuario    alumno       │  Binarios                       │
│  Contraseña alumno       │    C:\...\pgportable\pgsql\bin  │
│  Base       postgres     │  Tamaño        45 MB            │
│  [📋 URI]  [📋 JDBC]     │  Archivos      823              │
│  [📋 psql] [📋 ENV vars] │  [📁 Abrir carpeta data]        │
├──────────────────────────┴─────────────────────────────────┤
│  EXPLORADOR (bases de datos · esquemas · tablas)           │
│  ▾ postgres                                                │
│    ▾ public                                                │
│      • alumnos       TABLA · ~12 filas                     │
│      • cursos        TABLA · ~3 filas                      │
│  ▸ clase                                                   │
├────────────────────────────────────────────────────────────┤
│  CONSOLA                                                   │
│  Inicializando cluster ...                                 │
│  PostgreSQL listo en localhost:5432                        │
└────────────────────────────────────────────────────────────┘
```

---

## 🚀 Cómo usarlo (paso a paso)

1. **Descomprime** el ZIP donde quieras (escritorio, USB, OneDrive — no importa).
2. **Doble clic** en `PgPortable.exe`.
3. Clic en **▶ Iniciar entorno**. La primera vez tarda ~15 segundos creando la base de datos.
4. Cuando el indicador pase a **🟢 Corriendo · puerto 5432**, ya está listo.
5. Abre **DBeaver** (o pgAdmin, o lo que uses) y conecta con los datos de la card **Conexión**:
   - Host: `localhost`
   - Puerto: `5432` (puede ser otro si tenías 5432 ocupado — mira lo que indica la app)
   - Usuario: `alumno`
   - Contraseña: `alumno`
   - Base: `postgres`
6. ¿No quieres escribir los datos? Clic en **📋 URI** o **📋 JDBC** y pega en DBeaver.
7. Al terminar tu sesión: **⏹ Detener entorno** o cierra la ventana (te pregunta si detener).

> 💡 Tus tablas y datos **se guardan** entre sesiones en la subcarpeta `data/` junto al `.exe`.
> Si quieres empezar desde cero, borra esa carpeta y la próxima vez se crea nuevita.

---

## ⚙️ Funciones adicionales

| Funciona así | Por qué te importa |
|---|---|
| 🔌 **Auto-puerto**. Si 5432 está ocupado (otro PostgreSQL, Docker, etc.), elige 5433 ó 5434. | Nunca tendrás conflictos. |
| 🔒 **Single-instance**. Si abres el `.exe` dos veces, el segundo se cierra y te avisa. | No habrá dos servidores peleando por el mismo puerto. |
| 🌳 **Explorador integrado**. Ves qué bases / esquemas / tablas tienes sin abrir psql. | Diagnóstico instantáneo. |
| 📋 **5 formatos de copia**. URI · JDBC · psql · ENV vars · cada campo individual. | Pegas la cadena en DBeaver, VS Code, dotenv, lo que sea. |
| 🧯 **Sobrevive a apagones**. Si tu PC se apagó de golpe con Postgres corriendo, al volver a abrir la app se autocura. | Tus datos siguen ahí. |
| 📝 **Log de diagnóstico**. Si algo falla, hay un `pgportable.log` que puedes mandar al profe. | Soporte sin misterios. |

---

## ❓ Problemas comunes

### "No pasa nada cuando le doy doble clic"

- **¿Windows Defender / antivirus lo bloqueó?** Mira en la bandeja del sistema; algunos AV piden permiso la primera vez.
- **¿Falta WebView2?** Está en Windows 10 22H2 y todo Windows 11. Si tienes algo más viejo, instálalo desde [aquí](https://developer.microsoft.com/microsoft-edge/webview2/) (no requiere admin).

### "Dice 🔴 Error puerto 5432 ocupado"

Otra cosa está usando 5432 (otra instalación de PostgreSQL, Docker, etc.).

- **Opción A** (fácil): cierra el otro programa.
- **Opción B**: deja que pgportable elija otro puerto — borra `data/` y arranca de nuevo. La primera vez detectará 5432 ocupado y usará 5433.

### "Dice 🟡 Iniciando..." y nunca cambia

- Espera 30 segundos. La primera vez tarda por `initdb`.
- Si después de 1 minuto sigue ahí: abre la card **Consola** y mira si hay un error en rojo.
- Mándale a tu profe el archivo `pgportable.log` que está junto al `.exe`.

### "Mi antivirus dice que el .exe es sospechoso"

Falso positivo común con apps compiladas en Go que no están firmadas digitalmente. Agrega una excepción para la carpeta donde lo descomprimiste.

### "Mis datos desaparecieron"

Si moviste el `.exe` a otra carpeta sin llevar `data/` con él, sí, los perdiste — `data/` debe estar **junto** al `.exe`. Si solo cambiaste el `.exe` (actualización), tus datos siguen ahí.

---

## 🆚 ¿Por qué no usar otra cosa?

| Alternativa | Problema |
|---|---|
| **Instalador oficial de EnterpriseDB** | Pide admin, deja servicios en Windows, configurar manualmente, no portable. |
| **Docker** | Necesita Docker Desktop instalado (admin), WSL2, varios GB. Demasiado para una clase de SQL. |
| **PostgreSQL en la nube (Supabase, Neon, RDS)** | Requiere cuenta, internet, posiblemente tarjeta de crédito. Datos privados de práctica en servidor ajeno. |
| **Postgres.app** | Solo macOS. |
| **pgportable** ✓ | Una carpeta. Cero dependencias del sistema. Funciona offline. Cero costo. |

---

## 🛠️ Para desarrolladores / profes que quieren modificarlo

### Stack

| Capa | Tecnología |
|---|---|
| Framework GUI | Wails v2.12.0 (Go ↔ WebView2) |
| Backend | Go 1.22+ |
| Frontend | HTML / CSS / JS **vanilla** — sin npm, sin Vite, sin React |
| Cliente Postgres | `github.com/jackc/pgx/v5` (puro Go) |
| Binarios PG | EnterpriseDB v17 Win x86-64 (no incluidos en el repo) |
| Tamaño final | `.exe` ~14.7 MB · `pgsql/` ~132 MB · total deploy ~147 MB |

### Estructura del repo

```
pgportable/
├── main.go                ← entrada + single-instance guard + wails.Run
├── app.go                 ← API expuesta a JS (bind), eventos, snapshot
├── db_manager.go          ← initdb / pg_ctl / port-finder / pid cleanup
├── explorer.go            ← queries pgx (databases / schemas / tables)
├── single_instance.go     ← lock por PID + processAlive
├── platform_windows.go    ← Win32 (hide cmd, MessageBox)
├── platform_unix.go       ← stubs Mac/Linux
├── frontend/
│   ├── index.html
│   ├── style.css          ← tokens CSS + componentes .c-*
│   └── app.js             ← ICONS inline (Lucide-style) + STR i18n
├── scripts/
│   ├── build-windows.ps1
│   ├── build-linux.sh                ← validado Ubuntu 24.04
│   ├── pgsql-portable-linux.sh       ← arma pgsql/ desde el .deb oficial
│   └── build-darwin.sh               ← scaffolding
├── wails.json
├── go.mod
├── AGENTS.md              ← guía prescriptiva para completar Linux/macOS
└── README.md
```

### Compilar (Windows)

Requiere: **Go 1.22+**, **MSYS2 gcc** en `C:\msys64\mingw64\bin`, **Wails CLI v2**.

```powershell
.\scripts\build-windows.ps1
# salida: build\bin\PgPortable.exe (~14.7 MB)
```

### Armar el `deploy/` final

1. Descarga PostgreSQL 17 Win x86-64 ZIP de https://www.enterprisedb.com/download-postgresql-binaries
2. Extrae el ZIP. Renombra/copia la carpeta `pgsql/` al raíz del proyecto.
3. Crea `deploy/`:
   ```
   deploy/
   ├── PgPortable.exe       (copia de build\bin\)
   ├── pgsql/               (solo bin/ + lib/ + share/  →  descarta pgAdmin 4/, doc/, include/, StackBuilder/)
   └── LEEME.txt            (instrucciones para alumno)
   ```
4. Tamaño esperado: **~143-147 MB**. Comprime a ZIP → ~50 MB.

### Compilar Linux (Ubuntu 24.04 amd64) — validado

```bash
sudo apt install -y build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev golang-go
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
export PATH="$HOME/go/bin:$PATH"

cd ~/pgportable
bash scripts/build-linux.sh                    # → build/bin/PgPortable (~13 MB)
bash scripts/pgsql-portable-linux.sh           # → ./pgsql/ (~47 MB)

mkdir deploy
cp build/bin/PgPortable deploy/
cp -r pgsql deploy/
cp <tu LEEME.txt linux>  deploy/LEEME.txt
# deploy/ total: ~57 MB
```

Detalles, troubleshooting y lecciones aprendidas en [`AGENTS.md`](AGENTS.md) §1.

### Compilar macOS — pendiente

Scaffolding listo. [`AGENTS.md`](AGENTS.md) §2 tiene los comandos y las
lecciones de Linux que aplican (§1.5) para acelerar el trabajo del agente.

### Flag para devs

```
PgPortable.exe --autostart
```
Arranca Postgres automáticamente al abrir la GUI (útil para tests y demos).

### Decisiones técnicas

- **Wails y no Fyne**: stack consistente con otros proyectos del autor; WebView2 ya viene preinstalado en Windows moderno; HTML/CSS dan más control visual con cero learning curve.
- **`pgx` y no `psql` por subproceso**: pgx es puro Go, sin parsear stdout, pool decente. ~3 MB al binario.
- **Puerto en `postgresql.conf` y no en JSON aparte**: es la fuente canónica; un usuario avanzado puede editarlo y la app lo respeta.
- **Single-instance con PID-file y no named mutex**: portable cross-platform sin syscalls específicas; el archivo también sirve como huella para diagnóstico.
- **Side-log a disco (`pgportable.log`)**: si la GUI falla, el alumno puede mandarle ese archivo al profe sin tener que reproducir el bug.

---

## 🚫 Qué NO hace pgportable

Para que no haya confusiones:

- ❌ **No es un editor SQL.** Usa DBeaver, pgAdmin, psql o tu cliente favorito.
- ❌ **No es un servidor productivo.** Está pensado para una sola máquina, un solo usuario.
- ❌ **No es multi-versión.** Empaqueta solo PostgreSQL 17.
- ❌ **No es un fork de PostgreSQL.** Son los binarios oficiales de EnterpriseDB sin modificar.
- ❌ **No respalda automáticamente.** Si quieres backups, usa `pg_dump` (incluido en `pgsql/bin/`).

---

## 📜 Licencia

[MIT](LICENSE). El código del wrapper es MIT; los binarios de PostgreSQL siguen la [PostgreSQL License](https://www.postgresql.org/about/licence/) (también permisiva).

---

_Hecho con 💙 para alumnos universitarios que solo quieren conectar DBeaver y aprender SQL._
