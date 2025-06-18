package app

import (
	"cuturl/internal/config"
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
		respJson, _ := json.Marshal(Response{Result: config.Get().BaseURL + "/" + shortID})
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		res.Write(respJson)
		return
	}
	u.mu.RUnlock()

	shortID := u.generateShorten(8)
	u.mu.Lock()
	u.shortToOriginal[shortID] = reqBody.URL
	u.originalToShort[reqBody.URL] = shortID
	u.mu.Unlock()
	respJson, err := json.Marshal(Response{Result: config.Get().BaseURL + "/" + shortID})
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	u.logger.Infof("created short url: %s -> %s", shortID, reqBody.URL)
	res.Write(respJson)

}
