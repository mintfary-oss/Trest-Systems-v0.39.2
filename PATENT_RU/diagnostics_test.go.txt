package diagnostics

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCheckPortFreeAndBusy(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	if status := CheckPort(port); status.Free {
		t.Errorf("CheckPort(%d) = free, want busy while listener is open", port)
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	if status := CheckPort(port); !status.Free {
		t.Errorf("CheckPort(%d) = busy, want free after listener closed", port)
	}
}

func TestCheckPorts(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	busyPort := ln.Addr().(*net.TCPAddr).Port

	statuses := CheckPorts([]int{busyPort})
	if len(statuses) != 1 {
		t.Fatalf("statuses = %d, want 1", len(statuses))
	}
	if statuses[0].Free {
		t.Errorf("statuses[0].Free = true, want false")
	}
}

func writeLog(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write log %q: %v", path, err)
	}
	return path
}

func TestScanLogFileFindsAlerts(t *testing.T) {
	dir := t.TempDir()
	path := writeLog(t, dir, "app.log", "starting up\nINFO ready\nERROR: db connection refused\nFATAL: crashing\nWARN: disk 90% full\n")

	alerts, err := ScanLogFile(path)
	if err != nil {
		t.Fatalf("ScanLogFile: %v", err)
	}
	if len(alerts) != 3 {
		t.Fatalf("alerts = %d, want 3: %+v", len(alerts), alerts)
	}
	if alerts[0].Level != "ERROR" || alerts[0].Line != 3 {
		t.Errorf("alerts[0] = %+v, want Level=ERROR Line=3", alerts[0])
	}
	if alerts[1].Level != "FATAL" || alerts[1].Line != 4 {
		t.Errorf("alerts[1] = %+v, want Level=FATAL Line=4", alerts[1])
	}
	if alerts[2].Level != "WARN" || alerts[2].Line != 5 {
		t.Errorf("alerts[2] = %+v, want Level=WARN Line=5", alerts[2])
	}
}

func TestScanLogFileNoAlerts(t *testing.T) {
	dir := t.TempDir()
	path := writeLog(t, dir, "app.log", "starting up\nINFO ready\nINFO all good\n")

	alerts, err := ScanLogFile(path)
	if err != nil {
		t.Fatalf("ScanLogFile: %v", err)
	}
	if len(alerts) != 0 {
		t.Errorf("alerts = %+v, want none", alerts)
	}
}

func TestScanLogDirAggregatesLogFilesOnly(t *testing.T) {
	dir := t.TempDir()
	writeLog(t, dir, "a.log", "ERROR: a broke\n")
	writeLog(t, dir, "b.log", "ERROR: b broke\n")
	writeLog(t, dir, "notes.txt", "ERROR: should be ignored, not a .log file\n")

	alerts, err := ScanLogDir(dir)
	if err != nil {
		t.Fatalf("ScanLogDir: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("alerts = %d, want 2: %+v", len(alerts), alerts)
	}
}

func TestScanLogDirMissingDirIsNotError(t *testing.T) {
	alerts, err := ScanLogDir(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("ScanLogDir: %v", err)
	}
	if alerts != nil {
		t.Errorf("alerts = %+v, want nil", alerts)
	}
}

func TestWaitForPortSucceedsOnceListening(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if err := WaitForPort(port, 2*time.Second); err != nil {
		t.Errorf("WaitForPort: %v", err)
	}
}

func TestWaitForPortTimesOutWhenNothingListens(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if err := ln.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	if err := WaitForPort(port, 500*time.Millisecond); err == nil {
		t.Error("WaitForPort: want timeout error, got nil")
	}
}
