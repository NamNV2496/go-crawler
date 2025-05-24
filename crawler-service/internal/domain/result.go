package domain

import (
	"gorm.io/gorm"
)

type Result struct {
	gorm.Model
	Id     int64  `gorm:"column:id;primaryKey" json:"id"`
	Url    string `gorm:"column:url;type:text" json:"url"`
	Method string `gorm:"column:method;type:text" json:"method"`
	Queue  string `gorm:"column:queue"  json:"queue"`
	Domain string `gorm:"column:domain"  json:"domain"`
	Result string `gorm:"column:result"  json:"result"`
}

func (Result) TableName() string {
	return "result"
}
