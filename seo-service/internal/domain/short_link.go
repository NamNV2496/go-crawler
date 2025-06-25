package domain

import (
	"time"
)

type ShortLink struct {
	Id          int64     `gorm:"column:id;primaryKey" json:"id"`
	Uri         string    `gorm:"column:uri;type:text" json:"uri"`
	Group       string    `gorm:"column:group;type:text" json:"group"`
	Tittle      string    `gorm:"column:tittle;type:text"  json:"tittle"`
	Description string    `gorm:"column:description;type:text"  json:"description"`
	Filter      string    `gorm:"column:filter;type:json"  json:"filter"`
	IsActive    bool      `gorm:"column:is_active"  json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u ShortLink) TableName() string {
	return "short_link"
}
