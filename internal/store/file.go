package store

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type Storage struct {
	Path      string
	URLS      []StoredURL
	urlsMutex sync.RWMutex
}

type StoredURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorage(path string) (*Storage, error) {
	storage := &Storage{
		Path: path,
		URLS: []StoredURL{},
	}
	if err := storage.Load(); err != nil {
		return nil, err
	}
	return storage, nil
}

func (storage *Storage) Load() error {
	storage.urlsMutex.Lock()
	defer storage.urlsMutex.Unlock()

	file, err := os.OpenFile(storage.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			storage.URLS = append(storage.URLS, entry)
		}
	}
	return scanner.Err()
}

func (storage *Storage) Save(entry StoredURL) error {
	storage.urlsMutex.Lock()
	defer storage.urlsMutex.Unlock()

	storage.URLS = append(storage.URLS, entry)

	file, err := os.OpenFile(storage.Path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(entry)
}
