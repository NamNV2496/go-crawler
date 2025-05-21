package domain

import (
	"time"
)

type Queue struct {
	Id        int64
	Queue     string
	Domain    string
	Cron      string
	Quantity  int
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
