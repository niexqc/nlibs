package rediscache_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
)

func TestMutex(t *testing.T) {

	mutex := rediscache.RedisNewMutex("lock1", "1", redisService)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
		//执行逻辑
		fmt.Println("第1个获取到了")
	}
}
func TestMutex02(t *testing.T) {
	op1 := rediscache.RedisMutexSetDelay(time.Minute)
	op2 := rediscache.RedisMutexSetDelay(time.Duration(5) * time.Second)
	mutex := rediscache.RedisNewMutex("lock1", "1", redisService, op1, op2)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
		slog.Info("第1个获取到了")
	}

	mutex2 := rediscache.RedisNewMutex("lock1", "1", redisService, op1, op2)
	if mutex2.RedisLock() {
		defer mutex2.RedisReleseLock()
		//执行逻辑
		slog.Info("第2个获取到了")
	}
}
