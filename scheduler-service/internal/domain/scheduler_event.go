package domain

import (
	"time"
)

type StatusEnum string

const (
	StatusPending   StatusEnum = "pending"
	StatusRunning   StatusEnum = "running"
	StatusFailed    StatusEnum = "failed"
	StatusSuccessed StatusEnum = "successed"
	StatusDelete    StatusEnum = "delete"
)

func GetStatusEnum(status string) StatusEnum {
	switch status {
	case string(StatusPending):
		return StatusPending
	case string(StatusRunning):
		return StatusRunning
	case string(StatusFailed):
		return StatusFailed
	case string(StatusSuccessed):
		return StatusSuccessed
	case string(StatusDelete):
		return StatusDelete
	default:
		return ""
	}
}

type SchedulerEvent struct {
	Id          int64      `gorm:"column:id;primaryKey" json:"id"`
	Url         string     `gorm:"column:url;type:text" json:"url"`
	Method      string     `gorm:"column:method;type:text" json:"method"`
	Description string     `gorm:"column:description"  json:"description"`
	Queue       string     `gorm:"column:queue"  json:"queue"`
	Domain      string     `gorm:"column:domain"  json:"domain"`
	IsActive    bool       `gorm:"column:is_active"  json:"is_active"`
	NextRunTime int64      `gorm:"column:next_run_time" json:"next_run_time"`
	RepeatTimes int64      `gorm:"column:repeat_times" json:"repeat_times"`
	SchedulerAt int64      `gorm:"column:scheduler_at" json:"scheduler_at"`
	Status      StatusEnum `gorm:"column:status" json:"status"`
	CronExp     string     `gorm:"column:cron_exp" json:"cron_exp"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (u SchedulerEvent) TableName() string {
	return "scheduler_events"
}
