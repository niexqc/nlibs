package nmysql

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
	TableSchema   string `db:"TABLE_SCHEMA"`
	TableName     string `db:"TABLE_NAME"`
	ColumnName    string `db:"COLUMN_NAME"`
	DataType      string `db:"DATA_TYPE"`
	ColumnComment string `db:"COLUMN_COMMENT"`
	AllowNull     string `db:"IS_NULLABLE"`
}

func (dbw *NMysqlWrapper) GetStructDoByTableStr(tableSchema, tableName string) (string, error) {
	tcSql := "SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.`TABLES` WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
	tableComment, _, _ := SelectOne[string](dbw, tcSql, tableSchema, tableName)

	sqlStr := `
	SELECT TABLE_SCHEMA ,TABLE_NAME , COLUMN_NAME , DATA_TYPE , COLUMN_COMMENT ,IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	dos := []columnSchemaDo{}
	dbw.SelectList(&dos, sqlStr, tableSchema, tableName)

	NsStr := &ntools.NString{S: tableName}

	resultStr := fmt.Sprintf("// %s %s.%s\n", *tableComment, tableSchema, tableName)
	resultStr += fmt.Sprintf("type %sDo struct {", NsStr.Under2Camel(true))

	for _, v := range dos {
		isNull := v.AllowNull == "YES"
		NsCStr := &ntools.NString{S: v.ColumnName}
		// String() 返回 sqlext.NullString
		// Name() 返回 NullString
		goTypeRef, err := mysqlTypeToGoType(v.DataType, isNull)
		if nil != err {
			return "", err
		}
		goType := goTypeRef.String()
		resultStr += fmt.Sprintf("\n  %s %s", NsCStr.Under2Camel(true), goType)
		resultStr += fmt.Sprintf(" `%s:\"%s\" %s:\"%s\" %s:\"%s\" json:\"%s\" zhdesc:\"%s\"`",
			sqlext.NdbTags.TableSchema, v.TableSchema,
			sqlext.NdbTags.TableName, v.TableName,
			sqlext.NdbTags.TableColumn, v.ColumnName,
			NsCStr.Under2Camel(false), v.ColumnComment)
	}
	resultStr += "\n}"
	return resultStr, nil
}

func mysqlTypeToGoType(mysqlType string, isNull bool) (reflect.Type, error) {
	goType, err := func(mtype string) (reflect.Type, error) {
		switch mtype {
		case "VARCHAR", "TEXT", "LONGTEXT":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullString{}), reflect.TypeOf("")), nil
		case "BIT":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullBool{}), reflect.TypeOf(true)), nil
		case "INT", "SMALLINT", "TINYINT":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullInt{}), reflect.TypeOf(int(1))), nil
		case "BIGINT":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullInt64{}), reflect.TypeOf(int64(1))), nil
		case "DATETIME", "DATE":
			return reflect.TypeOf(sqlext.NullTime{}), nil
		case "DOUBLE", "FLOAT":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullFloat64{}), reflect.TypeOf(float64(0.00))), nil
		case "DECIMAL":
			return ntools.If3(isNull, reflect.TypeOf(sqlext.NullDecimal{}), reflect.TypeOf(decimal.Decimal{})), nil
		default:
			return nil, nerror.NewRunTimeError(fmt.Sprintf("Mysql字段【%s】还没有做具体解析,需要对应处理", mtype))
		}
	}(strings.ToUpper(mysqlType))
	return goType, err
}
