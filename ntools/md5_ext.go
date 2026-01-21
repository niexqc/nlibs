package ntools

import (
	"crypto/md5"
	"fmt"
	"strings"
)

func MD5Str(str string, upper bool) string {
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	if upper {
		return strings.ToUpper(md5Str)
	} else {
		return md5Str
	}
}
