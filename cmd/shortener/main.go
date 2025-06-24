package main

import (
	"cuturl/internal/app"
	"cuturl/internal/config"
	"cuturl/internal/middleware"
	"cuturl/internal/store"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	config.Init()
	cfg := config.Get()

	log.Println("Starting server on", cfg.RunAddress)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	sugar := logger.Sugar()
	defer logger.Sync()

	storage, err := store.NewStorage(cfg.FileStoragePath)

	u := app.NewURLShortener(sugar, storage)
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(sugar))
	r.Use(middleware.GzipCompressMiddleware)
	r.Use(middleware.GzipDecompressMiddleware)
	r.Post("/", http.HandlerFunc(u.OrigURLHandler))
	r.Get("/{id}", http.HandlerFunc(u.ShortURLHandler))
	r.Post("/api/shorten", http.HandlerFunc(u.OrigURLJSONHandler))

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
