package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/namnv2496/crawler/internal/pkg/logging"
)

type IDistributedWorkerPool interface {
	ExecuteOnServerSync(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error)
	ExecuteOnServerAsync(ctx context.Context, req *ExecutionRequest, resultChan chan *ExecutionResponse) error
	ExecuteOnServerWithPolling(ctx context.Context, req *ExecutionRequest, resultChan chan *ExecutionResponse, pollInterval time.Duration) error
	ExecuteMultipleAsync(ctx context.Context, requests []*ExecutionRequest) (map[string]chan *ExecutionResponse, error)
	ExecuteParallel(ctx context.Context, requests []*ExecutionRequest) (map[string]*ExecutionResponse, error)
	ExecuteWithRetry(ctx context.Context, req *ExecutionRequest, maxRetries int) (*ExecutionResponse, error)
}

type DistributedWorkerPool struct {
	serverManager  IServerManager
	serverExecutor IServerExecutor
	requestPoint   int64
}

func NewDistributedWorkerPool(sm IServerManager, se IServerExecutor, requestPoint int64) *DistributedWorkerPool {
	return &DistributedWorkerPool{
		serverManager:  sm,
		serverExecutor: se,
		requestPoint:   requestPoint,
	}
}

func (dwp *DistributedWorkerPool) ExecuteOnServerSync(ctx context.Context, req *ExecutionRequest) (*ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteOnServerSync")
	defer deferFunc()

	// Get available server with required capacity
	server := dwp.serverManager.GetAvailableServerByPoint(dwp.requestPoint)
	if server == nil {
		return nil, fmt.Errorf("no available server with point >= %d", dwp.requestPoint)
	}
	defer func() {
		if err := dwp.serverManager.ReturnServer(server); err != nil {
			logging.Error(ctx, "Failed to return server: %v", err)
		}
	}()

	logging.Info(ctx, "Executing job %s synchronously on server %s", req.JobID, server.ID)

	// Execute on remote server synchronously
	resp, err := dwp.serverExecutor.Execute(ctx, server, req)
	if err != nil {
		logging.Error(ctx, "Failed to execute job %s on server %s: %v", req.JobID, server.ID, err)
		return nil, err
	}

	logging.Info(ctx, "Job %s completed on server %s with status %s", req.JobID, server.ID, resp.Status)
	return resp, nil
}

func (dwp *DistributedWorkerPool) ExecuteOnServerAsync(ctx context.Context, req *ExecutionRequest, resultChan chan *ExecutionResponse) error {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteOnServerAsync")
	defer deferFunc()

	server := dwp.serverManager.GetAvailableServerByPoint(dwp.requestPoint)
	if server == nil {
		return fmt.Errorf("no available server with point >= %d", dwp.requestPoint)
	}

	logging.Info(ctx, "Executing job %s asynchronously on server %s", req.JobID, server.ID)

	go func() {
		defer func() {
			if err := dwp.serverManager.ReturnServer(server); err != nil {
				logging.Error(ctx, "Failed to return server: %v", err)
			}
		}()

		resp, err := dwp.serverExecutor.Execute(ctx, server, req)
		if err != nil {
			logging.Error(ctx, "Failed to execute job %s on server %s: %v", req.JobID, server.ID, err)
			resultChan <- &ExecutionResponse{
				JobID:  req.JobID,
				Status: "failed",
				Error:  err.Error(),
			}
			return
		}

		logging.Info(ctx, "Job %s completed on server %s with status %s", req.JobID, server.ID, resp.Status)
		resultChan <- resp
	}()

	return nil
}

func (dwp *DistributedWorkerPool) ExecuteOnServerWithPolling(ctx context.Context, req *ExecutionRequest, resultChan chan *ExecutionResponse, pollInterval time.Duration) error {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteOnServerWithPolling")
	defer deferFunc()

	server := dwp.serverManager.GetAvailableServerByPoint(dwp.requestPoint)
	if server == nil {
		return fmt.Errorf("no available server with point >= %d", dwp.requestPoint)
	}

	logging.Info(ctx, "Executing job %s with polling on server %s", req.JobID, server.ID)

	go func() {
		defer func() {
			if err := dwp.serverManager.ReturnServer(server); err != nil {
				logging.Error(ctx, "Failed to return server: %v", err)
			}
		}()

		resp, err := dwp.serverExecutor.Execute(ctx, server, req)
		if err != nil {
			logging.Error(ctx, "Failed to submit job %s on server %s: %v", req.JobID, server.ID, err)
			resultChan <- &ExecutionResponse{
				JobID:  req.JobID,
				Status: "failed",
				Error:  err.Error(),
			}
			return
		}

		if resp.Status != "pending" {
			resultChan <- resp
			return
		}

		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logging.Info(ctx, "Polling cancelled for job %s", req.JobID)
				resultChan <- &ExecutionResponse{
					JobID:  req.JobID,
					Status: "cancelled",
					Error:  "context cancelled",
				}
				return

			case <-ticker.C:
				statusResp, err := dwp.serverExecutor.GetJobStatus(ctx, server, req.JobID)
				if err != nil {
					logging.Error(ctx, "Failed to get status for job %s: %v", req.JobID, err)
					continue
				}

				logging.Debug(ctx, "Job %s status: %s", req.JobID, statusResp.Status)

				if statusResp.Status != "pending" {
					resultChan <- statusResp
					return
				}
			}
		}
	}()

	return nil
}

func (dwp *DistributedWorkerPool) ExecuteMultipleAsync(ctx context.Context, requests []*ExecutionRequest) (map[string]chan *ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteMultipleAsync")
	defer deferFunc()

	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}

	resultChans := make(map[string]chan *ExecutionResponse)

	for _, req := range requests {
		resultChan := make(chan *ExecutionResponse, 1)
		resultChans[req.JobID] = resultChan

		if err := dwp.ExecuteOnServerAsync(ctx, req, resultChan); err != nil {
			logging.Error(ctx, "Failed to execute job %s: %v", req.JobID, err)
			resultChan <- &ExecutionResponse{
				JobID:  req.JobID,
				Status: "failed",
				Error:  err.Error(),
			}
		}
	}

	logging.Info(ctx, "Submitted %d jobs for async execution", len(requests))
	return resultChans, nil
}

func (dwp *DistributedWorkerPool) ExecuteParallel(ctx context.Context, requests []*ExecutionRequest) (map[string]*ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteParallel")
	defer deferFunc()

	if len(requests) == 0 {
		return nil, fmt.Errorf("no requests provided")
	}

	resultChans, err := dwp.ExecuteMultipleAsync(ctx, requests)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*ExecutionResponse)
	var wg sync.WaitGroup

	for jobID, resultChan := range resultChans {
		wg.Add(1)
		go func(jID string, rChan chan *ExecutionResponse) {
			defer wg.Done()
			resp := <-rChan
			results[jID] = resp
		}(jobID, resultChan)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logging.Info(ctx, "All %d jobs completed", len(requests))
		return results, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled while waiting for jobs")
	}
}

func (dwp *DistributedWorkerPool) ExecuteWithRetry(ctx context.Context, req *ExecutionRequest, maxRetries int) (*ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("DistributedWorkerPool.ExecuteWithRetry")
	defer deferFunc()

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		logging.Info(ctx, "Executing job %s (attempt %d/%d)", req.JobID, attempt+1, maxRetries+1)

		resp, err := dwp.ExecuteOnServerSync(ctx, req)
		if err == nil && resp.Status == "success" {
			return resp, nil
		}

		lastErr = err
		if resp != nil && resp.Status == "failed" {
			logging.Warn(ctx, "Job %s failed: %s", req.JobID, resp.Error)
		}

		if attempt < maxRetries {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	logging.Error(ctx, "Job %s failed after %d retries", req.JobID, maxRetries+1)
	return nil, lastErr
}
