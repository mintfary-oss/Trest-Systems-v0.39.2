// Package errorlog appends fatal orchestration errors to error_log.txt,
// giving operators a persistent, append-only record of failures even after
// state.json and report.json have been overwritten by later runs.
package errorlog

import (
	"fmt"
	"os"
	"time"
)

// Append writes a timestamped entry for err to path, creating the file if
// it does not exist and preserving any prior entries. It is best-effort:
// a failure to write the log is returned but should never mask the
// original orchestration error that triggered it.
func Append(path string, err error) error {
	f, openErr := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if openErr != nil {
		return fmt.Errorf("open error log %q: %w", path, openErr)
	}
	defer f.Close()

	if _, writeErr := fmt.Fprintf(f, "%s %s\n", time.Now().UTC().Format(time.RFC3339), err); writeErr != nil {
		return fmt.Errorf("write error log %q: %w", path, writeErr)
	}
	return nil
}
