package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Seb0sti0n/astrophage/backend/internal/analysis"
	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	apihttp "github.com/Seb0sti0n/astrophage/backend/internal/http"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	store := db.NewStore(pool)
	if err := store.FailStaleRuns(ctx); err != nil { // runs left in progress by a previous process
		log.Fatalf("cleanup: %v", err)
	}
	server := apihttp.NewServer(store, analysis.NewRunner(store, cfg), cfg)

	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, server.Router()))
}
