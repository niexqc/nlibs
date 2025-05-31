package nmsql_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/shopspring/decimal"
)

var tableName = "test01"
var schameName = "niexq01"

var delTableStr = fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
var crtTableStr = fmt.Sprintf(`CREATE TABLE %s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  t02_int int(11) NULL DEFAULT NULL,
  t03_varchar varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  t04_text text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  t05_longtext longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  t06_decimal decimal(64, 2) NULL DEFAULT NULL,
  t07_float float NULL DEFAULT NULL,
  t08_double double NULL DEFAULT NULL,
  t09_datetime datetime NULL DEFAULT NULL,
  t10_bool bit(1) NULL DEFAULT NULL,
  PRIMARY KEY (id)
)  ENGINE=InnoDB DEFAULT CHARSET=utf8mb4  COMMENT='Test'`, tableName)

var dbconf = &nyaml.YamlConfMysqlDb{
	DbHost: "8.137.54.220",
	DbPort: 3306,
	DbUser: "root",
	DbPwd:  "Nxq@198943",
	DbName: schameName,
}

var sqlPrintConf = &nyaml.YamlConfSqlPrint{
	DbSqlLogPrint:    true,
	DbSqlLogLevel:    "debug",
	DbSqlLogCompress: false,
}

var IDbWrapper *nmysql.NMysqlWrapper

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	IDbWrapper = nmysql.NewNMysqlWrapper(dbconf, sqlPrintConf)

}

func TestGenStruct(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	str := IDbWrapper.GetStructDoByTableStr(schameName, tableName)
	if !strings.Contains(str, "T04Text sqlext.NullString") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T04Text sqlext.NullString")
	}
	if !strings.Contains(str, "T06Decimal decimal.NullDecimal") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T06Decimal decimal.NullDecimal")
	}
	if !strings.Contains(str, "T08Double sqlext.NullFloat64") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T08Double sqlext.NullFloat64")
	}
	if !strings.Contains(str, "Test `niexq01`.test01") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "Test `niexq01`.test01")
	}
	t.Log("TestGenStruct 执行成功")
}

func TestInsert(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	IDbWrapper.InsertWithLastId("INSERT into test01(t03_varchar) VALUES('aaa2')")
	lasetId, _ := IDbWrapper.InsertWithLastId("INSERT into test01(t03_varchar) VALUES('aaa2')")
	if lasetId != 2 {
		t.Error("InsertWithLastId 应该返回2")
	}
	rowEff, _ := IDbWrapper.InsertWithRowsAffected("INSERT into test01(t03_varchar) VALUES('aaa1'),('aaa2'),('aaa2'),('aaa2')")
	if rowEff != 4 {
		t.Error("InsertWithRowsAffected应该返回4")
	}
}

func TestSelectOne(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if res, _, err := nmysql.SelectOne[sqlext.NullString](IDbWrapper, "SELECT t03_varchar FROM test01 WHERE id=1"); nil != err {
		t.Error(err)
	} else {
		if res.NullString.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}

	if res, _, err := nmysql.SelectOne[int64](IDbWrapper, "SELECT id FROM test01 WHERE id=1"); nil != err {
		t.Error(err)
	} else {
		if *res != 1 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectObj(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	type Test01Do struct {
		Id          int64               `db:"id" json:"id" zhdesc:""`
		T01Bigint   sqlext.NullInt64    `db:"t01_bigint" json:"t01Bigint" zhdesc:""`
		T02Int      sqlext.NullInt      `db:"t02_int" json:"t02Int" zhdesc:""`
		T03Varchar  sqlext.NullString   `db:"t03_varchar" json:"t03Varchar" zhdesc:""`
		T04Text     sqlext.NullString   `db:"t04_text" json:"t04Text" zhdesc:""`
		T05Longtext sqlext.NullString   `db:"t05_longtext" json:"t05Longtext" zhdesc:""`
		T06Decimal  decimal.NullDecimal `db:"t06_decimal" json:"t06Decimal" zhdesc:""`
		T07Float    sqlext.NullFloat64  `db:"t07_float" json:"t07Float" zhdesc:""`
		T08Double   sqlext.NullFloat64  `db:"t08_double" json:"t08Double" zhdesc:""`
		T09Datetime sqlext.NullTime     `db:"t09_datetime" json:"t09Datetime" zhdesc:""`
		T10Bool     sqlext.NullBool     `db:"t10_bool" json:"t10Bool" zhdesc:""`
	}

	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if obj, _, err := nmysql.SelectObj[Test01Do](IDbWrapper, "SELECT * FROM test01 where id=1"); nil != err {
		println(err.Error())
	} else {
		if obj.Id != 1 || obj.T03Varchar.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectList(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	type Test01Do struct {
		Id          int64               `db:"id" json:"id" zhdesc:""`
		T01Bigint   sqlext.NullInt64    `db:"t01_bigint" json:"t01Bigint" zhdesc:""`
		T02Int      sqlext.NullInt      `db:"t02_int" json:"t02Int" zhdesc:""`
		T03Varchar  sqlext.NullString   `db:"t03_varchar" json:"t03Varchar" zhdesc:""`
		T04Text     sqlext.NullString   `db:"t04_text" json:"t04Text" zhdesc:""`
		T05Longtext sqlext.NullString   `db:"t05_longtext" json:"t05Longtext" zhdesc:""`
		T06Decimal  decimal.NullDecimal `db:"t06_decimal" json:"t06Decimal" zhdesc:""`
		T07Float    sqlext.NullFloat64  `db:"t07_float" json:"t07Float" zhdesc:""`
		T08Double   sqlext.NullFloat64  `db:"t08_double" json:"t08Double" zhdesc:""`
		T09Datetime sqlext.NullTime     `db:"t09_datetime" json:"t09Datetime" zhdesc:""`
		T10Bool     sqlext.NullBool     `db:"t10_bool" json:"t10Bool" zhdesc:""`
	}

	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if list, err := nmysql.SelectList[sqlext.NullString](IDbWrapper, "SELECT t03_varchar FROM test01 ORDER BY id"); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].String != "aaa1" || list[1].String != "aaa2" {
			t.Error("返回值不匹配")
		}
	}

	if list, err := nmysql.SelectList[Test01Do](IDbWrapper, "SELECT * FROM test01 ORDER BY id"); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].Id != 1 || list[1].Id != 2 {
			t.Error("返回值不匹配")
		}
	}

}

func TestSelectDyObj(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")

	if dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1"); nil != err {
		println(err.Error())
	} else {
		val, err := nmysql.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
		if nil != err {
			panic(err)
		}
		fmt.Println(val.String)
	}

	if dyObjList, err := IDbWrapper.SelectDyObjList("SELECT * FROM test01 "); nil != err {
		println(err.Error())
	} else {
		val, err := nmysql.GetFiledVal[sqlext.NullString](dyObjList[1], "T03Varchar")
		if nil != err {
			panic(err)
		}
		fmt.Println(val.String)
	}
}

func TestSqlInNotExist(t *testing.T) {

	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(1,'aaa1')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(2,'aaa2')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(3,'aaa3')")
	IDbWrapper.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(4,'aaa4')")

	ids := []int64{1, 2, 6, 7}

	sqlStr, allArgs, _ := sqlext.SqlFmtSqlInNotExist("test01", "id", ids)
	notExistIdds := []int64{}

	err := IDbWrapper.SelectList(&notExistIdds, sqlStr, nlibs.Arr2ArrAny(allArgs)...)

	if nil != err {
		panic(err)
	}

	exResult := "[6,7]"
	acResult := njson.SonicObj2Str(notExistIdds)
	if acResult != exResult {
		panic(nerror.NewRunTimeErrorFmt("查询结果不匹配，期望:%s,实际:%s", exResult, acResult))
	}

}

func TestTx(t *testing.T) {
	IDbWrapper.Exec(delTableStr)
	IDbWrapper.Exec(crtTableStr)

	time.Sleep(time.Second)
	ntools.SlogSetTraceId("1111")
	// time.Sleep(6 * time.Second)
	txr, err := IDbWrapper.NdbTxBgn(3)
	if nil != err {
		panic(err)
	}
	defer txr.NdbTxCommit()

	r, err := txr.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(5,'aaa1')")
	if err != nil {
		panic(err)
	}

	slog.Info("TestTx", "lastInsertId", r)

	r, err = txr.InsertWithLastId("INSERT into test01(id,t03_varchar) VALUES(6,'aaa2')")
	if nil != err {
		// panic(err)
	}
	fmt.Sprintln(r, err)

}
