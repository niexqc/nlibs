package ndb

import (
	"github.com/niexqc/nlibs/njson"
)

type INdbWrapper interface {
	SelectNwNode(sqlStr string, args ...any) (nwNode *njson.NwNode, err error)
	SelectNwNodeList(sqlStr string, args ...any) (nodeList []*njson.NwNode, err error)
	SelectOne(dest any, sqlStr string, args ...any) error
	SelectList(dest any, sqlStr string, args ...any) error
	Exec(sqlStr string, args ...any) (rowsAffected int64, err error)
	Insert(sqlStr string, args ...any) (lastInsertId int64, err error)
}
