package sqlext

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var blankRegexp = regexp.MustCompile(`\s+`)
var argsRegexp = regexp.MustCompile(`\?`)

func PrintSql(dbConf *nyaml.YamlConfDb, start time.Time, sqlStr string, args ...any) {
	if !dbConf.DbSqlLogPrint {
		return
	}
	costTime := time.Now().UnixMilli() - start.UnixMilli()
	//去除换行符
	if dbConf.DbSqlLogCompress {
		sqlStr = string(blankRegexp.ReplaceAllString(sqlStr, " "))
	}
	sqlStr = SqlFmt(sqlStr, args...)
	//打印日志
	slog.Log(context.Background(), ntools.SlogLevelStr2Level(dbConf.DbSqlLogLevel), fmt.Sprintf("[%dms] %s", costTime, sqlStr))
}

// Sql参数格式化.只支持?格式
func SqlFmt(sqlStr string, args ...any) string {
	if len(args) > 0 {
		splTexts := []string{}
		argsRange := argsRegexp.FindAllStringIndex(sqlStr, -1)
		splTexts = append(splTexts, sqlStr[0:argsRange[0][0]])
		for idx := 1; idx < len(argsRange); idx++ {
			splTexts = append(splTexts, sqlStr[argsRange[idx-1][1]:argsRange[idx][0]])
		}
		splTexts = append(splTexts, sqlStr[argsRange[len(argsRange)-1][1]:])
		sqlStr = splTexts[0]
		for idx, v := range args {
			sqlStr += sqlAnyArg(v) + splTexts[idx+1]
		}
	}
	return sqlStr
}

func sqlAnyArg(arg any) string {
	switch v := arg.(type) {
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
		if rt == reflect.TypeOf(NullTime{}) {
			nnullTime := arg.(NullTime)
			if nnullTime.Valid {
				return fmt.Sprintf("'%v'", ntools.Time2Str(nnullTime.Time))
			} else {
				return "NULL"
			}
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
