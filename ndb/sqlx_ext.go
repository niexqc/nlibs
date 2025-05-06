package ndb

import (
	"fmt"
	"log/slog"
	"strings"

	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

func InitMysqlConnPool(conf *nyaml.YamlConfDb) *NDbWrapper {
	//开始连接数据库
	mysqlUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.DbUser, conf.DbPwd, conf.DbHost, conf.DbPort, conf.DbName)
	mysqlUrl = mysqlUrl + "?loc=Local&parseTime=true&charset=utf8mb4"
	slog.Debug(mysqlUrl)
	db, err := sqlx.Open("mysql", mysqlUrl)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	return &NDbWrapper{sqlxDb: db, conf: conf}
}

type ColumnSchemaDo struct {
	TableName     string `db:"TABLE_NAME"`
	ColumnName    string `db:"COLUMN_NAME"`
	DataType      string `db:"DATA_TYPE"`
	ColumnComment string `db:"COLUMN_COMMENT"`
	IsNullable    string `db:"IS_NULLABLE"`
}

func (dbw *NDbWrapper) GenDoByTable(tableSchema, tableName string) {
	sqlStr := `
	SELECT TABLE_NAME , COLUMN_NAME , DATA_TYPE , COLUMN_COMMENT ,IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	dos := []ColumnSchemaDo{}
	dbw.SelectList(&dos, sqlStr, tableSchema, tableName)

	NsStr := &ntools.NString{S: tableName}
	resultStr := fmt.Sprintf("// 表名 `%s`.%s", tableSchema, tableName)
	resultStr += fmt.Sprintf("\ntype %sDo struct {", NsStr.UnderscoreToCamelcase(true))
	for _, v := range dos {
		isNull := v.IsNullable == "YES"
		NsCStr := &ntools.NString{S: v.ColumnName}
		resultStr += fmt.Sprintf("\n  %s %s", NsCStr.UnderscoreToCamelcase(true), mysqlTypeToGoType(v.DataType, isNull))
		resultStr += fmt.Sprintf(" `db:\"%s\" json:\"%s\" zhdesc:\"%s\"`", v.ColumnName, NsCStr.UnderscoreToCamelcase(false), v.ColumnComment)
	}
	resultStr += "\n}"

	slog.Info("\n" + resultStr)
}

// 示例：将 VARCHAR 转为 string
func mysqlTypeToGoType(mysqlType string, isNull bool) string {
	mysqlType = strings.ToUpper(mysqlType)
	switch mysqlType {
	case "VARCHAR", "TEXT", "LONGTEXT":
		return ntools.If3(isNull, "ndb.NullString", "string")
	case "BIT":
		return ntools.If3(isNull, "ndb.NullBool", "bool")
	case "INT":
		return ntools.If3(isNull, "ndb.NullInt", "int")
	case "BIGINT":
		return ntools.If3(isNull, "ndb.NullInt64", "int64")
	case "DATETIME":
		return ntools.If3(isNull, "ndb.NullTime", "time.Time")
	case "DOUBLE", "FLOAT", "DECIMAL":
		return ntools.If3(isNull, "ndb.NullFloat64", "float64")
	default:
		return "interface{}"
	}
}

// 表名 nab_user
// type NbaUserDo struct {
// 	UserAcc string `db:"user_acc" json:"userAcc" zhdesc:"用户账号" binding:"required" `
// }
