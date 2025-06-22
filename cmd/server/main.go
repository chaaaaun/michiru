package main

import (
	"log"
	"net/http"

	"michiru/config"
	"michiru/handlers"
)

func main() {
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		log.Fatalf("FATAL: Could not load configuration.\n%v", err)
	}

	fs := http.FileServer(http.Dir(cfg.WebUIPath))

	http.Handle("GET /", fs)
	http.HandleFunc("GET /search", handlers.HandleSearch(cfg))
	http.HandleFunc("GET /metadata", handlers.HandleMetadata(cfg))
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
