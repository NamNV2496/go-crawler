package service

import (
	"github.com/namnv2496/crawler/internal/repository"
)

const (
	MaxWorker = 10
	MaxQueue  = 100
)

type IUrlWorker interface {
	Start()
}

type UrlWorker struct {
	urlRepo   repository.IUrlRepository
	queueRepo repository.IQueueRepository
}

func NewUrlWorker(
	urlRepo repository.IUrlRepository,
	queueRepo repository.IQueueRepository,
) *UrlWorker {
	return &UrlWorker{
		urlRepo:   urlRepo,
		queueRepo: queueRepo,
	}
}

var _ IUrlWorker = &UrlWorker{}

func (w *UrlWorker) Start() {
	// ctx := context.Background()
	// numberOfQueue, err := w.queueRepo.CountQueue(ctx)
	// if err != nil {
	// 	return
	// }
	// for i := 0; i < int(numberOfQueue/MaxQueue); i++ {
	// 	go w.startWorker(ctx, i)
	// }
}

// func (w *UrlWorker) startWorker(ctx context.Context, index int) {
// queues, err := w.queueRepo.GetQueues(ctx, MaxQueue, int32(index*MaxQueue))
// if err != nil {
// 	return
// }
// for _, queue := range queues {
// 	go w.startCrawler(ctx, queue)
// }
// }
