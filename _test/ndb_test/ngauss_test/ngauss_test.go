package ngauss_test

import (
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/ndb/ngauss"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/shopspring/decimal"
)

var tableName = "tb01"
var schameName = "ndb_test"
var caussdbCreateTableStr = ""

var gaussConf *nyaml.YamlConfGaussDb
var sqlPrintConf *nyaml.YamlConfSqlPrint

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	gaussConf = &nyaml.YamlConfGaussDb{
		DbHost: "8.137.54.220",
		DbPort: 15432,
		DbUser: "gaussdb",
		DbPwd:  "Niexq@198943",
		DbName: "ndb_test",
	}
	sqlPrintConf = &nyaml.YamlConfSqlPrint{
		DbSqlLogPrint:    true,
		DbSqlLogLevel:    "debug",
		DbSqlLogCompress: false,
	}

	// 采取文本替换的形式
	caussdbCreateTableSrcStr := `CREATE TABLE "tb01" (
  "id" bigserial,
  "col_varchar" varchar(20),
  "col_int1" int1,
  "col_int2" int2,
  "col_int4" int4,
  "col_int8" int8,
  "col_bool" bool,
  "col_text" text,
  "col_date" date,
  "col_time" time,
  "col_float4" float4,
  "col_float8" float8,
  "col_numeric" numeric(20,2),
  PRIMARY KEY ("id")
);
CREATE INDEX "idx_col_varchar" ON "tb01" USING btree ("col_varchar");
COMMENT ON COLUMN "tb01"."id" IS '主键';
COMMENT ON COLUMN "tb01"."col_varchar" IS 'varchar空';
COMMENT ON COLUMN "tb01"."col_int1" IS 'int1空';
COMMENT ON COLUMN "tb01"."col_int2" IS 'int2空';
COMMENT ON COLUMN "tb01"."col_int4" IS 'int4空';
COMMENT ON COLUMN "tb01"."col_int8" IS 'int8空';
COMMENT ON COLUMN "tb01"."col_bool" IS 'bool空';
COMMENT ON COLUMN "tb01"."col_text" IS 'text空';
COMMENT ON COLUMN "tb01"."col_date" IS 'date空';
COMMENT ON COLUMN "tb01"."col_time" IS 'time空';
COMMENT ON COLUMN "tb01"."col_float4" IS 'float4空';
COMMENT ON COLUMN "tb01"."col_float8" IS 'float8空';
COMMENT ON COLUMN "tb01"."col_numeric" IS 'decimal空';
COMMENT ON TABLE "tb01" IS '测试表';`

	caussdbCreateTableStr = strings.ReplaceAll(caussdbCreateTableSrcStr, `"tb01"`, fmt.Sprintf(`%s.%s`, schameName, tableName))

}

func TestSqlFmtSqlStr2Gauss(t *testing.T) {

	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)

	sqlFmtStr := dbWrapper.SqlFmtSqlStr2Gauss(" WHERE name='nixq'")
	ntools.TestEq(t, "TestSqlFmt dbWrapper.SqlFmtSqlStr2Gauss", ` WHERE name='nixq'`, sqlFmtStr)

	sqlFmtStr = dbWrapper.SqlFmtSqlStr2Gauss("? WHERE name=? ORDER BY id desc")
	ntools.TestEq(t, "TestSqlFmt dbWrapper.SqlFmtSqlStr2Gauss", `$1 WHERE name=$2 ORDER BY id desc`, sqlFmtStr)

	sqlFmtStr = dbWrapper.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=?")
	ntools.TestEq(t, "TestSqlFmt dbWrapper.SqlFmtSqlStr2Gauss", ` WHERE name=$1 AND id=$2`, sqlFmtStr)

	sqlFmtStr = dbWrapper.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=? AND no=?")
	ntools.TestEq(t, "TestSqlFmt dbWrapper.SqlFmtSqlStr2Gauss", ` WHERE name=$1 AND id=$2 AND no=$3`, sqlFmtStr)

	sqlFmtStr = dbWrapper.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=? AND no=? AND time>?")
	ntools.TestEq(t, "TestSqlFmt dbWrapper.SqlFmtSqlStr2Gauss", ` WHERE name=$1 AND id=$2 AND no=$3 AND time>$4`, sqlFmtStr)

}

func TestCrateTable(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	_, err := dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	ntools.TestErrPainic(t, "TestCrateTable DROP TABLE ", err)

	_, err = dbWrapper.Exec(caussdbCreateTableStr)
	ntools.TestErrPainic(t, "TestCrateTable CREATE TABLE", err)

	tcSql := fmt.Sprintf("SELECT obj_description('%s.%s'::regclass) tableComment", schameName, tableName)

	comment, findOk, err := ngauss.SelectOne[string](dbWrapper, tcSql)
	ntools.TestErrPainic(t, "TestCrateTable SELECT tableComment ", err)

	if !findOk {
		ntools.TestErrPanicMsg(t, "TestCrateTable SELECT tableComment 未获取到注释 ")
	}
	ntools.TestEq(t, "TestCrateTable SELECT tableComment ", "测试表", *comment)
}

func TestGenStruct(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	str, err := dbWrapper.GetStructDoByTableStr(schameName, tableName)
	ntools.TestErrPainic(t, "TestGenStruct ", err)
	slog.Info(str)
	if !strings.Contains(str, "Id int64") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "Id int64")
	}
	if !strings.Contains(str, "ColVarchar sqlext.NullString") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "ColVarchar sqlext.NullString")
	}
	if !strings.Contains(str, "ColTime sqlext.NullTime") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "ColTime sqlext.NullTime")
	}
	if !strings.Contains(str, "ColNumeric decimal.NullDecimal") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "ColNumeric decimal.NullDecimal")
	}
	if !strings.Contains(str, "测试表 ndb_test.tb01") {
		t.Errorf("TestGenStruct 生成的结果中，没有包含:%s", "Test `niexq01`.test01")
	}
	t.Log("TestGenStruct 执行成功")
}

func TestInsert(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))
	lasetId, _ := dbWrapper.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2') RETURNING id", schameName, tableName))
	if lasetId != 2 {
		t.Error("InsertWithLastId 应该返回2")
	}
	rowEff, _ := dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into  %s.%s(col_varchar) VALUES('aaa3'),('aaa4'),('aa5'),('aaa6')", schameName, tableName))
	if rowEff != 4 {
		t.Error("InsertWithRowsAffected应该返回4")
	}
}

func TestSelectOne(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))

	querySql := fmt.Sprintf("SELECT col_varchar FROM %s.%s WHERE id=1", schameName, tableName)
	if res, _, err := ngauss.SelectOne[sqlext.NullString](dbWrapper, querySql); nil != err {
		t.Error(err)
	} else {
		if res.NullString.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
	querySql = fmt.Sprintf("SELECT id FROM %s.%s WHERE id=1", schameName, tableName)
	if res, _, err := ngauss.SelectOne[int64](dbWrapper, querySql); nil != err {
		t.Error(err)
	} else {
		if *res != 1 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectObj(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))

	// 测试表 ndb_test.tb01
	type Tb01Do struct {
		Id         int64               `schm:"ndb_test" tbn:"tb01" db:"id" json:"id" zhdesc:"主键"`
		ColVarchar sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"col_varchar" json:"colVarchar" zhdesc:"varchar空"`
		ColInt1    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int1" json:"colInt1" zhdesc:"int1空"`
		ColInt2    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int2" json:"colInt2" zhdesc:"int2空"`
		ColInt4    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int4" json:"colInt4" zhdesc:"int4空"`
		ColInt8    sqlext.NullInt64    `schm:"ndb_test" tbn:"tb01" db:"col_int8" json:"colInt8" zhdesc:"int8空"`
		ColBool    sqlext.NullBool     `schm:"ndb_test" tbn:"tb01" db:"col_bool" json:"colBool" zhdesc:"bool空"`
		ColText    sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"col_text" json:"colText" zhdesc:"text空"`
		ColDate    sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"col_date" json:"colDate" zhdesc:"date空"`
		ColTime    sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"col_time" json:"colTime" zhdesc:"time空"`
		ColFloat4  sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"col_float4" json:"colFloat4" zhdesc:"float4空"`
		ColFloat8  sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"col_float8" json:"colFloat8" zhdesc:"float8空"`
		ColNumeric decimal.NullDecimal `schm:"ndb_test" tbn:"tb01" db:"col_numeric" json:"colNumeric" zhdesc:"decimal空"`
	}

	querySql := fmt.Sprintf("SELECT * FROM %s.%s WHERE id=1", schameName, tableName)

	if obj, _, err := ngauss.SelectObj[Tb01Do](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if obj.Id != 1 || obj.ColVarchar.String != "aaa1" {
			t.Error("返回值不匹配")
		}
	}
}

func TestSelectList(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2')", schameName, tableName))

	// 测试表 ndb_test.tb01
	type Tb01Do struct {
		Id         int64               `schm:"ndb_test" tbn:"tb01" db:"id" json:"id" zhdesc:"主键"`
		ColVarchar sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"col_varchar" json:"colVarchar" zhdesc:"varchar空"`
		ColInt1    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int1" json:"colInt1" zhdesc:"int1空"`
		ColInt2    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int2" json:"colInt2" zhdesc:"int2空"`
		ColInt4    sqlext.NullInt      `schm:"ndb_test" tbn:"tb01" db:"col_int4" json:"colInt4" zhdesc:"int4空"`
		ColInt8    sqlext.NullInt64    `schm:"ndb_test" tbn:"tb01" db:"col_int8" json:"colInt8" zhdesc:"int8空"`
		ColBool    sqlext.NullBool     `schm:"ndb_test" tbn:"tb01" db:"col_bool" json:"colBool" zhdesc:"bool空"`
		ColText    sqlext.NullString   `schm:"ndb_test" tbn:"tb01" db:"col_text" json:"colText" zhdesc:"text空"`
		ColDate    sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"col_date" json:"colDate" zhdesc:"date空"`
		ColTime    sqlext.NullTime     `schm:"ndb_test" tbn:"tb01" db:"col_time" json:"colTime" zhdesc:"time空"`
		ColFloat4  sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"col_float4" json:"colFloat4" zhdesc:"float4空"`
		ColFloat8  sqlext.NullFloat64  `schm:"ndb_test" tbn:"tb01" db:"col_float8" json:"colFloat8" zhdesc:"float8空"`
		ColNumeric decimal.NullDecimal `schm:"ndb_test" tbn:"tb01" db:"col_numeric" json:"colNumeric" zhdesc:"decimal空"`
	}

	querySql := fmt.Sprintf("SELECT col_varchar FROM %s.%s ORDER BY id ASC", schameName, tableName)

	if list, err := ngauss.SelectList[sqlext.NullString](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].String != "aaa1" || list[1].String != "aaa2" {
			t.Error("返回值不匹配")
		}
	}

	querySql = fmt.Sprintf("SELECT * FROM %s.%s ORDER BY id ASC", schameName, tableName)
	if list, err := ngauss.SelectList[Tb01Do](dbWrapper, querySql); nil != err {
		println(err.Error())
	} else {
		if len(list) != 2 || list[0].Id != 1 || list[1].Id != 2 {
			t.Error("返回值不匹配")
		}
	}
}

func TestSqlInNotExist(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa3')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa4')", schameName, tableName))

	ids := []int64{1, 2, 6, 7}
	sqlStr, allArgs, err := sqlext.SqlFmtSqlInNotExist(fmt.Sprintf("%s.%s", schameName, tableName), "id", ids)
	ntools.TestErrPainic(t, "TestSqlInNotExist ", err)

	notExistIdds, err := ngauss.SelectList[int64](dbWrapper, sqlStr, nlibs.Arr2ArrAny(allArgs)...)
	ntools.TestErrPainic(t, "TestSqlInNotExist ", err)

	acResult := njson.SonicObj2StrWithPanicError(notExistIdds)
	ntools.TestEq(t, "TestSqlInNotExist ", "[6,7]", acResult)

}

func TestSelectDyObjAndList(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2')", schameName, tableName))

	dyObj, err := dbWrapper.SelectDyObj(fmt.Sprintf("SELECT * FROM %s.%s where id=1", schameName, tableName))
	ntools.TestErrPainic(t, "TestSelectDyObj ", err)

	val, err := ngauss.GetFiledVal[sqlext.NullString](dyObj, dyObj.DbNameFiledsMap["col_varchar"].StructFieldName)
	ntools.TestErrPainic(t, "TestSelectDyObj ", err)
	ntools.TestEq(t, "TestSelectDyObj ", "aaa1", val.String)

	// 列表
	dyObjList, err := dbWrapper.SelectDyObjList(fmt.Sprintf("SELECT * FROM %s.%s ORDER BY id ASC", schameName, tableName))
	ntools.TestErrPainic(t, "TestSelectDyList ", err)
	ntools.TestEq(t, "TestSelectDyList 列表数", 2, len(dyObjList))

	jsonStr, err := ngauss.DyObjList2Json(dyObjList)
	ntools.TestErrPainic(t, "TestSelectDyList 转Json失败 ", err)

	expJson := `[{"id":1,"colVarchar":"aaa1","colInt1":null,"colInt2":null,"colInt4":null,"colInt8":null,"colBool":null,"colText":null,"colDate":null,"colTime":null,"colFloat4":null,"colFloat8":null,"colNumeric":null},{"id":2,"colVarchar":"aaa2","colInt1":null,"colInt2":null,"colInt4":null,"colInt8":null,"colBool":null,"colText":null,"colDate":null,"colTime":null,"colFloat4":null,"colFloat8":null,"colNumeric":null}]`
	ntools.TestEq(t, "TestSelectDyList ", expJson, jsonStr)

}

func TestNdbTx(t *testing.T) {
	dbWrapper, _ := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	dbWrapper.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", schameName, tableName))
	dbWrapper.Exec(caussdbCreateTableStr)

	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa1')", schameName, tableName))
	dbWrapper.InsertWithRowsAffected(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2')", schameName, tableName))

	ntools.SlogSetTraceId("TestNdbTx")

	// time.Sleep(6 * time.Second)
	txr, err := dbWrapper.NdbTxBgn(3)
	ntools.TestErrPainic(t, "TestNdbTx", err)
	defer txr.NdbTxCommit()

	lasetId, _ := txr.InsertWithLastId(fmt.Sprintf("INSERT into %s.%s(col_varchar) VALUES('aaa2') RETURNING id", schameName, tableName))
	if lasetId != 3 {
		t.Error("InsertWithLastId 应该返回2")
	}
	rowEff, _ := txr.InsertWithRowsAffected(fmt.Sprintf("INSERT into  %s.%s(col_varchar) VALUES('aaa3'),('aaa4'),('aa5'),('aaa6')", schameName, tableName))
	if rowEff != 4 {
		t.Error("InsertWithRowsAffected应该返回4")
	}

}
