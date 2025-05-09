package nmysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
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
	// tcSql := "SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.`TABLES` WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"

	// tableComment, _ := SelectOne[string](dbw, tcSql, tableSchema, tableName)
	tableComment := ""
	sqlStr := `
	SELECT TABLE_NAME , COLUMN_NAME , DATA_TYPE , COLUMN_COMMENT ,IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	dos := []columnSchemaDo{}
	dbw.SelectList(&dos, sqlStr, tableSchema, tableName)

	NsStr := &ntools.NString{S: tableName}
	resultStr := fmt.Sprintf("//%s `%s`.%s\n", tableComment, tableSchema, tableName)
	resultStr += fmt.Sprintf("type %sDo struct {", NsStr.UnderscoreToCamelcase(true))

	clmSql := ""
	for _, v := range dos {
		isNull := v.IsNullable == "YES"
		NsCStr := &ntools.NString{S: v.ColumnName}
		// String() 返回 sqlext.NullString
		// Name() 返回 NullString
		goType := mysqlTypeToGoType(v.DataType, isNull).String()
		resultStr += fmt.Sprintf("\n  %s %s", NsCStr.UnderscoreToCamelcase(true), goType)
		resultStr += fmt.Sprintf(" `db:\"%s\" json:\"%s\" zhdesc:\"%s\"`", v.ColumnName, NsCStr.UnderscoreToCamelcase(false), v.ColumnComment)
		clmSql += (ntools.If3(len(clmSql) > 0, ",", "") + v.ColumnName)
	}
	resultStr += "\n}"

	resultStr += fmt.Sprintf("var %sDoClmStr=\"%s\"", NsStr.UnderscoreToCamelcase(true), clmSql)

	println(resultStr)

}

func mysqlTypeToGoType(mysqlType string, isNull bool) reflect.Type {
	mysqlType = strings.ToUpper(mysqlType)
	switch mysqlType {
	case "VARCHAR", "TEXT", "LONGTEXT":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullString{}), reflect.TypeOf(""))
	case "BIT":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullBool{}), reflect.TypeOf(true))
	case "INT":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullInt{}), reflect.TypeOf(int(1)))
	case "BIGINT":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullInt64{}), reflect.TypeOf(int64(1)))
	case "DATETIME":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullTime{}), reflect.TypeOf(time.Now()))
	case "DOUBLE", "FLOAT", "DECIMAL":
		return ntools.If3(isNull, reflect.TypeOf(sqlext.NullFloat64{}), reflect.TypeOf(float64(0.00)))
	default:
		panic(nerror.NewRunTimeError(fmt.Sprintf("Mysql字段【%s】还没有做具体解析,需要对应处理", mysqlType)))
	}
}

func createDyStruct(cols []*sql.ColumnType) reflect.Type {
	fields := []reflect.StructField{}
	mysqlType2GoType := func(col *sql.ColumnType) reflect.Type {
		nullable, ok := col.Nullable()
		if !ok {
			nullable = false
		}
		return mysqlTypeToGoType(col.DatabaseTypeName(), nullable)
	}
	for _, v := range cols {
		fName := v.Name()
		field := reflect.StructField{
			Name: (&ntools.NString{S: fName}).UnderscoreToCamelcase(true),
			Type: mysqlType2GoType(v),
			Tag:  reflect.StructTag(fmt.Sprintf(`db:"%s" json:"%s"`, fName, fName)),
		}
		fields = append(fields, field)
	}
	// 创建动态结构体类型
	return reflect.StructOf(fields)
}
