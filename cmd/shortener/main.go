package main

import (
	"cuturl/internal/app"
	"cuturl/internal/config"
	"cuturl/internal/middleware"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	config.Init()
	cfg := config.Get()

	log.Println("Starting server on", cfg.RunAddress)

	u := app.NewURLShortener()

	r := chi.NewRouter()
	r.Use(middleware.LoggingWiddleware)
	r.Post("/", http.HandlerFunc(u.OrigURLHandler))
	r.Get("/{id}", http.HandlerFunc(u.ShortURLHandler))

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
