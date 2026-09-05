// Package pgx exposes the small compatibility surface used by Trest Systems.
package pgx

import (
	"database/sql"
	"fmt"
)

// ErrNoRows matches the error returned by QueryRow.Scan when no row exists.
var ErrNoRows = sql.ErrNoRows

// Row is the query-row scan contract used by pgxpool.
type Row interface {
	Scan(dest ...any) error
}

// CommandTag contains the affected-row count for a command.
type CommandTag struct {
	Affected int64
}

func (c CommandTag) RowsAffected() int64 { return c.Affected }
func (c CommandTag) String() string      { return fmt.Sprintf("ROWS %d", c.Affected) }
