package nredis_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/niexqc/nlibs/nredis"
)

// RedisPool Redis连接池
var redisPool *redis.Pool
var redisService *nredis.RedisService

func init() {
	redisHost := "8.137.54.220"
	redisPort := 6379
	redisPwd := "Nxq@198943"

	redisIdleTimeout := 100
	redisMaxidle := 10
	redisMaxactive := 2

	address := fmt.Sprintf("%s:%d", redisHost, redisPort)
	dbOption := redis.DialDatabase(0)
	pwOption := redis.DialPassword(redisPwd)
	// **重要** 设置读写超时
	readTimeout := redis.DialReadTimeout(time.Second * time.Duration(30))
	writeTimeout := redis.DialWriteTimeout(time.Second * time.Duration(30))
	conTimeout := redis.DialConnectTimeout(time.Second * time.Duration(30))

	// 建立连接池
	redisPool = &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     redisMaxidle,
		MaxActive:   redisMaxactive,
		IdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
		//如果空闲列表中没有可用的连接,且当前Active连接数 < MaxActive,则等待
		Wait: true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", address, dbOption, pwOption, readTimeout, writeTimeout, conTimeout)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	redisService = &nredis.RedisService{RedisPool: redisPool}
}

func TestMutex(t *testing.T) {

	mutex := nredis.NewMutex("lock111", redisService)
	if mutex.Lock() {
		defer mutex.ReleseLock()
		//执行逻辑
		fmt.Println("第1个获取到了")
	}
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
	err := redisService.ExpireKey("s1", 1000)
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
