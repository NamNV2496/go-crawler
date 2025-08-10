package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/redis/go-redis/v9"
)

type CrawlerEvent struct {
	Id          int64     `json:"id"`
	Url         string    `json:"url"`
	Method      string    `json:"method"`
	Description string    `json:"description"`
	Queue       string    `json:"queue"`
	Quantity    int64     `json:"quantity"`
	Domain      string    `json:"domain"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
