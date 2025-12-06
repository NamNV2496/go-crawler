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

type ISchedulerEventService interface {
	CreateSchedulerEvent(ctx context.Context, SchedulerEvent *entity.SchedulerEvent) (int64, error)
	GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*entity.SchedulerEvent, error)
	UpdateSchedulerEvent(ctx context.Context, id int64, SchedulerEvent *entity.SchedulerEvent) error
	UpdateEventStatus(ctx context.Context, id int64, status domain.StatusEnum) error
}

type SchedulerEventService struct {
	repo  repository.ISchedulerEventRepository
	cache cache.ICache[entity.SchedulerEvent]
}

func NewSchedulerEventService(
	repo repository.ISchedulerEventRepository,
) *SchedulerEventService {
	return &SchedulerEventService{
		repo: repo,
	}
}

func (_self *SchedulerEventService) CreateSchedulerEvent(ctx context.Context, SchedulerEvent *entity.SchedulerEvent) (int64, error) {
	var request domain.SchedulerEvent
	err := utils.Copy(&request, SchedulerEvent)
	if err != nil {
		return 0, err
	}
	id, err := _self.repo.CreateSchedulerEvent(ctx, &request)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (_self *SchedulerEventService) GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*entity.SchedulerEvent, error) {
	var resp []*entity.SchedulerEvent
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
	urls, err := _self.repo.GetSchedulerEvents(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	for _, SchedulerEvent := range urls {
		var elem entity.SchedulerEvent
		err = utils.Copy(&elem, SchedulerEvent)
		if err != nil {
			return nil, err
		}
		resp = append(resp, &elem)
	}
	return resp, nil
}

func (_self *SchedulerEventService) UpdateSchedulerEvent(ctx context.Context, id int64, SchedulerEvent *entity.SchedulerEvent) error {
	existingUrl, err := _self.repo.GetSchedulerEventByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Url = SchedulerEvent.Url
	existingUrl.Description = SchedulerEvent.Description
	existingUrl.Queue = SchedulerEvent.Queue
	existingUrl.IsActive = SchedulerEvent.IsActive
	existingUrl.Method = SchedulerEvent.Method
	existingUrl.CronExp = SchedulerEvent.CronExp
	existingUrl.NextRunTime = SchedulerEvent.NextRunTime
	existingUrl.RepeatTimes = SchedulerEvent.RepeatTimes
	existingUrl.SchedulerAt = SchedulerEvent.SchedulerAt

	err = _self.repo.UpdateSchedulerEvent(ctx, existingUrl)
	if err != nil {
		return err
	}
	return nil
}

func (_self *SchedulerEventService) UpdateEventStatus(ctx context.Context, id int64, status domain.StatusEnum) error {
	existingUrl, err := _self.repo.GetSchedulerEventByID(ctx, id)
	if err != nil {
		return err
	}
	existingUrl.Status = status

	err = _self.repo.UpdateSchedulerEvent(ctx, existingUrl)
	if err != nil {
		return err
	}
	return nil
}
