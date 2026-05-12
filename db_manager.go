package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Config struct {
	BaseDir  string
	BinDir   string
	DataDir  string
	LogFile  string
	Port     int
	User     string
	Database string
	Password string
}

type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusError
)

type Manager struct {
	cfg      Config
	mu       sync.RWMutex
	status   Status
	subPhase string // texto fino que el UI muestra junto al estado (ej. "initdb")
	logCh    chan string
}

func NewManager(cfg Config) *Manager {
	return &Manager{cfg: cfg, logCh: make(chan string, 1024)}
}

func (m *Manager) SubPhase() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.subPhase
}

func (m *Manager) setSubPhase(s string) {
	m.mu.Lock()
	m.subPhase = s
	m.mu.Unlock()
}

func (m *Manager) Config() Config      { return m.cfg }
func (m *Manager) Logs() <-chan string { return m.logCh }

func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status
}

func (m *Manager) setStatus(s Status) {
	m.mu.Lock()
	m.status = s
	m.mu.Unlock()
}

func (m *Manager) log(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	// Side log a disco para diagnostico post-mortem (util si la GUI no abre).
	if f, err := os.OpenFile(filepath.Join(m.cfg.BaseDir, "pgportable.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05.000"), msg)
		_ = f.Close()
	}
	select {
	case m.logCh <- msg:
	default:
	}
}

func (m *Manager) binary(name string) string {
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(m.cfg.BinDir, name)
}

func (m *Manager) BinariesPresent() bool {
	for _, b := range []string{"initdb", "pg_ctl", "postgres"} {
		if _, err := os.Stat(m.binary(b)); err != nil {
			return false
		}
	}
	return true
}

func (m *Manager) IsInitialized() bool {
	_, err := os.Stat(filepath.Join(m.cfg.DataDir, "PG_VERSION"))
	return err == nil
}

func (m *Manager) IsRunning() bool {
	return isListening(m.cfg.Port)
}

func (m *Manager) IsPortFree() bool {
	return isPortFree(m.cfg.Port)
}

func isListening(port int) bool {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func isPortFree(port int) bool {
	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// findFreePort intenta a partir de preferred y crece hasta preferred+maxScan-1.
func findFreePort(preferred, maxScan int) (int, error) {
	for p := preferred; p < preferred+maxScan; p++ {
		if isPortFree(p) {
			return p, nil
		}
	}
	return 0, fmt.Errorf("ningun puerto libre en %d..%d", preferred, preferred+maxScan-1)
}

// readPortFromConf parsea postgresql.conf y devuelve la ultima directiva `port = N`
// (las nuestras se agregan al final como override).
func (m *Manager) readPortFromConf() int {
	data, err := os.ReadFile(filepath.Join(m.cfg.DataDir, "postgresql.conf"))
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(`(?m)^\s*port\s*=\s*(\d+)`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	if len(matches) == 0 {
		return 0
	}
	p, _ := strconv.Atoi(matches[len(matches)-1][1])
	return p
}

// SyncPortFromDisk se llama al construir el Manager para que el puerto
// configurado en disco gane sobre el default del Config.
func (m *Manager) SyncPortFromDisk() {
	if p := m.readPortFromConf(); p > 0 {
		m.cfg.Port = p
	}
}

// cleanupStalePidFile recupera el cluster de un cierre abrupto: detecta
// postmaster.pid huerfanos cuando el proceso ya no vive o el puerto esta libre.
func (m *Manager) cleanupStalePidFile() error {
	pidFile := filepath.Join(m.cfg.DataDir, "postmaster.pid")
	data, err := os.ReadFile(pidFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	lines := strings.SplitN(string(data), "\n", 2)
	pid, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		m.log("postmaster.pid corrupto, eliminando")
		return os.Remove(pidFile)
	}
	if !processAlive(pid) {
		m.log("postmaster.pid huerfano (PID %d no existe), limpiando", pid)
		return os.Remove(pidFile)
	}
	if !m.IsRunning() {
		m.log("postmaster.pid presente pero sin escucha en %d, limpiando", m.cfg.Port)
		return os.Remove(pidFile)
	}
	return nil
}

func (m *Manager) Initialize(ctx context.Context) error {
	if m.IsInitialized() {
		return nil
	}
	if !m.BinariesPresent() {
		return fmt.Errorf("binarios PostgreSQL no encontrados en %s", m.cfg.BinDir)
	}
	if err := os.MkdirAll(m.cfg.DataDir, 0o755); err != nil {
		return fmt.Errorf("creando data dir: %w", err)
	}
	// Puerto inteligente: si 5432 esta ocupado, escala. Esto solo aplica al
	// initdb inicial; el puerto elegido queda persistido en postgresql.conf.
	if !isPortFree(m.cfg.Port) {
		alt, err := findFreePort(m.cfg.Port, 50)
		if err != nil {
			return err
		}
		m.log("Puerto %d ocupado, usando %d como alternativa.", m.cfg.Port, alt)
		m.cfg.Port = alt
	}
	pwfile := filepath.Join(m.cfg.BaseDir, ".pgpw.tmp")
	if err := os.WriteFile(pwfile, []byte(m.cfg.Password), 0o600); err != nil {
		return fmt.Errorf("escribiendo password file: %w", err)
	}
	defer os.Remove(pwfile)

	args := []string{
		"-D", m.cfg.DataDir,
		"-U", m.cfg.User,
		"-A", "scram-sha-256",
		"--pwfile=" + pwfile,
		"-E", "UTF8",
		"--locale=C",
	}
	m.log("Inicializando cluster en %s ...", m.cfg.DataDir)
	m.setSubPhase("initdb")
	if err := m.runStream(ctx, m.binary("initdb"), args...); err != nil {
		return fmt.Errorf("initdb fallo: %w", err)
	}
	if err := m.applyConfig(); err != nil {
		return err
	}
	m.log("Cluster inicializado correctamente.")
	return nil
}

func (m *Manager) applyConfig() error {
	confPath := filepath.Join(m.cfg.DataDir, "postgresql.conf")
	extra := fmt.Sprintf(
		"\n# portable overrides\nlisten_addresses = '127.0.0.1'\nport = %d\n",
		m.cfg.Port,
	)
	// En Linux, los binarios de las distros (Debian/Ubuntu/Fedora) tienen
	// compilado unix_socket_directories='/var/run/postgresql', carpeta que
	// solo existe con el paquete oficial instalado y no es escribible por
	// usuarios sin sudo. Forzamos /tmp (default de upstream PostgreSQL)
	// para que el deploy portable arranque en cualquier sistema.
	if runtime.GOOS == "linux" {
		extra += "unix_socket_directories = '/tmp'\n"
	}
	f, err := os.OpenFile(confPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(extra)
	return err
}

func (m *Manager) Start(ctx context.Context) error {
	if m.IsRunning() {
		m.setStatus(StatusRunning)
		return nil
	}
	if !m.BinariesPresent() {
		m.setStatus(StatusError)
		return fmt.Errorf("binarios PostgreSQL no encontrados en %s (¿copiaste pgsql/?)", m.cfg.BinDir)
	}
	m.setStatus(StatusStarting)

	if !m.IsInitialized() {
		if err := m.Initialize(ctx); err != nil {
			m.setStatus(StatusError)
			return err
		}
	}
	if err := m.cleanupStalePidFile(); err != nil {
		m.log("Aviso limpiando PID: %v", err)
	}
	// Si el cluster ya estaba inicializado, el puerto vive en postgresql.conf.
	// Lo re-leemos por si cambio externamente.
	if p := m.readPortFromConf(); p > 0 && p != m.cfg.Port {
		m.cfg.Port = p
	}
	if !m.IsPortFree() {
		m.setStatus(StatusError)
		return fmt.Errorf("puerto %d ocupado por otro proceso (no es nuestro postgres). Cierra esa app o cambia el puerto en %s\\postgresql.conf", m.cfg.Port, m.cfg.DataDir)
	}

	args := []string{
		"-D", m.cfg.DataDir,
		"-l", m.cfg.LogFile,
		"-w",
		"-o", fmt.Sprintf("-p %d", m.cfg.Port),
		"start",
	}
	m.log("Arrancando PostgreSQL en puerto %d ...", m.cfg.Port)
	m.setSubPhase("pg_ctl")
	if err := m.runStream(ctx, m.binary("pg_ctl"), args...); err != nil {
		m.setStatus(StatusError)
		m.setSubPhase("")
		return err
	}

	m.setSubPhase("esperando puerto")
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if m.IsRunning() {
			m.setStatus(StatusRunning)
			m.setSubPhase("")
			m.log("PostgreSQL listo en localhost:%d", m.cfg.Port)
			return nil
		}
		time.Sleep(300 * time.Millisecond)
	}
	m.setStatus(StatusError)
	m.setSubPhase("")
	return fmt.Errorf("PostgreSQL no respondio en 30s")
}

func (m *Manager) Stop(ctx context.Context) error {
	if !m.IsRunning() {
		m.setStatus(StatusStopped)
		return nil
	}
	m.setStatus(StatusStopping)
	m.setSubPhase("pg_ctl stop")
	args := []string{"-D", m.cfg.DataDir, "-m", "fast", "-w", "stop"}
	m.log("Deteniendo PostgreSQL ...")
	if err := m.runStream(ctx, m.binary("pg_ctl"), args...); err != nil {
		m.setStatus(StatusError)
		m.setSubPhase("")
		return err
	}
	m.setStatus(StatusStopped)
	m.setSubPhase("")
	m.log("PostgreSQL detenido.")
	return nil
}

func (m *Manager) runStream(ctx context.Context, bin string, args ...string) error {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = append(os.Environ(),
		"PGCLIENTENCODING=UTF8",
		"LC_ALL=C",
	)
	hideWindow(cmd)

	// Critico en Windows: pg_ctl arranca postgres detachado, y los workers heredan
	// los handles de stdout/stderr de pg_ctl. Tras salir pg_ctl, esas pipes siguen
	// abiertas mientras viva postgres; nuestros scanners quedarian bloqueados para
	// siempre. WaitDelay (Go 1.20+) cierra las pipes a la fuerza N segundos despues
	// de que el proceso principal salga, desbloqueando los scanners.
	cmd.WaitDelay = 2 * time.Second

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); m.scan(stdout) }()
	go func() { defer wg.Done(); m.scan(stderr) }()

	// IMPORTANTE: cmd.Wait() ANTES de wg.Wait(). En Windows pg_ctl arranca
	// postgres detachado y los workers heredan los handles, asi que las pipes
	// no cierran solas: necesitamos que WaitDelay las cierre. WaitDelay solo
	// dispara dentro de cmd.Wait(); si esperaramos wg.Wait() primero, deadlock.
	errWait := cmd.Wait()
	wg.Wait()
	return errWait
}

func (m *Manager) scan(r io.Reader) {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for s.Scan() {
		m.log("%s", s.Text())
	}
}

func (m *Manager) ConnectionString() string {
	return fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s",
		m.cfg.User, m.cfg.Password, m.cfg.Port, m.cfg.Database)
}

func (m *Manager) JDBCString() string {
	return fmt.Sprintf("jdbc:postgresql://localhost:%d/%s?user=%s&password=%s",
		m.cfg.Port, m.cfg.Database, m.cfg.User, m.cfg.Password)
}

func (m *Manager) PsqlCommand() string {
	return fmt.Sprintf(`psql -h localhost -p %d -U %s -d %s`,
		m.cfg.Port, m.cfg.User, m.cfg.Database)
}

// EnvBlock devuelve un bloque listo para pegar en .env / shell.
func (m *Manager) EnvBlock() string {
	return fmt.Sprintf(
		"PGHOST=localhost\nPGPORT=%d\nPGUSER=%s\nPGPASSWORD=%s\nPGDATABASE=%s\n",
		m.cfg.Port, m.cfg.User, m.cfg.Password, m.cfg.Database,
	)
}
