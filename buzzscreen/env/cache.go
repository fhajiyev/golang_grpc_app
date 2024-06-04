package env

import (
	"time"

	"github.com/go-redis/cache"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var (
	codec *cache.Codec
)

func getCodec() *cache.Codec {
	if codec == nil {
		codec = &cache.Codec{
			Redis: GetRedis(),
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

// SetCacheAsync func definition
func SetCacheAsync(key string, obj interface{}, expiration time.Duration) {
	go SetCache(key, obj, expiration)
}

// SetCache func definition
func SetCache(key string, obj interface{}, expiration time.Duration) error {
	//lock.Lock()
	//defer lock.Unlock()
	err := getCodec().Set(&cache.Item{
		Key:        key,
		Object:     obj,
		Expiration: expiration,
	})
	return err
}

// GetCache func definition
func GetCache(key string, obj interface{}) error {
	return getCodec().Get(key, obj)
}
