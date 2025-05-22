package njson

import (
	"log/slog"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/niexqc/nlibs/nerror"
)

type NwNode struct {
	*ast.Node
}

func (s *NwNode) GetString(key string) string {
	node := s.Node.Get(key)
	val, err := node.String()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetString", err))
	}
	return val
}

func (s *NwNode) GetStringByPath(paths []string) string {
	val, err := s.GetByPath(paths).String()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetStringByPath", err))
	}
	return val
}

func (s *NwNode) GetInt64(key string) int64 {
	val, err := s.Get(key).Int64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetInt64", err))
	}
	return val
}

func (s *NwNode) GetInt64ByPath(paths []string) int64 {
	val, err := s.GetByPath(paths).Int64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetInt64ByPath", err))
	}
	return val
}

func (s *NwNode) GetFloat64(key string) float64 {
	val, err := s.Get(key).Float64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetFloat64", err))
	}
	return val
}

func (s *NwNode) GetFloat64ByPath(paths []string) float64 {
	val, err := s.GetByPath(paths).Float64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetFloat64ByPath", err))
	}
	return val
}

func (s *NwNode) GetBool(key string) bool {
	nd := s.Get(key)
	val, err := nd.Bool()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetBool", err))
	}
	return val
}

func (s *NwNode) GetBoolByPath(paths []string) bool {
	val, err := s.GetByPath(paths).Bool()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetBoolByPath", err))
	}
	return val
}

func (s *NwNode) ToString() string {
	return SonicObj2Str(s)
}

func SonicObj2Str(obj any) string {
	bytes := SonicObj2Bytes(obj)
	return string(bytes)
}

func SonicObj2Bytes(obj any) []byte {
	bytes, err := sonic.Marshal(obj)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicObj2Bytes", err))
	}
	return bytes
}

func SonicNode2Obj[T any](node *ast.Node) *T {
	bytes, err := node.MarshalJSON()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicNode2Obj-node2Bytes", err))
	}
	t := new(T)
	err = sonic.Unmarshal(bytes, t)
	if err != nil {
		slog.Error(string(bytes))
		panic(nerror.NewRunTimeErrorWithError("SonicNode2Obj-bytes2Obj", err))
	}
	return t
}

func SonicStr2Obj(str *string, t any) {
	err := sonic.UnmarshalString(*str, &t)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicStr2Obj", err))
	}
}

func SonicStr2NwNode(str string) *NwNode {
	root, err := sonic.GetFromString(str)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicStr2NwNode失败", err))
	}
	return &NwNode{&root}
}

func SonicMap2NwNode(data map[string]any) *NwNode {
	return SonicStr2NwNode(SonicObj2Str(data))
}

func SonicMap2Obj(data map[string]any, t any) {
	str := SonicObj2Str(data)
	SonicStr2Obj(&str, &t)
}
