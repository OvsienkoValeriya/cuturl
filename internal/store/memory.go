package store

import (
	"context"
	"sync"
)

type InMemoryRepository struct {
	data map[string]StoredURL
	mu   *sync.Mutex
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{data: make(map[string]StoredURL), mu: &sync.Mutex{}}
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

func (r *InMemoryRepository) GetURLsByUserID(ctx context.Context, userID string) ([]StoredURL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var result []StoredURL
	for _, u := range r.data {
		if u.UserID == userID {
			result = append(result, u)
		}
	}
	return result, nil
}

func (r *InMemoryRepository) MarkDeleted(ctx context.Context, userID string, ids []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	for key, entry := range r.data {
		if entry.UserID == userID {
			if _, ok := idSet[entry.UUID]; ok {
				entry.IsDeleted = true
				r.data[key] = entry
			}
		}
	}

	return nil
}
