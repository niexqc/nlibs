package ncache

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/niexqc/nlibs/nyaml"
)

func RedisInitPool(conf *nyaml.YamlConfRedis) *redis.Pool {

	redisIdleTimeout := 100
	redisMaxidle := 10
	redisMaxactive := 2

	address := fmt.Sprintf("%s:%d", conf.RedisHost, conf.RedisPort)

	dbOption := redis.DialDatabase(0)
	pwOption := redis.DialPassword(conf.RedisPwd)
	// **重要** 设置读写超时
	readTimeout := redis.DialReadTimeout(time.Second * time.Duration(30))
	writeTimeout := redis.DialWriteTimeout(time.Second * time.Duration(30))
	conTimeout := redis.DialConnectTimeout(time.Second * time.Duration(30))

	// 建立连接池
	redisPool := &redis.Pool{
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
	slog.Debug(address + " " + conf.RedisPwd)
	return redisPool
}

func NewRedisService(redisPool *redis.Pool) *RedisService {
	return &RedisService{RedisPool: redisPool}
}
