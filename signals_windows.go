//go:build windows

package main

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/ssh"
)

// handleSignals handles SIGINT only on Windows (no SIGWINCH support).
func handleSignals(session *ssh.Session) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	for sig := range sigChan {
		if sig == syscall.SIGINT {
			_ = session.Signal(ssh.SIGINT)
		}
	}
}
