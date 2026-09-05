// Package stdlib registers the pgx-compatible database/sql driver name.
package stdlib

import (
	"database/sql"

	"github.com/jackc/pgx/v5/internal/libpqdriver"
)

func init() {
	sql.Register("pgx", libpqdriver.Driver{})
}
