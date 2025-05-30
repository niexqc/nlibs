package ndb_test

import (
	"github.com/niexqc/nlibs/ndb/ngauss"
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var NGaussWrapper *ngauss.NGaussWrapper

var NMysqlWrapper *nmysql.NMysqlWrapper

var gaussConf = &nyaml.YamlConfGaussDb{
	DbHost: "192.168.0.253",
	DbPort: 15432,
	DbUser: "gaussdb",
	DbPwd:  "Cdwts@2025",
	DbName: "ndb_test",
}
var mysqlConf = &nyaml.YamlConfMysqlDb{
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
	NGaussWrapper = ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
	NMysqlWrapper = nmysql.NewNMysqlWrapper(mysqlConf, sqlPrintConf)
}
