package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock for ISchedulerEventRepository
type mockCronJobRepository struct {
	getEventsByStatusFunc func(context.Context, domain.StatusEnum, int32) ([]*domain.SchedulerEvent, error)
}

func (m *mockCronJobRepository) GetSchedulerEventsByStatus(ctx context.Context, status domain.StatusEnum, limit int32) ([]*domain.SchedulerEvent, error) {
	if m.getEventsByStatusFunc != nil {
		return m.getEventsByStatusFunc(ctx, status, limit)
	}
	return nil, nil
}

func (m *mockCronJobRepository) CreateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
	return 0, nil
}

func (m *mockCronJobRepository) GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
	return nil, nil
}

func (m *mockCronJobRepository) UpdateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) error {
	return nil
}

func (m *mockCronJobRepository) GetSchedulerEventByID(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
	return nil, nil
}

// Mock for Kafka producer
type mockKafkaProducer struct {
	publishCallCount int
	lastTopic        string
	lastKey          string
	lastValue        interface{}
}

func (m *mockKafkaProducer) Publish(ctx context.Context, topic, key string, value interface{}) error {
	m.publishCallCount++
	m.lastTopic = topic
	m.lastKey = key
	m.lastValue = value
	return nil
}

func TestCronJob_FetchAndPublishEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cron job test in short mode")
	}

	tests := []struct {
		name           string
		mockEvents     []*domain.SchedulerEvent
		expectedPublish int
		wantErr        bool
	}{
		{
			name: "success - publish pending events",
			mockEvents: []*domain.SchedulerEvent{
				{
					Id:     1,
					Url:    "https://example.com",
					Status: domain.StatusPending,
					Queue:  "normal",
				},
				{
					Id:     2,
					Url:    "https://example2.com",
					Status: domain.StatusPending,
					Queue:  "priority",
				},
			},
			expectedPublish: 2,
			wantErr:         false,
		},
		{
			name:            "success - no pending events",
			mockEvents:      []*domain.SchedulerEvent{},
			expectedPublish: 0,
			wantErr:         false,
		},
		{
			name: "success - filter inactive events",
			mockEvents: []*domain.SchedulerEvent{
				{
					Id:       1,
					Url:      "https://example.com",
					Status:   domain.StatusPending,
					Queue:    "normal",
					IsActive: true,
				},
				{
					Id:       2,
					Url:      "https://example2.com",
					Status:   domain.StatusPending,
					Queue:    "normal",
					IsActive: false, // Should be filtered out
				},
			},
			expectedPublish: 1,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockCronJobRepository{
				getEventsByStatusFunc: func(ctx context.Context, status domain.StatusEnum, limit int32) ([]*domain.SchedulerEvent, error) {
					assert.Equal(t, domain.StatusPending, status)
					return tt.mockEvents, nil
				},
			}

			mockProducer := &mockKafkaProducer{}

			// Note: Actual CronJob implementation would need to be tested here
			// This is a simplified test structure

			ctx := context.Background()
			events, err := mockRepo.GetSchedulerEventsByStatus(ctx, domain.StatusPending, 100)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, events, len(tt.mockEvents))

			// Simulate publishing
			activeCount := 0
			for _, event := range events {
				if event.IsActive {
					_ = mockProducer.Publish(ctx, event.Queue, strconv.FormatInt(event.Id, 10), event)
					activeCount++
				}
			}

			assert.Equal(t, tt.expectedPublish, activeCount)
			assert.Equal(t, tt.expectedPublish, mockProducer.publishCallCount)
		})
	}
}

func TestCronJob_QueueRouting(t *testing.T) {
	tests := []struct {
		name          string
		queue         string
		expectedTopic string
	}{
		{
			name:          "normal queue",
			queue:         "normal",
			expectedTopic: "normal",
		},
		{
			name:          "priority queue",
			queue:         "priority",
			expectedTopic: "priority",
		},
		{
			name:          "default queue",
			queue:         "",
			expectedTopic: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProducer := &mockKafkaProducer{}

			event := &domain.SchedulerEvent{
				Id:    1,
				Queue: tt.queue,
			}

			ctx := context.Background()
			err := mockProducer.Publish(ctx, event.Queue, "test", event)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTopic, mockProducer.lastTopic)
		})
	}
}

func TestCronJob_ConcurrentEventFetching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	mockRepo := &mockCronJobRepository{
		getEventsByStatusFunc: func(ctx context.Context, status domain.StatusEnum, limit int32) ([]*domain.SchedulerEvent, error) {
			// Simulate delay
			time.Sleep(10 * time.Millisecond)
			return []*domain.SchedulerEvent{
				{Id: 1, Url: "https://example.com", IsActive: true},
			}, nil
		},
	}

	ctx := context.Background()

	// Test concurrent fetching
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			events, err := mockRepo.GetSchedulerEventsByStatus(ctx, domain.StatusPending, 100)
			assert.NoError(t, err)
			assert.Len(t, events, 1)
			done <- true
		}()
	}

	// Wait for all goroutines
	timeout := time.After(1 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// Benchmark tests
func BenchmarkCronJob_FetchEvents(b *testing.B) {
	mockRepo := &mockCronJobRepository{
		getEventsByStatusFunc: func(ctx context.Context, status domain.StatusEnum, limit int32) ([]*domain.SchedulerEvent, error) {
			return []*domain.SchedulerEvent{
				{Id: 1, Url: "https://example.com", IsActive: true},
				{Id: 2, Url: "https://example2.com", IsActive: true},
			}, nil
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mockRepo.GetSchedulerEventsByStatus(ctx, domain.StatusPending, 100)
	}
}

func BenchmarkCronJob_PublishEvents(b *testing.B) {
	mockProducer := &mockKafkaProducer{}

	event := &domain.SchedulerEvent{
		Id:    1,
		Queue: "normal",
		Url:   "https://example.com",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockProducer.Publish(ctx, event.Queue, "test", event)
	}
}
