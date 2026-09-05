package main

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mintfary-oss/trest-sistems/internal/bim/queue"
	"log"
	"net/http"
	"os"
	"time"
)

type health struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(health{"ok", "worker", time.Now().UTC().Format(time.RFC3339)})
	})
	addr := os.Getenv("WORKER_ADDR")
	if addr == "" {
		addr = ":8090"
	}
	go runQueue()
	log.Printf("trest worker health listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func runQueue() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return
	}
	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Printf("bim queue disabled: %v", err)
		return
	}
	p := &queue.Processor{DB: db, WorkDir: os.Getenv("BIM_WORKDIR")}
	go func() {
		defer db.Close()
		for {
			ok, err := p.RunOnce(context.Background())
			if err != nil {
				log.Printf("bim queue: %v", err)
			}
			if !ok {
				time.Sleep(2 * time.Second)
			}
		}
	}()
}
