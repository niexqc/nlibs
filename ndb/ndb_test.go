package ndb_test

import (
	"log/slog"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/njson"
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
		DbName:           "tb-jsc",
		DbSqlLogPrint:    true,
		DbSqlLogLevel:    "debug",
		DbSqlLogCompress: false,
	}
	IDbWrapper = ndb.InitMysqlConnPool(dbconf)

}

type UserDto struct {
	UserId     int64          `json:"userId" db:"user_id"`
	UserAcc    ndb.NullString `json:"userAcc" db:"user_acc"`
	UserPwd    ndb.NullString `json:"userPwd" db:"user_pwd"`
	UserName   ndb.NullString `json:"userName" db:"user_name"`
	UserRemark ndb.NullString `json:"userRemark" db:"user_remark"`
	CreateTime ndb.NullTime   `json:"createTime" db:"create_time"`
}

func TestDbSelect(t *testing.T) {
	sqlStr := `
	SELECT user_id,user_acc,user_pwd,user_name,user_remark,create_time 
	 FROM jsc_user 	 WHERE user_acc=?
	`
	var users []UserDto
	err := IDbWrapper.SelectList(&users, sqlStr, "niexq")
	if err != nil {
		panic(err)
	}
	slog.Info(njson.SonicObj2Str(users))
}

func TestUserService(t *testing.T) {
	IDbWrapper.PrintSql(time.Now(), " WHERE name='nixq'")
	IDbWrapper.PrintSql(time.Now(), "? WHERE name=? ORDER BY id desc", "aaa", "niexq2")
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=?", "niexq", 1)
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=?", "niexq", 1, int64(2))
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>?", "niexq", 1, int64(2), time.Now())
	IDbWrapper.PrintSql(time.Now(), " WHERE name=? AND id=? AND no=? AND time>? AND bool_true=? AND bool_false=?", "niexq", 1, int64(2), time.Now(), true, false)
}
