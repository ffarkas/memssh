//go:build !windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// handleSignals handles SIGWINCH (resize) and SIGINT on Unix systems.
func handleSignals(session *ssh.Session) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGINT)
	for sig := range sigChan {
		switch sig {
		case syscall.SIGWINCH:
			fd := int(syscall.Stdin)
			if width, height, err := term.GetSize(fd); err == nil {
				_ = session.WindowChange(height, width)
			}
		case syscall.SIGINT:
			_ = session.Signal(ssh.SIGINT)
		}
	}
}
