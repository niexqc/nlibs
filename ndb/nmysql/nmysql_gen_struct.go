package nmysql

import (
	"fmt"
	"strings"

	"github.com/niexqc/nlibs/ntools"
)

type columnSchemaDo struct {
	TableName     string `db:"TABLE_NAME"`
	ColumnName    string `db:"COLUMN_NAME"`
	DataType      string `db:"DATA_TYPE"`
	ColumnComment string `db:"COLUMN_COMMENT"`
	IsNullable    string `db:"IS_NULLABLE"`
}

func (dbw *NMysqlWrapper) PrintStructDoByTable(tableSchema, tableName string) {
	sqlStr := `
	SELECT TABLE_NAME , COLUMN_NAME , DATA_TYPE , COLUMN_COMMENT ,IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	dos := []columnSchemaDo{}
	dbw.SelectList(&dos, sqlStr, tableSchema, tableName)

	NsStr := &ntools.NString{S: tableName}
	resultStr := fmt.Sprintf("// 表名 `%s`.%s", tableSchema, tableName)
	resultStr += fmt.Sprintf("type %sDo struct {", NsStr.UnderscoreToCamelcase(true))

	clmSql := ""
	for _, v := range dos {
		isNull := v.IsNullable == "YES"
		NsCStr := &ntools.NString{S: v.ColumnName}
		resultStr += fmt.Sprintf("\n  %s %s", NsCStr.UnderscoreToCamelcase(true), mysqlTypeToGoType(v.DataType, isNull))
		resultStr += fmt.Sprintf(" `db:\"%s\" json:\"%s\" zhdesc:\"%s\"`", v.ColumnName, NsCStr.UnderscoreToCamelcase(false), v.ColumnComment)
		clmSql += (ntools.If3(len(clmSql) > 0, ",", "") + v.ColumnName)
	}
	resultStr += "\n}\n"
	resultStr += fmt.Sprintf("var %sDoClmStr=\"%s\"", NsStr.UnderscoreToCamelcase(true), clmSql)

	println(resultStr)

}

func mysqlTypeToGoType(mysqlType string, isNull bool) string {
	mysqlType = strings.ToUpper(mysqlType)
	switch mysqlType {
	case "VARCHAR", "TEXT", "LONGTEXT":
		return ntools.If3(isNull, "sqlext.NullString", "string")
	case "BIT":
		return ntools.If3(isNull, "sqlext.NullBool", "bool")
	case "INT":
		return ntools.If3(isNull, "sqlext.NullInt", "int")
	case "BIGINT":
		return ntools.If3(isNull, "sqlext.NullInt64", "int64")
	case "DATETIME":
		return ntools.If3(isNull, "sqlext.NullTime", "time.Time")
	case "DOUBLE", "FLOAT", "DECIMAL":
		return ntools.If3(isNull, "sqlext.NullFloat64", "float64")
	default:
		return "interface{}"
	}
}

// 表名 nab_user
// type NbaUserDo struct {
// 	UserAcc string `db:"user_acc" json:"userAcc" zhdesc:"用户账号" binding:"required" `
// }
