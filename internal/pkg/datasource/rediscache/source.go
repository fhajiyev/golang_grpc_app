package rediscache

import (
	"time"
)

// RedisSource type definition
type RedisSource interface {
	SetCacheAsync(key string, obj interface{}, expiration time.Duration)
	SetCache(key string, obj interface{}, expiration time.Duration) error
	GetCache(key string, obj interface{}) error
	DeleteCache(key string) error
}
