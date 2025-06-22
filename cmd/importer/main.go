package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"michiru/config"
	"michiru/handlers"
	"michiru/internal/clients"
)

func main() {
	// Setup signal context for graceful shutdown
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	// Load env
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("FATAL: Could not load configuration.\n%v", err)
	}

	// Initialise Meilisearch indexes if they don't exist
	if err := clients.InitIndexes(ctx, cfg); err != nil {
		log.Fatalf("FATAL: Could not initialise Meilisearch.\n%v", err)
	}

	pastMeta, err := clients.GetMetadata(ctx, cfg)
	if err != nil {
		log.Fatalf("FATAL: Could not get metadata from Meilisearch.\n%v", err)
	}

	if err = handlers.ValidateImportInterval(pastMeta); err != nil {
		log.Fatalf("FATAL: Could not validate import interval.\n%v", err)
	}

	b, err := handlers.FetchDump(ctx, cfg)
	if err != nil {
		log.Fatalf("FATAL: Could not fetch title dump.\n%v", err)
	}

	if err = handlers.ValidateXml(b); err != nil {
		log.Fatalf("FATAL: Could not validate title dump.\n%v", err)
	}

	anime, meta, err := handlers.ParseDump(b)
	if err != nil {
		log.Fatalf("FATAL: Could not parse title dump.\n%v", err)
	}

	if err = clients.AddAnime(ctx, cfg, anime); err != nil {
		log.Fatalf("FATAL: Could not add anime to Meilisearch.\n%v", err)
	}

	if err = clients.UpdateMetadata(ctx, cfg, meta); err != nil {
		log.Fatalf("FATAL: Could not update metadata in Meilisearch.\n%v", err)
	}
}
