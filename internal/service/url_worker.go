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
)

const (
	MaxWorker = 10
	MaxQueue  = 100
)

type IUrlWorker interface {
	Start() error
}

type UrlWorker struct {
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
		domains:   conf.AppConfig.Domains,
		urlRepo:   urlRepo,
		queueRepo: queueRepo,
		producers: producers,
	}
}

var _ IUrlWorker = &UrlWorker{}

func (w *UrlWorker) Start() error {
	ctx := context.Background()
	numberOfQueue, err := w.queueRepo.CountQueue(ctx)
	if err != nil {
		return err
	}
	for i := 0; i < int(numberOfQueue/MaxQueue)+1; i++ {
		go w.startWorker(ctx, i)
	}
	return nil
}

func (w *UrlWorker) startWorker(ctx context.Context, index int) {
	queues, err := w.queueRepo.GetQueuesByDomain(ctx, w.domains, MaxQueue, int32(index*MaxQueue))
	if err != nil {
		return
	}
	for _, queue := range queues {
		go w.startCrawler(ctx, queue)
	}
}

func (w *UrlWorker) startCrawler(ctx context.Context, queue *domain.Queue) {
	numberOfUrls, err := w.urlRepo.CountUrlByDomainAndQueue(ctx, queue.Domain, queue.Queue)
	if err != nil {
		return
	}
	for i := range int(numberOfUrls/MaxWorker) + 1 {
		go func() {
			urls, err := w.urlRepo.GetUrlByDomainAndQueue(ctx, queue.Domain, queue.Queue, MaxWorker, i*MaxWorker)
			if err != nil {
				return
			}
			for _, url := range urls {
				data := entity.Url{
					Url:         url.Url,
					Description: url.Description,
					Queue:       url.Queue,
					Domain:      url.Domain,
					IsActive:    url.IsActive,
				}
				go func() {
					err := w.producers.Publish(ctx, url.Queue, strconv.Itoa(int(url.Id)), data)
					if err != nil {
						// retry
						log.Println(err)
						return
					}
				}()
			}
		}()
	}
}
