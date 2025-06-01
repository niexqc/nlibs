package sqlext

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/shopspring/decimal"
)

var SqlParamArgsRegexp = regexp.MustCompile(`\?`)

// Sql参数格式化.只支持?格式
func SqlFmt(sqlStr string, args ...any) (string, error) {
	if len(args) > 0 {
		splTexts := []string{}
		argsRange := SqlParamArgsRegexp.FindAllStringIndex(sqlStr, -1)
		splTexts = append(splTexts, sqlStr[0:argsRange[0][0]])
		for idx := 1; idx < len(argsRange); idx++ {
			splTexts = append(splTexts, sqlStr[argsRange[idx-1][1]:argsRange[idx][0]])
		}
		splTexts = append(splTexts, sqlStr[argsRange[len(argsRange)-1][1]:])
		sqlStr = splTexts[0]
		for idx, v := range args {
			//判断args是否是数组
			if nlibs.IsArrayOrSlice(v) {
				return "", nerror.NewRunTimeErrorFmt("参数【%v】不能为Array|Slice", v)
			}
			sqlStr += sqlFmtSqlAnyArg(v) + splTexts[idx+1]
		}
	}
	return sqlStr, nil
}

// 使用In查询返回没有记录的 参数
// 例如,数据库中存在1,2,3两条记录,如果参数传入[1,5,6],则结果为[5,6]
func SqlFmtSqlInNotExist[T NdbBasicType](tableName, dbFieldName string, args []T) (sqlStr string, allArgs []T, err error) {
	if len(args) < 1 {
		return sqlStr, allArgs, nerror.NewRunTimeError("参数个数必须大于0")
	}

	sqlStr = `SELECT t1.%s  
FROM (%s) t1 
LEFT JOIN (%s) t2 ON t1.%s=t2.%s 
WHERE t2.%s IS NULL ORDER BY t1.%s ASC`

	t1SqlStr := ""
	for idx := range args {
		if nlibs.IsArrayOrSlice(args[idx]) {
			return sqlStr, allArgs, nerror.NewRunTimeErrorFmt("参数【%v】不能为Array|Slice", args[idx])
		}
		if idx > 0 {
			t1SqlStr += " UNION ALL "
		}
		t1SqlStr += fmt.Sprintf(" SELECT ? AS %s", dbFieldName)
	}

	t2SqlStr := fmt.Sprintf(" SELECT %s FROM  %s WHERE %s IN (?)", dbFieldName, tableName, dbFieldName)
	t2SqlStr, t2Args, err := sqlx.In(t2SqlStr, args) // []T 和 []interface{}（即 []any）类型不兼容，无法直接赋值。
	if nil != err {
		return sqlStr, allArgs, nil
	}
	//将t2的参数追加到参数中
	for _, v := range t2Args {
		args = append(args, v.(T))
	}

	sqlStr = fmt.Sprintf(sqlStr, dbFieldName, t1SqlStr, t2SqlStr, dbFieldName, dbFieldName, dbFieldName, dbFieldName)
	return sqlStr, args, nil
}

func sqlFmtSqlAnyArg(arg any) string {
	//如果是指针需要解引用
	argv := ntools.AnyElem(arg)
	switch v := argv.(type) {
	case nil:
		return "NULL"
	case bool:
		return fmt.Sprintf("%v", v)
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int, uint, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	case time.Time:
		return fmt.Sprintf("'%v'", ntools.Time2Str(v))
	default:
		// 反射检查底层类型（例如处理自定义类型）
		rt := reflect.TypeOf(arg)
		if nullType, str := sqlFmtNNullTypeResult(arg, rt); nullType {
			return str
		} else {
			rv := reflect.ValueOf(arg)
			switch rv.Kind() {
			case reflect.String:
				return fmt.Sprintf("'%s'", rv.String())
			case reflect.Int, reflect.Int64, reflect.Float64:
				return fmt.Sprintf("%v", rv.Interface())
			default:
				return fmt.Sprintf("'%v'", arg)
			}
		}
	}
}

func sqlFmtNNullTypeResult(arg any, rt reflect.Type) (bool, string) {
	if rt == reflect.TypeOf(NullTime{}) {
		nullv := arg.(NullTime)
		if nullv.Valid {
			return true, fmt.Sprintf("'%v'", ntools.Time2Str(nullv.Time))
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(NullString{}) {
		nullv := arg.(NullString)
		if nullv.Valid {
			return true, fmt.Sprintf("'%v'", nullv.String)
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(NullInt{}) {
		nullv := arg.(NullInt)
		if nullv.Valid {
			return true, fmt.Sprintf("%v", nullv.Int32)
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(NullInt64{}) {
		nullv := arg.(NullInt64)
		if nullv.Valid {
			return true, fmt.Sprintf("%v", nullv.Int64)
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(NullFloat64{}) {
		nullv := arg.(NullFloat64)
		if nullv.Valid {
			return true, fmt.Sprintf("%v", nullv.Float64)
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(NullBool{}) {
		nullv := arg.(NullBool)
		if nullv.Valid {
			return true, fmt.Sprintf("%v", nullv.Bool)
		} else {
			return true, "NULL"
		}
	} else if rt == reflect.TypeOf(decimal.NullDecimal{}) {
		nullv := arg.(decimal.NullDecimal)
		if nullv.Valid {
			return true, fmt.Sprintf("%v", nullv.Decimal)
		} else {
			return true, "NULL"
		}
	}

	return false, ""
}
