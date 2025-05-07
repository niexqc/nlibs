package nlibs

import (
	"reflect"

	"github.com/niexqc/nlibs/ntools"
)

var FileDirExt = ntools.GetFileDirExt()

var HttpClientExt = ntools.GetHttpClientExt()

func IsArrayOrSlice(v any) bool {
	rv := reflect.ValueOf(v)
	kind := rv.Kind()
	return kind == reflect.Array || kind == reflect.Slice
}
