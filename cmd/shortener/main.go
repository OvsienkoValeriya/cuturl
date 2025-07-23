package main

import (
	"cuturl/internal/app"
	"cuturl/internal/auth"
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

	var repo store.Repository
	if cfg.DBConnection != "" {
		db, err := store.NewPostgresRepository(cfg.DBConnection)
		if err == nil {
			repo = db
			log.Println("Using PostgreSQL as storage")
		} else {
			log.Printf("Postgres connection failed: %v", err)
		}
	} else {
		log.Println("No DB connection string set; falling back to file or memory storage")
	}

	if repo == nil && cfg.FileStoragePath != "" {
		repo = store.NewFileRepository(cfg.FileStoragePath)
		log.Println("Using FileStorage as storage")
	}

	if repo == nil {
		repo = store.NewInMemoryRepository()
		log.Println("Using Memory as storage")
	}

	auth.Init(cfg.AuthSecret)

	u := app.NewURLShortener(sugar, repo)
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(sugar))
	r.Use(middleware.GzipCompressMiddleware)
	r.Use(middleware.GzipDecompressMiddleware)
	r.Use(middleware.AuthMiddleware)
	r.Post("/", http.HandlerFunc(u.OrigURLHandler))
	r.Get("/{id}", http.HandlerFunc(u.ShortURLHandler))
	r.Get("/ping", http.HandlerFunc(u.PingHandler))

	r.Route("/api/shorten", func(r chi.Router) {
		r.Post("/", http.HandlerFunc(u.OrigURLJSONHandler))
		r.Post("/batch", http.HandlerFunc(u.ShortenBatchHandler))
	})

	r.Route("/api/user", func(r chi.Router) {
		r.Get("/urls", http.HandlerFunc(u.UserURLsHandler))
		r.Delete("/urls", http.HandlerFunc(u.DeleteUserURLSHandler))
	})

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
