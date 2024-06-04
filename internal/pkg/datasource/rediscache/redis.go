package rediscache

import (
	"time"

	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// RedisCache type definition
type RedisCache struct {
	client *redis.Client
}

// SetCacheAsync func definition
func (r *RedisCache) SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	go r.SetCache(key, obj, expiration)
}

// SetCache func definition
func (r *RedisCache) SetCache(key string, obj interface{}, expiration time.Duration) error {
	//lock.Lock()
	//defer lock.Unlock()
	err := r.getCodec().Set(&cache.Item{
		Key:        key,
		Object:     obj,
		Expiration: expiration,
	})
	return err
}

// GetCache func definition
func (r *RedisCache) GetCache(key string, obj interface{}) error {
	return r.getCodec().Get(key, obj)
}

// DeleteCache func definition
func (r *RedisCache) DeleteCache(key string) error {
	return r.getCodec().Delete(key)
}

var (
	codec *cache.Codec
)

func (r *RedisCache) getCodec() *cache.Codec {
	if codec == nil {
		codec = &cache.Codec{
			Redis: r.client,
			Marshal: func(v interface{}) ([]byte, error) {
				return msgpack.Marshal(v)
			},
			Unmarshal: func(b []byte, v interface{}) error {
				return msgpack.Unmarshal(b, v)
			},
		}
		codec.UseLocalCache(1000, time.Minute)
	}
	return codec
}

// NewSource func definition
func NewSource(client *redis.Client) *RedisCache {
	return &RedisCache{client}
}
