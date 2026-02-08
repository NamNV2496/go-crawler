package server

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/namnv2496/crawler/internal/pkg/logging"
)

type Server struct {
	ID           string
	Point        int64
	Status       string
	LastUsed     time.Time
	JobsExecuted int64
	Index        int
}

type ServerHeap []*Server

func (h ServerHeap) Len() int           { return len(h) }
func (h ServerHeap) Less(i, j int) bool { return h[i].Point < h[j].Point } // Min-heap, but we'll use it as max by comparing differently
func (h ServerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}
func (h *ServerHeap) Push(x any) {
	server := x.(*Server)
	server.Index = len(*h)
	*h = append(*h, server)
}
func (h *ServerHeap) Pop() any {
	old := *h
	n := len(old)
	server := old[n-1]
	server.Index = -1
	*h = old[0 : n-1]
	return server
}

type IServerManager interface {
	AddServer(id string, point int64) error
	GetAvailableServerByPoint(requestPoint int64) *Server
	ReturnServer(server *Server) error
	GetServerStats() []ServerStats
	GetTotalCapacity() int64
	RemoveServer(id string) error
	GetServerCount() int
}

type ServerStats struct {
	ServerID     string
	Point        int64
	CPU          int64 // manage in future
	RAM          int64 // manage in future
	Status       string
	LastUsed     time.Time
	JobsExecuted int64
}

type ServerManager struct {
	heap          ServerHeap
	mutex         sync.Mutex
	servers       map[string]*Server
	inUse         map[string]bool
	totalCapacity int64
}

func NewServerManager() IServerManager {
	defaultServers := []*Server{
		{
			ID:     "server-1",
			Point:  5,
			Status: "available",
		},
		{
			ID:     "server-2",
			Point:  10,
			Status: "available",
		},
		{
			ID:     "server-3",
			Point:  15,
			Status: "available",
		},
		{
			ID:     "server-4",
			Point:  15,
			Status: "available",
		},
		{
			ID:     "server-5",
			Point:  20,
			Status: "available",
		},
	}
	return &ServerManager{
		heap:    defaultServers,
		servers: make(map[string]*Server),
		inUse:   make(map[string]bool),
	}
}

func (sm *ServerManager) AddServer(id string, point int64) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.servers[id]; exists {
		return fmt.Errorf("server %s already exists", id)
	}

	server := &Server{
		ID:       id,
		Point:    point,
		Status:   "available",
		LastUsed: time.Now(),
	}
	sm.servers[id] = server
	sm.inUse[id] = false
	sm.totalCapacity += point
	heap.Push(&sm.heap, server)

	logging.Debug(context.TODO(), "Added server %s with point %d", id, point)
	return nil
}

func (sm *ServerManager) ReturnServer(server *Server) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if server == nil {
		return fmt.Errorf("cannot return nil server")
	}

	if _, exists := sm.servers[server.ID]; !exists {
		return fmt.Errorf("server %s not found", server.ID)
	}

	if !sm.inUse[server.ID] {
		return fmt.Errorf("server %s is not in use", server.ID)
	}

	server.Status = "available"
	server.JobsExecuted++
	sm.inUse[server.ID] = false

	heap.Push(&sm.heap, server)
	logging.Debug(context.TODO(), "Returned server %s to pool (jobs executed: %d)", server.ID, server.JobsExecuted)
	return nil
}

func (sm *ServerManager) RemoveServer(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	server, exists := sm.servers[id]
	if !exists {
		return fmt.Errorf("server %s not found", id)
	}

	if sm.inUse[id] {
		return fmt.Errorf("cannot remove server %s while in use", id)
	}

	if server.Index >= 0 && server.Index < len(sm.heap) {
		heap.Remove(&sm.heap, server.Index)
	}

	delete(sm.servers, id)
	delete(sm.inUse, id)
	sm.totalCapacity -= server.Point

	logging.Debug(context.TODO(), "Removed server %s from pool", id)
	return nil
}

func (sm *ServerManager) GetServerStats() []ServerStats {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	stats := make([]ServerStats, 0, len(sm.servers))
	for _, server := range sm.servers {
		stats = append(stats, ServerStats{
			ServerID:     server.ID,
			Point:        server.Point,
			Status:       server.Status,
			LastUsed:     server.LastUsed,
			JobsExecuted: server.JobsExecuted,
		})
	}
	return stats
}

func (sm *ServerManager) GetTotalCapacity() int64 {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return sm.totalCapacity
}

func (sm *ServerManager) GetServerCount() int {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	return len(sm.servers)
}

func (sm *ServerManager) GetAvailableServerByPoint(requestPoint int64) *Server {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	var candidates []*Server
	for i := 0; i < sm.heap.Len(); i++ {
		server := sm.heap[i]
		if !sm.inUse[server.ID] && server.Point >= requestPoint {
			candidates = append(candidates, server)
		}
	}

	if len(candidates) == 0 {
		logging.Debug(context.TODO(), "No available server with point >= %d", requestPoint)
		return nil
	}

	strongest := candidates[0]
	for _, s := range candidates[1:] {
		if s.Point > strongest.Point {
			strongest = s
		}
	}

	heap.Remove(&sm.heap, strongest.Index)
	sm.inUse[strongest.ID] = true
	strongest.Status = "in-use"
	strongest.LastUsed = time.Now()

	logging.Debug(context.TODO(), "Allocated server %s (point: %d) for request point %d", strongest.ID, strongest.Point, requestPoint)
	return strongest
}
