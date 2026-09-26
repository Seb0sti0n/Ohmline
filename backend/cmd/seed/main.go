// Command seed applies migrations and loads the CSV data plus the demo user.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Seb0sti0n/astrophage/backend/internal/config"
	"github.com/Seb0sti0n/astrophage/backend/internal/db"
	"github.com/Seb0sti0n/astrophage/backend/internal/seed"
)

func main() {
	cfg := config.Load()
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "../data"
	}

	if err := seed.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	fmt.Println("migrations applied")

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	s, err := seed.Run(ctx, pool, dataDir)
	if err != nil {
		log.Fatalf("seed: %v", err)
	}
	fmt.Printf("loaded %d meters, %d readings, %d events\n", s.Meters, s.Readings, s.Events)
	fmt.Printf("demo user: %s / %s\n", seed.DemoEmail, seed.DemoPassword)
	if len(s.Issues) == 0 {
		fmt.Println("data validation: no issues found")
		return
	}
	fmt.Printf("data validation: %d issue(s)\n", len(s.Issues))
	for _, i := range s.Issues {
		fmt.Println("  -", i)
	}
}
