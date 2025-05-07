package ndb_test

import (
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var IDbWrapper *ndb.NDbWrapper
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
	Id          int64           `db:"id" json:"id" zhdesc:""`
	T01Bigint   ndb.NullInt64   `db:"t01_bigint" json:"t01Bigint" zhdesc:""`
	T02Int      ndb.NullInt     `db:"t02_int" json:"t02Int" zhdesc:""`
	T03Varchar  ndb.NullString  `db:"t03_varchar" json:"t03Varchar" zhdesc:""`
	T04Text     ndb.NullString  `db:"t04_text" json:"t04Text" zhdesc:""`
	T05Longtext ndb.NullString  `db:"t05_longtext" json:"t05Longtext" zhdesc:""`
	T06Decimal  ndb.NullFloat64 `db:"t06_decimal" json:"t06Decimal" zhdesc:""`
	T07Float    ndb.NullFloat64 `db:"t07_float" json:"t07Float" zhdesc:""`
	T08Double   ndb.NullFloat64 `db:"t08_double" json:"t08Double" zhdesc:""`
	T09Datetime ndb.NullTime    `db:"t09_datetime" json:"t09Datetime" zhdesc:""`
	T10Bool     ndb.NullBool    `db:"t10_bool" json:"t10Bool" zhdesc:""`
}

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	dbconf := &nyaml.YamlConfDb{
		DbHost:           "8.137.54.220",
		DbPort:           3306,
		DbUser:           "root",
		DbPwd:            "Nxq@198943",
		DbName:           "niexq01",
		DbSqlLogPrint:    true,
		DbSqlLogLevel:    "debug",
		DbSqlLogCompress: false,
	}
	IDbWrapper = ndb.InitMysqlConnPool(dbconf)

}

func Test002(t *testing.T) {
	IDbWrapper.PrintSql(time.Now(), " WHERE name='nixq'")
	IDbWrapper.PrintSql(time.Now(), "? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=?", "niexq", 1)
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), time.Now())
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), time.Now(), true, false)
}

func TestIDbWrapper(t *testing.T) {
	IDbWrapper.Exec("DROP TABLE IF EXISTS test01 ")
	IDbWrapper.Exec(ctTableSql)
	IDbWrapper.PrintStructDoByTable("niexq01", "test01")

	IDbWrapper.Insert("INSERT into test01(id) VALUES(1)")
	IDbWrapper.Exec("INSERT into test01(t01_bigint) VALUES(1),(2)")

	if _, err := ndb.SelectOne[Test01Do](IDbWrapper, "SELECT * FROM test01"); nil != err {
		println(err.Error())
	}
	if d, _ := ndb.SelectOne[Test01Do](IDbWrapper, "SELECT * FROM test01 WHERE id=0"); nil != d {
		println("没有数据")
	}
	d, _ := ndb.SelectOne[Test01Do](IDbWrapper, "SELECT * FROM test01 WHERE id=1")
	println(njson.SonicObj2Str(d))
	// IDbWrapper.GenDoByTable("niexq01", "nba_user")
}

// 表名 `niexq01`.test01
