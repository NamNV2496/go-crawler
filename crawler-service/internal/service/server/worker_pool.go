package server

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/namnv2496/crawler/internal/configs"
	"github.com/namnv2496/crawler/internal/pkg/logging"
)

type IWorkerPool interface {
	Execute(crawlFunc func() (any, error), depth int64, statscallback StatsCallback, outputCallback OuputCallback)
	Start(ctx context.Context) error
	Stop()
	// GetServerManager() IServerManager // returns the server manager
}

type workerPool struct {
	workers       int
	pagesCrawled  atomic.Int32
	activeWorkers atomic.Int32
	queueSize     atomic.Int32
	waitGroup     sync.WaitGroup
	queue         chan WorkerJob
	serverManager IServerManager // Server pool for job execution
	done          chan struct{}
	started       atomic.Bool
	executor      IServerExecutor
}

// WorkerJob encapsulates a job with its metadata
type WorkerJob struct {
	crawlFunc      func() (any, error)
	depth          int64
	statsCallback  StatsCallback
	outputCallback OuputCallback
}

type StatsCallback func(crawled, active, queued int32)
type OuputCallback func(output any, err error)

func NewWorkerPool(
	conf *configs.Config,
	executor IServerExecutor,
) IWorkerPool {
	return &workerPool{
		workers:       conf.AppConfig.Workers,
		queue:         make(chan WorkerJob, 1000),
		serverManager: NewServerManager(),
		done:          make(chan struct{}),
		executor:      executor,
	}
}

func (_self *workerPool) Start(ctx context.Context) error {
	if !_self.started.CompareAndSwap(false, true) {
		return fmt.Errorf("worker pool already started")
	}

	deferFunc := logging.AppendPrefix("WorkerPool.Start")
	defer deferFunc()

	// Start all workers once
	for i := 0; i < _self.workers; i++ {
		_self.waitGroup.Add(1)
		go _self.worker(ctx)
	}

	logging.Debug(ctx, "Worker pool started with %d workers", _self.workers)
	return nil
}

func (_self *workerPool) Stop() {
	close(_self.done)
	_self.waitGroup.Wait()
}

func (_self *workerPool) Execute(crawlFunc func() (any, error), depth int64, statscallback StatsCallback, outputCallback OuputCallback) {
	if !_self.started.Load() {
		outputCallback(nil, fmt.Errorf("worker pool not started, call Start() first"))
		return
	}

	_self.queueSize.Add(1)

	select {
	case _self.queue <- WorkerJob{
		crawlFunc:      crawlFunc,
		depth:          depth,
		statsCallback:  statscallback,
		outputCallback: outputCallback,
	}:
		// Job queued successfully
	case <-_self.done:
		outputCallback(nil, fmt.Errorf("worker pool is shutting down"))
		_self.queueSize.Add(-1)
	}
}

func (_self *workerPool) worker(ctx context.Context) {
	defer _self.waitGroup.Done()

	for {
		select {
		case <-_self.done:
			// Shutdown signal received
			return

		case job := <-_self.queue:
			// calculate requestPoint
			requestPoint := calculateRequestPoint()
			// Get available server from heap
			server := _self.serverManager.GetAvailableServerByPoint(requestPoint)
			if server == nil {
				logging.Debug(ctx, "No available servers, retrying after 100ms")
				// Wait and retry
				time.Sleep(100 * time.Millisecond)
				server = _self.serverManager.GetAvailableServerByPoint(requestPoint)
				if server == nil {
					job.outputCallback(nil, fmt.Errorf("no available servers after retry"))
					_self.queueSize.Add(-1)
					continue
				}
			}

			_self.activeWorkers.Add(1)
			if job.statsCallback != nil {
				job.statsCallback(_self.pagesCrawled.Load(), _self.activeWorkers.Load(), _self.queueSize.Load())
			}

			logging.Debug(ctx, "Executing task on server: %s (point: %d)", server.ID, server.Point)
			// execute task with retry
			var output any
			var err error
			for retryCount := int64(0); retryCount <= job.depth; retryCount++ {
				output, err = job.crawlFunc()
				if err == nil || retryCount == job.depth {
					break
				}
				logging.Debug(ctx, "Retry attempt %d/%d for execution on server %s: %v", retryCount+1, job.depth, server.ID, err)
			}

			if err != nil {
				job.outputCallback(nil, fmt.Errorf("error executing on server %s: %v", server.ID, err))
			} else {
				job.outputCallback(output, nil)
			}

			// Return server to pool
			if err := _self.serverManager.ReturnServer(server); err != nil {
				logging.Error(ctx, "failed to return server %s: %v", server.ID, err)
			}

			_self.pagesCrawled.Add(1)
			_self.queueSize.Add(-1)
			_self.activeWorkers.Add(-1)
			if job.statsCallback != nil {
				job.statsCallback(_self.pagesCrawled.Load(), _self.activeWorkers.Load(), _self.queueSize.Load())
			}
		}
	}
}

// GetServerManager returns the server manager for this worker pool
func (_self *workerPool) GetServerManager() IServerManager {
	return _self.serverManager
}

func calculateRequestPoint() int64 {
	return rand.Int63n(20)
}
