package npg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/niexqc/nlibs/ndb"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
)

const (
	txInActive   = int32(1)
	txActive     = int32(2)
	txCommitted  = int32(3)
	txRolledBack = int32(4)
)

type NPgWrapper struct {
	sqlxDb                  *sqlx.DB
	conf                    *nyaml.YamlConfPgDb
	sqlPrintConf            *nyaml.YamlConfSqlPrint
	bgnTx                   bool
	sqlxTx                  *sqlx.Tx
	sqlxTxContext           context.Context
	sqlxTxContextCancelFunc context.CancelFunc
	txState                 int32 //
	txMutx                  *sync.Mutex
}

func NewNPgWrapper(conf *nyaml.YamlConfPgDb, sqlPrintConf *nyaml.YamlConfSqlPrint) (*NPgWrapper, error) {
	connStr := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable "
	connStr = fmt.Sprintf(connStr, conf.DbHost, conf.DbPort, conf.DbUser, conf.DbPwd, conf.DbName)
	slog.Debug(connStr)
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, nerror.NewRunTimeErrorWithError("连接到Pg失败", err)
	}
	db.SetConnMaxLifetime(time.Second * time.Duration(conf.ConnMaxLifetime))
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	return &NPgWrapper{sqlxDb: db, conf: conf, sqlPrintConf: sqlPrintConf, bgnTx: false}, nil
}

//	 查询单个字段单个值
//		 sqlStr:=select id from table where id=?
//		 str:=ndb.SelectOne[string](ndbw,sql,id)
func SelectOne[T sqlext.NdbBasicType](ndbw *NPgWrapper, sqlStr string, args ...any) (t *T, findOk bool, err error) {
	obj := new(T)
	findOk, err = ndbw.SelectOne(obj, sqlStr, args...)
	return obj, findOk, err
}

//	 查询单行记录返回Struct实例
//		 sqlStr:=select * from table where id=?
//		 user:=ndb.SelectObj[UserDo](ndbw,sql,id)
func SelectObj[T any](ndbw *NPgWrapper, sqlStr string, args ...any) (t *T, findOk bool, err error) {
	obj := new(T)
	findOk, err = ndbw.SelectObj(obj, sqlStr, args...)
	return obj, findOk, err
}

// 查询多行记录，支持值和Struct
func SelectList[T any](ndbw *NPgWrapper, sqlStr string, args ...any) (tlist []*T, err error) {
	objs := new([]*T)
	err = ndbw.SelectList(objs, sqlStr, args...)
	return *objs, err
}

// SqlLimitStr
// pageNo 页码从1开始
func (ndbw *NPgWrapper) SqlLimitStr(pageNo, pageSize int) string {
	result, _ := ndb.SqlFmt(" LIMIT ? OFFSET ? ", pageSize, (pageNo-1)*pageSize)
	return result
}

func (ndbw *NPgWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(pgSqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(pgSqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

// 实现返回ID需要
func (ndbw *NPgWrapper) InsertWithRowsAffected(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(pgSqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(pgSqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

// 实现返回ID需要其他方法
// Sql示例:INSERT INTO users (name) VALUES ($1) RETURNING id
func (ndbw *NPgWrapper) InsertWithLastId(sqlStr string, args ...any) (lastInsertId int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	if !strings.Contains(strings.ToUpper(sqlStr), "RETURNING") {
		return 0, nerror.NewRunTimeError("InsertWithLastId 必须包含 RETURNING")
	}
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var id int64
	if ndbw.bgnTx {
		err = ndbw.sqlxTx.QueryRow(pgSqlStr, args...).Scan(&id)
	} else {
		err = ndbw.sqlxDb.QueryRow(pgSqlStr, args...).Scan(&id)
	}
	if nil != err {
		return 0, err
	}
	return id, err
}

func (ndbw *NPgWrapper) InsertFor(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(pgSqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(pgSqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NPgWrapper) SelectOne(dest any, sqlStr string, args ...any) (findOk bool, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)

	var rows *sqlx.Rows

	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(pgSqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(pgSqlStr, args...)
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

func (ndbw *NPgWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)

	if ndbw.bgnTx {
		return ndbw.sqlxTx.Select(dest, pgSqlStr, args...)
	} else {
		return ndbw.sqlxDb.Select(dest, pgSqlStr, args...)
	}
}

//	 查询并生成动态对象返回
//		 dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1")
//		 val, err := sqlext.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
func (ndbw *NPgWrapper) SelectDyObj(sqlStr string, args ...any) (dyObj *NPgDyObj, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var rows *sqlx.Rows
	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(pgSqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(pgSqlStr, args...)
	}
	if nil != err {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.ColumnTypes()
	if nil != err {
		return nil, err
	}
	// 创建动态Struct
	dyStructType, dbNameFieldsMap, err := CreateDyStruct(cols)
	if nil != err {
		return nil, err
	}
	if rows.Next() {
		// 创建动态Struct的实例
		instance := reflect.New(dyStructType).Interface()
		// 对动态Struct的实例赋值
		err := rows.StructScan(instance)
		if nil != err {
			return nil, err
		}
		if rows.Next() {
			return nil, nerror.NewRunTimeError("查询结果中包含多个值")
		}
		return &NPgDyObj{Data: instance, DbNameFiledsMap: dbNameFieldsMap}, err
		// return instance, err
	} else {
		return nil, nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NPgWrapper) SelectDyObjList(sqlStr string, args ...any) (objValList []*NPgDyObj, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)

	var rows *sqlx.Rows
	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(pgSqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(pgSqlStr, args...)
	}

	if nil != err {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.ColumnTypes()
	if nil != err {
		return nil, err
	}
	// 创建动态Struct
	dyStructType, dbNameFieldsMap, err := CreateDyStruct(cols)
	if nil != err {
		return nil, err
	}
	objValList = make([]*NPgDyObj, 0)
	for rows.Next() {
		// 创建动态Struct的实例
		instance := reflect.New(dyStructType).Interface()
		// 对动态Struct的实例赋值
		err := rows.StructScan(instance)
		if nil != err {
			return nil, err
		}
		objValList = append(objValList, &NPgDyObj{Data: instance, DbNameFiledsMap: dbNameFieldsMap})
	}
	return objValList, err
}

func (ndbw *NPgWrapper) SelectObj(dest any, sqlStr string, args ...any) (bool, error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	pgSqlStr := ndbw.SqlFmtSqlStr2Pg(sqlStr)
	var rows *sqlx.Rows
	var err error

	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(pgSqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(pgSqlStr, args...)
	}

	if nil != err {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(dest); nil != err {
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

func (ndbw *NPgWrapper) NdbTxBgn(timeoutSecond int) (txWrper *NPgWrapper, err error) {
	if timeoutSecond > 60 {
		slog.Warn("事务时长超过60秒,判断下业务")
	}
	txCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	mysqlTxWrapper := new(NPgWrapper)
	mysqlTxWrapper.sqlxDb = ndbw.sqlxDb
	mysqlTxWrapper.sqlPrintConf = ndbw.sqlPrintConf
	mysqlTxWrapper.conf = ndbw.conf
	mysqlTxWrapper.bgnTx = true
	mysqlTxWrapper.sqlxTxContext = txCtx

	sqlTx, err := ndbw.sqlxDb.BeginTxx(mysqlTxWrapper.sqlxTxContext, nil)
	if err == nil {
		mysqlTxWrapper.txState = txActive
		mysqlTxWrapper.sqlxTxContextCancelFunc = cancel
		mysqlTxWrapper.sqlxTx = sqlTx
		mysqlTxWrapper.txMutx = &sync.Mutex{}
		return mysqlTxWrapper, nil
	} else {
		mysqlTxWrapper.txState = txInActive
		cancel() // 立即释放
		return nil, err
	}
}

// 错误回滚使用  defer txr.NdbTxCommit(recover())
// 不捕获错误使用   txr.NdbTxCommit(nil)
func (ndbw *NPgWrapper) NdbTxCommit(recoveResult any) error {
	ndbw.txMutx.Lock()
	defer ndbw.txMutx.Unlock()
	defer ndbw.sqlxTxContextCancelFunc()
	// 执行提交的时候检查是否有异常， 如果有异常就直接回滚
	if recoveResult != nil {
		err := recoveResult.(error)
		slog.Error(fmt.Sprintf("提交事务前,捕获到异常【%v】,执行回滚", err))
		// 回滚事务
		ndbw.sqlxTx.Rollback()
		return err
	}
	// 提交时原子检查状态
	if !atomic.CompareAndSwapInt32(&ndbw.txState, txActive, txCommitted) {
		return errors.New("提交事务前,检查事务状态,事务已结束")
	}
	err := ndbw.sqlxTx.Commit()
	if nil != err {
		slog.Error("事务提交时,捕获到异常", "异常原因", err)
	} else {
		slog.Debug("事务提交成功")
	}
	return err
}

func (ndbw *NPgWrapper) NdbTxRollBack(err error) error {
	ndbw.txMutx.Lock()
	defer ndbw.txMutx.Unlock()
	defer ndbw.sqlxTxContextCancelFunc()
	if nil != err {
		slog.Error("异常回滚事务", "原始错误", err)
	}
	// 提交时原子检查状态
	if !atomic.CompareAndSwapInt32(&ndbw.txState, txActive, txRolledBack) {
		slog.Error("回滚事务前,检查事务状态,事务已结束")
		return nerror.NewRunTimeError("回滚事务前,检查事务状态,事务已结束")
	}
	rollbackErr := ndbw.sqlxTx.Rollback()
	if rollbackErr != nil {
		slog.Error("事务回滚失败", "原始错误", err, "回滚失败错误", rollbackErr)
		return rollbackErr
	}
	return nil
}

// Gauss驱动不支持?参数，需要将?参数全部替换为$1的格式
func (ndbw *NPgWrapper) SqlFmtSqlStr2Pg(sqlStr string) string {
	splTexts := []string{}
	argsRange := sqlext.SqlParamArgsRegexp.FindAllStringIndex(sqlStr, -1)
	if len(argsRange) > 0 {
		splTexts = append(splTexts, sqlStr[0:argsRange[0][0]])

		for idx := 1; idx < len(argsRange); idx++ {
			splTexts = append(splTexts, sqlStr[argsRange[idx-1][1]:argsRange[idx][0]])
		}
		splTexts = append(splTexts, sqlStr[argsRange[len(argsRange)-1][1]:])
		sqlStr = splTexts[0]

		for idx := range len(splTexts) - 1 {
			sqlStr += fmt.Sprintf("$%d", idx+1) + splTexts[idx+1]
		}
	}
	return sqlStr
}
