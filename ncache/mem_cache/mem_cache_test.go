package mencache_test

import (
	"fmt"
	"testing"
	"time"

	mencache "github.com/niexqc/nlibs/ncache/mem_cache"
	"github.com/patrickmn/go-cache"
)

var ncahceService *mencache.MemCacheService

func init() {
	ncahceService = &mencache.MemCacheService{
		Cache: cache.New(0, time.Minute),
	}
}

func Test1(t *testing.T) {
	ncahceService.PutExStr("aaa", "111", 61)
	v, _ := ncahceService.GetStr("aaa")
	fmt.Printf("%v -> %v", "111", v)
}
