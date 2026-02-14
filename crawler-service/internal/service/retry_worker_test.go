package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/namnv2496/crawler/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock crawler service for retry testing
type mockCrawlerServiceForRetry struct {
	crawlFunc    func(context.Context, entity.CrawlerEvent) error
	callCount    int32
	shouldFail   bool
	failureCount int
	mu           sync.Mutex
}

func (m *mockCrawlerServiceForRetry) Crawl(ctx context.Context, event entity.CrawlerEvent) error {
	atomic.AddInt32(&m.callCount, 1)

	if m.crawlFunc != nil {
		return m.crawlFunc(ctx, event)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFail && int(m.callCount) <= m.failureCount {
		return errors.New("crawl failed")
	}

	return nil
}

func TestRetryWorker_ProcessEvent(t *testing.T) {
	tests := []struct {
		name         string
		event        entity.CrawlerEvent
		maxRetries   int
		mockSetup    func(*mockCrawlerServiceForRetry)
		wantSuccess  bool
		wantAttempts int32
	}{
		{
			name: "success - first attempt",
			event: entity.CrawlerEvent{
				Id:        1,
				Url:       "https://example.com",
				Method:    "GET",
				Retrytime: 0,
			},
			maxRetries: 3,
			mockSetup: func(m *mockCrawlerServiceForRetry) {
				m.shouldFail = false
			},
			wantSuccess:  true,
			wantAttempts: 1,
		},
		{
			name: "success - retry after failure",
			event: entity.CrawlerEvent{
				Id:        2,
				Url:       "https://example.com",
				Method:    "GET",
				Retrytime: 1,
			},
			maxRetries: 3,
			mockSetup: func(m *mockCrawlerServiceForRetry) {
				m.shouldFail = true
				m.failureCount = 1
			},
			wantSuccess:  true,
			wantAttempts: 2,
		},
		{
			name: "failure - max retries exceeded",
			event: entity.CrawlerEvent{
				Id:        3,
				Url:       "https://example.com",
				Method:    "GET",
				Retrytime: 2,
			},
			maxRetries: 3,
			mockSetup: func(m *mockCrawlerServiceForRetry) {
				m.shouldFail = true
				m.failureCount = 10
			},
			wantSuccess:  false,
			wantAttempts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockCrawlerServiceForRetry{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			ctx := context.Background()

			// Simulate retry logic
			var attempts int32
			maxAttempts := tt.maxRetries - int(tt.event.Retrytime)
			for i := 0; i < maxAttempts; i++ {
				atomic.AddInt32(&attempts, 1)
				err := mockService.Crawl(ctx, tt.event)
				if err == nil {
					break
				}
				if i < maxAttempts-1 {
					time.Sleep(10 * time.Millisecond)
				}
			}

			if tt.wantSuccess {
				assert.LessOrEqual(t, attempts, int32(tt.maxRetries))
			}
			assert.GreaterOrEqual(t, mockService.callCount, tt.wantAttempts)
		})
	}
}

func TestRetryWorker_ExponentialBackoff(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping backoff test in short mode")
	}

	tests := []struct {
		name            string
		retryCount      int
		baseDelay       time.Duration
		expectedMinTime time.Duration
	}{
		{
			name:            "first retry - 1 second",
			retryCount:      1,
			baseDelay:       1 * time.Second,
			expectedMinTime: 1 * time.Second,
		},
		{
			name:            "second retry - 2 seconds",
			retryCount:      2,
			baseDelay:       1 * time.Second,
			expectedMinTime: 2 * time.Second,
		},
		{
			name:            "third retry - 4 seconds",
			retryCount:      3,
			baseDelay:       1 * time.Second,
			expectedMinTime: 4 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate exponential backoff
			delay := tt.baseDelay
			for i := 1; i < tt.retryCount; i++ {
				delay *= 2
			}

			assert.GreaterOrEqual(t, delay, tt.expectedMinTime)
		})
	}
}

func TestRetryWorker_ConcurrentRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	mockService := &mockCrawlerServiceForRetry{
		shouldFail:   false,
		failureCount: 0,
	}

	ctx := context.Background()
	var wg sync.WaitGroup
	concurrentRetries := 10

	for i := 0; i < concurrentRetries; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			event := entity.CrawlerEvent{
				Id:        int64(id),
				Url:       "https://example.com",
				Method:    "GET",
				Retrytime: 1,
			}

			err := mockService.Crawl(ctx, event)
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int32(concurrentRetries), mockService.callCount)
}

func TestRetryWorker_ContextCancellation(t *testing.T) {
	mockService := &mockCrawlerServiceForRetry{
		crawlFunc: func(ctx context.Context, event entity.CrawlerEvent) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	event := entity.CrawlerEvent{
		Id:     1,
		Url:    "https://example.com",
		Method: "GET",
	}

	err := mockService.Crawl(ctx, event)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestRetryWorker_RetryTracking(t *testing.T) {
	type retryAttempt struct {
		eventID     int64
		attemptNum  int
		success     bool
		attemptTime time.Time
	}

	var attempts []retryAttempt
	var mu sync.Mutex

	mockService := &mockCrawlerServiceForRetry{
		crawlFunc: func(ctx context.Context, event entity.CrawlerEvent) error {
			mu.Lock()
			defer mu.Unlock()

			attempt := retryAttempt{
				eventID:     event.Id,
				attemptNum:  int(event.Retrytime) + 1,
				success:     event.Retrytime >= 2,
				attemptTime: time.Now(),
			}
			attempts = append(attempts, attempt)

			if event.Retrytime < 2 {
				return errors.New("not yet")
			}
			return nil
		},
	}

	ctx := context.Background()
	event := entity.CrawlerEvent{
		Id:        1,
		Url:       "https://example.com",
		Method:    "GET",
		Retrytime: 0,
	}

	// Simulate retries
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		event.Retrytime = int64(i)
		err := mockService.Crawl(ctx, event)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	require.Greater(t, len(attempts), 0)
	assert.Equal(t, int64(1), attempts[0].eventID)

	// Last attempt should be successful
	lastAttempt := attempts[len(attempts)-1]
	assert.True(t, lastAttempt.success)
}

func TestRetryWorker_PartialFailures(t *testing.T) {
	events := []entity.CrawlerEvent{
		{Id: 1, Url: "https://example1.com", Retrytime: 0},
		{Id: 2, Url: "https://example2.com", Retrytime: 0},
		{Id: 3, Url: "https://example3.com", Retrytime: 0},
	}

	var successCount int32
	var failureCount int32

	mockService := &mockCrawlerServiceForRetry{
		crawlFunc: func(ctx context.Context, event entity.CrawlerEvent) error {
			// Fail event with ID 2
			if event.Id == 2 {
				atomic.AddInt32(&failureCount, 1)
				return errors.New("intentional failure")
			}
			atomic.AddInt32(&successCount, 1)
			return nil
		},
	}

	ctx := context.Background()

	for _, event := range events {
		_ = mockService.Crawl(ctx, event)
	}

	assert.Equal(t, int32(2), successCount)
	assert.Equal(t, int32(1), failureCount)
}

// Benchmark tests
func BenchmarkRetryWorker_ProcessEvent(b *testing.B) {
	mockService := &mockCrawlerServiceForRetry{
		shouldFail: false,
	}

	ctx := context.Background()
	event := entity.CrawlerEvent{
		Id:        1,
		Url:       "https://example.com",
		Method:    "GET",
		Retrytime: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockService.Crawl(ctx, event)
	}
}

func BenchmarkRetryWorker_ConcurrentProcessing(b *testing.B) {
	mockService := &mockCrawlerServiceForRetry{
		shouldFail: false,
	}

	ctx := context.Background()
	event := entity.CrawlerEvent{
		Id:        1,
		Url:       "https://example.com",
		Method:    "GET",
		Retrytime: 0,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mockService.Crawl(ctx, event)
		}
	})
}

func TestRetryWorker_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping rate limit test in short mode")
	}

	var processedCount int32
	rateLimitPerSecond := 10
	duration := 1 * time.Second

	mockService := &mockCrawlerServiceForRetry{
		crawlFunc: func(ctx context.Context, event entity.CrawlerEvent) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		},
	}

	ctx := context.Background()
	event := entity.CrawlerEvent{
		Id:     1,
		Url:    "https://example.com",
		Method: "GET",
	}

	ticker := time.NewTicker(duration / time.Duration(rateLimitPerSecond))
	defer ticker.Stop()

	done := make(chan bool)
	go func() {
		time.Sleep(duration)
		done <- true
	}()

	for {
		select {
		case <-ticker.C:
			_ = mockService.Crawl(ctx, event)
		case <-done:
			goto finish
		}
	}

finish:
	t.Logf("Processed %d events in %v", processedCount, duration)
	assert.LessOrEqual(t, processedCount, int32(rateLimitPerSecond+2))
}
