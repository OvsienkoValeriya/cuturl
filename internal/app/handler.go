package app

import (
	"crypto/rand"
	"cuturl/internal/config"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
)

type URLShortener struct {
	shortToOriginal map[string]string
	originalToShort map[string]string
	letters         []rune
	mu              sync.RWMutex
}

func NewURLShortener() *URLShortener {
	return &URLShortener{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
		letters:         []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
	}
}

func (u *URLShortener) generateShorten(n int) string {
	b := make([]rune, n)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(u.letters))))
		if err != nil {
			panic(err)
		}
		b[i] = u.letters[num.Int64()]
	}
	return string(b)
}

func (u *URLShortener) OrigURLHandler(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		log.Println("error reading body:", err)
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	origURL := strings.TrimSpace(string(body))
	if origURL == "" {
		log.Println("empty URL received")
		http.Error(res, "empty url", http.StatusBadRequest)
		return
	}

	u.mu.RLock()
	if shortID, ok := u.originalToShort[origURL]; ok {
		u.mu.RUnlock()
		shortURL := config.Get().BaseURL + "/" + shortID
		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(shortURL))
		return
	}
	u.mu.RUnlock()

	shortID := u.generateShorten(8)

	u.mu.Lock()
	u.shortToOriginal[shortID] = origURL
	u.originalToShort[origURL] = shortID
	u.mu.Unlock()

	shortURL := config.Get().BaseURL + "/" + shortID
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	log.Printf("created short url: %s -> %s", shortID, origURL)
	res.Write([]byte(shortURL))
}

func (u *URLShortener) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		log.Println("invalid short url received: missing 'id' param")
		http.Error(res, "invalid short url", http.StatusBadRequest)
		return
	}

	u.mu.RLock()
	originalURL, ok := u.shortToOriginal[id]
	u.mu.RUnlock()

	if !ok {
		log.Println("short url not found: ")
		http.Error(res, "short url not found", http.StatusNotFound)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
