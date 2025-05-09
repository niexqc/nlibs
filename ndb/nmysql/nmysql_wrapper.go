package nmysql

import (
	"fmt"
	"log/slog"
	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nyaml"
)

type NMysqlWrapper struct {
	sqlxDb *sqlx.DB
	conf   *nyaml.YamlConfDb
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
	return &NMysqlWrapper{sqlxDb: db, conf: conf}
}

func (ndbw *NMysqlWrapper) SelectDyObj(sqlStr string, args ...any) (dyObj any, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.ColumnTypes()
	if nil != err {
		return nil, err
	}
	// 创建动态Struct
	dyStructType := createDyStruct(cols)
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
		return instance, err
		// return instance, err
	} else {
		return nil, nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectDyObjList(sqlStr string, args ...any) (objValList []any, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.ColumnTypes()
	if nil != err {
		return nil, err
	}
	// 创建动态Struct
	dyStructType := createDyStruct(cols)

	results := []any{}
	for rows.Next() {
		// 创建动态Struct的实例
		instance := reflect.New(dyStructType).Interface()
		// 对动态Struct的实例赋值
		err := rows.StructScan(instance)
		if nil != err {
			return nil, err
		}
		results = append(results, instance)
	}
	return results, err
}

func (ndbw *NMysqlWrapper) SelectOne(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if nil != err {
		return err
	}
	if len(cols) != 1 {
		return nerror.NewRunTimeError("查询结果包含多个列")
	}
	if rows.Next() {
		if err := rows.Scan(dest); nil != err {
			return err
		}
		if rows.Next() {
			dest = nil
			return nerror.NewRunTimeError("查询结果中包含多个值")
		}
		return err
	} else {
		return nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectObj(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.StructScan(dest); nil != err {
			return err
		}
		if rows.Next() {
			dest = nil
			return nerror.NewRunTimeError("查询结果中包含多个值")
		}
		return err
	} else {
		return nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	return ndbw.sqlxDb.Select(dest, sqlStr, args...)
}

func (ndbw *NMysqlWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	r, err := ndbw.sqlxDb.Exec(sqlStr, args...)
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NMysqlWrapper) Insert(sqlStr string, args ...any) (lastInsertId int64, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	r, err := ndbw.sqlxDb.Exec(sqlStr, args...)
	if nil != err {
		return lastInsertId, err
	}
	lastInsertId, _ = r.LastInsertId()
	return lastInsertId, err
}
