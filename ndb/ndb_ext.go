package ndb

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/niexqc/nlibs/ndb/nmysql"
	"github.com/niexqc/nlibs/nyaml"
)

func NewNMysqlWrapper(conf *nyaml.YamlConfDb) *nmysql.NMysqlWrapper {
	return nmysql.NewNMysqlWrapper(conf)
}

func SelectOne[T any](ndbw INdbWrapper, sqlStr string, args ...any) (t *T, err error) {
	dest := new(T)
	err = ndbw.SelectOne(dest, sqlStr, args...)
	if nil != err {
		return nil, err
	}
	return dest, nil
}

func SelectList[T any](ndbw INdbWrapper, sqlStr string, args ...any) (t []*T, err error) {
	dest := new([]*T)
	err = ndbw.SelectList(dest, sqlStr, args...)
	if nil != err {
		return nil, err
	}
	return *dest, nil
}
