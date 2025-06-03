package njson

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/niexqc/nlibs/nerror"
)

type NwNode struct {
	*ast.Node
}

func NewNwNodeByJsonStr(str string) (*NwNode, error) {
	root, err := sonic.GetFromString(str)
	return &NwNode{&root}, err
}

func NewNwNodeByMap(data map[string]any) (*NwNode, error) {
	jsonStr, err := Obj2JsonStr(data)
	if nil != err {
		return nil, err
	}
	return NewNwNodeByJsonStr(jsonStr)
}

func (s *NwNode) GetStringByPath(paths ...any) string {
	val, err := s.GetByPath(paths...).String()
	if nil != err {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return val
}

func (s *NwNode) GetInt64ByPath(paths ...any) int64 {
	val, err := s.GetByPath(paths...).Int64()
	if nil != err {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return val
}

func (s *NwNode) GetFloat64ByPath(paths ...any) float64 {
	val, err := s.GetByPath(paths...).Float64()
	if nil != err {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return val
}

func (s *NwNode) GetBoolByPath(paths ...any) bool {
	val, err := s.GetByPath(paths...).Bool()
	if nil != err {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return val
}

func (s *NwNode) ToString() (string, error) {
	return Obj2JsonStr(s)
}
