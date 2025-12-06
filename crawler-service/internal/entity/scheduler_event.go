package entity

type UpdateSchedulerEventRequest struct {
	Id    string          `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Event *SchedulerEvent `protobuf:"bytes,2,opt,name=event,proto3" json:"event,omitempty"`
}

type SchedulerEvent struct {
	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Url         string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	Method      string `protobuf:"bytes,3,opt,name=method,proto3" json:"method,omitempty"`
	Description string `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	Queue       string `protobuf:"bytes,5,opt,name=queue,proto3" json:"queue,omitempty"`
	Domain      string `protobuf:"bytes,6,opt,name=domain,proto3" json:"domain,omitempty"`
	IsActive    bool   `protobuf:"varint,7,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty"`
	NextRunTime int64  `protobuf:"varint,8,opt,name=next_run_time,json=nextRunTime,proto3" json:"next_run_time,omitempty"`
	RepeatTimes int64  `protobuf:"varint,9,opt,name=repeat_times,json=repeatTimes,proto3" json:"repeat_times,omitempty"`
	SchedulerAt int64  `protobuf:"varint,10,opt,name=scheduler_at,json=schedulerAt,proto3" json:"scheduler_at,omitempty"`
	Status      string `protobuf:"bytes,11,opt,name=status,proto3" json:"status,omitempty"`
	CronExp     string `protobuf:"bytes,12,opt,name=cron_exp,json=cronExp,proto3" json:"cron_exp,omitempty"`
	CreatedAt   string `protobuf:"bytes,13,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt   string `protobuf:"bytes,14,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}

type StatusEnum string

const (
	StatusPending   StatusEnum = "pending"
	StatusRunning   StatusEnum = "running"
	StatusFailed    StatusEnum = "failed"
	StatusSuccessed StatusEnum = "successed"
	StatusDelete    StatusEnum = "delete"
)
