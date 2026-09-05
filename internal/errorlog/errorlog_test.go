package errorlog

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendCreatesFileWithTimestampAndMessage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "error_log.txt")

	if err := Append(path, errors.New("clone/update repo: boom")); err != nil {
		t.Fatalf("Append: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error log: %v", err)
	}
	line := strings.TrimSpace(string(data))

	if !strings.Contains(line, "clone/update repo: boom") {
		t.Errorf("line = %q, want it to contain the error message", line)
	}
	fields := strings.SplitN(line, " ", 2)
	if len(fields) != 2 {
		t.Fatalf("line = %q, want a timestamp followed by the message", line)
	}
	if !strings.HasSuffix(fields[0], "Z") {
		t.Errorf("timestamp %q is not UTC (RFC3339 with Z suffix)", fields[0])
	}
}

func TestAppendPreservesPriorEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "error_log.txt")

	if err := Append(path, errors.New("first failure")); err != nil {
		t.Fatalf("Append (first): %v", err)
	}
	if err := Append(path, errors.New("second failure")); err != nil {
		t.Fatalf("Append (second): %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "first failure") || !strings.Contains(lines[1], "second failure") {
		t.Errorf("lines = %v, want first/second failure in order", lines)
	}
}
