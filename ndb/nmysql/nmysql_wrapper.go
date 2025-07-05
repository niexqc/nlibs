package nmysql

import (
	"context"
	"crypto/tls"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
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

type NMysqlWrapper struct {
	sqlxDb                  *sqlx.DB
	conf                    *nyaml.YamlConfMysqlDb
	sqlPrintConf            *nyaml.YamlConfSqlPrint
	bgnTx                   bool
	sqlxTx                  *sqlx.Tx
	sqlxTxContext           context.Context
	sqlxTxContextCancelFunc context.CancelFunc
	txState                 int32 //
	txMutx                  *sync.Mutex
}

func NewNMysqlWrapper(conf *nyaml.YamlConfMysqlDb, sqlPrintConf *nyaml.YamlConfSqlPrint) (*NMysqlWrapper, error) {
	//开始连接数据库
	mysqlUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.DbUser, conf.DbPwd, conf.DbHost, conf.DbPort, conf.DbName)
	mysqlUrl = mysqlUrl + "?loc=Local&parseTime=true&charset=utf8mb4"
	if conf.UseSsl {
		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS12, //最低1.2
			InsecureSkipVerify: true,             // 跳过验证（与skip-verify行为一致）
			CipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // 现代浏览器支持
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		}
		mysql.RegisterTLSConfig("nssl", tlsConfig)
		mysqlUrl += "&tls=nssl"
	}
	slog.Debug(mysqlUrl)
	db, err := sqlx.Open("mysql", mysqlUrl)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Second * time.Duration(conf.ConnMaxLifetime))
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetMaxIdleConns(conf.MaxIdleConns)
	return &NMysqlWrapper{sqlxDb: db, conf: conf, sqlPrintConf: sqlPrintConf, bgnTx: false}, nil
}

//	 查询单个字段单个值
//		 sqlStr:=select id from table where id=?
//		 str:=ndb.SelectOne[string](ndbw,sql,id)
func SelectOne[T sqlext.NdbBasicType](ndbw *NMysqlWrapper, sqlStr string, args ...any) (t *T, findOk bool, err error) {
	obj := new(T)
	findOk, err = ndbw.SelectOne(obj, sqlStr, args...)
	return obj, findOk, err
}

//	 查询单行记录返回Struct实例
//		 sqlStr:=select * from table where id=?
//		 user:=ndb.SelectObj[UserDo](ndbw,sql,id)
func SelectObj[T any](ndbw *NMysqlWrapper, sqlStr string, args ...any) (t *T, findOk bool, err error) {
	obj := new(T)
	findOk, err = ndbw.SelectObj(obj, sqlStr, args...)
	return obj, findOk, err
}

// 查询多行记录，支持值和Struct
func SelectList[T any](ndbw *NMysqlWrapper, sqlStr string, args ...any) (tlist []*T, err error) {
	objs := new([]*T)
	err = ndbw.SelectList(objs, sqlStr, args...)
	return *objs, err
}

// SqlLimitStr
// pageNo 页码从1开始
func (ndbw *NMysqlWrapper) SqlLimitStr(pageNo, pageSize int) string {
	result, _ := ndb.SqlFmt(" LIMIT ?,? ", (pageNo-1)*pageSize, pageSize)
	return result
}

func (ndbw *NMysqlWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(sqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(sqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NMysqlWrapper) InsertWithRowsAffected(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(sqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(sqlStr, args...)
	}
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NMysqlWrapper) InsertWithLastId(sqlStr string, args ...any) (lastInsertId int64, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var r sql.Result
	if ndbw.bgnTx {
		r, err = ndbw.sqlxTx.Exec(sqlStr, args...)
	} else {
		r, err = ndbw.sqlxDb.Exec(sqlStr, args...)
	}
	if nil != err {
		return lastInsertId, err
	}
	lastInsertId, _ = r.LastInsertId()
	return lastInsertId, err
}

func (ndbw *NMysqlWrapper) SelectOne(dest any, sqlStr string, args ...any) (findOk bool, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var rows *sqlx.Rows

	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(sqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(sqlStr, args...)
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

func (ndbw *NMysqlWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	if ndbw.bgnTx {
		return ndbw.sqlxTx.Select(dest, sqlStr, args...)
	} else {
		return ndbw.sqlxDb.Select(dest, sqlStr, args...)
	}
}

//	 查询并生成动态对象返回
//		 dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1")
//		 val, err := sqlext.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
func (ndbw *NMysqlWrapper) SelectDyObj(sqlStr string, args ...any) (dyObj *NMysqlDyObj, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var rows *sqlx.Rows
	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(sqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(sqlStr, args...)
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
	dyStructType, fieldsInfo, err := createDyStruct(cols)
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
		return &NMysqlDyObj{Data: instance, DbNameFiledsMap: fieldsInfo}, err
		// return instance, err
	} else {
		return nil, nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectDyObjList(sqlStr string, args ...any) (objValList []*NMysqlDyObj, err error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)

	var rows *sqlx.Rows
	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(sqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(sqlStr, args...)
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
	dyStructType, fieldsInfo, err := createDyStruct(cols)
	if nil != err {
		return nil, err
	}
	objValList = make([]*NMysqlDyObj, 0)
	for rows.Next() {
		// 创建动态Struct的实例
		instance := reflect.New(dyStructType).Interface()
		// 对动态Struct的实例赋值
		err := rows.StructScan(instance)
		if nil != err {
			return nil, err
		}
		objValList = append(objValList, &NMysqlDyObj{Data: instance, DbNameFiledsMap: fieldsInfo})
	}
	return objValList, err
}

func (ndbw *NMysqlWrapper) SelectObj(dest any, sqlStr string, args ...any) (bool, error) {
	defer sqlext.PrintSql(ndbw.sqlPrintConf, time.Now(), sqlStr, args...)
	var rows *sqlx.Rows
	var err error

	if ndbw.bgnTx {
		rows, err = ndbw.sqlxTx.Queryx(sqlStr, args...)
	} else {
		rows, err = ndbw.sqlxDb.Queryx(sqlStr, args...)
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

func (ndbw *NMysqlWrapper) NdbTxBgn(timeoutSecond int) (txWrper *NMysqlWrapper, err error) {
	if timeoutSecond > 60 {
		slog.Warn("事务时长超过60秒,判断下业务")
	}

	txCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecond)*time.Second)
	mysqlTxWrapper := new(NMysqlWrapper)
	mysqlTxWrapper.sqlxDb = ndbw.sqlxDb
	mysqlTxWrapper.conf = ndbw.conf
	mysqlTxWrapper.sqlPrintConf = ndbw.sqlPrintConf
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
func (ndbw *NMysqlWrapper) NdbTxCommit(recoveResult any) error {
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

func (ndbw *NMysqlWrapper) NdbTxRollBack(err error) error {
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
