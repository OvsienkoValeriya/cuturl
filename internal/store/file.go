package store

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
)

type Repository interface {
	Load() ([]StoredURL, error)
	Save(entry StoredURL) error
	Ping() error
	FindByShortID(id string) (*StoredURL, error)
	FindByOriginalURL(orig string) (*StoredURL, error)
	BatchSave(ctx context.Context, urls []StoredURL) error
	GetURLsByUserID(ctx context.Context, userId string) ([]StoredURL, error)
	MarkDeleted(ctx context.Context, userId string, ids []string) error
}

type FileRepository struct {
	Path      string
	urlsMutex *sync.Mutex
}

type StoredURL struct {
	UUID        string `json:"uuid" db:"uuid"`
	ShortURL    string `json:"short_url" db:"short_url"`
	OriginalURL string `json:"original_url" db:"original_url"`
	UserID      string `json:"user_id" db:"user_id"`
	IsDeleted   bool   `json:"is_deleted" db:"is_deleted"`
}

func NewFileRepository(path string) *FileRepository {
	return &FileRepository{Path: path, urlsMutex: &sync.Mutex{}}
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

func (fr *FileRepository) Ping() error {
	return nil
}

func (fr *FileRepository) FindByShortID(id string) (*StoredURL, error) {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.Open(fr.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			if entry.ShortURL == id {
				return &entry, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (fr *FileRepository) FindByOriginalURL(orig string) (*StoredURL, error) {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.Open(fr.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			if entry.OriginalURL == orig {
				return &entry, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (fr *FileRepository) BatchSave(ctx context.Context, urls []StoredURL) error {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	file, err := os.OpenFile(fr.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var existing []StoredURL
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			existing = append(existing, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	existing = append(existing, urls...)

	tmpPath := fr.Path + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	enc := json.NewEncoder(tmpFile)
	for _, u := range existing {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := enc.Encode(u); err != nil {
			return err
		}
	}

	return os.Rename(tmpPath, fr.Path)

}

func (fr *FileRepository) GetURLsByUserID(ctx context.Context, userID string) ([]StoredURL, error) {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.Open(fr.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var result []StoredURL
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			if entry.UserID == userID {
				result = append(result, entry)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (fr *FileRepository) MarkDeleted(ctx context.Context, userID string, ids []string) error {
	fr.urlsMutex.Lock()
	defer fr.urlsMutex.Unlock()

	file, err := os.OpenFile(fr.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var all []StoredURL
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry StoredURL
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			all = append(all, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	for i, url := range all {
		if url.UserID == userID {
			if _, ok := idSet[url.UUID]; ok {
				all[i].IsDeleted = true
			}
		}
	}

	tmpPath := fr.Path + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	enc := json.NewEncoder(tmpFile)
	for _, entry := range all {
		if err := enc.Encode(entry); err != nil {
			return err
		}
	}

	return os.Rename(tmpPath, fr.Path)
}
