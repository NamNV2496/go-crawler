package cache

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheOperation struct {
	OpType string // "SET" or "DEL"
	Key    string
	Value  interface{}
	TTL    time.Duration
}

type IBatchRedisCache interface {
	SetAsync(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	DeleteAsync(ctx context.Context, key string) error
	Flush(ctx context.Context) error
	Close()
}

type BatchRedisCache struct {
	client        *redis.Client
	batchSize     int
	flushInterval time.Duration
	buffer        []CacheOperation
	mu            sync.Mutex
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func NewBatchRedisCache(
	client *redis.Client,
	batchSize int,
	flushInterval time.Duration,
) IBatchRedisCache {
	brc := &BatchRedisCache{
		client:        client,
		batchSize:     batchSize,
		flushInterval: flushInterval,
		buffer:        make([]CacheOperation, 0, batchSize),
		stopCh:        make(chan struct{}),
	}

	brc.wg.Add(1)
	go brc.flushWorker()

	return brc
}

func (brc *BatchRedisCache) SetAsync(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	brc.mu.Lock()
	brc.buffer = append(brc.buffer, CacheOperation{
		OpType: "SET",
		Key:    key,
		Value:  value,
		TTL:    ttl,
	})
	shouldFlush := len(brc.buffer) >= brc.batchSize
	brc.mu.Unlock()

	if shouldFlush {
		return brc.Flush(ctx)
	}

	return nil
}

func (brc *BatchRedisCache) DeleteAsync(ctx context.Context, key string) error {
	brc.mu.Lock()
	brc.buffer = append(brc.buffer, CacheOperation{
		OpType: "DEL",
		Key:    key,
	})
	shouldFlush := len(brc.buffer) >= brc.batchSize
	brc.mu.Unlock()

	if shouldFlush {
		return brc.Flush(ctx)
	}

	return nil
}

func (brc *BatchRedisCache) flushWorker() {
	defer brc.wg.Done()
	ticker := time.NewTicker(brc.flushInterval)
	defer ticker.Stop()

	ctx := context.Background()

	for {
		select {
		case <-ticker.C:
			brc.Flush(ctx)
		case <-brc.stopCh:
			brc.Flush(ctx)
			return
		}
	}
}

func (brc *BatchRedisCache) Flush(ctx context.Context) error {
	brc.mu.Lock()
	if len(brc.buffer) == 0 {
		brc.mu.Unlock()
		return nil
	}

	toFlush := brc.buffer
	brc.buffer = make([]CacheOperation, 0, brc.batchSize)
	brc.mu.Unlock()

	// Use Redis pipeline for batch operations
	pipe := brc.client.Pipeline()

	for _, op := range toFlush {
		switch op.OpType {
		case "SET":
			jsonData, err := json.Marshal(op.Value)
			if err != nil {
				log.Printf("[ERROR] failed to marshal cache value for key %s: %v", op.Key, err)
				continue
			}
			pipe.Set(ctx, op.Key, jsonData, op.TTL)

		case "DEL":
			pipe.Del(ctx, op.Key)
		}
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		log.Printf("[ERROR] redis pipeline failed: %v", err)
		return err
	}

	log.Printf("[INFO] executed %d redis operations in pipeline", len(toFlush))
	return nil
}

func (brc *BatchRedisCache) Close() {
	close(brc.stopCh)
	brc.wg.Wait()
}
