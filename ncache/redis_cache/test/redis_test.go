package rediscache_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ncache"
	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

// RedisPool Redis连接池
var redisService *rediscache.RedisService

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	conf := &nyaml.YamlConfRedis{
		RedisHost:      "8.137.54.220",
		RedisPort:      6379,
		RedisPwd:       "Nxq@198943",
		DataBaseIdx:    0,
		ConnectTimeout: 3,
		ReadTimeout:    30,
		WriteTimeout:   30,
		MaxIdle:        30,
		MaxActive:      100,
		IdleTimeout:    100,
	}
	pool := ncache.NewRedisPool(conf)
	redisService = &rediscache.RedisService{RedisPool: pool}
}

func TestScanByCur(t *testing.T) {
	for i := 1; i < 51; i++ {
		redisService.PutExStr(fmt.Sprintf("test%d", i), fmt.Sprintf("test%d", i), 500)
	}
	fmt.Println("已添加测试数据")
	clearNum, err := redisService.ClearByKeyPrefix("test")
	if nil != err {
		fmt.Println("删除失败", err.Error())
	} else {
		fmt.Println("删除成功", clearNum)
	}
}

func TestPutExStr(t *testing.T) {
	err := redisService.PutExStr("s1", "123", 10)
	if nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println("写入成功")
	}
}

func TestPutStr(t *testing.T) {
	err := redisService.PutStr("s1", "123")
	if nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println("写入成功")
	}
}

func TestGetStr(t *testing.T) {
	x0, err := redisService.GetStr("s1")
	if nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println(x0)
	}
}

func TestExist(t *testing.T) {
	x0 := redisService.ExistWithoutErr("s1")
	fmt.Println(x0)
}

func TestExpireKey(t *testing.T) {
	err := redisService.KeySetExpire("s1", 1000)
	if nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println("更新成功")
	}
}

func TestIncr(t *testing.T) {
	num, err := redisService.Int64Incr("aaaa", 10000)
	if nil != err {
		fmt.Println(err.Error())
	} else {
		fmt.Println(num)
	}
}

func TestMutexCreate(t *testing.T) {
	mutex := rediscache.RedisNewMutex("TestMutexCreate", "1", redisService)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
	}
}

func TestMutexCreateWithParams(t *testing.T) {
	op1 := rediscache.RedisMutexSetDelay(time.Minute)
	op2 := rediscache.RedisMutexSetDelay(time.Duration(500) * time.Millisecond)
	mutex := rediscache.RedisNewMutex("TestMutexCreateWithParams", "1", redisService, op1, op2)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
	}
}

func TestMutex3(t *testing.T) {
	for i := range 3 {
		go func(idx int) {
			ntools.SlogSetTraceId(fmt.Sprintf("v%d", idx))
			result, err := redisService.LockRun("lock1", fmt.Sprintf("v%d", idx), 5, 6, 300, func() any {
				time.Sleep(2 * time.Second)
				return fmt.Sprintf("这是[%d]返回的", idx)
			})
			if nil != err {
				slog.Error(err.Error())
			} else {
				slog.Info(result.(string))
			}
		}(i)
	}
	time.Sleep(5 * time.Second)
}

func TestProducerConsumer(t *testing.T) {
	key := "aaaa"
	for i := 1; i < 51; i++ {
		redisService.Producer(key, fmt.Sprintf("%d", i))
	}
	reciveChan := make(chan string, 10)
	go func() {
		for v := range reciveChan {
			fmt.Println(v)
		}
	}()
	redisService.Consumer(key, reciveChan)
}
