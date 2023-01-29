//go:build linux
// +build linux

package main

import "os/exec"
import "path/filepath"

func openFile(f string) {
	exec.Command("xdg-open", f).Run()
}

func highlightFile(f string) {
	exec.Command("xdg-open", filepath.Dir(f)).Run()
}
