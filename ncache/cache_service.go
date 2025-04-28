package ncache

import (
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
