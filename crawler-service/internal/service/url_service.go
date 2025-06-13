package service

import (
	"context"
	"fmt"

	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/repository/cache"
	"github.com/namnv2496/crawler/pkg/utils"
)

type IUrlService interface {
	CreateUrl(ctx context.Context, url *entity.Url) (int64, error)
	GetUrls(ctx context.Context, limit, offset int32) ([]*entity.Url, error)
	UpdateUrl(ctx context.Context, id int64, url *entity.Url) error
}

type UrlService struct {
	repo  repository.IUrlRepository
	cache cache.ICache[entity.Url]
}

// NewUrlService creates a new UrlService instance
func NewUrlService(
	repo repository.IUrlRepository,
) *UrlService {
	return &UrlService{
		repo: repo,
	}
}

func (_self *UrlService) CreateUrl(ctx context.Context, url *entity.Url) (int64, error) {
	var request *domain.Url
	err := utils.Copy(&request, url)
	if err != nil {
		return 0, err
	}
	id, err := _self.repo.CreateUrl(ctx, request)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (_self *UrlService) GetUrls(ctx context.Context, limit, offset int32) ([]*entity.Url, error) {
	var resp []*entity.Url
	// caching. Actually not necessary
	if _self.cache != nil {
		urls, err := _self.cache.Get(ctx, fmt.Sprintf("%d:%d", limit, offset))
		if err == nil {
			return resp, nil
		}
		err = utils.Copy(resp, urls)
		if err != nil {
			return nil, err
		}
		return resp, nil
	}
	urls, err := _self.repo.GetUrls(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, url := range urls {
		var elem *entity.Url
		err = utils.Copy(&elem, url)
		if err != nil {
			return nil, err
		}
		resp = append(resp, elem)
	}
	return resp, nil
}

func (_self *UrlService) UpdateUrl(ctx context.Context, id int64, url *entity.Url) error {
	existingUrl, err := _self.repo.GetUrlByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Url = url.Url
	existingUrl.Description = url.Description
	existingUrl.Queue = url.Queue
	existingUrl.IsActive = url.IsActive
	existingUrl.Method = url.Method

	err = _self.repo.UpdateUrl(ctx, existingUrl)
	if err != nil {
		return err
	}
	return nil
}
