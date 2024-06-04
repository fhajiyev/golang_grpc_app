package env

import (
	"sync"

	"fmt"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzlib-go/datetime"
	"github.com/go-redis/redis"
)

type singletonRedis struct {
	Client *redis.Client
}

var instanceRedis *singletonRedis
var onceRedis sync.Once

// InitRedis func definition
func InitRedis() *redis.Client {
	client := GetRedis()
	if _, err := client.Ping().Result(); err != nil {
		core.Logger.Fatal(err)
	}
	return client
}

// GetRedis func definition
func GetRedis() *redis.Client {
	return getRedisInstance().Client
}

func getRedisInstance() *singletonRedis {
	onceRedis.Do(func() {
		instanceRedis = &singletonRedis{}
		instanceRedis.Client = redis.NewClient(&redis.Options{
			Addr:     Config.Redis.Endpoint,
			Password: "",              // no password set
			DB:       Config.Redis.DB, // use default DB
		})
		core.Logger.Debugf("Redis.getRedisInstance() - Address: %s, DB: %d", Config.Redis.Endpoint, Config.Redis.DB)
	})
	return instanceRedis
}

// SetRedisDeviceDau func definition
func SetRedisDeviceDau(deviceID int64) {
	GetRedis().SetBit(fmt.Sprintf("stat:device:dau:%s", datetime.GetDate("2006-01-02", "Asia/Tokyo")), deviceID, 1)
}
