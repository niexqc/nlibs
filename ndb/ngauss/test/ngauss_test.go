package ngauss_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/niexqc/nlibs/ndb/ngauss"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/shopspring/decimal"
)

var NGaussWrapper *ngauss.NGaussWrapper

var dbconf = &nyaml.YamlConfGaussDb{
	DbHost: "192.168.0.253",
	DbPort: 15432,
	DbUser: "gaussdb",
	DbPwd:  "Cdwts@2025",
	DbName: "ndb_test",
}
var sqlPrintConf = &nyaml.YamlConfSqlPrint{
	DbSqlLogPrint:    true,
	DbSqlLogLevel:    "debug",
	DbSqlLogCompress: false,
}

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
	NGaussWrapper = ngauss.NewNGaussWrapper(dbconf, sqlPrintConf)
}

func TestGetStructDoByTableStr(t *testing.T) {
	str := NGaussWrapper.GetStructDoByTableStr("public", "test0")
	slog.Info("\n" + str)
}

func TestInsert(t *testing.T) {
	// _, err := NGaussWrapper.Insert("INSERT into public.test0(id,name) VALUES(1,?)", "TestInsert")
	// if nil != err {
	// 	slog.Info("错误发生,测试通过:", "err", err)
	// }
	rowsAffected, err := NGaussWrapper.InsertWithRowsAffected("INSERT into public.test0(name) VALUES(?)", "TestInsert")
	if nil != err {
		t.Error(err)
	}
	slog.Info("rowsAffected:", "rowsAffected", rowsAffected)

	lastInsertId, err := NGaussWrapper.InsertWithLastId("INSERT into public.test0(name) VALUES(?) RETURNING id", "InsertAndLastId")
	if nil != err {
		t.Error(err)
	}
	slog.Info("lastInsertId:", "lastInsertId", lastInsertId)

}

func TestSelectDyObj(t *testing.T) {
	// NGaussWrapper.Exec(sql, "")

	if dyObj, err := NGaussWrapper.SelectDyObj("SELECT * FROM public.test0 where id=1"); nil != err {
		println(err.Error())
	} else {
		val, err := ngauss.GetFiledVal[sqlext.NullString](dyObj, dyObj.DbNameFiledsMap["name"].StructFieldName)
		if nil != err {
			panic(err)
		}
		fmt.Println(val.String)
	}

	if dyObjList, err := NGaussWrapper.SelectDyObjList("SELECT * FROM public.test0"); nil != err {
		println(err.Error())
	} else {
		for _, v := range dyObjList {
			val, err := ngauss.GetFiledVal[sqlext.NullString](v, "Name")
			if nil != err {
				panic(err)
			}
			fmt.Println(val.String)
		}
	}
}

func TestSelectObj(t *testing.T) {
	// NGaussWrapper.Exec(sql, "")

	// 测试 public.test0
	type Test0Do struct {
		Id          int64               `dbtb:"test0" db:"id" json:"id" zhdesc:"ID"`
		Name        string              `dbtb:"test0" db:"name" json:"name" zhdesc:"名称"`
		ClInt1      sqlext.NullInt      `dbtb:"test0" db:"cl_int1" json:"clInt1" zhdesc:"测试byte"`
		ClInt2      sqlext.NullInt      `dbtb:"test0" db:"cl_int2" json:"clInt2" zhdesc:""`
		ClInt4      sqlext.NullInt      `dbtb:"test0" db:"cl_int4" json:"clInt4" zhdesc:""`
		ClInt8      sqlext.NullInt64    `dbtb:"test0" db:"cl_int8" json:"clInt8" zhdesc:""`
		ClText      sqlext.NullString   `dbtb:"test0" db:"cl_text" json:"clText" zhdesc:""`
		ClDecimal   decimal.NullDecimal `dbtb:"test0" db:"cl_decimal" json:"clDecimal" zhdesc:""`
		ClFloat8    sqlext.NullFloat64  `dbtb:"test0" db:"cl_float8" json:"clFloat8" zhdesc:""`
		ClBool      sqlext.NullBool     `dbtb:"test0" db:"cl_bool" json:"clBool" zhdesc:""`
		ClDate      sqlext.NullTime     `dbtb:"test0" db:"cl_date" json:"clDate" zhdesc:""`
		ClTime      sqlext.NullTime     `dbtb:"test0" db:"cl_time" json:"clTime" zhdesc:""`
		ClDatetime  sqlext.NullTime     `dbtb:"test0" db:"cl_datetime" json:"clDatetime" zhdesc:""`
		ClDatetimez sqlext.NullTime     `dbtb:"test0" db:"cl_datetimez" json:"clDatetimez" zhdesc:""`
	}

	do := Test0Do{}
	_, err := NGaussWrapper.SelectObj(&do, "SELECT * FROM public.test0 where id=1")
	if nil != err {
		t.Error(err)
	}
	slog.Info(njson.SonicObj2Str(do))

}

func TestSelectObjList(t *testing.T) {
	// NGaussWrapper.Exec(sql, "")

	// 测试 public.test0
	type Test0Do struct {
		Id          int64               `dbtb:"test0" db:"id" json:"id" zhdesc:"ID"`
		Name        string              `dbtb:"test0" db:"name" json:"name" zhdesc:"名称"`
		ClInt1      sqlext.NullInt      `dbtb:"test0" db:"cl_int1" json:"clInt1" zhdesc:"测试byte"`
		ClInt2      sqlext.NullInt      `dbtb:"test0" db:"cl_int2" json:"clInt2" zhdesc:""`
		ClInt4      sqlext.NullInt      `dbtb:"test0" db:"cl_int4" json:"clInt4" zhdesc:""`
		ClInt8      sqlext.NullInt64    `dbtb:"test0" db:"cl_int8" json:"clInt8" zhdesc:""`
		ClText      sqlext.NullString   `dbtb:"test0" db:"cl_text" json:"clText" zhdesc:""`
		ClDecimal   decimal.NullDecimal `dbtb:"test0" db:"cl_decimal" json:"clDecimal" zhdesc:""`
		ClFloat8    sqlext.NullFloat64  `dbtb:"test0" db:"cl_float8" json:"clFloat8" zhdesc:""`
		ClBool      sqlext.NullBool     `dbtb:"test0" db:"cl_bool" json:"clBool" zhdesc:""`
		ClDate      sqlext.NullTime     `dbtb:"test0" db:"cl_date" json:"clDate" zhdesc:""`
		ClTime      sqlext.NullTime     `dbtb:"test0" db:"cl_time" json:"clTime" zhdesc:""`
		ClDatetime  sqlext.NullTime     `dbtb:"test0" db:"cl_datetime" json:"clDatetime" zhdesc:""`
		ClDatetimez sqlext.NullTime     `dbtb:"test0" db:"cl_datetimez" json:"clDatetimez" zhdesc:""`
	}

	dos := []*Test0Do{}
	err := NGaussWrapper.SelectList(&dos, "SELECT * FROM public.test0 ")
	if nil != err {
		t.Error(err)
	}
	slog.Info(njson.SonicObj2Str(dos))

}

func TestTx(t *testing.T) {

	ntools.SlogSetTraceId("1111")
	// time.Sleep(6 * time.Second)
	txr, err := NGaussWrapper.NdbTxBgn(3)
	if nil != err {
		panic(err)
	}
	defer txr.NdbTxCommit()

	rowsAffected, err := txr.InsertWithRowsAffected("INSERT into public.test0(name) VALUES(?)", "TestInsert")
	if nil != err {
		t.Error(err)
	}
	slog.Info("rowsAffected:", "rowsAffected", rowsAffected)

	lastInsertId, err := txr.InsertWithLastId("INSERT into public.test0(name) VALUES(?) RETURNING id", "InsertAndLastId")
	if nil != err {
		t.Error(err)
	}
	slog.Info("lastInsertId:", "lastInsertId", lastInsertId)

}
