package njson

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/niexqc/nlibs/nerror"
)

type nwNode struct {
	*ast.Node
}

func (s *nwNode) GetString(key string) string {
	val, err := s.Node.Get(key).String()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetString", err))
	}
	return val
}

func (s *nwNode) GetStringByPath(paths []string) string {
	val, err := s.GetByPath(paths).String()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetStringByPath", err))
	}
	return val
}

func (s *nwNode) GetInt64(key string) int64 {
	val, err := s.Get(key).Int64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetInt64", err))
	}
	return val
}

func (s *nwNode) GetInt64ByPath(paths []string) int64 {
	val, err := s.GetByPath(paths).Int64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetInt64ByPath", err))
	}
	return val
}

func (s *nwNode) GetFloat64(key string) float64 {
	val, err := s.Get(key).Float64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetFloat64", err))
	}
	return val
}

func (s *nwNode) GetFloat64ByPath(paths []string) float64 {
	val, err := s.GetByPath(paths).Float64()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetFloat64ByPath", err))
	}
	return val
}

func (s *nwNode) GetBool(key string) bool {
	val, err := s.Get(key).Bool()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetBool", err))
	}
	return val
}

func (s *nwNode) GetBoolByPath(paths []string) bool {
	val, err := s.GetByPath(paths).Bool()
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("GetBoolByPath", err))
	}
	return val
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

func SonicStr2Obj(str *string, t any) {
	err := sonic.UnmarshalString(*str, &t)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicStr2Obj", err))
	}
}

func SonicStr2nwNode(str string) *nwNode {
	root, err := sonic.GetFromString(str)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("SonicStr2nwNode失败", err))
	}
	return &nwNode{&root}
}

func SonicMap2nwNode(data map[string]any) *nwNode {
	return SonicStr2nwNode(SonicObj2Str(data))
}

func SonicMap2Obj(data map[string]any, t any) {
	str := SonicObj2Str(data)
	SonicStr2Obj(&str, &t)
}
