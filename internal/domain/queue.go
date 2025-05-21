package domain

import (
	"time"

	"gorm.io/gorm"
)

type Queue struct {
	gorm.Model
	Id        int64     `gorm:"column:id;primaryKey" json:"id"`
	Queue     string    `gorm:"column:queue"  json:"queue"`
	Domain    string    `gorm:"column:domain"  json:"domain"`
	Cron      string    `gorm:"column:cron"  json:"cron"`
	Quantity  int64     `gorm:"column:quantity"  json:"quantity"`
	IsActive  bool      `gorm:"column:is_active"  json:"is_active"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
