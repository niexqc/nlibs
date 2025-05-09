package ndb_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var ctTableSql = `CREATE TABLE IF NOT EXISTS test01  (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  t01_bigint bigint(20) NULL DEFAULT NULL,
  t02_int int(11) NULL DEFAULT NULL,
  t03_varchar varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  t04_text text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  t05_longtext longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  t06_decimal decimal(64, 2) NULL DEFAULT NULL,
  t07_float float NULL DEFAULT NULL,
  t08_double double NULL DEFAULT NULL,
  t09_datetime datetime NULL DEFAULT NULL,
  t10_bool bit(1) NULL DEFAULT NULL,
  PRIMARY KEY (id) USING BTREE)`

type Test01Do struct {
	Id          int64              `db:"id" json:"id" zhdesc:""`
	T01Bigint   sqlext.NullInt64   `db:"t01_bigint" json:"t01Bigint" zhdesc:""`
	T02Int      sqlext.NullInt     `db:"t02_int" json:"t02Int" zhdesc:""`
	T03Varchar  sqlext.NullString  `db:"t03_varchar" json:"t03Varchar" zhdesc:""`
	T04Text     sqlext.NullString  `db:"t04_text" json:"t04Text" zhdesc:""`
	T05Longtext sqlext.NullString  `db:"t05_longtext" json:"t05Longtext" zhdesc:""`
	T06Decimal  sqlext.NullFloat64 `db:"t06_decimal" json:"t06Decimal" zhdesc:""`
	T07Float    sqlext.NullFloat64 `db:"t07_float" json:"t07Float" zhdesc:""`
	T08Double   sqlext.NullFloat64 `db:"t08_double" json:"t08Double" zhdesc:""`
	T09Datetime sqlext.NullTime    `db:"t09_datetime" json:"t09Datetime" zhdesc:""`
	T10Bool     sqlext.NullBool    `db:"t10_bool" json:"t10Bool" zhdesc:""`
}

var ndbWrapper ndb.INdbWrapper
var dbconf = &nyaml.YamlConfDb{
	DbHost:           "8.137.54.220",
	DbPort:           3306,
	DbUser:           "root",
	DbPwd:            "Nxq@198943",
	DbName:           "niexq01",
	DbSqlLogPrint:    true,
	DbSqlLogLevel:    "debug",
	DbSqlLogCompress: false,
}

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	ndbWrapper = ndb.NewNMysqlWrapper(dbconf)
	ndbWrapper.Exec("DROP TABLE IF EXISTS test01 ")
	ndbWrapper.Exec(ctTableSql)

	ndbWrapper.Insert("INSERT into test01(id) VALUES(1)")
	ndbWrapper.Exec("INSERT into test01(t01_bigint) VALUES(1),(2)")
	ndbWrapper.Insert("INSERT into test01(t09_datetime) VALUES(?)", time.Now())
}

func Test001(t *testing.T) {

	if vo, err := ndb.SelectObj[Test01Do](ndbWrapper, "SELECT * FROM test01 WHERE id=4"); nil == err {
		slog.Info(njson.SonicObj2Str(vo))
	} else {
		t.Error(err)
	}

	if vos, err := ndb.SelectList[Test01Do](ndbWrapper, "SELECT * FROM test01"); nil == err {
		slog.Info(njson.SonicObj2Str(vos))
	} else {
		t.Error(err)
	}

}
