// Command trest-bootstrap launches the platform-native automatic installer and provides a minimal doctor.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const version = "v0.39.2"

type check struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func look(name string, args ...string) check {
	p, e := exec.LookPath(name)
	if e != nil {
		return check{name, "FAIL", e.Error()}
	}
	if len(args) > 0 {
		o, e := exec.Command(p, args...).CombinedOutput()
		if e != nil {
			return check{name, "FAIL", strings.TrimSpace(string(o))}
		}
		return check{name, "PASS", strings.TrimSpace(string(o))}
	}
	return check{name, "PASS", p}
}
func port(p string) check {
	c, e := net.DialTimeout("tcp", net.JoinHostPort("127.0.0.1", p), 2*time.Second)
	if e != nil {
		return check{"port " + p, "FAIL", e.Error()}
	}
	c.Close()
	return check{"port " + p, "PASS", "listening"}
}
func rootDir() (string, error) {
	exe, e := os.Executable()
	if e != nil {
		return "", e
	}
	candidates := []string{filepath.Dir(exe), filepath.Join(filepath.Dir(exe), "..", "..", "..", ".."), "."}
	for _, c := range candidates {
		a, _ := filepath.Abs(c)
		if _, e := os.Stat(filepath.Join(a, "install.sh")); e == nil {
			return a, nil
		}
	}
	return "", errors.New("install.sh not found near executable")
}
func install(args []string) error {
	root, e := rootDir()
	if e != nil {
		return e
	}
	if runtime.GOOS == "windows" {
		translated := []string{}
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--tls", "--domain", "--email":
				if i+1 >= len(args) {
					return errors.New("missing installer argument")
				}
				key := map[string]string{"--tls": "-Tls", "--domain": "-Domain", "--email": "-AcmeEmail"}[args[i]]
				i++
				translated = append(translated, key, args[i])
			case "--non-interactive":
				translated = append(translated, "-NonInteractive")
			case "--dry-run":
				translated = append(translated, "-DryRun")
			default:
				return fmt.Errorf("Windows launcher: unsupported flag %s; use Linux installer under WSL for advanced options", args[i])
			}
		}
		all := append([]string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(root, "install.ps1")}, translated...)
		cmd := exec.Command("powershell.exe", all...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		return cmd.Run()
	}
	all := append([]string{filepath.Join(root, "install.sh")}, args...)
	cmd := exec.Command("bash", all...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
func doctor() int {
	cs := []check{look("docker", "version", "--format", "{{.Server.Version}}"), look("docker", "compose", "version"), port("80")}
	fail := false
	for _, c := range cs {
		fmt.Printf("%-18s %-5s %s\n", c.Name, c.Status, c.Detail)
		fail = fail || c.Status == "FAIL"
	}
	b, _ := json.MarshalIndent(cs, "", "  ")
	_ = os.WriteFile("trest-doctor-report.json", b, 0600)
	if fail {
		return 1
	}
	return 0
}
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("trest-bootstrap", version, "\ncommands: install, doctor, report, version")
		os.Exit(2)
	}
	switch args[0] {
	case "install":
		if e := install(args[1:]); e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
	case "doctor":
		os.Exit(doctor())
	case "version":
		fmt.Println(version)
	case "report":
		f, e := os.Open("/var/lib/trest-systems/last-install-report.txt")
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		defer f.Close()
		s := bufio.NewScanner(f)
		for s.Scan() {
			fmt.Println(s.Text())
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(2)
	}
}
