package distributedlock

import (
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	goredis "github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/namnv2496/scheduler/internal/configs"
	goredislib "github.com/redis/go-redis/v9"
)

type IDistributedLock interface {
	Lock(lockName string, expiry time.Duration) (*redsync.Mutex, error)
	Unlock(lock *redsync.Mutex) error
}

type DistributedLock struct {
	client *redsync.Redsync
}

func NewDistributedLock(
	conf *configs.Config,
) IDistributedLock {
	client := goredislib.NewClient(&goredislib.Options{
		Addr: conf.Redis.Addr,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion lock
	rs := redsync.New(pool)
	return &DistributedLock{
		client: rs,
	}
}

func (_self *DistributedLock) Lock(lockName string, expiry time.Duration) (*redsync.Mutex, error) {
	mutex := _self.client.NewMutex(lockName, redsync.WithExpiry(expiry))
	if err := mutex.Lock(); err != nil {
		return nil, err
	}
	return mutex, nil
}

func (_self *DistributedLock) Unlock(lock *redsync.Mutex) error {
	if ok, err := lock.Unlock(); !ok || err != nil {
		return fmt.Errorf("unlock failed: %s", err)
	}
	return nil
}
