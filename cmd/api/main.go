package main

import (
	"context"
	"flag"
	"github.com/mintfary-oss/trest-sistems/internal/api"
	"github.com/mintfary-oss/trest-sistems/internal/auth"
	"github.com/mintfary-oss/trest-sistems/internal/db"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "apply checked transactional migrations and exit")
	verifyOnly := flag.Bool("verify-migrations", false, "verify migration checksums and exit")
	flag.Parse()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("TREST_AUTH_SECRET")
	}
	if secret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database, err := db.Open(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer database.Close()
	dir := getenv("MIGRATIONS_DIR", "/app/migrations")
	if *verifyOnly || (!*migrateOnly && getenv("SKIP_API_MIGRATIONS", "") == "1") {
		err = db.CheckMigrations(ctx, database, dir)
	} else {
		err = db.Migrate(ctx, database, dir)
	}
	if err != nil {
		log.Fatalf("migrations: %v", err)
	}
	if *migrateOnly || *verifyOnly {
		log.Print("migration check: PASS")
		return
	}
	svc := auth.New(secret)
	if email, pass := os.Getenv("ADMIN_EMAIL"), os.Getenv("ADMIN_PASSWORD"); email != "" && pass != "" {
		var exists bool
		if err = database.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1))`, email).Scan(&exists); err != nil {
			log.Fatal("admin lookup failed: ", err)
		}
		if !exists {
			hash, e := svc.HashPassword(pass)
			if e != nil {
				log.Fatal(e)
			}
			if _, e = database.ExecContext(ctx, `INSERT INTO users(email,name,role,password_hash) VALUES($1,'Administrator','admin',$2) ON CONFLICT(email) DO NOTHING`, email, hash); e != nil {
				log.Fatal("admin creation failed: ", e)
			}
		}
	}
	srv := &http.Server{Addr: getenv("API_ADDR", ":8080"), Handler: api.NewServer(database, svc).Handler(), ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 120 * time.Second, WriteTimeout: 180 * time.Second, IdleTimeout: 60 * time.Second}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-signals
		c, stop := context.WithTimeout(context.Background(), 20*time.Second)
		defer stop()
		_ = srv.Shutdown(c)
	}()
	log.Printf("trest api listening on %s", srv.Addr)
	if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
func getenv(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}
