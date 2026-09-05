package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runInstall(force bool) int {
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	args := []string{filepath.Join(root, "install.sh")}
	if force {
		args = append(args, "--rebuild")
	}
	cmd := exec.Command("bash", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
