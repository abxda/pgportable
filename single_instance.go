package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// acquireSingleInstance escribe pgportable.lock con el PID actual.
// Devuelve (ok=true) si pudo tomar el lock. Si ya existe y la otra instancia
// sigue viva, devuelve (false, otraPID).
func acquireSingleInstance(baseDir string) (bool, int) {
	lockPath := filepath.Join(baseDir, "pgportable.lock")
	if data, err := os.ReadFile(lockPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
			if pid != os.Getpid() && processAlive(pid) {
				return false, pid
			}
		}
	}
	_ = os.WriteFile(lockPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0o644)
	return true, 0
}

func releaseSingleInstance(baseDir string) {
	_ = os.Remove(filepath.Join(baseDir, "pgportable.lock"))
}
