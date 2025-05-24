package domain

import (
	"time"

	"gorm.io/gorm"
)

type Url struct {
	gorm.Model
	Id          int64     `gorm:"column:id;primaryKey" json:"id"`
	Url         string    `gorm:"column:url;type:text" json:"url"`
	Method      string    `gorm:"column:method;type:text" json:"method"`
	Description string    `gorm:"column:description"  json:"description"`
	Queue       string    `gorm:"column:queue"  json:"queue"`
	Domain      string    `gorm:"column:domain"  json:"domain"`
	IsActive    bool      `gorm:"column:is_active"  json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}
