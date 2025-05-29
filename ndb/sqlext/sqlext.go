package sqlext

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

type NdbBasicType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string | ~bool |
		time.Time | NullBool | NullFloat64 | NullInt | NullInt64 | NullString | NullTime
}

var blankRegexp = regexp.MustCompile(`\s+`)

func PrintSql(dbConf *nyaml.YamlConfSqlPrint, start time.Time, sqlStr string, args ...any) {
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

	mapVals := map[string]any{}
	sb := strings.Builder{}
	for i := range objType.NumField() {
		field := objType.Field(i)
		tagDb := field.Tag.Get("db")
		//解析字段类型
		valV := objVal.Field(i).Interface()
		mapVals[tagDb] = valV
		if sb.Len() > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("?")
	}

	dbFieldStrs := strings.SplitSeq(insertField, ",")
	for v := range dbFieldStrs {
		vals = append(vals, mapVals[v])
	}
	return sb.String(), vals, nil
}

// 基础类型切片展开为为any切片
func ArrBaseTypeExpand2ArrAny[T NdbBasicType](args []T) []any {
	anyArgs := make([]any, len(args))
	for i, v := range args {
		anyArgs[i] = v
	}
	return anyArgs
}

func StructDoTableName(doType reflect.Type) string {
	if doType.NumField() <= 0 {
		panic(nerror.NewRunTimeErrorFmt("%s没有字段", doType.Name()))
	}
	dbtbTag := doType.Field(0).Tag
	tbname := dbtbTag.Get("dbtb")
	if tbname == "" {
		panic(nerror.NewRunTimeErrorFmt("%s字段的Tag没有标识[dbtb]", doType.Name()))
	}
	return tbname
}

func StructDoDbColList(doType reflect.Type, tableAlias string) []string {
	if doType.NumField() <= 0 {
		panic(nerror.NewRunTimeErrorFmt("%s没有字段", doType.Name()))
	}
	result := []string{}
	//字段
	for idx := range doType.NumField() {
		dbTag := doType.Field(idx).Tag
		dbcol := dbTag.Get("db")
		if dbcol == "" {
			panic(nerror.NewRunTimeErrorFmt("%s字段的Tag没有标识[db]", doType.Name()))
		}
		if tableAlias == "" {
			result = append(result, dbcol)
		} else {
			result = append(result, fmt.Sprintf("%s.%s", tableAlias, dbcol))
		}
	}
	return result
}

func StructDoDbColStr(doType reflect.Type, tableAlias string) string {
	sb := &strings.Builder{}
	cols := StructDoDbColList(doType, tableAlias)
	for idx, v := range cols {
		if idx > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(v)
	}
	return sb.String()
}
