package ndb

import (
	"fmt"
	"reflect"
	"strings"

	"slices"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
)

// Sql参数格式化.只支持?格式
// 暂时只简单转换后续再处理或过滤其他字符
func SqlFmt(str string, arg ...any) string {
	return sqlext.SqlFmt(str, arg...)
}

func StructDoTableSchema(doType reflect.Type) string {
	if doType.Kind() == reflect.Pointer {
		doType = doType.Elem() //解引用
	}
	if doType.NumField() <= 0 {
		panic(nerror.NewRunTimeErrorFmt("%s没有字段", doType.Name()))
	}
	dbtbTag := doType.Field(0).Tag
	tbname := dbtbTag.Get(sqlext.NdbTags.TableSchema)
	if tbname == "" {
		panic(nerror.NewRunTimeErrorFmt("%s字段的Tag没有标识[%s]", doType.Name(), sqlext.NdbTags.TableSchema))
	}
	return tbname
}

func StructDoTableName(doType reflect.Type) string {
	if doType.Kind() == reflect.Pointer {
		doType = doType.Elem() //解引用
	}
	if doType.NumField() <= 0 {
		panic(nerror.NewRunTimeErrorFmt("%s没有字段", doType.Name()))
	}
	dbtbTag := doType.Field(0).Tag
	tbname := dbtbTag.Get(sqlext.NdbTags.TableName)
	if tbname == "" {
		panic(nerror.NewRunTimeErrorFmt("%s字段的Tag没有标识[%s]", doType.Name(), sqlext.NdbTags.TableName))
	}
	return tbname
}

func StructDoDbColList(doType reflect.Type, tableAlias string, excludeCols ...string) []string {
	if doType.Kind() == reflect.Pointer {
		doType = doType.Elem() //解引用
	}

	if doType.NumField() <= 0 {
		panic(nerror.NewRunTimeErrorFmt("%s没有字段", doType.Name()))
	}
	result := []string{}
	//字段
	for idx := range doType.NumField() {
		dbTag := doType.Field(idx).Tag
		dbcol := dbTag.Get(sqlext.NdbTags.TableColumn)
		if dbcol == "" {
			panic(nerror.NewRunTimeErrorFmt("%s字段的Tag没有标识[%s]", doType.Name(), sqlext.NdbTags.TableColumn))
		}

		if slices.Contains(excludeCols, dbcol) {
			continue
		}

		if tableAlias == "" {
			result = append(result, dbcol)
		} else {
			result = append(result, fmt.Sprintf("%s.%s", tableAlias, dbcol))
		}
	}
	return result
}

func StructDoDbColStr(doType reflect.Type, tableAlias string, excludeCols ...string) string {
	if doType.Kind() == reflect.Pointer {
		doType = doType.Elem() //解引用
	}
	sb := &strings.Builder{}
	cols := StructDoDbColList(doType, tableAlias)
	for _, v := range cols {
		if slices.Contains(excludeCols, v) {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(v)
	}
	return sb.String()
}
