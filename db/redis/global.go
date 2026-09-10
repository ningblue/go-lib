package redis

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// 全局客户端持有（与 db/gorm 的 Init/Get 模式一致，便于各服务 dal 层统一初始化）。
// 业务代码经 Get() 取；可选能力（OIDC state、缓存）使用，未初始化即 panic 快速暴露装配缺失。

var globalClient *redis.Client

// ErrNotInitialized 表示尚未调用 Init。
var ErrNotInitialized = errors.New("redis: not initialized (call Init first)")

// Init 以地址初始化全局客户端。addr 形如 "127.0.0.1:6379"。
func Init(addr, username, password string, db int) error {
	if strings.TrimSpace(addr) == "" {
		return errors.New("redis: addr is required")
	}
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       db,
		PoolSize: DefaultPoolSize,
	})
	if err := c.Ping(context.Background()).Err(); err != nil {
		_ = c.Close()
		return err
	}
	globalClient = c
	return nil
}

// Get 返回全局客户端；未初始化 panic（装配错误应启动期暴露，不静默兜底）。
func Get() *redis.Client {
	if globalClient == nil {
		panic(ErrNotInitialized)
	}
	return globalClient
}

// Close 关闭全局客户端。
func Close() error {
	if globalClient == nil {
		return nil
	}
	err := globalClient.Close()
	globalClient = nil
	return err
}

// SetEx 写入带过期时间的键。
func SetEx(ctx context.Context, key, value string, ttl time.Duration) error {
	return Get().Set(ctx, key, value, ttl).Err()
}

// GetDel 读取并删除（一次性凭证：state / callback code 消费语义）。
// 键不存在返回 ("", false, nil)——调用方据此判定凭证失效，不区分“从未存在/已消费”。
func GetDel(ctx context.Context, key string) (string, bool, error) {
	v, err := Get().GetDel(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}
