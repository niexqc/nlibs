package ndb_test

import (
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var IDbWrapper *ndb.NDbWrapper

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
func Test001(t *testing.T) {
	IDbWrapper.GenDoByTable("niexq01", "nba_user")
}

func Test002(t *testing.T) {
	IDbWrapper.PrintSql(time.Now(), " WHERE name='nixq'")
	IDbWrapper.PrintSql(time.Now(), "? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=?", "niexq", 1)
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), time.Now())
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), time.Now(), true, false)
}
