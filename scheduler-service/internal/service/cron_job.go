package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/namnv2496/scheduler/internal/repository/distributedlock"
	"github.com/namnv2496/scheduler/internal/service/mq"
	"github.com/namnv2496/scheduler/pkg/logging"

	"github.com/robfig/cron/v3"
)

const (
	MaxUrls   = 1000
	MaxQueue  = 100
	MaxWorker = 10
)

type ICrawlerCronJob interface {
	Start() error
}

type CrawlerCronJob struct {
	conf             *configs.Config
	domains          []string
	crawlerEventRepo repository.ICrawlerEventRepository
	distributedLock  distributedlock.IDistributedLock
	producers        mq.IProducer
}

func NewUrlCronJob(
	conf *configs.Config,
	crawlerEventRepo repository.ICrawlerEventRepository,
	distributedLock distributedlock.IDistributedLock,
	producers mq.IProducer,
) ICrawlerCronJob {
	return &CrawlerCronJob{
		conf:             conf,
		domains:          conf.AppConfig.Domains,
		crawlerEventRepo: crawlerEventRepo,
		distributedLock:  distributedLock,
		producers:        producers,
	}
}

func (_self *CrawlerCronJob) Start() error {
	cronJob := cron.New()
	ctx := context.Background()
	_, err := cronJob.AddFunc(
		_self.conf.Cron.CronExpression,
		_self.ExecuteEvent(ctx),
	)
	logging.ResetPrefix(ctx, "Start")
	logging.Infof(ctx, "Cron job queue %s, every 1 minutes is started", entity.QueueTypeNormal)
	if err != nil {
		return err
	}
	cronJob.Start()
	return nil
}

func (_self *CrawlerCronJob) ExecuteEvent(ctx context.Context) func() {
	ctx = logging.AppendPrefix(ctx, "ExecuteEvent")

	return func() {
		logging.Infof(ctx, "Acquire and execute event")
		now := time.Now().UnixMilli()
		events, err := _self.crawlerEventRepo.GetCrawlerEventByStatusAndSchedulerAt(ctx, domain.StatusPending, now)
		if err != nil {
			logging.Errorf(ctx, "Failed to get crawler events: %v", err)
			return
		}

		updateEventsChan := make(chan *domain.CrawlerEvent, len(events))
		semaphore := make(chan struct{}, MaxWorker)
		var wg sync.WaitGroup

		for _, event := range events {
			// separate go routine to work as much as possible
			wg.Add(1)
			go func(e *domain.CrawlerEvent) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				mutex, err := _self.distributedLock.Lock(fmt.Sprint(e.Id), time.Second*10)
				if err != nil {
					logging.Errorf(ctx, "Failed to acquire lock for URL %s: %v", e.Url, err)
					return
				}
				defer mutex.Unlock()

				if err := _self.publishToCrawler(ctx, entity.CrawlerEvent(*e)); err != nil {
					return
				}

				e.Status = domain.StatusRunning
				if e.RepeatTimes > 0 {
					e.RepeatTimes = e.RepeatTimes - 1
					var nextTime int64
					for i := 1; i <= 10; i++ {
						if e.SchedulerAt+e.NextRunTime*int64(i) > now {
							nextTime = e.NextRunTime * int64(i)
							break
						}
					}
					if nextTime == 0 {
						nextTime = now + e.NextRunTime - e.SchedulerAt
					}
					e.SchedulerAt = e.SchedulerAt + nextTime
				} else {
					e.Status = domain.StatusFailed
				}
				updateEventsChan <- e
			}(event)
		}

		go func() {
			wg.Wait()
			close(updateEventsChan)
		}()

		updateEvents := make([]*domain.CrawlerEvent, 0)
		for event := range updateEventsChan {
			updateEvents = append(updateEvents, event)
		}

		logging.Infof(ctx, "update events: %d", len(updateEvents))
		if err := _self.crawlerEventRepo.Updates(ctx, updateEvents); err != nil {
			logging.Errorf(ctx, "error update events: %s", err)
		}
	}
}

func (_self *CrawlerCronJob) publishToCrawler(ctx context.Context, eventData entity.CrawlerEvent) error {
	// deferFunc := logging.AppendPrefix("publishToCrawler")
	// defer deferFunc()
	err := _self.producers.Publish(ctx, eventData.Queue, strconv.Itoa(int(eventData.Id)), eventData)
	if err != nil {
		// Can apply retry at here
		logging.Infof(ctx, "Publish message to kafka failed: %+v, err: %s", eventData, err)
		return err
	}
	return nil
}
