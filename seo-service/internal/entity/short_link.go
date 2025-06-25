package entity

import "time"

type ShortLink struct {
	Id          int64     `json:"id,omitempty"`
	Uri         string    `json:"uri,omitempty"`
	Group       string    `json:"group,omitempty"`
	Tittle      string    `json:"tittle,omitempty"`
	Description string    `json:"description,omitempty"`
	Filter      string    `sql:"type:JSON" json:"filter,omitempty"`
	IsActive    bool      `json:"is_active,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
