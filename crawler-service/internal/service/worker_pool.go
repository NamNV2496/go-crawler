package service

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/namnv2496/crawler/internal/configs"
)

type IWorkerPool interface {
	Crawl(crawlFunc func() (any, error), depth int, statscallback StatsCallback, outputCallback OuputCallback)
}

type WorkerPool struct {
	workers       int
	pagesCrawled  atomic.Int32
	activeWorkers atomic.Int32
	queueSize     atomic.Int32
	waitGroup     sync.WaitGroup
	queue         chan func() (any, error)
}

type StatsCallback func(crawled, active, queued int32)
type OuputCallback func(output any, err error)

func NewWorkerPool(
	conf *configs.Config,
) *WorkerPool {
	return &WorkerPool{
		workers: conf.AppConfig.Workers,
		queue:   make(chan func() (any, error), 1000),
	}
}

func (wp *WorkerPool) Crawl(crawlFunc func() (any, error), depth int, statscallback StatsCallback, outputCallback OuputCallback) {
	// Initialize the pool
	wp.queue <- crawlFunc
	wp.queueSize.Add(1)

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.waitGroup.Add(1)
		go wp.worker(depth, statscallback, outputCallback)
	}

	// Start a goroutine to wait for completion
	go func() {
		wp.waitGroup.Wait()
		close(wp.queue)
	}()
}

func (wp *WorkerPool) worker(depth int, statscallback StatsCallback, outputCallback OuputCallback) {
	defer wp.waitGroup.Done()

	for urlExecute := range wp.queue {
		wp.activeWorkers.Add(1)
		// statscallback(wp.pagesCrawled.Load(), wp.activeWorkers.Load(), wp.queueSize.Load())
		var output any
		var err error
		for retryCount := 0; retryCount <= depth; retryCount++ {
			output, err = urlExecute()
			if err == nil || retryCount == depth {
				break
			}
			fmt.Printf("Retry attempt %d/%d for execution: %v\n", retryCount+1, depth, err)
		}

		if err != nil {
			outputCallback(nil, fmt.Errorf("error executing curl command: %v", err))
		} else {
			outputCallback(output, nil)
		}
		// Process URL
		// links := wp.crawlURL(url, depth)
		// // Add new URLs to queue
		// for _, link := range links {
		// 	if _, visited := wp.visited.LoadOrStore(link, true); !visited {
		// 		wp.queue <- link
		// 		wp.queueSize.Add(1)
		// 	}
		// }

		wp.pagesCrawled.Add(1)
		wp.queueSize.Add(-1)
		wp.activeWorkers.Add(-1)
		// statscallback(wp.pagesCrawled.Load(), wp.activeWorkers.Load(), wp.queueSize.Load())
	}
}
