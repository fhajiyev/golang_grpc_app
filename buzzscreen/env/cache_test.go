package env_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Buzzvil/buzzscreen-api/buzzscreen/env"
)

func TestRedisCache(t *testing.T) {
	now := time.Now().Unix()
	cacheKey := fmt.Sprintf("TEMP_CACHE_KEY_%d", now)
	env.SetCache(cacheKey, now, time.Second*10)
	var nowFromCache int64
	if err := env.GetCache(cacheKey, &nowFromCache); err == nil {
		t.Logf("TestRedisCache() - now: %v", now)
	} else {
		t.Fatalf("TestRedisCache() - cache shouldn't expired. nowFromCache - %d", nowFromCache)
	}
}
