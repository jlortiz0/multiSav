//go:build linux
// +build linux

package main

import "os/exec"
import "path"

func openFile(f string) {
	exec.Command("xdg-open", f).Run()
}

func highlightFile(f string) {
	exec.Command("xdg-open", path.Dir(f)).Run()
}
