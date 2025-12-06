package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/service"
	internalvalidator "github.com/namnv2496/scheduler/internal/validator"

	// Import service

	schedulerv1 "github.com/namnv2496/scheduler/pkg/generated/pkg/proto"
	"github.com/namnv2496/scheduler/pkg/logging"
	"github.com/namnv2496/scheduler/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SchedulerEventController struct {
	schedulerv1.UnimplementedSchedulerEventServiceServer
	SchedulerEventService service.ISchedulerEventService
	ratelimter            utils.IRateLimit
	internalvalidator     internalvalidator.IValidate
}

func NewSchedulerEventController(
	SchedulerEventService service.ISchedulerEventService,
	ratelimter utils.IRateLimit,
	internalvalidator internalvalidator.IValidate,
) schedulerv1.SchedulerEventServiceServer {
	return &SchedulerEventController{
		SchedulerEventService: SchedulerEventService,
		ratelimter:            ratelimter,
		internalvalidator:     internalvalidator,
	}
}

func (_self *SchedulerEventController) CreateSchedulerEvent(
	ctx context.Context,
	req *schedulerv1.CreateSchedulerEventRequest,
) (*schedulerv1.CreateSchedulerEventResponse, error) {
	ctx = logging.InjectTraceId(ctx)
	logging.SetName("scheduler")
	ctx = logging.ResetPrefix(ctx, "CreateSchedulerEvent")

	logging.Infof(ctx, "CreateSchedulerEvent is called before")
	// // rate limit
	if err := _self.checkInserRateLimit(ctx, req.Event.Id); err != nil {
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
	}

	logging.Infof(ctx, "CreateSchedulerEvent is called after")
	if req == nil || req.Event == nil {
		return nil, status.Errorf(codes.InvalidArgument, "request or url is nil")
	}
	newEvent := &entity.SchedulerEvent{
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
	eventFields := newEvent.ToMap()
	if err := _self.internalvalidator.ValidateRequire(ctx, "insert", eventFields); err != nil {
		return nil, err
	}

	if err := _self.internalvalidator.ValidateValue(ctx, eventFields); err != nil {
		return nil, err
	}
	if err := _self.internalvalidator.ValidateCustomeRules(eventFields); err != nil {
		return nil, err
	}

	id, err := _self.SchedulerEventService.CreateSchedulerEvent(ctx, newEvent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create url: %v", err)
	}

	return &schedulerv1.CreateSchedulerEventResponse{
		Id:     strconv.FormatInt(id, 10),
		Status: strconv.Itoa(http.StatusCreated),
	}, nil
}

func (_self *SchedulerEventController) GetSchedulerEvents(
	ctx context.Context,
	req *schedulerv1.GetSchedulerEventsRequest,
) (*schedulerv1.GetSchedulerEventsResponse, error) {
	ctx = logging.InjectTraceId(ctx)
	logging.SetName("scheduler")
	ctx = logging.ResetPrefix(ctx, "GetSchedulerEvents")

	logging.Infof(ctx, "GetSchedulerEvents is called")
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

	events, err := _self.SchedulerEventService.GetSchedulerEvents(ctx, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get urls: %v", err)
	}

	SchedulerEvents := make([]*schedulerv1.SchedulerEvent, len(events))
	for i, event := range events {
		SchedulerEvents[i] = &schedulerv1.SchedulerEvent{
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
	return &schedulerv1.GetSchedulerEventsResponse{
		Events: SchedulerEvents,
	}, nil
}

func (_self *SchedulerEventController) UpdateSchedulerEvent(
	ctx context.Context,
	req *schedulerv1.UpdateSchedulerEventRequest,
) (*schedulerv1.UpdateSchedulerEventResponse, error) {
	ctx = logging.InjectTraceId(ctx)
	ctx = logging.ResetPrefix(ctx, "UpdateSchedulerEvent")
	if req == nil || req.Event == nil || req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "request, url, or id is nil/empty")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ID format")
	}
	domainUrl := &entity.SchedulerEvent{
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

	err = _self.SchedulerEventService.UpdateSchedulerEvent(ctx, id, domainUrl)
	if err != nil {
		if errors.Is(err, errors.New("url not found")) { // Check for specific error from repository/service
			return nil, status.Errorf(codes.NotFound, "url with ID %s not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "failed to update url: %v", err)
	}

	return &schedulerv1.UpdateSchedulerEventResponse{
		Id:     req.Id,
		Status: "updated",
	}, nil
}

func (_self *SchedulerEventController) checkRateLimit(ctx context.Context, key string) error {
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

func (_self *SchedulerEventController) checkInserRateLimit(ctx context.Context, key string) error {
	ctx = logging.AppendPrefix(ctx, "checkInserRateLimit")

	logging.Infof(ctx, "rate limit is called")

	// rate limit 50 request/ second you can use LimitSecond(50 /*rate*/, 1)
	pass, err := _self.ratelimter.Allow(ctx, "create_event", key, utils.LimitCustom(50 /*rate*/, 1 /*burst*/, time.Second))
	if err != nil {
		return err
	}
	logging.Infof(ctx, "rate limit is called after")
	if !pass {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (_self *SchedulerEventController) UpdateEventStatus(ctx context.Context, req *schedulerv1.UpdateEventStatusRequest) (*schedulerv1.UpdateEventStatusResponse, error) {
	ctx = logging.InjectTraceId(ctx)
	logging.ResetPrefix(ctx, "UpdateEventStatus")
	logging.Infof(ctx, "update status of event: %s", req)
	err := _self.SchedulerEventService.UpdateEventStatus(ctx, req.Id, domain.StatusEnum(req.Status))
	if err != nil {
		return nil, err
	}
	return &schedulerv1.UpdateEventStatusResponse{}, nil
}
