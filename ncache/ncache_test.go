package ncache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ncache"
)

var ncahceService *ncache.NcahceService

func init() {
	ncahceService = ncache.NewNcahceService(60 * time.Second)
}

func Test1(t *testing.T) {
	ncahceService.PutExStr("aaa", "111", 61)
	v, _ := ncahceService.GetStr("aaa")
	fmt.Printf("%v -> %v", "111", v)
}
