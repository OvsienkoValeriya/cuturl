package store

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
)

type Repository interface {
	Load() ([]StoredURL, error)
	Save(entry StoredURL) error
}

type FileRepository struct {
	Path      string
	urlsMutex sync.RWMutex
}

type StoredURL struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileRepository(path string) *FileRepository {
	return &FileRepository{Path: path}
}

func (fr *FileRepository) Load() ([]StoredURL, error) {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.OpenFile(fr.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var urls []StoredURL
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			urls = append(urls, entry)
		}
	}
	return urls, scanner.Err()
}

func (fr *FileRepository) Save(entry StoredURL) error {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.OpenFile(fr.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var urls []StoredURL
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var e StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &e); err == nil {
			urls = append(urls, e)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	urls = append(urls, entry)

	tmpPath := fr.Path + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(tmpFile)
	for _, u := range urls {
		if err := enc.Encode(u); err != nil {
			tmpFile.Close()
			return err
		}
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, fr.Path)
}
