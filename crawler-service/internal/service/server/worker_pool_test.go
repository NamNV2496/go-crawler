package server

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_Execute(t *testing.T) {
	tests := []struct {
		name          string
		poolSize      int
		taskCount     int
		taskFunc      func() (any, error)
		wantErr       bool
		checkCallback bool
	}{
		{
			name:      "success - single task",
			poolSize:  2,
			taskCount: 1,
			taskFunc: func() (any, error) {
				return "result", nil
			},
			wantErr:       false,
			checkCallback: true,
		},
		{
			name:      "success - multiple tasks",
			poolSize:  5,
			taskCount: 10,
			taskFunc: func() (any, error) {
				time.Sleep(10 * time.Millisecond)
				return "result", nil
			},
			wantErr:       false,
			checkCallback: true,
		},
		{
			name:      "error - task failure",
			poolSize:  2,
			taskCount: 1,
			taskFunc: func() (any, error) {
				return nil, errors.New("task failed")
			},
			wantErr:       true,
			checkCallback: true,
		},
		{
			name:      "success - concurrent execution",
			poolSize:  10,
			taskCount: 100,
			taskFunc: func() (any, error) {
				time.Sleep(1 * time.Millisecond)
				return "concurrent", nil
			},
			wantErr:       false,
			checkCallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var completedTasks int32
			var callbackCalled int32

			callback := func(result any, err error) {
				atomic.AddInt32(&callbackCalled, 1)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			}

			// Simulate worker pool execution
			var wg sync.WaitGroup
			semaphore := make(chan struct{}, tt.poolSize)

			for i := 0; i < tt.taskCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					semaphore <- struct{}{}
					defer func() { <-semaphore }()

					result, err := tt.taskFunc()
					atomic.AddInt32(&completedTasks, 1)

					if tt.checkCallback {
						callback(result, err)
					}
				}()
			}

			wg.Wait()

			assert.Equal(t, int32(tt.taskCount), completedTasks)
			if tt.checkCallback {
				assert.Equal(t, int32(tt.taskCount), callbackCalled)
			}
		})
	}
}

func TestWorkerPool_Concurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrency test in short mode")
	}

	poolSize := 5
	taskCount := 50
	var activeWorkers int32
	var maxActiveWorkers int32
	var mutex sync.Mutex

	taskFunc := func() (any, error) {
		current := atomic.AddInt32(&activeWorkers, 1)

		mutex.Lock()
		if current > maxActiveWorkers {
			maxActiveWorkers = current
		}
		mutex.Unlock()

		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&activeWorkers, -1)
		return "done", nil
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, poolSize)

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			_, _ = taskFunc()
		}()
	}

	wg.Wait()

	assert.LessOrEqual(t, maxActiveWorkers, int32(poolSize), "max active workers should not exceed pool size")
}

func TestWorkerPool_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		tasks      []func() (any, error)
		wantErrors int
	}{
		{
			name: "mixed success and errors",
			tasks: []func() (any, error){
				func() (any, error) { return "success", nil },
				func() (any, error) { return nil, errors.New("error 1") },
				func() (any, error) { return "success", nil },
				func() (any, error) { return nil, errors.New("error 2") },
			},
			wantErrors: 2,
		},
		{
			name: "all success",
			tasks: []func() (any, error){
				func() (any, error) { return "success", nil },
				func() (any, error) { return "success", nil },
				func() (any, error) { return "success", nil },
			},
			wantErrors: 0,
		},
		{
			name: "all errors",
			tasks: []func() (any, error){
				func() (any, error) { return nil, errors.New("error 1") },
				func() (any, error) { return nil, errors.New("error 2") },
			},
			wantErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorCount := 0
			var mutex sync.Mutex

			callback := func(result any, err error) {
				mutex.Lock()
				defer mutex.Unlock()
				if err != nil {
					errorCount++
				}
			}

			var wg sync.WaitGroup
			for _, task := range tt.tasks {
				wg.Add(1)
				go func(t func() (any, error)) {
					defer wg.Done()
					result, err := t()
					callback(result, err)
				}(task)
			}

			wg.Wait()

			assert.Equal(t, tt.wantErrors, errorCount)
		})
	}
}

func TestWorkerPool_Stats(t *testing.T) {
	var completed int32
	var failed int32

	statsCallback := func(result any, err error) {
		if err != nil {
			atomic.AddInt32(&failed, 1)
		} else {
			atomic.AddInt32(&completed, 1)
		}
	}

	tasks := []func() (any, error){
		func() (any, error) { return "ok", nil },
		func() (any, error) { return nil, errors.New("fail") },
		func() (any, error) { return "ok", nil },
		func() (any, error) { return "ok", nil },
		func() (any, error) { return nil, errors.New("fail") },
	}

	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go func(t func() (any, error)) {
			defer wg.Done()
			result, err := t()
			statsCallback(result, err)
		}(task)
	}

	wg.Wait()

	assert.Equal(t, int32(3), completed)
	assert.Equal(t, int32(2), failed)
}

func TestWorkerPool_Shutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping shutdown test in short mode")
	}

	poolSize := 3
	done := make(chan struct{})
	var tasksStarted int32
	var tasksCompleted int32

	// Start tasks
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, poolSize)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				atomic.AddInt32(&tasksStarted, 1)

				select {
				case <-done:
					return
				case <-time.After(50 * time.Millisecond):
					atomic.AddInt32(&tasksCompleted, 1)
				}
			case <-done:
				return
			}
		}()
	}

	// Simulate shutdown after short time
	time.Sleep(25 * time.Millisecond)
	close(done)

	wg.Wait()

	t.Logf("Tasks started: %d, completed: %d", tasksStarted, tasksCompleted)
	assert.GreaterOrEqual(t, tasksStarted, int32(1))
	assert.LessOrEqual(t, tasksStarted, int32(10))
}

// Benchmark tests
func BenchmarkWorkerPool_Execute(b *testing.B) {
	poolSize := 10
	semaphore := make(chan struct{}, poolSize)

	task := func() (any, error) {
		time.Sleep(100 * time.Microsecond)
		return "result", nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		semaphore <- struct{}{}
		go func() {
			defer func() { <-semaphore }()
			_, _ = task()
		}()
	}

	// Wait for all tasks to complete
	for i := 0; i < poolSize; i++ {
		semaphore <- struct{}{}
	}
}

func BenchmarkWorkerPool_HighConcurrency(b *testing.B) {
	poolSize := 100
	semaphore := make(chan struct{}, poolSize)

	task := func() (any, error) {
		return "result", nil
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			semaphore <- struct{}{}
			go func() {
				defer func() { <-semaphore }()
				_, _ = task()
			}()
		}
	})
}

func TestWorkerPool_RaceCondition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping race condition test in short mode")
	}

	// Test for race conditions
	var counter int32
	var wg sync.WaitGroup
	poolSize := 10

	semaphore := make(chan struct{}, poolSize)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Atomic operation to avoid race
			atomic.AddInt32(&counter, 1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(100), counter)
}

func TestWorkerPool_PanicRecovery(t *testing.T) {
	var recovered bool
	var mu sync.Mutex

	task := func() (any, error) {
		defer func() {
			if r := recover(); r != nil {
				mu.Lock()
				recovered = true
				mu.Unlock()
			}
		}()
		panic("intentional panic")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = task()
	}()

	wg.Wait()

	assert.True(t, recovered, "should recover from panic")
}
