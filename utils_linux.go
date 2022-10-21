package main

import "os/exec"
import "path"

// +build linux

func openFile(f string) {
	exec.Command("xdg-open", f).Run()
}


func highlightFile(f string) {
    exec.Command("xdg-open", path.Dir(f)).Run()
}

