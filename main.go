package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	baseDir := filepath.Dir(exe)

	// Single-instance: si ya hay otra instancia con el lock vigente, avisamos y salimos.
	if ok, other := acquireSingleInstance(baseDir); !ok {
		showMessageBox(
			"PostgreSQL Portable",
			fmt.Sprintf("Ya está corriendo otra instancia de PgPortable (PID %d).\n\nBuscala en la barra de tareas.", other),
		)
		return
	}
	defer releaseSingleInstance(baseDir)

	app := NewApp(baseDir)

	if err := wails.Run(&options.App{
		Title:     "PostgreSQL Portable — Entorno de Alumno",
		Width:     920,
		Height:    760,
		MinWidth:  720,
		MinHeight: 560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 20, B: 28, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableWindowIcon:                 false,
			DisableFramelessWindowDecorations: false,
		},
	}); err != nil {
		log.Fatal(err)
	}
}
