package mencache

import (
	"strings"
	"sync"
	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/patrickmn/go-cache"
)

// 内存缓存
// MemCacheService ...
type MemCacheService struct {
	Cache *cache.Cache
	nmu   sync.RWMutex
}

// Int64Incr implements INcache.
func (service *MemCacheService) Int64Incr(key string, expireMillisecond int64) (num int64, err error) {
	service.nmu.Lock()
	defer service.nmu.Unlock()
	val, fund := service.Cache.Get(key)
	if !fund {
		err = service.Cache.Add(key, int64(1), time.Duration(expireMillisecond)*time.Second)
		return int64(1), err
	} else {
		if v, cok := val.(int64); cok {
			newv := v + 1
			err = service.Cache.Replace(key, newv, time.Duration(expireMillisecond)*time.Second)
			return newv, err
		}
		panic(nerror.NewRunTimeError("自增的值不是int64"))
	}
}

// PutStr ...
func (service *MemCacheService) PutStr(key string, val string) error {
	service.Cache.SetDefault(key, val)
	return nil
}

// GetStr ...
func (service *MemCacheService) GetStr(key string) (string, error) {
	val, ok := service.Cache.Get(key)
	if ok {
		if str, cok := val.(string); cok {
			return str, nil
		} else {
			return "", nerror.NewRunTimeError("缓存的值非string")
		}
	}
	return "", nerror.NewRunTimeError("缓存不存在")
}

// ExistWithoutErr ...
func (service *MemCacheService) ExistWithoutErr(key string) bool {
	_, ok := service.Cache.Get(key)
	return ok
}

func (service *MemCacheService) KeySetExpire(key string, sencond int) error {
	service.nmu.Lock()
	defer service.nmu.Unlock()
	v, found := service.Cache.Get(key)
	if !found {
		return nerror.NewRunTimeError("key not found")
	}
	return service.Cache.Replace(key, v, time.Duration(sencond)*time.Second)
}

// ClearByKeyPrefix 清理指定前缀的KEY
func (service *MemCacheService) ClearByKeyPrefix(keyPrefix string) (int, error) {
	maps := service.Cache.Items()
	count := 0
	for k, _ := range maps {
		if strings.HasPrefix(k, keyPrefix) {
			count++
			service.ClearKey(k)
		}
	}
	return count, nil
}

// 设置键值对并指定过期时间（​​原子性操作​​）
// 无论键是否存在，都会​​覆盖旧值​​并设置新的过期时间
func (service *MemCacheService) PutExStr(key string, val string, sencond int) error {
	service.nmu.Lock()
	defer service.nmu.Unlock()
	service.ClearKey(key)
	err := service.Cache.Add(key, val, time.Duration(sencond)*time.Second)
	return err
}

// ClearKey 清理KEY
func (service *MemCacheService) ClearKey(key string) error {
	service.Cache.Delete(key)
	return nil
}
