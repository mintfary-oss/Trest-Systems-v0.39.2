package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mintfary-oss/trest-sistems/internal/api"
	"github.com/mintfary-oss/trest-sistems/internal/auth"
	"github.com/mintfary-oss/trest-sistems/internal/db"
)

func runAPI() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	database, err := db.Open(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "database:", err)
		return 1
	}
	defer database.Close()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("TREST_AUTH_SECRET")
	}
	if secret == "" {
		fmt.Fprintln(os.Stderr, "JWT_SECRET is required")
		return 1
	}
	srv := &http.Server{Addr: getenv("TREST_API_ADDR", ":8081"), Handler: api.NewServer(database, auth.New(secret)).Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = srv.Shutdown(shutdownCtx)
	}()
	fmt.Println("trest API listening on", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
