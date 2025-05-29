package ngauss_test

import (
	"log/slog"
	"testing"

	"github.com/niexqc/nlibs/ndb/ngauss"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
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
