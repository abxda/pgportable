//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func hideWindow(_ *exec.Cmd) {}

func showMessageBox(title, text string) {
	fmt.Fprintf(os.Stderr, "[%s] %s\n", title, text)
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}
