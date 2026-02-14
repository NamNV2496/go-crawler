package repository

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/namnv2496/scheduler/internal/domain"
)

type StatusUpdate struct {
	ID     int64
	Status domain.StatusEnum
}

type IBatchStatusUpdater interface {
	UpdateStatus(ctx context.Context, id int64, status domain.StatusEnum) error
	Flush(ctx context.Context) error
	Close()
}

type BatchStatusUpdater struct {
	repo          ISchedulerEventRepository
	batchSize     int
	flushInterval time.Duration
	buffer        []StatusUpdate
	mu            sync.Mutex
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewBatchStatusUpdater(
	repo ISchedulerEventRepository,
	batchSize int,
	flushInterval time.Duration,
) IBatchStatusUpdater {
	bsu := &BatchStatusUpdater{
		repo:          repo,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]StatusUpdate, 0, batchSize),
		stopCh:        make(chan struct{}),
	}

	bsu.wg.Add(1)
	go bsu.flushWorker()

	return bsu
}

func (bsu *BatchStatusUpdater) UpdateStatus(ctx context.Context, id int64, status domain.StatusEnum) error {
	bsu.mu.Lock()
	bsu.buffer = append(bsu.buffer, StatusUpdate{ID: id, Status: status})
	shouldFlush := len(bsu.buffer) >= bsu.batchSize
	bsu.mu.Unlock()

	if shouldFlush {
		return bsu.Flush(ctx)
	}

	return nil
}

func (bsu *BatchStatusUpdater) flushWorker() {
	defer bsu.wg.Done()
	ticker := time.NewTicker(bsu.flushInterval)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		select {
		case <-ticker.C:
			bsu.Flush(ctx)
		case <-bsu.stopCh:
			bsu.Flush(ctx)
			return
		}
	}
}

func (bsu *BatchStatusUpdater) Flush(ctx context.Context) error {
	bsu.mu.Lock()
	if len(bsu.buffer) == 0 {
		bsu.mu.Unlock()
		return nil
	}

	toFlush := bsu.buffer
	bsu.buffer = make([]StatusUpdate, 0, bsu.batchSize)
	bsu.mu.Unlock()

	// Group by status for efficient batch updates
	eventsByStatus := make(map[domain.StatusEnum][]int64)
	for _, update := range toFlush {
		eventsByStatus[update.Status] = append(eventsByStatus[update.Status], update.ID)
	}

	// Batch update using raw SQL for efficiency
	for status, ids := range eventsByStatus {
		if err := bsu.repo.(*SchedulerEventRepository).GetDB().
			WithContext(ctx).
			Table("scheduler_events").
			Where("id IN ?", ids).
			Update("status", status).Error; err != nil {
			log.Printf("[ERROR] batch status update failed: status=%s, count=%d, error=%v", status, len(ids), err)

			// Fallback: update one-by-one
			for _, id := range ids {
				event, err := bsu.repo.GetSchedulerEventByID(ctx, id)
				if err != nil {
					log.Printf("[ERROR] failed to get event %d: %v", id, err)
					continue
				}
				event.Status = status
				if err := bsu.repo.UpdateSchedulerEvent(ctx, event); err != nil {
					log.Printf("[ERROR] failed to update event %d: %v", id, err)
				}
			}
		} else {
			log.Printf("[INFO] batch updated %d events to status %s", len(ids), status)
		}
	}

	return nil
}

func (bsu *BatchStatusUpdater) Close() {
	close(bsu.stopCh)
	bsu.wg.Wait()
}
