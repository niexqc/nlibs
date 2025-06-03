package nmysql_test

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
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
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

func TestGenStruct(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	str, err := dbWrapper.GetStructDoByTableStr(schameName, tableName)
	ntools.TestErrPainic(t, "TestGenStruct ", err)
	slog.Info(str)
	if !strings.Contains(str, "T04Text sqlext.NullString") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T04Text sqlext.NullString")
	}
	if !strings.Contains(str, "T06Decimal decimal.NullDecimal") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T06Decimal decimal.NullDecimal")
	}
	if !strings.Contains(str, "T08Double sqlext.NullFloat64") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "T08Double sqlext.NullFloat64")
	}
	if !strings.Contains(str, "测试表 ndb_test.tb01") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "Test `niexq01`.test01")
	}
	t.Log("TestGenStruct 执行成功")
}

func TestInsert(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa1')", schameName, tableName))
	lasetId, _ := dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa2')", schameName, tableName))
	if lasetId != 2 {
		t.Error("InsertWithLastId 应该返回2")
	}
	rowEff, _ := dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into  %s.%s(t03_varchar) VALUES('aaa3'),('aaa4'),('aa5'),('aaa6')", schameName, tableName))
	if rowEff != 4 {
		t.Error("InsertWithRowsAffected应该返回4")
	}
}

func TestSelectOne(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa1')", schameName, tableName))

	querySql := fmt.Sprintf("SELECT t03_varchar FROM %s.%s WHERE id=1", schameName, tableName)
	if res, _, err := nmysql.SelectOne[sqlext.NullString](dbWrapper, querySql); nil != err {
		t.Error(err)
	} else {
		if res.NullString.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
	querySql = fmt.Sprintf("SELECT id FROM %s.%s WHERE id=1", schameName, tableName)
	if res, _, err := nmysql.SelectOne[int64](dbWrapper, querySql); nil != err {
		t.Error(err)
	} else {
		if *res != 1 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectObj(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa1')", schameName, tableName))

	// 测试表 ndb_test.tb01
	type Tb01Do struct {
		Id          int64               `schm:"ndb_test" tbn:"tb01" db:"id" json:"id" zhdesc:"主键"`
		T02Int      sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"t02_int" json:"t02Int" zhdesc:"NullInt"`
		T03Varchar  sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t03_varchar" json:"t03Varchar" zhdesc:"NullVarchar"`
		T04Text     sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t04_text" json:"t04Text" zhdesc:"NullText"`
		T05Longtext sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t05_longtext" json:"t05Longtext" zhdesc:"NullLongText"`
		T06Decimal  decimal.NullDecimal `schm:"ndb_test" tbn:"tb01" db:"t06_decimal" json:"t06Decimal" zhdesc:"NullDecimal"`
		T07Float    sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"t07_float" json:"t07Float" zhdesc:"NullFloat"`
		T08Double   sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"t08_double" json:"t08Double" zhdesc:"NullDouble"`
		T09Datetime sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"t09_datetime" json:"t09Datetime" zhdesc:"NullDateTime"`
		T10Bool     sqlext.NullBool     `schm:"ndb_test" tbn:"tb01" db:"t10_bool" json:"t10Bool" zhdesc:"NullBool"`
	}

	querySql := fmt.Sprintf("SELECT * FROM %s.%s WHERE id=1", schameName, tableName)

	if obj, _, err := nmysql.SelectObj[Tb01Do](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if obj.Id != 1 || obj.T03Varchar.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectList(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(t03_varchar) VALUES('aaa2')", schameName, tableName))

	// 测试表 ndb_test.tb01
	type Tb01Do struct {
		Id          int64               `schm:"ndb_test" tbn:"tb01" db:"id" json:"id" zhdesc:"主键"`
		T02Int      sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"t02_int" json:"t02Int" zhdesc:"NullInt"`
		T03Varchar  sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t03_varchar" json:"t03Varchar" zhdesc:"NullVarchar"`
		T04Text     sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t04_text" json:"t04Text" zhdesc:"NullText"`
		T05Longtext sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"t05_longtext" json:"t05Longtext" zhdesc:"NullLongText"`
		T06Decimal  decimal.NullDecimal `schm:"ndb_test" tbn:"tb01" db:"t06_decimal" json:"t06Decimal" zhdesc:"NullDecimal"`
		T07Float    sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"t07_float" json:"t07Float" zhdesc:"NullFloat"`
		T08Double   sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"t08_double" json:"t08Double" zhdesc:"NullDouble"`
		T09Datetime sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"t09_datetime" json:"t09Datetime" zhdesc:"NullDateTime"`
		T10Bool     sqlext.NullBool     `schm:"ndb_test" tbn:"tb01" db:"t10_bool" json:"t10Bool" zhdesc:"NullBool"`
	}

	querySql := fmt.Sprintf("SELECT t03_varchar FROM %s.%s ORDER BY id ASC", schameName, tableName)

	if list, err := nmysql.SelectList[sqlext.NullString](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].String != "aaa1" || list[1].String != "aaa2" {
			t.Error("返回值不匹配")
		}
	}

	querySql = fmt.Sprintf("SELECT * FROM %s.%s ORDER BY id ASC", schameName, tableName)
	if list, err := nmysql.SelectList[Tb01Do](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].Id != 1 || list[1].Id != 2 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSqlInNotExist(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(1,'aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(2,'aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(3,'aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(4,'aaa1')", schameName, tableName))

	ids := []int64{1, 2, 6, 7}
	sqlStr, allArgs, err := sqlext.SqlFmtSqlInNotExist(fmt.Sprintf("%s.%s", schameName, tableName), "id", ids)
	ntools.TestErrPainic(t, "TestSqlInNotExist ", err)

	notExistIdds, err := nmysql.SelectList[int64](dbWrapper, sqlStr, nlibs.Arr2ArrAny(allArgs)...)
	ntools.TestErrPainic(t, "TestSqlInNotExist ", err)

	acResult := njson.Obj2StrWithPanicError(notExistIdds)
	ntools.TestEq(t, "TestSqlInNotExist ", "[6,7]", acResult)

}

func TestSelectDyObjAndList(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(1,'aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(2,'aaa2')", schameName, tableName))

	dyObj, err := dbWrapper.SelectDyObj(fmt.Sprintf("SELECT * FROM %s.%s where id=1", schameName, tableName))
	ntools.TestErrPainic(t, "TestSelectDyObj ", err)

	val, err := nmysql.GetFiledVal[sqlext.NullString](dyObj, dyObj.DbNameFiledsMap["t03_varchar"].StructFieldName)
	ntools.TestErrPainic(t, "TestSelectDyObj ", err)
	ntools.TestEq(t, "TestSelectDyObj ", "aaa1", val.String)

	// 列表
	dyObjList, err := dbWrapper.SelectDyObjList(fmt.Sprintf("SELECT * FROM %s.%s ", schameName, tableName))
	ntools.TestErrPainic(t, "TestSelectDyList ", err)
	ntools.TestEq(t, "TestSelectDyList 列表数", 2, len(dyObjList))

	jsonStr, err := nmysql.DyObjList2Json(dyObjList)
	ntools.TestErrPainic(t, "TestSelectDyList 转Json失败 ", err)

	expJson := `[{"id":1,"t02Int":null,"t03Varchar":"aaa1","t04Text":null,"t05Longtext":null,"t06Decimal":null,"t07Float":null,"t08Double":null,"t09Datetime":null,"t10Bool":null},{"id":2,"t02Int":null,"t03Varchar":"aaa2","t04Text":null,"t05Longtext":null,"t06Decimal":null,"t07Float":null,"t08Double":null,"t09Datetime":null,"t10Bool":null}]`
	ntools.TestEq(t, "TestSelectDyList ", expJson, jsonStr)

}

func TestNdbTx(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(1,'aaa1')", schameName, tableName))
	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(2,'aaa2')", schameName, tableName))

	ntools.SlogSetTraceId("TestNdbTx")

	// time.Sleep(6 * time.Second)
	txr, err := dbWrapper.NdbTxBgn(3)
	ntools.TestErrPainic(t, "TestNdbTx", err)
	defer txr.NdbTxCommit(recover())

	_, err = txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(3,'aaa1')", schameName, tableName))
	ntools.TestErrPainic(t, "TestNdbTx", err)

	_, err = txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(4,'aaa1')", schameName, tableName))
	ntools.TestErrPainic(t, "TestNdbTx", err)

}

func TestNdbTxTimeOut(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(1,'aaa1')", schameName, tableName))

	ntools.SlogSetTraceId("TestNdbTxTimeOut")

	txr, err := dbWrapper.NdbTxBgn(1)
	ntools.TestErrPainic(t, "TestNdbTxTimeOut", err)
	defer func() {
		err := txr.NdbTxCommit(recover())
		ntools.TestErrNotNil(t, "此时应该捕获到事务超时", err)
		ntools.TestStrContains(t, "此时应该捕获到事务超时", "transaction has already been committed or rolled back", err.Error())
		//SQL验证数据未被写入
		count, _, _ := nmysql.SelectOne[int64](dbWrapper, fmt.Sprintf("SELECT COUNT(id) FROM  %s.%s ", schameName, tableName))
		ntools.TestEq(t, "此时应该捕获到事务超时-数据未被写入", int64(1), *count)
	}()

	txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(2,'aaa2')", schameName, tableName))
	time.Sleep(1200 * time.Millisecond)
	txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(3,'aaa3')", schameName, tableName))

}

func TestNdbTxErrRollback(t *testing.T) {
	dbWrapper, _ := nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(mysqlCreateTableStr)

	dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(1,'aaa1')", schameName, tableName))

	ntools.SlogSetTraceId("TestNdbTxErrRollback")

	txr, err := dbWrapper.NdbTxBgn(30)
	ntools.TestErrPainic(t, "TestNdbTxErrRollback", err)
	defer func() {
		// 执行提交的时候检查是否有异常， 如果有异常就直接回滚
		err = txr.NdbTxCommit(recover())
		ntools.TestErrPainic(t, "TestNdbTxErrRollback", err)
		//SQL验证数据未被写入
		count, _, _ := nmysql.SelectOne[int64](dbWrapper, fmt.Sprintf("SELECT COUNT(id) FROM  %s.%s ", schameName, tableName))
		ntools.TestEq(t, "此时应该捕获到事务超时-数据未被写入", int64(1), *count)
	}()

	txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(2,'aaa2')", schameName, tableName))
	txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(id,t03_varchar) VALUES(3,'aaa3')", schameName, tableName))
	panic(nerror.NewRunTimeError("主动回滚事务"))
}
