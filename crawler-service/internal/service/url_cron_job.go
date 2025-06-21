package service

import (
	"context"
	"log"
	"strconv"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/domain"
	"github.com/namnv2496/crawler/internal/entity"
	"github.com/namnv2496/crawler/internal/repository"
	"github.com/namnv2496/crawler/internal/service/mq"
	"github.com/robfig/cron/v3"
)

const (
	MaxUrls   = 1000
	MaxQueue  = 100
	MaxWorker = 10
)

type IUrlCronJob interface {
	Start() error
}

type UrlCronJob struct {
	conf      *configs.Config
	domains   []string
	urlRepo   repository.IUrlRepository
	queueRepo repository.IQueueRepository
	producers mq.IProducer
}

func NewUrlCronJob(
	conf *configs.Config,
	urlRepo repository.IUrlRepository,
	queueRepo repository.IQueueRepository,
	producers mq.IProducer,
) *UrlCronJob {
	return &UrlCronJob{
		conf:      conf,
		domains:   conf.AppConfig.Domains,
		urlRepo:   urlRepo,
		queueRepo: queueRepo,
		producers: producers,
	}
}

var _ IUrlCronJob = &UrlCronJob{}

func (_self *UrlCronJob) Start() error {
	cronJob := cron.New()
	ctx := context.Background()
	_, err := cronJob.AddFunc(_self.conf.Queue.Normal, func() {
		_self.startJobWithQueue(ctx, entity.QueueTypeNormal)
	})
	log.Printf("Cron job queue %s, every %s minutes is started", entity.QueueTypeNormal, _self.conf.Queue.Normal)
	if err != nil {
		return err
	}
	_, err = cronJob.AddFunc(_self.conf.Queue.Priority, func() {
		_self.startJobWithQueue(ctx, entity.QueueTypePriority)
	})
	log.Printf("Cron job queue %s, every %s minutes is started", entity.QueueTypePriority, _self.conf.Queue.Priority)
	if err != nil {
		return err
	}
	cronJob.Start()
	return nil
}

func (_self *UrlCronJob) startJobWithQueue(ctx context.Context, queueName string) {
	log.Println("start job with queue: ", queueName)
	var numberOfQueues int64
	var reqErr error
	if len(_self.domains) != 0 {
		numberOfQueues, reqErr = _self.queueRepo.CountQueueByDomainsAndQueueName(ctx, _self.domains, queueName)
	} else {
		numberOfQueues, reqErr = _self.queueRepo.CountQueueByQueueName(ctx, queueName)
	}
	if reqErr != nil {
		return
	}
	for i := range int(numberOfQueues/MaxUrls) + 1 {
		var queueInfors []*domain.Queue
		var reqErr error
		if len(_self.domains) != 0 {
			queueInfors, reqErr = _self.queueRepo.GetQueuesByDomainsAndQueueName(ctx, _self.domains, queueName, MaxQueue, int32(i*MaxQueue))
		} else {
			queueInfors, reqErr = _self.queueRepo.GetQueuesByQueueName(ctx, queueName, MaxQueue, int32(i*MaxQueue))
		}
		if reqErr != nil {
			return
		}
		go _self.getUrlByQueueInfo(ctx, queueInfors)
	}
}

func (_self *UrlCronJob) getUrlByQueueInfo(ctx context.Context, queueInfors []*domain.Queue) {
	// NEED TO IMPROVE: RC: Use query inside for loop
	for _, queue := range queueInfors {
		go func() {
			urls, err := _self.urlRepo.GetUrlByDomainAndQueue(ctx, queue.Domain, queue.Queue, int(queue.Quantity), 0)
			if err != nil {
				return
			}
			for _, url := range urls {
				urlData := entity.Url{
					Url:         url.Url,
					Method:      url.Method,
					Description: url.Description,
					Queue:       url.Queue,
					Quantity:    queue.Quantity,
					Domain:      url.Domain,
					IsActive:    url.IsActive,
				}
				log.Printf("publish to crawler queue: %s, request: %+v", url.Queue, url)
				_self.publishToCrawler(ctx, urlData)
			}
		}()
	}
}

func (_self *UrlCronJob) publishToCrawler(ctx context.Context, urlData entity.Url) {
	err := _self.producers.Publish(ctx, urlData.Queue, strconv.Itoa(int(urlData.Id)), urlData)
	if err != nil {
		// Can apply retry at here
		log.Printf("Publish message to kafka failed: %+v, err: %s", urlData, err)
		return
	}
}
