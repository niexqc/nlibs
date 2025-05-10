package nmsql_test

import (
	"fmt"
	"testing"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var IDbWrapper *nmysql.NMysqlWrapper
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
  PRIMARY KEY (id) USING BTREE)  COMMENT='Test'`

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

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	IDbWrapper = ndb.NewNMysqlWrapper(dbconf)
	IDbWrapper.Exec("DROP TABLE IF EXISTS test01 ")
	IDbWrapper.Exec(ctTableSql)
}

func TestSelectOne(t *testing.T) {
	IDbWrapper.Exec("DELETE FROM test01")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if res, err := ndb.SelectOne[sqlext.NullString](IDbWrapper, "SELECT t03_varchar FROM test01 WHERE id=1"); nil != err {
		t.Error(err)
	} else {
		if res.NullString.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}

	if res, err := ndb.SelectOne[int64](IDbWrapper, "SELECT id FROM test01 WHERE id=1"); nil != err {
		t.Error(err)
	} else {
		if *res != 1 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectObj(t *testing.T) {
	IDbWrapper.Exec("DELETE FROM test01")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if obj, err := ndb.SelectObj[Test01Do](IDbWrapper, "SELECT * FROM test01 where id=1"); nil != err {
		println(err.Error())
	} else {
		if obj.Id != 1 || obj.T03Varchar.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectList(t *testing.T) {
	IDbWrapper.Exec("DELETE FROM test01")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if list, err := ndb.SelectList[sqlext.NullString](IDbWrapper, "SELECT t03_varchar FROM test01 ORDER BY id"); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].String != "aaa1" || list[1].String != "aaa2" {
			t.Error("返回值不匹配")
		}
	}

	if list, err := ndb.SelectList[Test01Do](IDbWrapper, "SELECT * FROM test01 ORDER BY id"); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].Id != 1 || list[1].Id != 2 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectDyObj(t *testing.T) {
	IDbWrapper.Exec("DELETE FROM test01")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.Insert("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1"); nil != err {
		println(err.Error())
	} else {
		val, err := sqlext.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
		if nil != err {
			panic(err)
		}
		fmt.Println(val.String)
	}

	if dyObjList, err := IDbWrapper.SelectDyObjList("SELECT * FROM test01 "); nil != err {
		println(err.Error())
	} else {
		val, err := sqlext.GetFiledVal[sqlext.NullString](dyObjList[1], "T03Varchar")
		if nil != err {
			panic(err)
		}
		fmt.Println(val.String)
	}
}
