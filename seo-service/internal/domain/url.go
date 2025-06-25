package domain

import (
	"time"
)

type Url struct {
	Id          int64     `gorm:"column:id;primaryKey" json:"id"`
	Url         string    `gorm:"column:url;type:text" json:"url"`
	Name        string    `gorm:"column:name"  json:"name"`
	Tittle      string    `gorm:"column:tittle;type:text" json:"tittle"`
	Description string    `gorm:"column:description;type:text"  json:"description"`
	Template    string    `gorm:"column:template"  json:"template"`
	Prefix      string    `gorm:"column:prefix"  json:"prefix"`
	Suffix      string    `gorm:"column:suffix"  json:"suffix"`
	Domain      string    `gorm:"column:domain"  json:"domain"`
	IsActive    bool      `gorm:"column:is_active"  json:"is_active"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u Url) TableName() string {
	return "url"
}
