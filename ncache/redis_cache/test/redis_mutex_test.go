package rediscache_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
	"github.com/niexqc/nlibs/ntools"
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

func TestMutex3(t *testing.T) {

	for i := 0; i < 3; i++ {
		go func(idx int) {
			ntools.SlogSetTraceId(fmt.Sprintf("v%d", idx))
			result, err := redisService.LockRun("lock1", fmt.Sprintf("v%d", idx), 10, 3, 1, func() any {
				time.Sleep(5 * time.Second)
				return fmt.Sprintf("这是[%d]返回的", idx)
			})
			if nil != err {
				slog.Error(err.Error())
			} else {
				slog.Info(result.(string))
			}
		}(i)
	}
	time.Sleep(16 * time.Second)

}
