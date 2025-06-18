package app

import (
	"crypto/rand"
	"cuturl/internal/config"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLShortener struct {
	shortToOriginal map[string]string
	originalToShort map[string]string
	letters         []rune
	mu              sync.RWMutex
	logger          *zap.SugaredLogger
}

func NewURLShortener(logger *zap.SugaredLogger) *URLShortener {
	return &URLShortener{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
		letters:         []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		logger:          logger,
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
		u.logger.Errorf("Error reading request body: %v", err)
		http.Error(res, "invalid request body", http.StatusBadRequest)
		return
	}

	origURL := strings.TrimSpace(string(body))
	if origURL == "" {
		u.logger.Errorf("Empty URL received")
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
	u.logger.Infof("created short url: %s -> %s", shortID, origURL)
	res.Write([]byte(shortURL))
}

func (u *URLShortener) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		u.logger.Error("invalid short url received: missing 'id' param")
		http.Error(res, "invalid short url", http.StatusBadRequest)
		return
	}

	u.mu.RLock()
	originalURL, ok := u.shortToOriginal[id]
	u.mu.RUnlock()

	if !ok {
		u.logger.Error("short url not found for id: " + id)
		http.Error(res, "short url not found", http.StatusNotFound)
		return
	}

	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
