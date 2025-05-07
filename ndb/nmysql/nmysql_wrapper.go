package nmysql

import (
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
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

func (ndbw *NMysqlWrapper) SelectNwNode(sqlStr string, args ...any) (nwNode *njson.NwNode, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		row := map[string]any{}
		err := rows.MapScan(row)
		if nil != err {
			return nil, err
		}
		nwNode = njson.SonicMap2NwNode(row)
		if rows.Next() {
			return nil, nerror.NewRunTimeError("查询结果中包含多个值")
		}
		return nwNode, err
	} else {
		return nil, nerror.NewRunTimeError("未查询到结果")
	}
}

func (ndbw *NMysqlWrapper) SelectNwNodeList(sqlStr string, args ...any) (nodeList []*njson.NwNode, err error) {
	defer sqlext.PrintSql(ndbw.conf, time.Now(), sqlStr, args...)
	rows, err := ndbw.sqlxDb.Queryx(sqlStr, args...)
	if nil != err {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		row := map[string]any{}
		err := rows.MapScan(row)
		if nil != err {
			return nil, err
		}
		nodeList = append(nodeList, njson.SonicMap2NwNode(row))
	}
	return nodeList, nil
}

func (ndbw *NMysqlWrapper) SelectOne(dest any, sqlStr string, args ...any) error {
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
