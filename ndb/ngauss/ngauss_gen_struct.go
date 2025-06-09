package ngauss

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/shopspring/decimal"
)

type columnSchemaDo struct {
	TableSchema   string            `db:"table_schema"`
	TableName     string            `db:"table_name"`
	ColumnName    string            `db:"column_name"`
	ColumnComment sqlext.NullString `db:"column_comment"`
	ColumnOrdinal int               `db:"column_ordinal"`
	UdtName       string            `db:"udt_name"`
	AllowNull     string            `db:"allow_null"`
	VarcharMaxLen sqlext.NullInt    `db:"varchar_max_len"`
}

func (dbw *NGaussWrapper) GetStructDoByTableStr(tableSchema, tableName string) (string, error) {
	tcSql := fmt.Sprintf("SELECT obj_description('%s.%s'::regclass) tableComment", tableSchema, tableName)
	tableComment := ""
	findOk, err := dbw.SelectOne(&tableComment, tcSql)
	if nil != err {
		return "", nerror.NewRunTimeErrorFmt("查询表[%s.%s]注释异常:%v", tableSchema, tableName, err)
	}
	if !findOk {
		return "", nerror.NewRunTimeErrorFmt("未获取到表[%s.%s]注释", tableSchema, tableName)
	}

	colStr := `
		SELECT isc.table_schema,isc.table_name,isc.column_name,pg_catalog.col_description(c.oid, isc.ordinal_position) column_comment
		,isc.ordinal_position column_ordinal,isc.udt_name,isc.is_nullable allow_null,isc.character_maximum_length varchar_max_len
		FROM information_schema.columns isc
		LEFT JOIN pg_catalog.pg_class c ON c.relname = isc.table_name::text
			AND c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = isc.table_schema::text)
		WHERE isc.table_schema=? AND isc.table_name=?
 `
	dos := []columnSchemaDo{}
	err = dbw.SelectList(&dos, colStr, tableSchema, tableName)
	if nil != err {
		return "", nerror.NewRunTimeErrorFmt("查询表[%s.%s]字段异常:%v", tableSchema, tableName, err)
	}

	NsStr := &ntools.NString{S: tableName}
	resultStr := fmt.Sprintf("// %s %s.%s\n", tableComment, tableSchema, tableName)
	resultStr += fmt.Sprintf("type %sDo struct {", NsStr.Under2Camel(true))

	for _, v := range dos {
		isNull := v.AllowNull == "YES"
		NsCStr := &ntools.NString{S: v.ColumnName}
		// String() 返回 sqlext.NullString
		// Name() 返回 NullString
		goTypeRef, err := gaussDbUdtNameToGoType(v.UdtName, isNull)
		if nil != err {
			return "", err
		}
		goType := goTypeRef.String()
		resultStr += fmt.Sprintf("\n  %s %s", NsCStr.Under2Camel(true), goType)
		// resultStr += fmt.Sprintf(" `schm:\"%s\" tbn:\"%s\" db:\"%s\" json:\"%s\" zhdesc:\"%s\"`", v.TableSchema, v.TableName, v.ColumnName, NsCStr.Under2Camel(false), v.ColumnComment.String)
		resultStr += fmt.Sprintf(" `%s:\"%s\" %s:\"%s\" %s:\"%s\" json:\"%s\" zhdesc:\"%s\"`",
			sqlext.NdbTags.TableSchema, v.TableSchema,
			sqlext.NdbTags.TableName, v.TableName,
			sqlext.NdbTags.TableColumn, v.ColumnName,
			NsCStr.Under2Camel(false), v.ColumnComment.String)
	}
	resultStr += "\n}"
	return resultStr, nil
}

func gaussDbUdtNameToGoType(gaussUdtName string, allowNull bool) (reflect.Type, error) {
	goType, err := func(gaussUdtName string) (reflect.Type, error) {
		switch gaussUdtName {
		case "bool":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullBool{}), reflect.TypeOf(true)), nil
		case "varchar", "text":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullString{}), reflect.TypeOf("")), nil
		case "int1", "int2", "int4":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullInt{}), reflect.TypeOf(int(1))), nil
		case "int8":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullInt64{}), reflect.TypeOf(int64(1))), nil
		case "date", "time", "timestamp", "timestamptz":
			return reflect.TypeOf(sqlext.NullTime{}), nil
		case "float4", "float8":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullFloat64{}), reflect.TypeOf(float64(0.00))), nil
		case "numeric":
			return ntools.If3(allowNull, reflect.TypeOf(sqlext.NullDecimal{}), reflect.TypeOf(decimal.Decimal{})), nil
		default:
			return nil, nerror.NewRunTimeError(fmt.Sprintf("GaussDb字段【%s】还没有做具体解析,需要对应处理", gaussUdtName))
		}
	}(strings.ToLower(gaussUdtName))
	return goType, err
}
