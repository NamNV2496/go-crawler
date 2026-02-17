package service

import (
	"context"
	"errors"
	"testing"

	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/namnv2496/scheduler/internal/entity"
	"github.com/namnv2496/scheduler/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple mock for testing (no gomock needed for basic tests)
type mockSchedulerEventRepository struct {
	createEventFunc       func(context.Context, *domain.SchedulerEvent) (int64, error)
	getEventsFunc         func(context.Context, int32, int32) ([]*domain.SchedulerEvent, error)
	updateEventFunc       func(context.Context, *domain.SchedulerEvent) error
	getEventByIDFunc      func(context.Context, int64) (*domain.SchedulerEvent, error)
	getEventsByStatusFunc func(context.Context, domain.StatusEnum, int32) ([]*domain.SchedulerEvent, error)
}

func (m *mockSchedulerEventRepository) CreateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
	if m.createEventFunc != nil {
		return m.createEventFunc(ctx, event)
	}
	return 0, errors.New("not implemented")
}

func (m *mockSchedulerEventRepository) GetSchedulerEvents(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
	if m.getEventsFunc != nil {
		return m.getEventsFunc(ctx, limit, offset)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSchedulerEventRepository) UpdateSchedulerEvent(ctx context.Context, event *domain.SchedulerEvent) error {
	if m.updateEventFunc != nil {
		return m.updateEventFunc(ctx, event)
	}
	return errors.New("not implemented")
}

func (m *mockSchedulerEventRepository) GetSchedulerEventByID(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
	if m.getEventByIDFunc != nil {
		return m.getEventByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockSchedulerEventRepository) GetSchedulerEventsByStatus(ctx context.Context, status domain.StatusEnum, limit int32) ([]*domain.SchedulerEvent, error) {
	if m.getEventsByStatusFunc != nil {
		return m.getEventsByStatusFunc(ctx, status, limit)
	}
	return nil, errors.New("not implemented")
}

// Implement all remaining IRepository methods as stubs
func (m *mockSchedulerEventRepository) InsertOnce(ctx context.Context, entity *domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) Inserts(ctx context.Context, entities []*domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) UpdateOnce(ctx context.Context, entity *domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) Updates(ctx context.Context, entities []*domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) DeleteOnce(ctx context.Context, entity *domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) DeleteById(ctx context.Context, entity *domain.SchedulerEvent, opts ...repository.QueryOptionFunc) error {
	return nil
}

func (m *mockSchedulerEventRepository) Finds(ctx context.Context, opts ...repository.QueryOptionFunc) ([]*domain.SchedulerEvent, error) {
	return nil, nil
}

func (m *mockSchedulerEventRepository) Find(ctx context.Context, opts ...repository.QueryOptionFunc) (*domain.SchedulerEvent, error) {
	return nil, nil
}

func (m *mockSchedulerEventRepository) CountOnce(ctx context.Context, opts ...repository.QueryOptionFunc) (int64, error) {
	return 0, nil
}

func (m *mockSchedulerEventRepository) RunWithTransaction(ctx context.Context, txName string, funcs ...repository.FunctionExec) error {
	return nil
}

// Additional methods that might be required
func (m *mockSchedulerEventRepository) GetSchedulerEventByStatusAndSchedulerAt(ctx context.Context, status domain.StatusEnum, schedulerAt int64) ([]*domain.SchedulerEvent, error) {
	return nil, nil
}

func (m *mockSchedulerEventRepository) CountSchedulerEventByDomainsAndQueues(ctx context.Context, domains []string, queues []string) (int64, error) {
	return 0, nil
}

func (m *mockSchedulerEventRepository) GetSchedulerEventByDomainAndQueue(ctx context.Context, domain, queue string, limit, offset int) ([]*domain.SchedulerEvent, error) {
	return nil, nil
}

func (m *mockSchedulerEventRepository) UpdateSchedulerEvents(ctx context.Context, events []*domain.SchedulerEvent) error {
	return nil
}

// Tests
func TestSchedulerEventService_CreateSchedulerEvent(t *testing.T) {
	tests := []struct {
		name    string
		input   *entity.SchedulerEvent
		mockFn  func(*mockSchedulerEventRepository)
		want    int64
		wantErr bool
	}{
		{
			name: "success - create event",
			input: &entity.SchedulerEvent{
				Url:         "https://example.com",
				Method:      "GET",
				Queue:       "normal",
				Description: "Test crawler",
				IsActive:    true,
			},
			mockFn: func(m *mockSchedulerEventRepository) {
				m.createEventFunc = func(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
					assert.Equal(t, "https://example.com", event.Url)
					return 123, nil
				}
			},
			want:    123,
			wantErr: false,
		},
		{
			name: "error - repository error",
			input: &entity.SchedulerEvent{
				Url:    "https://example.com",
				Method: "GET",
			},
			mockFn: func(m *mockSchedulerEventRepository) {
				m.createEventFunc = func(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
					return 0, errors.New("database error")
				}
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSchedulerEventRepository{}
			tt.mockFn(mockRepo)

			service := &SchedulerEventService{
				repo: mockRepo,
			}

			ctx := context.Background()
			got, err := service.CreateSchedulerEvent(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestSchedulerEventService_GetSchedulerEvents(t *testing.T) {
	tests := []struct {
		name    string
		limit   int32
		offset  int32
		mockFn  func(*mockSchedulerEventRepository)
		want    int
		wantErr bool
	}{
		{
			name:   "success - get events",
			limit:  10,
			offset: 0,
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventsFunc = func(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
					return []*domain.SchedulerEvent{
						{Id: 1, Url: "https://example.com"},
						{Id: 2, Url: "https://example2.com"},
					}, nil
				}
			},
			want:    2,
			wantErr: false,
		},
		{
			name:   "success - empty result",
			limit:  10,
			offset: 100,
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventsFunc = func(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
					return []*domain.SchedulerEvent{}, nil
				}
			},
			want:    0,
			wantErr: false,
		},
		{
			name:   "error - repository error",
			limit:  10,
			offset: 0,
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventsFunc = func(ctx context.Context, limit, offset int32) ([]*domain.SchedulerEvent, error) {
					return nil, errors.New("database error")
				}
			},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSchedulerEventRepository{}
			tt.mockFn(mockRepo)

			service := &SchedulerEventService{
				repo: mockRepo,
			}

			ctx := context.Background()
			got, err := service.GetSchedulerEvents(ctx, tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.want)
			}
		})
	}
}

func TestSchedulerEventService_UpdateSchedulerEvent(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		input   *entity.SchedulerEvent
		mockFn  func(*mockSchedulerEventRepository)
		wantErr bool
	}{
		{
			name: "success - update event",
			id:   123,
			input: &entity.SchedulerEvent{
				Url:         "https://updated.com",
				Description: "Updated description",
				IsActive:    false,
			},
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventByIDFunc = func(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
					return &domain.SchedulerEvent{
						Id:  123,
						Url: "https://old.com",
					}, nil
				}

				m.updateEventFunc = func(ctx context.Context, event *domain.SchedulerEvent) error {
					assert.Equal(t, "https://updated.com", event.Url)
					assert.Equal(t, "Updated description", event.Description)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "error - event not found",
			id:   999,
			input: &entity.SchedulerEvent{
				Url: "https://example.com",
			},
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventByIDFunc = func(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
					return nil, errors.New("not found")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSchedulerEventRepository{}
			tt.mockFn(mockRepo)

			service := &SchedulerEventService{
				repo: mockRepo,
			}

			ctx := context.Background()
			err := service.UpdateSchedulerEvent(ctx, tt.id, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchedulerEventService_UpdateEventStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  domain.StatusEnum
		mockFn  func(*mockSchedulerEventRepository)
		wantErr bool
	}{
		{
			name:   "success - update to success status",
			id:     123,
			status: domain.StatusSuccessed,
			mockFn: func(m *mockSchedulerEventRepository) {
				m.getEventByIDFunc = func(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
					return &domain.SchedulerEvent{
						Id:     123,
						Status: domain.StatusPending,
					}, nil
				}

				m.updateEventFunc = func(ctx context.Context, event *domain.SchedulerEvent) error {
					assert.Equal(t, domain.StatusSuccessed, event.Status)
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockSchedulerEventRepository{}
			tt.mockFn(mockRepo)

			service := &SchedulerEventService{
				repo: mockRepo,
			}

			ctx := context.Background()
			err := service.UpdateEventStatus(ctx, tt.id, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkSchedulerEventService_CreateSchedulerEvent(b *testing.B) {
	mockRepo := &mockSchedulerEventRepository{
		createEventFunc: func(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
			return 1, nil
		},
	}

	service := &SchedulerEventService{repo: mockRepo}

	ctx := context.Background()
	event := &entity.SchedulerEvent{
		Url:    "https://example.com",
		Method: "GET",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.CreateSchedulerEvent(ctx, event)
	}
}

// Integration-style workflow test
func TestSchedulerEventService_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockRepo := &mockSchedulerEventRepository{
		createEventFunc: func(ctx context.Context, event *domain.SchedulerEvent) (int64, error) {
			return 1, nil
		},
		getEventByIDFunc: func(ctx context.Context, id int64) (*domain.SchedulerEvent, error) {
			return &domain.SchedulerEvent{
				Id:     1,
				Url:    "https://example.com",
				Status: domain.StatusPending,
			}, nil
		},
		updateEventFunc: func(ctx context.Context, event *domain.SchedulerEvent) error {
			return nil
		},
	}

	service := &SchedulerEventService{repo: mockRepo}
	ctx := context.Background()

	// Create event
	id, err := service.CreateSchedulerEvent(ctx, &entity.SchedulerEvent{
		Url:    "https://example.com",
		Method: "GET",
	})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Update status
	err = service.UpdateEventStatus(ctx, id, domain.StatusSuccessed)
	require.NoError(t, err)
}
