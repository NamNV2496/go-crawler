package service

import (
	"context"

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/repository"
)

type IUrlService interface {
	CreateUrl(ctx context.Context, url *domain.Url) (int64, error)
	GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error)
	UpdateUrl(ctx context.Context, id int64, url *domain.Url) error
}

type UrlService struct {
	repo repository.IUrlRepository
}

// NewUrlService creates a new UrlService instance
func NewUrlService(
	repo repository.IUrlRepository,
) *UrlService {
	return &UrlService{
		repo: repo,
	}
}

// CreateUrl handles the business logic for creating a URL
func (s *UrlService) CreateUrl(ctx context.Context, url *domain.Url) (int64, error) {
	id, err := s.repo.CreateUrl(ctx, url)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetUrls handles the business logic for retrieving URLs
func (s *UrlService) GetUrls(ctx context.Context, limit, offset int32) ([]*domain.Url, error) {
	urls, err := s.repo.GetUrls(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

// UpdateUrl handles the business logic for updating a URL
func (s *UrlService) UpdateUrl(ctx context.Context, id int64, url *domain.Url) error {
	existingUrl, err := s.repo.GetUrlByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Url = url.Url
	existingUrl.Description = url.Description
	existingUrl.Queue = url.Queue
	existingUrl.Domain = url.Domain
	existingUrl.IsActive = url.IsActive
	existingUrl.Method = url.Method

	err = s.repo.UpdateUrl(ctx, existingUrl)
	if err != nil {
		return err
	}

	return nil
}
