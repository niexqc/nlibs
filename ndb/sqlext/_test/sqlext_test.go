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
