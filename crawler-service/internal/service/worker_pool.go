package service

import (
	"sync"
	"sync/atomic"
)

type IWorkerPool interface {
	Start(startURL string, depth int, callback StatsCallback)
}

type WorkerPool struct {
	workers       int
	pagesCrawled  atomic.Int32
	activeWorkers atomic.Int32
	queueSize     atomic.Int32
	waitGroup     sync.WaitGroup
	queue         chan string
	visited       sync.Map
}

type StatsCallback func(crawled, active, queued int32)

func NewWorkerPool(workers int) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		queue:   make(chan string, 1000),
	}
}

func (wp *WorkerPool) Start(startURL string, depth int, callback StatsCallback) {
	// Initialize the pool
	wp.queue <- startURL
	wp.queueSize.Add(1)

	// Start workers
	for i := 0; i < wp.workers; i++ {
		wp.waitGroup.Add(1)
		go wp.worker(depth, callback)
	}

	// Start a goroutine to wait for completion
	go func() {
		wp.waitGroup.Wait()
		close(wp.queue)
	}()
}

func (wp *WorkerPool) worker(depth int, callback StatsCallback) {
	defer wp.waitGroup.Done()

	for url := range wp.queue {
		wp.activeWorkers.Add(1)
		callback(wp.pagesCrawled.Load(), wp.activeWorkers.Load(), wp.queueSize.Load())

		// Process URL
		links := wp.crawlURL(url, depth)

		// Add new URLs to queue
		for _, link := range links {
			if _, visited := wp.visited.LoadOrStore(link, true); !visited {
				wp.queue <- link
				wp.queueSize.Add(1)
			}
		}

		wp.pagesCrawled.Add(1)
		wp.queueSize.Add(-1)
		wp.activeWorkers.Add(-1)
		callback(wp.pagesCrawled.Load(), wp.activeWorkers.Load(), wp.queueSize.Load())
	}
}

func (wp *WorkerPool) crawlURL(url string, depth int) []string {
	return []string{}
}
