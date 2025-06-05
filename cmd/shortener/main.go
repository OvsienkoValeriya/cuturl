package main

import (
	"cuturl/internal/app"
	"cuturl/internal/config"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {

	config.Init()
	cfg := config.Get()
	r := chi.NewRouter()

	r.Post("/", app.OrigUrlHandler)

	r.Get("/{id}", app.ShortUrlHandler)

	if err := http.ListenAndServe(cfg.RunAddress, r); err != nil {
		panic(err)
	}
}
