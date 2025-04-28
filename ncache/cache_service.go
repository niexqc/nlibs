package ncache

import (
	"strings"
	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/patrickmn/go-cache"
)

// 内存缓存
// NcahceService ...
type NcahceService struct {
	Cache *cache.Cache
}

// 默认永不过期，5分钟淘汰一次的缓存
// cleanupInterval  5*time.Minute
func NewNcahceService(cleanupInterval time.Duration) *NcahceService {
	return &NcahceService{
		Cache: cache.New(0, cleanupInterval),
	}
}

// PutStr ...
func (service *NcahceService) PutStr(key string, val string) error {
	service.Cache.SetDefault(key, val)
	return nil
}

// GetStr ...
func (service *NcahceService) GetStr(key string) (string, error) {
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
func (service *NcahceService) ExistWithoutErr(key string) bool {
	_, ok := service.Cache.Get(key)
	return ok
}

func (service *NcahceService) ExpireKey(key string, sencond int) error {
	v, found := service.Cache.Get(key)
	if !found {
		return nerror.NewRunTimeError("key not found")
	}
	return service.Cache.Replace(key, v, time.Duration(sencond)*time.Second)
}

// ClearByKeyPrefix 清理指定前缀的KEY
func (service *NcahceService) ClearByKeyPrefix(keyPrefix string) (int, error) {
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
func (service *NcahceService) PutExStr(key string, val string, sencond int) error {
	service.ClearKey(key)
	return service.Cache.Add(key, val, time.Duration(sencond)*time.Millisecond)
}

// ClearKey 清理KEY
func (service *NcahceService) ClearKey(key string) error {
	service.Cache.Delete(key)
	return nil
}
