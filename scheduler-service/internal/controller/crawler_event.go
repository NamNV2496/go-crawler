package controller

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/service"
	internalvalidator "github.com/namnv2496/scheduler/internal/validator"

	// Import service

	crawlerv1 "github.com/namnv2496/scheduler/pkg/generated/pkg/proto"
	"github.com/namnv2496/scheduler/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CrawlerEventController struct {
	crawlerv1.UnimplementedCrawlerEventServiceServer
	crawlerEventService service.ICrawlerEventService
	ratelimter          utils.IRateLimit
	internalvalidator   internalvalidator.IValidate
}

func NewCrawlerEventController(
	crawlerEventService service.ICrawlerEventService,
	ratelimter utils.IRateLimit,
	internalvalidator internalvalidator.IValidate,
) crawlerv1.CrawlerEventServiceServer {
	return &CrawlerEventController{
		crawlerEventService: crawlerEventService,
		ratelimter:          ratelimter,
		internalvalidator:   internalvalidator,
	}
}

func (_self *CrawlerEventController) CreateCrawlerEvent(
	ctx context.Context,
	req *crawlerv1.CreateCrawlerEventRequest,
) (*crawlerv1.CreateCrawlerEventResponse, error) {

	// // rate limit
	// if err := _self.checkInserRateLimit(ctx, req.Event.Id); err != nil {
	// 	return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
	// }

	if req == nil || req.Event == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request or url is nil")
	}
	newEvent := &entity.CrawlerEvent{
		Url:         req.Event.Url,
		Method:      req.Event.Method,
		Description: req.Event.Description,
		Queue:       req.Event.Queue,
		Domain:      req.Event.Domain,
		IsActive:    true,
		NextRunTime: req.Event.NextRunTime,
		RepeatTimes: req.Event.RepeatTimes,
		SchedulerAt: req.Event.SchedulerAt,
		Status:      domain.StatusPending,
		CronExp:     req.Event.CronExp,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := _self.internalvalidator.ValidateRequire(ctx, "insert", newEvent); err != nil {
		return nil, err
	}

	if err := _self.internalvalidator.ValidateValue(ctx, newEvent); err != nil {
		return nil, err
	}
	// id, err := _self.crawlerEventService.CreateCrawlerEvent(ctx, newEvent)
	// if err != nil {
	// 	return nil, status.Errorf(codes.Internal, "failed to create url: %v", err)
	// }

	return &crawlerv1.CreateCrawlerEventResponse{
		// Id:     strconv.FormatInt(id, 10),
		Id:     "12",
		Status: "created",
	}, nil
}

func (_self *CrawlerEventController) GetCrawlerEvents(
	ctx context.Context,
	req *crawlerv1.GetCrawlerEventsRequest,
) (*crawlerv1.GetCrawlerEventsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request is nil")
	}
	// rate limit
	if err := _self.checkRateLimit(ctx, "test"); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
	}
	if req.Limit == 0 {
		req.Limit = 20
	}

	events, err := _self.crawlerEventService.GetCrawlerEvents(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get urls: %v", err)
	}

	crawlerEvents := make([]*crawlerv1.CrawlerEvent, len(events))
	for i, event := range events {
		crawlerEvents[i] = &crawlerv1.CrawlerEvent{
			Id:          fmt.Sprintf("%d", event.Id),
			Url:         event.Url,
			Method:      event.Method,
			Description: event.Description,
			Queue:       event.Queue,
			Domain:      event.Domain,
			IsActive:    event.IsActive,
			NextRunTime: event.NextRunTime,
			RepeatTimes: event.RepeatTimes,
			SchedulerAt: event.SchedulerAt,
			CronExp:     event.CronExp,
			CreatedAt:   event.CreatedAt.String(),
			UpdatedAt:   event.UpdatedAt.String(),
		}
	}
	return &crawlerv1.GetCrawlerEventsResponse{
		Events: crawlerEvents,
	}, nil
}

func (_self *CrawlerEventController) UpdateCrawlerEvent(
	ctx context.Context,
	req *crawlerv1.UpdateCrawlerEventRequest,
) (*crawlerv1.UpdateCrawlerEventResponse, error) {
	if req == nil || req.Event == nil || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request, url, or id is nil/empty")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	domainUrl := &entity.CrawlerEvent{
		Id:          id,
		Url:         req.Event.Url,
		Method:      req.Event.Method,
		Description: req.Event.Description,
		Queue:       req.Event.Queue,
		Domain:      req.Event.Domain,
		IsActive:    req.Event.IsActive,
		NextRunTime: req.Event.NextRunTime,
		RepeatTimes: req.Event.RepeatTimes,
		SchedulerAt: req.Event.SchedulerAt,
		Status:      domain.GetStatusEnum(req.Event.Status),
		CronExp:     req.Event.CronExp,
	}

	err = _self.crawlerEventService.UpdateCrawlerEvent(ctx, id, domainUrl)
	if err != nil {
		if errors.Is(err, errors.New("url not found")) { // Check for specific error from repository/service
			return nil, status.Errorf(codes.NotFound, "url with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to update url: %v", err)
	}

	return &crawlerv1.UpdateCrawlerEventResponse{
		Id:     req.Id,
		Status: "updated",
	}, nil
}

func (_self *CrawlerEventController) checkRateLimit(ctx context.Context, key string) error {
	// rate limit 10 request/ minute
	pass, err := _self.ratelimter.Allow(ctx, "query_url", key, utils.LimitPerMinute(10, 1))
	if err != nil {
		return err
	}
	if !pass {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (_self *CrawlerEventController) checkInserRateLimit(ctx context.Context, key string) error {
	// rate limit 50 request/ second you can use LimitSecond(50 /*rate*/, 1)
	pass, err := _self.ratelimter.Allow(ctx, "create_event", key, utils.LimitCustom(50 /*rate*/, 1 /*burst*/, time.Second))
	if err != nil {
		return err
	}
	if !pass {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (_self *CrawlerEventController) UpdateEventStatus(ctx context.Context, req *crawlerv1.UpdateEventStatusRequest) (*crawlerv1.UpdateEventStatusResponse, error) {
	log.Printf("update status of event: %s", req)
	err := _self.crawlerEventService.UpdateEventStatus(ctx, req.Id, domain.StatusEnum(req.Status))
	if err != nil {
		return nil, err
	}
	return &crawlerv1.UpdateEventStatusResponse{}, nil
}
