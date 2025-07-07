package redis

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

const DefaultPoolSize = 50

var (
	expireTime = time.Hour * 1
)

type Redis struct {
	hosts    string
	port     int
	name     string
	password string
	db       int
}

func NewRedis(hosts string, port int, name, password string, db int) *Redis {
	return &Redis{
		hosts:    hosts,
		port:     port,
		name:     name,
		password: password,
		db:       db,
	}
}

func NewRedisDriver(hosts string, port int, name, password string, db int, poolSize int) (rdb *redis.Client) {
	if poolSize == 0 {
		poolSize = DefaultPoolSize
	}
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", hosts, port),
		Username: name,
		Password: password,
		DB:       db,
		PoolSize: poolSize,
	})
	return rdb
}
