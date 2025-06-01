package njson

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
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

func (s *NwNode) GetString(key string) (string, error) {
	node := s.Node.Get(key)
	return node.String()
}

func (s *NwNode) GetStringByPath(paths ...any) (string, error) {
	val, err := s.GetByPath(paths...).String()
	return val, err
}

func (s *NwNode) GetInt64(key string) (int64, error) {
	val, err := s.Get(key).Int64()
	return val, err
}

func (s *NwNode) GetInt64ByPath(paths ...any) (int64, error) {
	val, err := s.GetByPath(paths...).Int64()
	return val, err
}

func (s *NwNode) GetFloat64(key string) (float64, error) {
	val, err := s.Get(key).Float64()
	return val, err
}

func (s *NwNode) GetFloat64ByPath(paths ...any) (float64, error) {
	val, err := s.GetByPath(paths...).Float64()
	return val, err
}

func (s *NwNode) GetBool(key string) (bool, error) {
	nd := s.Get(key)
	val, err := nd.Bool()
	return val, err
}

func (s *NwNode) GetBoolByPath(paths ...any) (bool, error) {
	val, err := s.GetByPath(paths...).Bool()
	return val, err
}

func (s *NwNode) ToString() (string, error) {
	return Obj2JsonStr(s)
}
