// Package diagnostics performs pre-flight and post-start health checks:
// verifying required TCP ports are available, and scanning service logs for
// ERROR/FATAL/WARN markers so problems surface immediately instead of being
// discovered only when a user notices a broken service.
package diagnostics

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PortStatus is the availability of a single TCP port.
type PortStatus struct {
	Port int
	Free bool
}

// CheckPort reports whether port is currently free to bind on all
// interfaces. A port already in use (by this or another process) is
// reported as not free.
func CheckPort(port int) PortStatus {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return PortStatus{Port: port, Free: false}
	}
	_ = ln.Close()
	return PortStatus{Port: port, Free: true}
}

// CheckPorts checks every port in ports and returns their statuses in the
// same order.
func CheckPorts(ports []int) []PortStatus {
	statuses := make([]PortStatus, 0, len(ports))
	for _, p := range ports {
		statuses = append(statuses, CheckPort(p))
	}
	return statuses
}

// alertLevels are the log markers treated as noteworthy, most severe first.
// Order matters: a line containing both "ERROR" and "WARN" is classified by
// whichever level appears first in this slice.
var alertLevels = []string{"FATAL", "ERROR", "WARN"}

// LogAlert is a single noteworthy line found while scanning a log source.
type LogAlert struct {
	// Source is the path of the file the line came from.
	Source string
	// Line is the 1-based line number within Source.
	Line int
	// Level is the marker that triggered the alert (FATAL, ERROR, WARN).
	Level string
	// Text is the full line content.
	Text string
}

// classify returns the highest-severity alert level present in line, or ""
// if none of the known markers appear.
func classify(line string) string {
	upper := strings.ToUpper(line)
	for _, level := range alertLevels {
		if strings.Contains(upper, level) {
			return level
		}
	}
	return ""
}

// ScanLogFile scans a single file for lines containing FATAL, ERROR, or
// WARN markers (case-insensitive) and returns one LogAlert per match.
func ScanLogFile(path string) ([]LogAlert, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log %q: %w", path, err)
	}
	defer f.Close()

	var alerts []LogAlert
	scanner := bufio.NewScanner(f)
	// Log lines can be long (stack traces, JSON payloads); grow the buffer
	// well beyond bufio's 64KiB default rather than failing the scan.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if level := classify(line); level != "" {
			alerts = append(alerts, LogAlert{Source: path, Line: lineNum, Level: level, Text: line})
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return alerts, fmt.Errorf("scan log %q: %w", path, err)
	}

	return alerts, nil
}

// ScanLogDir scans every *.log file directly inside dir (non-recursive) and
// returns their combined alerts. A missing directory yields no alerts and
// no error, since not every project writes logs to disk.
func ScanLogDir(dir string) ([]LogAlert, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read log dir %q: %w", dir, err)
	}

	var alerts []LogAlert
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".log" {
			continue
		}
		fileAlerts, err := ScanLogFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return alerts, err
		}
		alerts = append(alerts, fileAlerts...)
	}
	return alerts, nil
}

// WaitForPort polls port until it responds to a TCP connection or timeout
// elapses, returning an error in the latter case. It is used to confirm a
// service actually came up after being started, rather than trusting the
// start command's exit code alone.
func WaitForPort(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("port %d did not become reachable within %s: %w", port, timeout, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
