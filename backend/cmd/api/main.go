package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	apihttp "github.com/Seb0sti0n/astrophage/backend/internal/http"
)

func main() {
	cfg := config.Load()
	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, apihttp.NewRouter(pool, cfg.CORSOrigin)))
}
