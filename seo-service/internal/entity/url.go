package entity

import "time"

type Url struct {
	Id          int64          `json:"id"`
	Url         string         `json:"url"`
	Name        string         `json:"name"`
	Tittle      string         `json:"tittle"`
	Description string         `json:"description"`
	Template    string         `json:"template"`
	Prefix      string         `json:"prefix"`
	Suffix      string         `json:"suffix"`
	Metadata    []*UrlMetadata `json:"metadata"`
	Domain      string         `json:"domain"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
