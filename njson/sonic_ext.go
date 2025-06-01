package njson

import (
	"log/slog"
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/niexqc/nlibs/nerror"
)

type NwNode struct {
	*ast.Node
}

func (s *NwNode) GetString(key string) (string, error) {
	node := s.Node.Get(key)
	return node.String()
}

func (s *NwNode) GetStringByPath(paths []string) (string, error) {
	val, err := s.GetByPath(paths).String()
	return val, err
}

func (s *NwNode) GetInt64(key string) (int64, error) {
	val, err := s.Get(key).Int64()
	return val, err
}

func (s *NwNode) GetInt64ByPath(paths []string) (int64, error) {
	val, err := s.GetByPath(paths).Int64()
	return val, err
}

func (s *NwNode) GetFloat64(key string) (float64, error) {
	val, err := s.Get(key).Float64()
	return val, err
}

func (s *NwNode) GetFloat64ByPath(paths []string) (float64, error) {
	val, err := s.GetByPath(paths).Float64()
	return val, err
}

func (s *NwNode) GetBool(key string) (bool, error) {
	nd := s.Get(key)
	val, err := nd.Bool()
	return val, err
}

func (s *NwNode) GetBoolByPath(paths []string) (bool, error) {
	val, err := s.GetByPath(paths).Bool()
	return val, err
}

func (s *NwNode) ToString() (string, error) {
	return SonicObj2Str(s)
}

func SonicObj2Str(obj any) (string, error) {
	bytes, err := SonicObj2Bytes(obj)
	return string(bytes), err
}

func SonicObj2Bytes(obj any) ([]byte, error) {
	bytes, err := sonic.Marshal(obj)
	return bytes, err
}

func SonicObj2StrWithPanicError(obj any) string {
	bytes, err := SonicObj2Bytes(obj)
	if nil != err {
		panic(nerror.NewRunTimeErrorFmt("%s不能转换为JSON字符串", reflect.TypeOf(obj).Name()))
	}
	return string(bytes)
}

func SonicNode2Obj[T any](node *ast.Node) (*T, error) {
	bytes, err := node.MarshalJSON()
	if err != nil {
		return nil, err
	}
	t := new(T)
	err = sonic.Unmarshal(bytes, t)
	if err != nil {
		slog.Error(string(bytes))
		return nil, err
	}
	return t, nil
}

func SonicStr2Obj[T any](str *string) (*T, error) {
	t := new(T)
	err := sonic.UnmarshalString(*str, t)
	return t, err
}

func SonicStr2ObjArr[T any](str *string) (*[]T, error) {
	tarr := new([]T)
	err := sonic.UnmarshalString(*str, tarr)
	return tarr, err
}

func SonicStr2ObjWithPanicError[T any](str *string) *T {
	t, err := SonicStr2Obj[T](str)
	if nil != err {
		slog.Warn("JSON转对象失败", "jsonStr", str, "err", err)
		panic(nerror.NewRunTimeErrorWithError("JSON转对象失败", err))
	}
	return t
}

func SonicStr2ObjArrWithPanicError[T any](str *string) *[]T {
	t, err := SonicStr2ObjArr[T](str)
	if nil != err {
		slog.Warn("JSON转对象数组失败", "jsonStr", str, "err", err)
		panic(nerror.NewRunTimeErrorWithError("JSON转对象数组失败", err))
	}
	return t
}

func SonicStr2NwNode(str string) (*NwNode, error) {
	root, err := sonic.GetFromString(str)
	if err != nil {
		return nil, err
	}
	return &NwNode{&root}, nil
}

func SonicStr2NwNodeWithPanicError(str string) *NwNode {
	root, err := sonic.GetFromString(str)
	if err != nil {
		slog.Warn("JSON转 SonicNode失败", "jsonStr", str, "err", err)
		panic(nerror.NewRunTimeErrorWithError("JSON转 SonicNode失败", err))
	}
	return &NwNode{&root}
}

func SonicMap2NwNode(data map[string]any) (*NwNode, error) {
	jsonStr, err := SonicObj2Str(data)
	if nil != err {
		return nil, err
	}
	return SonicStr2NwNode(jsonStr)
}
