package service

import (
	"context"
	"cuturl/internal/store"
)

type URLService struct {
	repo store.Repository
}

func NewURLService(repo store.Repository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) SaveURL(ctx context.Context, url store.StoredURL) error {
	return s.repo.Save(url)
}

func (s *URLService) GetByShortID(ctx context.Context, id string) (*store.StoredURL, error) {
	return s.repo.FindByShortID(id)
}

func (s *URLService) GetByOriginalURL(ctx context.Context, url string) (*store.StoredURL, error) {
	return s.repo.FindByOriginalURL(url)
}

func (s *URLService) GetUserURLs(ctx context.Context, userID string) ([]store.StoredURL, error) {
	return s.repo.GetURLsByUserID(ctx, userID)
}

func (s *URLService) MarkDeleted(ctx context.Context, userID string, ids []string) error {
	return s.repo.MarkDeleted(ctx, userID, ids)
}

func (s *URLService) BatchSave(ctx context.Context, urls []store.StoredURL) error {
	return s.repo.BatchSave(ctx, urls)
}
