package main

import (
	"cuturl/internal/app"
	"cuturl/internal/config"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config.Init()
	cfg := config.Get()

	logger.Info("Starting server", "address", cfg.RunAddress)

	u := app.NewURLShortener()

	r := chi.NewRouter()
	r.Post("/", http.HandlerFunc(u.OrigURLHandler))
	r.Get("/{id}", http.HandlerFunc(u.ShortURLHandler))

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
