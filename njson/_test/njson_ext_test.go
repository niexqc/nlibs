package njson_ext_test

import (
	"testing"

	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}

func TestErrorExt(t *testing.T) {
	jstr := "1"
	if njson.ToStrOk(jstr) != "\"1\"" {
		t.Errorf("【%s】 njson.ToStrOk(jstr) 失败", jstr)
	}

}
