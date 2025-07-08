package store

import (
	"context"
	"sync"
)

type InMemoryRepository struct {
	data map[string]StoredURL
	mu   sync.Mutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]StoredURL)}
}

func (r *InMemoryRepository) Load() ([]StoredURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []StoredURL
	for _, entry := range r.data {
		result = append(result, entry)
	}
	return result, nil
}

func (r *InMemoryRepository) Save(entry StoredURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[entry.ShortURL] = entry
	return nil
}

func (r *InMemoryRepository) FindByShortID(id string) (*StoredURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if entry, ok := r.data[id]; ok {
		return &entry, nil
	}
	return nil, nil
}

func (r *InMemoryRepository) FindByOriginalURL(orig string) (*StoredURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.data {
		if entry.OriginalURL == orig {
			return &entry, nil
		}
	}
	return nil, nil
}

func (r *InMemoryRepository) Ping() error {
	return nil
}

func (r *InMemoryRepository) BatchSave(ctx context.Context, urls []StoredURL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, entry := range urls {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		r.data[entry.ShortURL] = entry
	}

	return nil
}
