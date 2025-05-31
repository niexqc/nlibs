package ngauss_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/niexqc/nlibs/ndb/ngauss"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
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

func TestCrateTable(t *testing.T) {
	dbWrapper := ngauss.NewNGaussWrapper(gaussConf, sqlPrintConf)
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
