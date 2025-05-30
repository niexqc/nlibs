package ndb_test

import (
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

func TestSqlFmt(t *testing.T) {
	sqlStr := " WHERE name='nixq'"
	if sqlext.SqlFmt(sqlStr) != sqlStr {
		t.Errorf("格式化失败:%s", sqlStr)
	}
	sqlext.PrintSql(sqlPrintConf, time.Now(), " WHERE name='nixq'")
	sqlext.PrintSql(sqlPrintConf, time.Now(), "? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	sqlext.PrintSql(sqlPrintConf, time.Now(), " WHERE name=? AND id=?", "niexq", 1)
	sqlext.PrintSql(sqlPrintConf, time.Now(), " WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	sqlext.PrintSql(sqlPrintConf, time.Now(), " WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), time.Now())
	sqlext.PrintSql(sqlPrintConf, time.Now(), " WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), time.Now(), true, false)
}

func TestSqlFmtSqlStr2Gauss(t *testing.T) {
	slog.Info(sqlext.SqlFmtSqlStr2Gauss(" WHERE name='nixq'"))
	slog.Info(sqlext.SqlFmtSqlStr2Gauss("? WHERE name=? ORDER BY id desc"))
	slog.Info(sqlext.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=?"))
	slog.Info(sqlext.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=? AND no=?"))
	slog.Info(sqlext.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=? AND no=? AND time>?"))
	slog.Info(sqlext.SqlFmtSqlStr2Gauss(" WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?"))
}

func TestNNullTime(t *testing.T) {
	type TimeA struct {
		T09Datetime sqlext.NullTime `json:"T09Datetime"`
	}
	jsonStr := `{"T09Datetime":"2025-05-07 13:09:43"}`
	timeA := &TimeA{}
	njson.SonicStr2Obj(&jsonStr, timeA)
	str := njson.SonicObj2Str(timeA)
	fmt.Println(str)
}

func TestInserSqlVals(t *testing.T) {
	type TestVo struct {
		TaskId     string          `db:"task_id" json:"taskId" zhdesc:"任务编号"`
		RmqMsgId   string          `db:"rmq_msg_id" json:"rmqMsgId" zhdesc:"RmqMsgId"`
		TaskStatus int             `db:"task_status" json:"taskStatus" zhdesc:"任务状态：0待执行-1成功-2失败"`
		MdTime     sqlext.NullTime `db:"md_time" json:"mdTime" zhdesc:"修改时间"`
	}

	colsStr := "task_id,rmq_msg_id,task_status,md_time"
	slog.Info("ColStr:" + colsStr)
	strcVo := &TestVo{TaskId: "111", RmqMsgId: "SSSS", MdTime: sqlext.NewNullTime(true, ntools.TimeStr2TimeByLayout("2025-05-01", "2006-01-02"))}

	dyColStr := ndb.StructDoDbColStr(reflect.TypeOf(TestVo{}), "", "task_id", "rmq_msg_id")
	slog.Info("dyColStr:" + dyColStr)
	if dyColStr != "task_status,md_time" {
		t.Error("动态SQL列失败")
	}

	zwf, vals, _ := sqlext.InserSqlVals(dyColStr, strcVo)
	slog.Info("zwf:", "len", len(strings.Split(zwf, ",")), "text", zwf)
	if zwf != "?,?" {
		t.Error("占位符返回失败")
	}
	valsJsonStr := njson.SonicObj2Str(vals)
	slog.Info("val:", "len", len(vals), "valsJson", valsJsonStr)
	if valsJsonStr != `[0,"2025-05-01 00:00:00"]` {
		t.Error("值列表返回失败")
	}

}
