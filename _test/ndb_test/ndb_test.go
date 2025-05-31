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
	DbHost: "8.137.54.220",
	DbPort: 15432,
	DbUser: "gaussdb",
	DbPwd:  "Niexq@198943",
	DbName: "niexq01",
}

var mysqlConf = &nyaml.YamlConfMysqlDb{
	DbHost: "8.137.54.220",
	DbPort: 3306,
	DbUser: "root",
	DbPwd:  "Nxq@198943",
	DbName: "niexq01",
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

var GaussDbCreateTable = `
CREATE TABLE "tb01" (
  "id" bigserial,
  "col_varchar" varchar(20) NOT NULL,
  "col_int1" int1 NOT NULL,
  "col_int2" int2 NOT NULL,
  "col_int4" int4 NOT NULL,
  "col_int8" int8 NOT NULL,
  "col_bool" bool NOT NULL,
  "col_text" text NOT NULL,
  "col_date" date NOT NULL,
  "col_time" time NOT NULL,
  "col_float4" float4 NOT NULL,
  "col_float8" float8 NOT NULL,
  "col_numeric" numeric(20,2) NOT NULL,
  PRIMARY KEY ("id")
);
CREATE INDEX "idx_col_varchar" ON "tb01" USING btree (
  "col_varchar"
);
COMMENT ON COLUMN "tb01"."id" IS '主键';
COMMENT ON COLUMN "tb01"."col_varchar" IS 'varchar非空';
COMMENT ON COLUMN "tb01"."col_int1" IS 'int1非空';
COMMENT ON COLUMN "tb01"."col_int2" IS 'int2非空';
COMMENT ON COLUMN "tb01"."col_int4" IS 'int4非空';
COMMENT ON COLUMN "tb01"."col_int8" IS 'int8非空';
COMMENT ON COLUMN "tb01"."col_bool" IS 'bool非空';
COMMENT ON COLUMN "tb01"."col_text" IS 'text非空';
COMMENT ON COLUMN "tb01"."col_date" IS 'date非空';
COMMENT ON COLUMN "tb01"."col_time" IS 'time非空';
COMMENT ON COLUMN "tb01"."col_float4" IS 'float4非空';
COMMENT ON COLUMN "tb01"."col_float8" IS 'float8非空';
COMMENT ON COLUMN "tb01"."col_numeric" IS 'decimal非空';
COMMENT ON TABLE "tb01" IS '测试表'
`
