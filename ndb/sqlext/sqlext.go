package sqlext

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var NdbTags = struct {
	TableSchema string
	TableName   string
	TableColumn string
}{TableSchema: "schm", TableName: "tbn", TableColumn: "db"}

type NdbBasicType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string | ~bool |
		time.Time | NullBool | NullFloat64 | NullInt | NullInt64 | NullString | NullTime
}

var blankRegexp = regexp.MustCompile(`\s+`)

func PrintSql(sqlPrintConf *nyaml.YamlConfSqlPrint, start time.Time, sqlStr string, args ...any) {
	if !sqlPrintConf.DbSqlLogPrint {
		return
	}
	costTime := time.Now().UnixMilli() - start.UnixMilli()
	//去除换行符
	if sqlPrintConf.DbSqlLogCompress {
		sqlStr = string(blankRegexp.ReplaceAllString(sqlStr, " "))
	}

	sqlStr, err := SqlFmt(sqlStr, args...)
	//打印日志
	if nil != err {
		slog.Log(context.Background(), ntools.SlogLevelStr2Level(sqlPrintConf.DbSqlLogLevel), fmt.Sprintf("[%dms] %s:%v", costTime, "Sql格式化错误", err))
	} else {
		slog.Log(context.Background(), ntools.SlogLevelStr2Level(sqlPrintConf.DbSqlLogLevel), fmt.Sprintf("[%dms] %s", costTime, sqlStr))
	}
}

// insertField 需要用逗号分隔如【aaa,bbb,ccc】
func InserSqlVals(insertField string, dostrcut any) (zwf string, vals []any, err error) {

	objVal := reflect.ValueOf(dostrcut)
	if objVal.Kind() == reflect.Pointer {
		objVal = objVal.Elem() //解引用
	}
	if objVal.Kind() != reflect.Struct {
		return "", nil, nerror.NewRunTimeError("不能获取非结构的值")
	}
	objType := objVal.Type()

	dbFieldIterSeq := strings.SplitSeq(insertField, ",")
	fieldArr := []string{}
	for v := range dbFieldIterSeq {
		fieldArr = append(fieldArr, v)
	}

	mapVals := map[string]any{}
	sb := strings.Builder{}
	for i := range objType.NumField() {
		field := objType.Field(i)
		tagDb := field.Tag.Get(NdbTags.TableColumn)
		if slices.Contains(fieldArr, tagDb) {
			//解析字段类型
			valV := objVal.Field(i).Interface()

			mapVals[tagDb] = valV
			if sb.Len() > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("?")
		}
	}
	for _, v := range fieldArr {
		vals = append(vals, mapVals[v])
	}
	return sb.String(), vals, nil
}
