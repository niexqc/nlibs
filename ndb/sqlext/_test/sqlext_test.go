package sqlext_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var IDbWrapper *nmysql.NMysqlWrapper

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
	IDbWrapper = ndb.NewNMysqlWrapper(dbconf)

}

func TestSqlFmt(t *testing.T) {
	sqlStr := " WHERE name='nixq'"
	if sqlext.SqlFmt(sqlStr) != sqlStr {
		t.Errorf("格式化失败:%s", sqlStr)
	}
	sqlext.PrintSql(dbconf, time.Now(), " WHERE name='nixq'")
	sqlext.PrintSql(dbconf, time.Now(), "? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	sqlext.PrintSql(dbconf, time.Now(), " WHERE name=? AND id=?", "niexq", 1)
	sqlext.PrintSql(dbconf, time.Now(), " WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	sqlext.PrintSql(dbconf, time.Now(), " WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), time.Now())
	sqlext.PrintSql(dbconf, time.Now(), " WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), time.Now(), true, false)
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
	type AAAA struct {
		TaskId     string          `db:"task_id" json:"taskId" zhdesc:"任务编号"`
		RmqMsgId   string          `db:"rmq_msg_id" json:"rmqMsgId" zhdesc:"RmqMsgId"`
		SysCode    string          `db:"sys_code" json:"sysCode" zhdesc:"发送人编码"`
		AreaCode   string          `db:"area_code" json:"areaCode" zhdesc:"地区编码"`
		SendTime   sqlext.NullTime `db:"send_time" json:"sendTime" zhdesc:"发送时间"`
		BizCode    string          `db:"biz_code" json:"bizCode" zhdesc:"业务编码"`
		BizData    string          `db:"biz_data" json:"bizData" zhdesc:"业务数据"`
		BizType    string          `db:"biz_type" json:"bizType" zhdesc:"操作类型"`
		TaskStatus int             `db:"task_status" json:"taskStatus" zhdesc:"任务状态：0待执行-1成功-2失败"`
		TaskMsg    string          `db:"task_msg" json:"taskMsg" zhdesc:"任务执行消息"`
		CtTime     sqlext.NullTime `db:"ct_time" json:"ctTime" zhdesc:"创建时间"`
		MdTime     sqlext.NullTime `db:"md_time" json:"mdTime" zhdesc:"修改时间"`
	}
	nowt := time.Now()
	aaa := &AAAA{TaskId: "111", RmqMsgId: "SSSS", SendTime: sqlext.NewNullTime(&nowt), CtTime: sqlext.NewNullTime(nil)}

	ColsStr := "task_id,rmq_msg_id,sys_code,area_code,send_time,biz_code,biz_data,biz_type,task_status,task_msg,ct_time,md_time"

	zwf, vals, err := sqlext.InserSqlVals(ColsStr, aaa)

	fmt.Println(zwf, vals, err)
}
