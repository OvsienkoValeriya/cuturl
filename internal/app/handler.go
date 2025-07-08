package app

import (
	"crypto/rand"
	"cuturl/internal/config"
	"cuturl/internal/store"
	"io"
	"math/big"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type URLShortener struct {
	letters []rune
	logger  *zap.SugaredLogger
	repo    store.Repository
}

type Request struct {
	URL string `json:"url"`
}
type Response struct {
	Result string `json:"result"`
}
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewURLShortener(logger *zap.SugaredLogger, repo store.Repository) *URLShortener {
	us := &URLShortener{
		letters: []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		logger:  logger,
		repo:    repo,
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

func (u *URLShortener) getOrCreateShortURL(originalURL string) (string, int, error) {
	existing, err := u.repo.FindByOriginalURL(originalURL)
	if err == nil && existing != nil {
		return existing.ShortURL, http.StatusConflict, nil
	}

	shortID := u.generateShorten(8)

	err = u.repo.Save(store.StoredURL{
		UUID:        shortID,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	})
	if err != nil {
		if existing, findErr := u.repo.FindByOriginalURL(originalURL); findErr == nil && existing != nil {
			return existing.ShortURL, http.StatusConflict, nil
		}
		return "", http.StatusInternalServerError, err
	}
	return shortID, http.StatusCreated, nil
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

	shortID, status, err := u.getOrCreateShortURL(origURL)
	if err != nil {
		u.logger.Errorf("failed to save to file: %v", err)
		http.Error(res, "internal error", http.StatusInternalServerError)
		return
	}
	shortURL := config.Get().BaseURL + "/" + shortID
	res.Header().Set("Content-Type", "text/plain")
	u.logger.Infof("created short url: %s -> %s", shortURL, origURL)
	res.WriteHeader(status)
	res.Write([]byte(shortURL))
}

func (u *URLShortener) ShortURLHandler(res http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	if id == "" {
		http.Error(res, "missing id", http.StatusBadRequest)
		return
	}

	entry, err := u.repo.FindByShortID(id)
	if err != nil {
		u.logger.Errorf("failed to find short ID: %v", err)
		http.Error(res, "not found", http.StatusNotFound)
		return
	}

	res.Header().Set("Location", entry.OriginalURL)
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

	shortID, status, err := u.getOrCreateShortURL(reqBody.URL)
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
	res.WriteHeader(status)
	res.Write(respJSON)
}

func (u *URLShortener) PingHandler(w http.ResponseWriter, r *http.Request) {
	if err := u.repo.Ping(); err != nil {
		u.logger.Errorw("DB ping failed", "error", err)
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (u *URLShortener) ShortenBatchHandler(w http.ResponseWriter, r *http.Request) {
	var batch []BatchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil || len(batch) == 0 {
		http.Error(w, "invalid batch", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	result := make([]BatchResponseItem, 0, len(batch))
	entries := make([]store.StoredURL, 0, len(batch))

	for _, item := range batch {
		if strings.TrimSpace(item.OriginalURL) == "" {
			continue
		}
		shortID := u.generateShorten(8)
		entries = append(entries, store.StoredURL{
			UUID:        shortID,
			ShortURL:    shortID,
			OriginalURL: item.OriginalURL,
		})
		result = append(result, BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      config.Get().BaseURL + "/" + shortID,
		})
	}

	if len(entries) == 0 {
		http.Error(w, "empty valid batch", http.StatusBadRequest)
		return
	}

	if err := u.repo.BatchSave(r.Context(), entries); err != nil {
		u.logger.Errorf("batch save failed: %v", err)
		http.Error(w, "could not save batch", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(result)
}
