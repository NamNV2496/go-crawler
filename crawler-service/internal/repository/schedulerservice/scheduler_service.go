package schedulerservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/pkg/logging"
	"github.com/sony/gobreaker/v2"
)

type ISchedulerService interface {
	UpdateSchedulerEvent(ctx context.Context, req *entity.UpdateSchedulerEventRequest) error
}

type schedulerService struct {
	client  *http.Client
	host    string
	breaker *gobreaker.CircuitBreaker[int]
}

func NewSchedulerService(conf *configs.Config) ISchedulerService {
	breaker := gobreaker.NewCircuitBreaker[int](gobreaker.Settings{
		Name:    "UpdateSchedulerEventCB",
		Timeout: 10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRate := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRate >= 0.5
		},
	})
	return &schedulerService{
		host: conf.SchedulerService.Host,
		client: &http.Client{
			Timeout: conf.SchedulerService.Timeout,
		},
		breaker: breaker,
	}
}

var _ ISchedulerService = &schedulerService{}

func (_self *schedulerService) UpdateSchedulerEvent(ctx context.Context, req *entity.UpdateSchedulerEventRequest) error {
	deferFunc := logging.AppendPrefix("UpdateSchedulerEvent")
	defer deferFunc()

	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// CB operation → must return (int, error)
	operation := func() (int, error) {
		httpReq, err := http.NewRequestWithContext(
			ctx,
			http.MethodPut,
			_self.host+"/api/v1/events/"+req.Id,
			bytes.NewReader(payload),
		)
		if err != nil {
			return 0, err
		}

		resp, err := _self.client.Do(httpReq)
		if err != nil {
			return 0, err
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// Treat 5xx as circuit-breaker errors
		if resp.StatusCode >= 500 {
			return 0, fmt.Errorf("server error %d: %s", resp.StatusCode, string(body))
		}

		fmt.Println("Status:", resp.Status)
		fmt.Println("Response:", string(body))
		return 1, nil
	}

	// CALL breaker (v2 API uses generics)
	_, err = _self.breaker.Execute(operation)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			fmt.Println("⚠ Circuit breaker OPEN → skip request")
		}
		return err
	}

	return nil
}
