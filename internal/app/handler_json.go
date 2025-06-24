package app

import (
	"cuturl/internal/config"
	"cuturl/internal/store"
	"encoding/json"
	"net/http"
)

type Request struct {
	URL string `json:"url"`
}
type Response struct {
	Result string `json:"result"`
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

	u.mu.RLock()
	if shortID, ok := u.originalToShort[reqBody.URL]; ok {
		u.mu.RUnlock()
		shortURL := config.Get().BaseURL + "/" + shortID
		respJSON, err := json.Marshal(Response{Result: shortURL})
		if err != nil {
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(respJSON)
		return
	}
	u.mu.RUnlock()

	shortID := u.generateShorten(8)

	u.mu.Lock()
	u.shortToOriginal[shortID] = reqBody.URL
	u.originalToShort[reqBody.URL] = shortID
	u.mu.Unlock()

	err := u.storage.Save(store.StoredURL{
		UUID:        shortID,
		ShortURL:    shortID,
		OriginalURL: reqBody.URL,
	})
	if err != nil {
		u.logger.Errorf("failed to save to file: %v", err)
	}

	shortURL := config.Get().BaseURL + "/" + shortID
	respJSON, err := json.Marshal(Response{Result: shortURL})
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	u.logger.Infof("created short url: %s -> %s", shortURL, reqBody.URL)
	res.Write(respJSON)
}
