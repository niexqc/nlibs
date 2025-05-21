package nmysql

import (
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
	"github.com/timandy/routine"
)

type NMysqlWrapper struct {
	sqlxDb     *sqlx.DB
	conf       *nyaml.YamlConfDb
	bgnTx      bool
	sqlxTx     *sqlx.Tx
	txDoneChan chan error
}

func NewNMysqlWrapper(conf *nyaml.YamlConfDb) *NMysqlWrapper {
	//开始连接数据库
	mysqlUrl := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.DbUser, conf.DbPwd, conf.DbHost, conf.DbPort, conf.DbName)
	mysqlUrl = mysqlUrl + "?loc=Local&parseTime=true&charset=utf8mb4"
	slog.Debug(mysqlUrl)
	db, err := sqlx.Open("mysql", mysqlUrl)
	if err != nil {
		panic(err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	return &NMysqlWrapper{sqlxDb: db, conf: conf, bgnTx: false}
}

//	 查询并生成动态对象返回
//		 dyObj, err := IDbWrapper.SelectDyObj("SELECT * FROM test01 where id=1")
//		 val, err := sqlext.GetFiledVal[sqlext.NullString](dyObj, dyObj.FiledsInfo["t03_varchar"].StructFieldName)
func (ndbw *NMysqlWrapper) SelectDyObj(sqlStr string, args ...any) (dyObj *sqlext.NdbDyObj, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
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
	dyStructType, fieldsInfo := createDyStruct(cols)
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
		return &sqlext.NdbDyObj{Data: instance, FiledsInfo: fieldsInfo}, err
		// return instance, err
	} else {
		return nil, nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectDyObjList(sqlStr string, args ...any) (objValList []*sqlext.NdbDyObj, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)

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
	dyStructType, fieldsInfo := createDyStruct(cols)
	objValList = make([]*sqlext.NdbDyObj, 0)
	for rows.Next() {
		// 创建动态Struct的实例
		instance := reflect.New(dyStructType).Interface()
		// 对动态Struct的实例赋值
		err := rows.StructScan(instance)
		if nil != err {
			return nil, err
		}
		objValList = append(objValList, &sqlext.NdbDyObj{Data: instance, FiledsInfo: fieldsInfo})
	}
	return objValList, err
}

func (ndbw *NMysqlWrapper) SelectOne(dest any, sqlStr string, args ...any) (findOk bool, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
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

func (ndbw *NMysqlWrapper) SelectObj(dest any, sqlStr string, args ...any) (bool, error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
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

func (ndbw *NMysqlWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	if ndbw.bgnTx {
		return ndbw.sqlxTx.Select(dest, sqlStr, args...)
	} else {
		return ndbw.sqlxDb.Select(dest, sqlStr, args...)
	}
}

func (ndbw *NMysqlWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
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

func (ndbw *NMysqlWrapper) Insert(sqlStr string, args ...any) (lastInsertId int64, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
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

func (ndbw *NMysqlWrapper) NdbTxBgn(timeoutSecond int) (txWrper *NMysqlWrapper, err error) {
	sqlTx, err := ndbw.sqlxDb.Beginx()
	if nil != err {
		return nil, err
	}
	mysqlTxWrapper := new(NMysqlWrapper)
	mysqlTxWrapper.sqlxDb = ndbw.sqlxDb
	mysqlTxWrapper.conf = ndbw.conf
	mysqlTxWrapper.bgnTx = true
	mysqlTxWrapper.sqlxTx = sqlTx

	//运行超时检监测的协程
	mysqlTxWrapper.txDoneChan = make(chan error, 1)
	routine.Go(func() {
		select {
		case <-time.After(time.Duration(timeoutSecond) * time.Second):
			slog.Error(fmt.Sprintf("事务执行超时:%ds", timeoutSecond))
			txWrper.sqlxTx.Rollback()
		case err := <-txWrper.txDoneChan:
			if err != nil {
				slog.Error("事务执行时发生错误:" + nerror.GenErrDetail(err))
			} else {
				slog.Info("事务执行并提交完成")
			}
		}
	})
	return mysqlTxWrapper, nil
}

func (ndbw *NMysqlWrapper) NdbTxCommit() error {
	if err := recover(); err != nil {
		slog.Info(fmt.Sprintf("即将提交事务时,捕获到异常【%v】,执行回滚", err))
		ndbw.NdbTxRollBack(err.(error))
		panic(err)
	} else {
		err := ndbw.sqlxTx.Commit()
		if nil != err {
			slog.Info(fmt.Sprintf("执行事务提交时,捕获到异常【%v】,执行回滚", err))
			ndbw.NdbTxRollBack(err)
		} else {
			ndbw.txDoneChan <- nil
		}
		return err
	}
}

func (ndbw *NMysqlWrapper) NdbTxRollBack(err error) {
	if err == nil {
		err = nerror.NewRunTimeError("手动回滚事务,但是没有传入错误")
	}
	ndbw.sqlxTx.Rollback()
	ndbw.txDoneChan <- err
}
