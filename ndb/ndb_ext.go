package ndb

import (
	"github.com/niexqc/nlibs/ndb/sqlext"
)

// Sql参数格式化.只支持?格式
// 暂时只简单转换后续再处理或过滤其他字符
func SqlFmt(str string, arg ...any) string {
	return sqlext.SqlFmt(str, arg...)
}
