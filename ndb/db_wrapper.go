package ndb

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
	"github.com/niexqc/nlibs/nyaml"
)

var blankRegexp = regexp.MustCompile(`\s+`)
var argsRegexp = regexp.MustCompile(`\?`)

type NDbWrapper struct {
	sqlxDb *sqlx.DB
	conf   *nyaml.YamlConfDb
}

func SelectOne[T any](ndbw *NDbWrapper, sqlStr string, args ...any) (t *T, err error) {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
	dest := new([]T)
	err = ndbw.SelectList(dest, sqlStr, args...)
	if nil != err {
		return nil, err
	}
	if len(*dest) == 0 {
		return nil, nil
	}
	if len(*dest) != 1 {
		return nil, nerror.NewRunTimeError("查询结果包含多个值")
	}
	return &(*dest)[0], nil
}

func (ndbw *NDbWrapper) SelectNwNode(sqlStr string, args ...any) (nwNode *njson.NwNode, err error) {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
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

func (ndbw *NDbWrapper) SelectNwNodeList(sqlStr string, args ...any) (nodeList []*njson.NwNode, err error) {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
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

func (ndbw *NDbWrapper) SelectList(dest any, sqlStr string, args ...any) error {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
	return ndbw.sqlxDb.Select(dest, sqlStr, args...)
}

func (ndbw *NDbWrapper) Exec(sqlStr string, args ...any) (rowsAffected int64, err error) {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
	r, err := ndbw.sqlxDb.Exec(sqlStr, args...)
	if nil != err {
		return rowsAffected, err
	}
	rowsAffected, _ = r.RowsAffected()
	return rowsAffected, err
}

func (ndbw *NDbWrapper) Insert(sqlStr string, args ...any) (lastInsertId int64, err error) {
	defer ndbw.PrintSql(time.Now(), sqlStr, args...)
	r, err := ndbw.sqlxDb.Exec(sqlStr, args...)
	if nil != err {
		return lastInsertId, err
	}
	lastInsertId, _ = r.LastInsertId()
	return lastInsertId, err
}

func (ndbw *NDbWrapper) PrintSql(start time.Time, sqlStr string, args ...any) {
	if !ndbw.conf.DbSqlLogPrint {
		return
	}
	costTime := time.Now().UnixMilli() - start.UnixMilli()
	//去除换行符
	if ndbw.conf.DbSqlLogCompress {
		sqlStr = string(blankRegexp.ReplaceAllString(sqlStr, " "))
	}
	if len(args) > 0 {
		splTexts := []string{}
		argsRange := argsRegexp.FindAllStringIndex(sqlStr, -1)
		splTexts = append(splTexts, sqlStr[0:argsRange[0][0]])
		for idx := 1; idx < len(argsRange); idx++ {
			splTexts = append(splTexts, sqlStr[argsRange[idx-1][1]:argsRange[idx][0]])
		}
		splTexts = append(splTexts, sqlStr[argsRange[len(argsRange)-1][1]:])
		sqlStr = splTexts[0]
		for idx, v := range args {
			sqlStr += sqlAnyArg(v) + splTexts[idx+1]
		}
	}
	//打印日志
	slog.Log(context.Background(), ntools.SlogLevelStr2Level(ndbw.conf.DbSqlLogLevel), fmt.Sprintf("[%dms] %s", costTime, sqlStr))
}

func sqlAnyArg(arg any) string {
	switch v := arg.(type) {
	case nil:
		return "NULL"
	case bool:
		return fmt.Sprintf("%v", v)
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case int, uint, int8, int16, int32, int64, uint8, uint16, uint32, uint64, float32, float64:
		return fmt.Sprintf("%v", v)
	case time.Time:
		return fmt.Sprintf("'%v'", ntools.Time2Str(v))
	default:
		// 反射检查底层类型（例如处理自定义类型）
		rv := reflect.ValueOf(arg)
		switch rv.Kind() {
		case reflect.String:
			return fmt.Sprintf("'%s'", rv.String())
		case reflect.Int, reflect.Int64, reflect.Float64:
			return fmt.Sprintf("%v", rv.Interface())
		default:
			return fmt.Sprintf("'%v'", arg)
		}
	}
}
