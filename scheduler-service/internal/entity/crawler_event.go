package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/namnv2496/scheduler/internal/domain"
	"github.com/redis/go-redis/v9"
)

type CrawlerEvent struct {
	Id          int64             `json:"id"`
	Url         string            `json:"url"`
	Method      string            `json:"method"`
	Description string            `json:"description"`
	Queue       string            `json:"queue"`
	Domain      string            `json:"domain"`
	IsActive    bool              `json:"is_active"`
	NextRunTime int64             `json:"next_run_time"`
	RepeatTimes int64             `json:"repeat_times"`
	SchedulerAt int64             `json:"scheduler_at"`
	Status      domain.StatusEnum `json:"status"`
	CronExp     string            `json:"cron_exp"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func (_self CrawlerEvent) HashKey(key any) string {
	hash, err := hashstructure.Hash(key, hashstructure.FormatV2, nil)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%d", hash)

}

func (_self CrawlerEvent) Seriablize(key any) string {
	data, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (_self CrawlerEvent) Deserialize(data any, output any) error {
	return json.Unmarshal([]byte(data.(string)), output)
}

func (_self CrawlerEvent) Incr(ctx context.Context, key any) *redis.IntCmd {
	return nil
}

func (_self CrawlerEvent) Decr(ctx context.Context, key any) *redis.IntCmd {
	return nil
}

func (_self CrawlerEvent) Expire(ctx context.Context, key any, expiredTime time.Duration) error {
	return nil
}

func (c CrawlerEvent) ToMap() map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(c)
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the json tag name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			continue
		}
		paramName := strings.Split(jsonTag, ",")[0]
		if field.Kind() == reflect.Struct || field.Kind() == reflect.Slice {
			continue
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				result[paramName] = ""
			} else {
				result[paramName] = fmt.Sprintf("%v", field.Elem().Interface())
			}
		} else {
			result[paramName] = fmt.Sprintf("%v", field.Interface())
		}
	}
	return result
}
