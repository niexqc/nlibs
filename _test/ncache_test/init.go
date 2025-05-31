package ncache_test

import (
	"time"

	"github.com/niexqc/nlibs/ncache"
	mencache "github.com/niexqc/nlibs/ncache/mem_cache"
	rediscache "github.com/niexqc/nlibs/ncache/redis_cache"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var memCacheService *mencache.MemCacheService
var redisService *rediscache.RedisService

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	memCacheService = ncache.NewMemCacheService(time.Minute)
	conf := &nyaml.YamlConfRedis{
		RedisHost:      "8.137.54.220",
		RedisPort:      6379,
		RedisPwd:       "Nxq@198943",
		DataBaseIdx:    0,
		ConnectTimeout: 2,
		ReadTimeout:    5,
		WriteTimeout:   5,
		MaxIdle:        30,
		MaxActive:      100,
		IdleTimeout:    100,
	}
	redisService = ncache.NewRedisService(ncache.NewRedisPool(conf))
}
