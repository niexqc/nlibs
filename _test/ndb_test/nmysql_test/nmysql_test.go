package nmysql_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var tableName = "tb01"
var schameName = "ndb_test"
var mysqlCreateTableStr = ""

var mysqlConf *nyaml.YamlConfMysqlDb
var sqlPrintConf *nyaml.YamlConfSqlPrint

func init() {
	ntools.SlogConf("test", "debug", 1, 2)

	mysqlConf = &nyaml.YamlConfMysqlDb{
		DbHost: "8.137.54.220",
		DbPort: 3306,
		DbUser: "root",
		DbPwd:  "Nxq@198943",
		DbName: "niexq01",
	}

	sqlPrintConf = &nyaml.YamlConfSqlPrint{
		DbSqlLogPrint:    true,
		DbSqlLogLevel:    "debug",
		DbSqlLogCompress: false,
	}

	// 采取文本替换的形式
	mysqlCreateTableSrcStr := `CREATE TABLE tb01 (
  id bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
  t02_int int(11) DEFAULT NULL COMMENT 'NullInt',
  t03_varchar varchar(255) DEFAULT NULL COMMENT 'NullVarchar',
  t04_text text COMMENT 'NullText',
  t05_longtext longtext COMMENT 'NullLongText',
  t06_decimal decimal(64,2) DEFAULT NULL COMMENT 'NullDecimal',
  t07_float float DEFAULT NULL COMMENT 'NullFloat',
  t08_double double DEFAULT NULL COMMENT 'NullDouble',
  t09_datetime datetime DEFAULT NULL COMMENT 'NullDateTime',
  t10_bool bit(1) DEFAULT NULL COMMENT 'NullBool',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='测试表'`
	mysqlCreateTableStr = strings.ReplaceAll(mysqlCreateTableSrcStr, `tb01`, fmt.Sprintf(`%s.%s`, schameName, tableName))

}

func TestCrateTable(t *testing.T) {
	dbWrapper := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	_, err := dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	ntools.TestErrPainic(t, "TestCrateTable DROP TABLE ", err)

	_, err = dbWrapper.Exec(mysqlCreateTableStr)
	ntools.TestErrPainic(t, "TestCrateTable CREATE TABLE", err)

	tcSql := "SELECT TABLE_COMMENT FROM INFORMATION_SCHEMA.`TABLES` WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
	comment, findOk, err := nmysql.SelectOne[string](dbWrapper, tcSql, schameName, tableName)
	ntools.TestErrPainic(t, "TestCrateTable SELECT tableComment ", err)

	if !findOk {
		ntools.TestErrPanicMsg(t, "TestCrateTable SELECT tableComment 未获取到注释 ")
	}
	ntools.TestEq(t, "TestCrateTable SELECT tableComment ", "测试表", *comment)

}
