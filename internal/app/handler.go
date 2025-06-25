package app

import (
	"crypto/rand"
	"cuturl/internal/config"
	"cuturl/internal/store"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLShortener struct {
	shortToOriginal map[string]string
	originalToShort map[string]string
	letters         []rune
	mu              sync.RWMutex
	logger          *zap.SugaredLogger
	repo            store.Repository
}

type Request struct {
	URL string `json:"url"`
}
type Response struct {
	Result string `json:"result"`
}

func NewURLShortener(logger *zap.SugaredLogger, repo store.Repository) *URLShortener {
	us := &URLShortener{
		shortToOriginal: make(map[string]string),
		originalToShort: make(map[string]string),
		letters:         []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		logger:          logger,
		repo:            repo,
	}

	saved, err := repo.Load()
	if err != nil {
		logger.Errorf("failed to load storage: %v", err)
	} else {
		for _, entry := range saved {
			us.shortToOriginal[entry.ShortURL] = entry.OriginalURL
			us.originalToShort[entry.OriginalURL] = entry.ShortURL
		}
	}

	return us
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

func (u *URLShortener) getOrCreateShortURL(originalURL string) (string, error) {
	u.mu.RLock()
	if shortID, ok := u.originalToShort[originalURL]; ok {
		u.mu.RUnlock()
		return shortID, nil
	}
	u.mu.RUnlock()

	shortID := u.generateShorten(8)

	u.mu.Lock()
	u.shortToOriginal[shortID] = originalURL
	u.originalToShort[originalURL] = shortID
	u.mu.Unlock()

	err := u.repo.Save(store.StoredURL{
		UUID:        shortID,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	})
	if err != nil {
		return "", err
	}
	return shortID, nil
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

	shortID, err := u.getOrCreateShortURL(origURL)
	if err != nil {
		u.logger.Errorf("failed to save to file: %v", err)
		http.Error(res, "internal error", http.StatusInternalServerError)
		return
	}
	shortURL := config.Get().BaseURL + "/" + shortID
	res.Header().Set("Content-Type", "text/plain")
	u.logger.Infof("created short url: %s -> %s", shortURL, origURL)
	res.WriteHeader(http.StatusCreated)
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

func (u *URLShortener) OrigURLJSONHandler(res http.ResponseWriter, req *http.Request) {
	var reqBody Request
	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		http.Error(res, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if reqBody.URL == "" {
		u.logger.Errorf("empty URL received")
		http.Error(res, "Empty URL", http.StatusBadRequest)
		return
	}

	shortID, err := u.getOrCreateShortURL(reqBody.URL)
	if err != nil {
		u.logger.Errorf("failed to save to file: %v", err)
		http.Error(res, "internal error", http.StatusInternalServerError)
		return
	}
	shortURL := config.Get().BaseURL + "/" + shortID
	respJSON, err := json.Marshal(Response{Result: shortURL})
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(respJSON)
}
