package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/namnv2496/scheduler/internal/configs"
	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/namnv2496/scheduler/internal/repository/distributedlock"
	"github.com/namnv2496/scheduler/internal/service/mq"
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
	log.Printf("Cron job queue %s, every 1 minutes is started", entity.QueueTypeNormal)
	if err != nil {
		return err
	}
	cronJob.Start()
	return nil
}

func (_self *CrawlerCronJob) ExecuteEvent(ctx context.Context) func() {
	return func() {
		log.Println("Acquire and execute event")
		now := time.Now().UnixMilli()
		events, err := _self.crawlerEventRepo.GetCrawlerEventByStatusAndSchedulerAt(ctx, domain.StatusPending, now)
		if err != nil {
			log.Printf("Failed to get crawler events: %v", err)
			return
		}
		updateEvents := make([]*domain.CrawlerEvent, 0)
		for _, event := range events {
			mutex, err := _self.distributedLock.Lock(fmt.Sprint(event.Id), time.Second*10)
			if err != nil {
				log.Printf("Failed to acquire lock for URL %s: %v", event.Url, err)
				continue
			}
			defer mutex.Unlock()
			if err := _self.publishToCrawler(ctx, entity.CrawlerEvent(*event)); err != nil {
				continue
			}
			event.Status = domain.StatusRunning
			if event.RepeatTimes > 0 {
				event.RepeatTimes = event.RepeatTimes - 1
				var nextTime int64
				for i := 1; i <= 10; i++ {
					if event.SchedulerAt+event.NextRunTime*int64(i) > now {
						nextTime = event.NextRunTime * int64(i)
						break
					}
				}
				if nextTime == 0 {
					nextTime = now + event.NextRunTime - event.SchedulerAt
				}
				event.SchedulerAt = event.SchedulerAt + nextTime
			} else {
				event.Status = domain.StatusFailed
			}
			updateEvents = append(updateEvents, event)
		}
		// update events
		log.Printf("update events: %d", len(events))
		if err := _self.crawlerEventRepo.Updates(ctx, updateEvents); err != nil {
			log.Printf("error update events: %s", err)
		}
	}
}

func (_self *CrawlerCronJob) publishToCrawler(ctx context.Context, eventData entity.CrawlerEvent) error {
	err := _self.producers.Publish(ctx, eventData.Queue, strconv.Itoa(int(eventData.Id)), eventData)
	if err != nil {
		// Can apply retry at here
		log.Printf("Publish message to kafka failed: %+v, err: %s", eventData, err)
		return err
	}
	return nil
}
