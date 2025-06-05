package app

import (
	"cuturl/internal/config"
	"io"
	"math/rand"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

var (
	shortToOriginal = make(map[string]string)
	originalToShort = make(map[string]string)
	letters         = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func generateShorten(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func OrigUrlHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	origUrl := strings.TrimSpace(string(body))
	if origUrl == "" {
		http.Error(res, "empty url", http.StatusBadRequest)
		return
	}

	if shortID, ok := originalToShort[origUrl]; ok {
		shortURL := config.Get().BaseURL + shortID
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
		return
	}

	shortID := generateShorten(8)
	shortToOriginal[shortID] = origUrl
	originalToShort[origUrl] = shortID

	shortURL := config.Get().BaseURL + shortID
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}
func ShortUrlHandler(res http.ResponseWriter, req *http.Request) {
	idString := chi.URLParam(req, "id")
	if idString == "" {
		http.Error(res, "invalid short url", http.StatusBadRequest)
		return
	}

	originalURL, ok := shortToOriginal[idString]
	if !ok {
		http.Error(res, "short url not found", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
