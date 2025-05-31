package ncache_test

import (
	"testing"
	"time"

	"github.com/niexqc/nlibs/ntools"
)

func TestMemPutExStr(t *testing.T) {
	memCacheService.PutExStr("aaa", "111", 1)
	v, _ := memCacheService.GetStr("aaa")
	ntools.TestEq(t, "MemCacheService TestPutExStr", "111", v)

	time.Sleep(2 * time.Second)

	_, err := memCacheService.GetStr("aaa")
	if err == nil {
		ntools.TestErrPainic(t, "MemCacheService TestPutExStr 此时应该返回错误", nil)
	}
	ntools.TestEq(t, "MemCacheService TestPutExStr 2秒后", "缓存不存在", err.Error())
}
