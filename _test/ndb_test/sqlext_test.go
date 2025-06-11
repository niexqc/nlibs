package ndb_test

import (
	"log/slog"
	"reflect"
	"testing"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/shopspring/decimal"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}
func TestSqlFmt(t *testing.T) {

	sqlFmtStr, err := sqlext.SqlFmt(" WHERE name='nixq'")
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", ` WHERE name='nixq'`, sqlFmtStr)

	sqlFmtStr, err = sqlext.SqlFmt("? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", `'aaa' WHERE name='niexq2' ORDER BY id desc`, sqlFmtStr)

	sqlFmtStr, err = sqlext.SqlFmt(" WHERE name=? AND id=?", "niexq", 1)
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", ` WHERE name='niexq' AND id=1`, sqlFmtStr)

	sqlFmtStr, err = sqlext.SqlFmt(" WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", ` WHERE name='niexq' AND id=1 AND no=2`, sqlFmtStr)

	str2Time, _ := ntools.TimeStr2TimeByLayout("20250501", "20060102")
	sqlFmtStr, err = sqlext.SqlFmt(" WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), str2Time)
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", ` WHERE name='niexq' AND id=1 AND no=2 AND time>'2025-05-01 00:00:00'`, sqlFmtStr)

	sqlFmtStr, err = sqlext.SqlFmt(" WHERE name=? AND id=? AND no=? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), true, false)
	ntools.TestErrPainic(t, "TestSqlFmt", err)
	ntools.TestEq(t, "TestSqlFmt sqlext.SqlFmt", ` WHERE name='niexq' AND id=1 AND no=2 AND bool_true=true AND bool_false=false`, sqlFmtStr)
}

func TestSqlFmtPainc(t *testing.T) {
	params := []int{1, 2, 3}
	_, err := sqlext.SqlFmt(" WHERE name=?", params)
	ntools.TestEq(t, "TestSqlFmtPainc", "参数【[1 2 3]】不能为Array|Slice", err.Error())
}

func TestInserSqlVals(t *testing.T) {
	type TestVo struct {
		TaskId     string            `schm:"scsz" tbn:"wts_send_task" db:"task_id" json:"taskId" zhdesc:"任务编号"`
		BizCode    sqlext.NullString `schm:"scsz" tbn:"wts_send_task" db:"biz_code" json:"bizCode" zhdesc:"业务编码"`
		TaskStatus int               `schm:"scsz" tbn:"wts_send_task" db:"task_status" json:"taskStatus" zhdesc:"任务状态：20待发送|40已发送|50发送失败|60上传成功|70上传失败"`
		CtTime     sqlext.NullTime   `schm:"scsz" tbn:"wts_send_task" db:"ct_time" json:"ctTime" zhdesc:"创建时间"`
	}

	colsStr := "task_id,biz_code,task_status,ct_time"
	slog.Info("ColStr:" + colsStr)
	colTime, _ := ntools.TimeStr2TimeByLayout("2025-05-01", "2006-01-02")

	strcVo := &TestVo{TaskId: "111", BizCode: sqlext.NewNullString(false, ""), CtTime: sqlext.NewNullTime(true, colTime)}

	dyColStr, err := ndb.StructDoDbColStr(reflect.TypeOf(TestVo{}), "", "task_id", "rmq_msg_id", "biz_code")
	ntools.TestErrPainic(t, "通过ndb.StructDoDbColStr获取动态列失败", err)
	slog.Info("dyColStr:" + dyColStr)
	ntools.TestEq(t, "通过ndb.StructDoDbColStr获取动态列失败", "task_status,ct_time", dyColStr)

	zwf, vals, _ := sqlext.InserSqlVals(dyColStr, strcVo)
	ntools.TestEq(t, "通过sqlext.InserSqlVals获取占位符失败", "?,?", zwf)
	ntools.TestEq(t, "通过sqlext.InserSqlVals获取值列表失败", `[0,"2025-05-01 00:00:00"]`, njson.Obj2StrWithPanicError(vals))

}

func TestSqlFmtSqlInNotExist(t *testing.T) {
	ids := []int64{1, 2, 6, 7}
	sqlStr, allArgs, err := sqlext.SqlFmtSqlInNotExist("test01", "id", ids)

	ntools.TestErrPainic(t, "TestSqlFmtSqlInNotExist", err)

	ntools.TestEq(t, "通过sqlext.SqlFmtSqlInNotExist 获取SQL失败", `SELECT t1.id 
FROM ( SELECT ? AS id UNION ALL  SELECT ? AS id UNION ALL  SELECT ? AS id UNION ALL  SELECT ? AS id) t1 
LEFT JOIN ( SELECT id FROM  test01 WHERE id IN (?, ?, ?, ?)) t2 ON t1.id=t2.id 
WHERE t2.id IS NULL ORDER BY t1.id ASC`, sqlStr)

	ntools.TestEq(t, "通过sqlext.SqlFmtSqlInNotExist 获取值列表失败", "[1,2,6,7,1,2,6,7]", njson.Obj2StrWithPanicError(allArgs))
}

func TestNNullVo(t *testing.T) {
	type TestNullVo struct {
		T01 sqlext.NullString   `json:"t01"`
		T02 sqlext.NullTime     `json:"t02"`
		T03 sqlext.NullInt      `json:"t03"`
		T04 sqlext.NullInt64    `json:"t04"`
		T05 sqlext.NullFloat64  `json:"t05"`
		T06 sqlext.NullBool     `json:"t06"`
		T07 decimal.NullDecimal `json:"t07"`
	}
	testVo := &TestNullVo{}
	ntools.TestEq(t, "TestNNullVo 失败", `{"t01":null,"t02":null,"t03":null,"t04":null,"t05":null,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	testVo = &TestNullVo{
		T01: sqlext.NewNullString(true, "1"),
	}
	ntools.TestEq(t, "TestNNullVo NullString 失败", `{"t01":"1","t02":null,"t03":null,"t04":null,"t05":null,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))
	str2Time, _ := ntools.TimeStr2TimeByLayout("2025-05-01", "2006-01-02")
	testVo = &TestNullVo{
		T02: sqlext.NewNullTime(true, str2Time),
	}
	ntools.TestEq(t, "TestNNullVo NullTime 失败", `{"t01":null,"t02":"2025-05-01 00:00:00","t03":null,"t04":null,"t05":null,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	testVo = &TestNullVo{
		T03: sqlext.NewNullInt(true, 1),
	}
	ntools.TestEq(t, "TestNNullVo NullInt 失败", `{"t01":null,"t02":null,"t03":1,"t04":null,"t05":null,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	testVo = &TestNullVo{
		T04: sqlext.NewNullInt64(true, 1),
	}
	ntools.TestEq(t, "TestNNullVo NullInt64 失败", `{"t01":null,"t02":null,"t03":null,"t04":1,"t05":null,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	testVo = &TestNullVo{
		T05: sqlext.NewNullFloat64(true, 1.1),
	}
	ntools.TestEq(t, "TestNNullVo NullFloat64 失败", `{"t01":null,"t02":null,"t03":null,"t04":null,"t05":1.1,"t06":null,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	testVo = &TestNullVo{
		T06: sqlext.NewNullBool(true, false),
	}
	ntools.TestEq(t, "TestNNullVo NullBool 失败", `{"t01":null,"t02":null,"t03":null,"t04":null,"t05":null,"t06":false,"t07":null}`, njson.Obj2StrWithPanicError(testVo))

	decimalVal, _ := decimal.NewFromString("1.47000")
	testVo = &TestNullVo{
		T07: decimal.NewNullDecimal(decimalVal),
	}
	ntools.TestEq(t, "TestNNullVo NullDecimal 失败", `{"t01":null,"t02":null,"t03":null,"t04":null,"t05":null,"t06":null,"t07":"1.47"}`, njson.Obj2StrWithPanicError(testVo))

}
