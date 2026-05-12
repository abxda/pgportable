package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App orquesta el ciclo de vida de Postgres y publica eventos al frontend.
type App struct {
	ctx     context.Context
	baseDir string
	mgr     *Manager
}

func NewApp(baseDir string) *App {
	cfg := Config{
		BaseDir:  baseDir,
		BinDir:   filepath.Join(baseDir, "pgsql", "bin"),
		DataDir:  filepath.Join(baseDir, "data"),
		LogFile:  filepath.Join(baseDir, "pgsql.log"),
		Port:     5432,
		User:     "alumno",
		Database: "postgres",
		Password: "alumno",
	}
	m := NewManager(cfg)
	m.SyncPortFromDisk()
	return &App{baseDir: baseDir, mgr: m}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Pipe logs del manager hacia el frontend como eventos "pg:log".
	go func() {
		for line := range a.mgr.Logs() {
			runtime.EventsEmit(a.ctx, "pg:log", line)
		}
	}()

	// Auto-start opcional: --autostart o variable PGPORTABLE_AUTOSTART=1
	if slices.Contains(os.Args, "--autostart") || os.Getenv("PGPORTABLE_AUTOSTART") == "1" {
		go func() {
			time.Sleep(300 * time.Millisecond)
			_ = a.Start()
		}()
	}

	// Heartbeat de estado cada segundo (autonomo: detecta cambios externos).
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		var last StatusSnapshot
		for {
			select {
			case <-a.ctx.Done():
				return
			case <-t.C:
				cur := a.GetStatus()
				if cur != last {
					runtime.EventsEmit(a.ctx, "pg:status", cur)
					last = cur
				}
			}
		}
	}()
}

func (a *App) shutdown(_ context.Context) {
	// Detener Postgres al cerrar la ventana evita postmaster.pid huerfano.
	if a.mgr != nil && a.mgr.IsRunning() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = a.mgr.Stop(ctx)
	}
}

// ──────────────────────────────────────────────────────────────
// Métodos expuestos a JavaScript (Bind en main.go)
// ──────────────────────────────────────────────────────────────

type StatusSnapshot struct {
	State           string `json:"state"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	Host            string `json:"host"`
	ConnectionURI   string `json:"connectionUri"`
	JdbcURL         string `json:"jdbcUrl"`
	PsqlCommand     string `json:"psqlCommand"`
	EnvBlock        string `json:"envBlock"`
	Initialized     bool   `json:"initialized"`
	DataDir         string `json:"dataDir"`
	BinDir          string `json:"binDir"`
	BinariesPresent bool   `json:"binariesPresent"`
}

type StorageInfo struct {
	DataDir   string `json:"dataDir"`
	Exists    bool   `json:"exists"`
	SizeBytes int64  `json:"sizeBytes"`
	SizeHuman string `json:"sizeHuman"`
	FileCount int    `json:"fileCount"`
	BinDir    string `json:"binDir"`
	LogFile   string `json:"logFile"`
}

func (a *App) GetStatus() StatusSnapshot {
	cfg := a.mgr.Config()
	state := "stopped"
	switch a.mgr.Status() {
	case StatusStarting:
		state = "starting"
	case StatusStopping:
		state = "stopping"
	case StatusError:
		state = "error"
	}
	if state == "stopped" && a.mgr.IsRunning() {
		state = "running"
	}
	return StatusSnapshot{
		State:           state,
		Port:            cfg.Port,
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		Host:            "localhost",
		ConnectionURI:   a.mgr.ConnectionString(),
		JdbcURL:         a.mgr.JDBCString(),
		PsqlCommand:     a.mgr.PsqlCommand(),
		EnvBlock:        a.mgr.EnvBlock(),
		Initialized:     a.mgr.IsInitialized(),
		DataDir:         cfg.DataDir,
		BinDir:          cfg.BinDir,
		BinariesPresent: a.mgr.BinariesPresent(),
	}
}

// GetStorageInfo recorre data/ y reporta tamaño + nro de archivos.
func (a *App) GetStorageInfo() StorageInfo {
	cfg := a.mgr.Config()
	info := StorageInfo{
		DataDir: cfg.DataDir,
		BinDir:  cfg.BinDir,
		LogFile: cfg.LogFile,
	}
	if st, err := os.Stat(cfg.DataDir); err == nil && st.IsDir() {
		info.Exists = true
		_ = filepath.Walk(cfg.DataDir, func(_ string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() {
				return nil
			}
			info.SizeBytes += fi.Size()
			info.FileCount++
			return nil
		})
		info.SizeHuman = humanBytes(info.SizeBytes)
	}
	return info
}

func humanBytes(b int64) string {
	const u = 1024
	if b < u {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(u), 0
	for n := b / u; n >= u; n /= u {
		div *= u
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// Start arranca Postgres en background. Retorna error si algo falla.
func (a *App) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	go func() {
		defer cancel()
		if err := a.mgr.Start(ctx); err != nil {
			runtime.LogErrorf(a.ctx, "start: %v", err)
			runtime.EventsEmit(a.ctx, "pg:error", err.Error())
		}
		runtime.EventsEmit(a.ctx, "pg:status", a.GetStatus())
	}()
	return nil
}

func (a *App) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	go func() {
		defer cancel()
		if err := a.mgr.Stop(ctx); err != nil {
			runtime.LogErrorf(a.ctx, "stop: %v", err)
			runtime.EventsEmit(a.ctx, "pg:error", err.Error())
		}
		runtime.EventsEmit(a.ctx, "pg:status", a.GetStatus())
	}()
	return nil
}

// OpenDataFolder abre el explorador en la carpeta data/.
func (a *App) OpenDataFolder() {
	target := a.mgr.Config().DataDir
	if _, err := os.Stat(target); err != nil {
		target = a.mgr.Config().BaseDir
	}
	runtime.BrowserOpenURL(a.ctx, "file:///"+filepath.ToSlash(target))
}

// ListDatabases / ListSchemas / ListTables — exposiciones al frontend.
func (a *App) ListDatabases() ([]DBNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.mgr.ListDatabases(ctx)
}

func (a *App) ListSchemas(db string) ([]SchemaNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.mgr.ListSchemas(ctx, db)
}

func (a *App) ListTables(db, schema string) ([]TableNode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.mgr.ListTables(ctx, db, schema)
}
