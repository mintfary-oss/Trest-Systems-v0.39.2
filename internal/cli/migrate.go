package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/mintfary-oss/trest-sistems/internal/db"
)

func runMigrate() int {
	ctx := context.Background()
	database, err := db.Open(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database:", err)
		return 1
	}
	defer database.Close()
	dir := getenv("TREST_MIGRATIONS_DIR", "migrations")
	if err := db.Migrate(ctx, database, dir); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		return 1
	}
	fmt.Println("migrations applied")
	return 0
}
