package ngauss

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "gitee.com/opengauss/openGauss-connector-go-pq"
	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
)

type NGaussWrapper struct {
	sqlxDb       *sqlx.DB
	conf         *nyaml.YamlConfGaussDb
	sqlPrintConf *nyaml.YamlConfSqlPrint
	bgnTx        bool
	sqlxTx       *sqlx.Tx
	// 	sqlxTxContext           context.Context
	// 	sqlxTxContextCancelFunc context.CancelFunc
	// 	txState                 int32 //
	// 	txMutx                  *sync.Mutex
}

func NewNGaussWrapper(conf *nyaml.YamlConfGaussDb, sqlPrintConf *nyaml.YamlConfSqlPrint) *NGaussWrapper {
	connStr := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable"
	connStr = fmt.Sprintf(connStr, conf.DbHost, conf.DbPort, conf.DbUser, conf.DbPwd, conf.DbName)
	slog.Debug(connStr)
	db, err := sqlx.Open("opengauss", connStr)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("连接到Gauss失败", err))
	}
	db.SetConnMaxLifetime(time.Second * time.Duration(conf.ConnMaxLifetime))
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	return &NGaussWrapper{sqlxDb: db, conf: conf, sqlPrintConf: sqlPrintConf, bgnTx: false}
}

func (ndbw *NGaussWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	gaussSqlStr := sqlext.SqlFmtSqlStr2Gauss(sqlStr)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(gaussSqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(gaussSqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NGaussWrapper) SelectOne(dest any, sqlStr string, args ...any) (findOk bool, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	gaussSqlStr := sqlext.SqlFmtSqlStr2Gauss(sqlStr)

	var rows *sqlx.Rows

	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(gaussSqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(gaussSqlStr, args...)
	}

	if nil != err {
		return false, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if nil != err {
		return false, err
	}
	if len(cols) != 1 {
		return false, nerror.NewRunTimeError("查询结果包含多个列")
	}
	if rows.Next() {
		if err := rows.Scan(dest); nil != err {
			return false, err
		}
		if rows.Next() {
			dest = nil
			return false, nerror.NewRunTimeError("查询结果中包含多个值")
		}
		return true, err
	} else {
		return false, nil
	}
}

func (ndbw *NGaussWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	gaussSqlStr := sqlext.SqlFmtSqlStr2Gauss(sqlStr)

	if ndbw.bgnTx {
		return ndbw.sqlxTx.Select(dest, gaussSqlStr, args...)
	} else {
		return ndbw.sqlxDb.Select(dest, gaussSqlStr, args...)
	}
}
