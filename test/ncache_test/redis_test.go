package ncache_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
	"github.com/niexqc/nlibs/ntools"
)

func TestRedisClearByKeyPrefix(t *testing.T) {
	prefix := "TestClearByKeyPrefix"
	for i := 1; i < 51; i++ {
		redisService.PutExStr(fmt.Sprintf(prefix+"%d", i), fmt.Sprintf(prefix+"%d", i), 500)
	}
	clearNum, err := redisService.ClearByKeyPrefix(prefix)
	ntools.TestErrPainic(t, "测试 TestClearByKeyPrefix", err)
	slog.Info("实际清理:", "num", clearNum)
	ntools.TestEq(t, "测试 TestClearByKeyPrefix  删除数量", 50, clearNum)
}

func TestRedisPutExStr(t *testing.T) {
	redisService.PutExStr("aaa", "111", 1)
	v, _ := redisService.GetStr("aaa")
	ntools.TestEq(t, "TestRedisPutExStr", "111", v)

	time.Sleep(2 * time.Second)

	_, err := redisService.GetStr("aaa")
	if err == nil {
		ntools.TestErrPanicMsg(t, "TestRedisPutExStr 此时应该返回错误")
	}
	ntools.TestEq(t, "TestRedisPutExStr 2秒后", "redigo: nil returned", err.Error())
}

func TestRedisPutStr(t *testing.T) {
	err := redisService.PutStr("s1", "123")
	ntools.TestErrPainic(t, "测试 TestRedisPutStr", err)
	v, _ := redisService.GetStr("s1")
	ntools.TestEq(t, "TestRedisPutStr", "123", v)

	//清理掉
	err = redisService.ClearKey("s1")
	ntools.TestErrPainic(t, "测试 TestRedisPutStr", err)
}

func TestRedisExist(t *testing.T) {
	redisService.PutStr("s1", "123")
	x0 := redisService.ExistWithoutErr("s1")
	ntools.TestEq(t, "TestRedisExist", true, x0)
	//清理掉
	redisService.ClearKey("s1")
	x0 = redisService.ExistWithoutErr("s1")
	ntools.TestEq(t, "TestRedisExist", false, x0)
}

func TestKeyRestExpire(t *testing.T) {
	redisService.PutStr("s1", "123")
	err := redisService.KeySetExpire("s1", 1)
	ntools.TestErrPainic(t, "测试 TestKeyRestExpire", err)
	//1.2后判断是否存在
	slog.Info("设置1秒后过去,1.2秒后去查看")
	time.Sleep(1200 * time.Millisecond)

	x0 := redisService.ExistWithoutErr("s1")
	ntools.TestEq(t, "TestKeyRestExpire", false, x0)
}

func TestRedisIncr(t *testing.T) {
	redisService.Int64Incr("TestRedisIncr", 2500)
	redisService.Int64Incr("TestRedisIncr", 2500)
	redisService.Int64Incr("TestRedisIncr", 2500)
	num, err := redisService.Int64Incr("TestRedisIncr", 2500)
	ntools.TestErrPainic(t, "测试 TestRedisIncr", err)
	ntools.TestEq(t, "TestKeyRestExpire", int64(4), num)
}

func TestRedisMutexCreate(t *testing.T) {
	mutexLock1 := rediscache.RedisNewMutex("TestMutexCreate", "1", redisService)
	if mutexLock1.RedisLock() {
		defer mutexLock1.RedisReleseLock()
	}
}

func TestRedisMutexCreateWithParams(t *testing.T) {
	op1 := rediscache.RedisMutexSetDelay(time.Minute)
	op2 := rediscache.RedisMutexSetDelay(time.Duration(500) * time.Millisecond)
	mutex := rediscache.RedisNewMutex("TestMutexCreateWithParams", "1", redisService, op1, op2)
	if mutex.RedisLock() {
		defer mutex.RedisReleseLock()
	}
}

func TestRedisMutexLockRun(t *testing.T) {
	//耗时操作设置为2秒,启动3个协程后等待1秒获取计数器结果只能为1，否则没有达到锁定的效果
	counter := 0
	for i := range 3 {
		go func(idx int) {
			ntools.SlogSetTraceId(fmt.Sprintf("v%d", idx))
			result, err := redisService.LockRun("lock1", fmt.Sprintf("v%d", idx), 5, 6, 300, func() any {
				counter++
				//模拟耗时操作
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
	time.Sleep(1 * time.Second)
	ntools.TestEq(t, "TestRedisMutexLockRun 耗时操作设置为2秒,启动3个协程后等待1秒获取计数器结果只能为1 ", 1, counter)
}

func TestRedisProducerConsumer(t *testing.T) {
	key := "aaaa"
	for i := 1; i < 3; i++ {
		redisService.Producer(key, fmt.Sprintf("%d", i))
	}
	reciveChan := make(chan string, 1)
	go func() {
		for v := range reciveChan {
			fmt.Println(v)
		}
	}()
	slog.Info("如果要接收所有的 序列测试方法会暂停")
	// redisService.Consumer(key, reciveChan)
}
