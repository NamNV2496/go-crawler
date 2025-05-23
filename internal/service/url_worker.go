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

type IUrlWorker interface {
	Start() error
}

type UrlWorker struct {
	conf      *configs.Config
	domains   []string
	urlRepo   repository.IUrlRepository
	queueRepo repository.IQueueRepository
	producers mq.IProducer
}

func NewUrlWorker(
	conf *configs.Config,
	urlRepo repository.IUrlRepository,
	queueRepo repository.IQueueRepository,
	producers mq.IProducer,
) *UrlWorker {
	return &UrlWorker{
		conf:      conf,
		domains:   conf.AppConfig.Domains,
		urlRepo:   urlRepo,
		queueRepo: queueRepo,
		producers: producers,
	}
}

var _ IUrlWorker = &UrlWorker{}

func (w *UrlWorker) Start() error {
	cronJob := cron.New()
	ctx := context.Background()
	_, err := cronJob.AddFunc(w.conf.Queue.Normal, func() {
		w.startWorker(ctx, "normal")
	})
	if err != nil {
		return err
	}
	_, err = cronJob.AddFunc(w.conf.Queue.Priority, func() {
		w.startWorker(ctx, "priority")
	})
	if err != nil {
		return err
	}
	cronJob.Start()
	return nil
}

func (w *UrlWorker) startWorker(ctx context.Context, queue string) {
	log.Println("start worker: ", queue)
	numberOfQueues, err := w.queueRepo.CountQueueByDomainsAndQueue(ctx, w.domains, queue)
	if err != nil {
		return
	}
	for i := range int(numberOfQueues/MaxUrls) + 1 {
		queues, err := w.queueRepo.GetQueuesByDomainsAndQueue(ctx, w.domains, queue, MaxQueue, int32(i*MaxQueue))
		if err != nil {
			return
		}
		go w.publishToCrawler(ctx, queues)
	}
}

func (w *UrlWorker) publishToCrawler(ctx context.Context, queues []*domain.Queue) {
	domains := make([]string, len(queues))
	queueUrls := make([]string, len(queues))
	for i, queue := range queues {
		domains[i] = queue.Domain
		queueUrls[i] = queue.Queue
	}

	numberOfUrls, err := w.urlRepo.CountUrlByDomainsAndQueues(ctx, domains, queueUrls)
	if err != nil {
		return
	}
	for i := range int(numberOfUrls/MaxUrls) + 1 {
		go func() {
			urls, err := w.urlRepo.GetUrlByDomainsAndQueues(ctx, domains, queueUrls, MaxWorker, i*MaxWorker)
			if err != nil {
				return
			}
			for _, url := range urls {
				data := entity.Url{
					Url:         url.Url,
					Method:      url.Method,
					Description: url.Description,
					Queue:       url.Queue,
					Domain:      url.Domain,
					IsActive:    url.IsActive,
				}
				log.Printf("publish to crawler queue: %s, url: %s", url.Queue, url.Url)
				err := w.producers.Publish(ctx, url.Queue, strconv.Itoa(int(url.Id)), data)
				if err != nil {
					// retry
					log.Println(err)
					return
				}
			}
		}()
	}
}
