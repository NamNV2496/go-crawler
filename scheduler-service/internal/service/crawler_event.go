package service

import (
	"context"
	"fmt"

	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/namnv2496/scheduler/internal/repository/cache"
	"github.com/namnv2496/scheduler/pkg/utils"
)

type ICrawlerEventService interface {
	CreateCrawlerEvent(ctx context.Context, crawlerEvent *entity.CrawlerEvent) (int64, error)
	GetCrawlerEvents(ctx context.Context, limit, offset int32) ([]*entity.CrawlerEvent, error)
	UpdateCrawlerEvent(ctx context.Context, id int64, crawlerEvent *entity.CrawlerEvent) error
	UpdateEventStatus(ctx context.Context, id int64, status domain.StatusEnum) error
}

type CrawlerEventService struct {
	repo  repository.ICrawlerEventRepository
	cache cache.ICache[entity.CrawlerEvent]
}

func NewCrawlerEventService(
	repo repository.ICrawlerEventRepository,
) *CrawlerEventService {
	return &CrawlerEventService{
		repo: repo,
	}
}

func (_self *CrawlerEventService) CreateCrawlerEvent(ctx context.Context, crawlerEvent *entity.CrawlerEvent) (int64, error) {
	var request domain.CrawlerEvent
	err := utils.Copy(&request, crawlerEvent)
	if err != nil {
		return 0, err
	}
	id, err := _self.repo.CreateCrawlerEvent(ctx, &request)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (_self *CrawlerEventService) GetCrawlerEvents(ctx context.Context, limit, offset int32) ([]*entity.CrawlerEvent, error) {
	var resp []*entity.CrawlerEvent
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
	urls, err := _self.repo.GetCrawlerEvents(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, crawlerEvent := range urls {
		var elem entity.CrawlerEvent
		err = utils.Copy(&elem, crawlerEvent)
		if err != nil {
			return nil, err
		}
		resp = append(resp, &elem)
	}
	return resp, nil
}

func (_self *CrawlerEventService) UpdateCrawlerEvent(ctx context.Context, id int64, crawlerEvent *entity.CrawlerEvent) error {
	existingUrl, err := _self.repo.GetCrawlerEventByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Url = crawlerEvent.Url
	existingUrl.Description = crawlerEvent.Description
	existingUrl.Queue = crawlerEvent.Queue
	existingUrl.IsActive = crawlerEvent.IsActive
	existingUrl.Method = crawlerEvent.Method
	existingUrl.CronExp = crawlerEvent.CronExp
	existingUrl.NextRunTime = crawlerEvent.NextRunTime
	existingUrl.RepeatTimes = crawlerEvent.RepeatTimes
	existingUrl.SchedulerAt = crawlerEvent.SchedulerAt

	err = _self.repo.UpdateCrawlerEvent(ctx, existingUrl)
	if err != nil {
		return err
	}
	return nil
}

func (_self *CrawlerEventService) UpdateEventStatus(ctx context.Context, id int64, status domain.StatusEnum) error {
	existingUrl, err := _self.repo.GetCrawlerEventByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Status = status

	err = _self.repo.UpdateCrawlerEvent(ctx, existingUrl)
	if err != nil {
		return err
	}
	return nil
}
