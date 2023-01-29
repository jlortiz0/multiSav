//go:build windows
// +build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func openFile(f string) {
	exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", f).Run()
}

func highlightFile(f string) {
	cwd, _ := os.Getwd()
	cmd := exec.Command("explorer", "/select,", filepath.Join(cwd, f))
	cwd = "explorer /select, " + cmd.Args[2]
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: cwd}
	cmd.Run()
}
