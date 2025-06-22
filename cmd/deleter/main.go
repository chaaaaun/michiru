package main

import (
	"context"
	"log"

	"michiru/config"
	"michiru/internal/clients"
)

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("FATAL: Could not load configuration.\n%v", err)
	}

	err := clients.ResetIndexes(context.Background(), cfg)
	if err != nil {
		log.Fatalf("FATAL: Could not delete indexes.\n%v", err)
	}
}
