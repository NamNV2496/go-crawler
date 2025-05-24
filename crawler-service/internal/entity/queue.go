package entity

import (
	"time"
)

const (
	QueueTypeNormal   string = "normal"
	QueueTypePriority string = "priority"
)

type Queue struct {
	Id        int64     `json:"id"`
	Queue     string    `json:"queue"`
	Domain    string    `json:"domain"`
	Cron      string    `json:"cron"`
	Quantity  int       `json:"quantity"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
