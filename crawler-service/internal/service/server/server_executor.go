package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/namnv2496/crawler/internal/pkg/logging"
)

// ExecutionRequest is sent to the server for execution
type ExecutionRequest struct {
	JobID      string            `json:"job_id"`
	JobType    string            `json:"job_type"` // "crawl", "curl", etc.
	URL        string            `json:"url"`
	Method     string            `json:"method"` // GET, POST, etc.
	Payload    map[string]string `json:"payload,omitempty"`
	Timeout    int               `json:"timeout"`     // seconds
	RetryCount int               `json:"retry_count"` // 0-3
}

// ExecutionResponse is returned from the server
type ExecutionResponse struct {
	JobID     string            `json:"job_id"`
	Status    string            `json:"status"` // "success", "failed", "pending"
	Result    string            `json:"result"`
	Error     string            `json:"error,omitempty"`
	Duration  int64             `json:"duration"` // milliseconds
	Timestamp int64             `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// IServerExecutor handles communication with remote servers
type IServerExecutor interface {
	Execute(ctx context.Context, server *Server, req *ExecutionRequest) (*ExecutionResponse, error)
	ExecuteAsync(ctx context.Context, server *Server, req *ExecutionRequest, resultChan chan *ExecutionResponse) error
	GetJobStatus(ctx context.Context, server *Server, jobID string) (*ExecutionResponse, error)
	CancelJob(ctx context.Context, server *Server, jobID string) error
}

// ServerExecutor implements IServerExecutor
type ServerExecutor struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewServerExecutor creates a new server executor
func NewServerExecutor() IServerExecutor {
	return &ServerExecutor{
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		timeout: 5 * time.Second,
	}
}

func (se *ServerExecutor) Execute(ctx context.Context, server *Server, req *ExecutionRequest) (*ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("ServerExecutor.Execute")
	defer deferFunc()

	if server == nil {
		return nil, fmt.Errorf("server is nil")
	}

	// Build server URL (endpoint for executing jobs)
	serverURL := fmt.Sprintf("http://%s:8080/api/v1/execute", server.ID)

	// Marshal request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	logging.Debug(ctx, "Sending job %s to server %s (point: %d)", req.JobID, server.ID, server.Point)
	start := time.Now()

	httpResp, err := se.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute on server %s: %v", server.ID, err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Unmarshal response
	var execResp ExecutionResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	duration := time.Since(start).Milliseconds()
	logging.Debug(ctx, "Job %s completed on server %s in %dms with status: %s",
		req.JobID, server.ID, duration, execResp.Status)

	return &execResp, nil
}

// ExecuteAsync sends a job to a server without waiting for result
func (se *ServerExecutor) ExecuteAsync(ctx context.Context, server *Server, req *ExecutionRequest, resultChan chan *ExecutionResponse) error {
	deferFunc := logging.AppendPrefix("ServerExecutor.ExecuteAsync")
	defer deferFunc()

	if server == nil {
		return fmt.Errorf("server is nil")
	}

	if resultChan == nil {
		return fmt.Errorf("result channel is nil")
	}

	// Send job asynchronously
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logging.Error(ctx, "Panic in ExecuteAsync: %v", r)
				resultChan <- &ExecutionResponse{
					JobID:  req.JobID,
					Status: "failed",
					Error:  fmt.Sprintf("panic: %v", r),
				}
			}
		}()

		result, err := se.Execute(ctx, server, req)
		if err != nil {
			logging.Error(ctx, "ExecuteAsync failed for job %s: %v", req.JobID, err)
			resultChan <- &ExecutionResponse{
				JobID:  req.JobID,
				Status: "failed",
				Error:  err.Error(),
			}
			return
		}

		resultChan <- result
	}()

	logging.Debug(ctx, "Job %s sent asynchronously to server %s", req.JobID, server.ID)
	return nil
}

func (se *ServerExecutor) GetJobStatus(ctx context.Context, server *Server, jobID string) (*ExecutionResponse, error) {
	deferFunc := logging.AppendPrefix("ServerExecutor.GetJobStatus")
	defer deferFunc()

	if server == nil {
		return nil, fmt.Errorf("server is nil")
	}

	statusURL := fmt.Sprintf("http://%s:8080/api/v1/jobs/%s/status", server.ID, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpResp, err := se.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to check status: %v", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", httpResp.StatusCode)
	}

	var result ExecutionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &result, nil
}

func (se *ServerExecutor) CancelJob(ctx context.Context, server *Server, jobID string) error {
	deferFunc := logging.AppendPrefix("ServerExecutor.CancelJob")
	defer deferFunc()

	if server == nil {
		return fmt.Errorf("server is nil")
	}

	cancelURL := fmt.Sprintf("http://%s:8080/api/v1/jobs/%s/cancel", server.ID, jobID)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cancelURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpResp, err := se.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %v", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", httpResp.StatusCode)
	}

	logging.Debug(ctx, "Job %s cancelled on server %s", jobID, server.ID)
	return nil
}
