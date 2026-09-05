package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// The install subcommand always uses the single canonical 0.39.2 installer.
func unifiedInstall() bool {
	if len(os.Args) < 2 || os.Args[1] != "install" {
		return false
	}
	exe, _ := os.Executable()
	roots := []string{os.Getenv("TREST_ROOT"), filepath.Dir(exe), "."}
	for _, start := range roots {
		if start == "" {
			continue
		}
		p, _ := filepath.Abs(start)
		for i := 0; i < 6; i++ {
			file := filepath.Join(p, "install.sh")
			if st, e := os.Stat(file); e == nil && !st.IsDir() {
				cmd := exec.Command("bash", append([]string{file}, os.Args[2:]...)...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if e = cmd.Run(); e != nil {
					fmt.Fprintln(os.Stderr, e)
					os.Exit(1)
				}
				return true
			}
			p = filepath.Dir(p)
		}
	}
	fmt.Fprintln(os.Stderr, "install.sh not found; set TREST_ROOT to the full package directory")
	os.Exit(2)
	return true
}
