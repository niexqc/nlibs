package ncache_test

import (
	"os"
	"time"

	"github.com/niexqc/nlibs/ncache"
	mencache "github.com/niexqc/nlibs/ncache/mem_cache"
	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var memCacheService *mencache.MemCacheService
var redisService *rediscache.RedisService

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	memCacheService = ncache.NewMemCacheService(time.Minute)
	conf := &nyaml.YamlConfRedis{
		RedisHost:      "",
		RedisPort:      6379,
		RedisPwd:       "",
		DataBaseIdx:    0,
		ConnectTimeout: 2,
		ReadTimeout:    5,
		WriteTimeout:   5,
		MaxIdle:        30,
		MaxActive:      100,
		IdleTimeout:    100,
	}

	conf.RedisHost = os.Getenv("Niexq_Test_Host")
	conf.RedisPwd = os.Getenv("Niexq_Test_Redis_Pwd")
	if conf.RedisHost == "" || conf.RedisPwd == "" {
		panic(nerror.NewRunTimeError("请在环境变量配置:Niexq_Test_Host,Niexq_Test_Redis_Pwd"))
	}

	redisService = ncache.NewRedisService(ncache.NewRedisPool(conf))
}
