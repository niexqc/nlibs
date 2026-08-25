package njson

import (
	"bytes"
	"errors"
	"log/slog"
	"math"
	"math/big"
	"reflect"
	"strconv"

	jsonv2 "encoding/json/v2"
	"encoding/json/jsontext"
	"github.com/niexqc/nlibs/nerror"
)

// ---------- 基础 JSON 序列化/反序列化 (encoding/json/v2) ----------

func Obj2StrWithPanicError(obj any) string {
	bytes, err := Obj2JsonBytes(obj)
	if nil != err {
		panic(nerror.NewRunTimeErrorFmt("%s不能转换为JSON字符串", reflect.TypeOf(obj).Name()))
	}
	return string(bytes)
}

func Obj2JsonStr(obj any) (string, error) {
	bytes, err := Obj2JsonBytes(obj)
	return string(bytes), err
}

func Obj2JsonBytes(obj any) ([]byte, error) {
	bytes, err := jsonv2.Marshal(obj)
	return bytes, err
}

func Str2Obj[T any, STR string | *string](str STR) (*T, error) {
	t := new(T)
	acStr := ""
	if reflect.TypeOf(str) == reflect.TypeOf("") {
		acStr = reflect.ValueOf(str).String()
	} else {
		acStr = reflect.ValueOf(str).Elem().String()
	}
	err := jsonv2.Unmarshal([]byte(acStr), t)
	return t, err
}

func Bytes2Obj[T any](bytes []byte) (*T, error) {
	t := new(T)
	err := jsonv2.Unmarshal(bytes, t)
	return t, err
}

func Str2ObjArr[T any, STR string | *string](str STR) (*[]T, error) {
	tarr := new([]T)
	acStr := ""
	if reflect.TypeOf(str) == reflect.TypeOf("") {
		acStr = reflect.ValueOf(str).String()
	} else {
		acStr = reflect.ValueOf(str).Elem().String()
	}
	err := jsonv2.Unmarshal([]byte(acStr), tarr)
	return tarr, err
}

func Str2ObjWithPanicError[T any, STR string | *string](str STR) *T {
	t, err := Str2Obj[T](str)
	if nil != err {
		slog.Warn("JSON转对象失败", "jsonStr", str, "err", err)
		panic(nerror.NewRunTimeErrorWithError("JSON转对象失败", err))
	}
	return t
}

func Str2ObjArrWithPanicError[T any, STR string | *string](str STR) *[]T {
	t, err := Str2ObjArr[T](str)
	if nil != err {
		slog.Warn("JSON转对象数组失败", "jsonStr", str, "err", err)
		panic(nerror.NewRunTimeErrorWithError("JSON转对象数组失败", err))
	}
	return t
}

// 以下为GO默认JSON转换 (由 encoding/json/v2 实现)
func ObjToJsonStrByGoJson(t any) (string, error) {
	jsonBytes, err := ObjToJSONBytesByGoJson(t)
	if err != nil {
		return "", err
	}
	return string(*jsonBytes), nil
}

func ObjToJsonStrByGoJsonWithPanicError(t any) string {
	result, err := ObjToJsonStrByGoJson(t)
	if nil != err {
		slog.Warn("对象JSON失败", "type", reflect.TypeOf(t).Name(), "err", err)
		panic(nerror.NewRunTimeErrorFmt("%s类型转JSON失败", reflect.TypeOf(t).Name()))
	}
	return result
}

func ObjToJSONBytesByGoJson(t any) (*[]byte, error) {
	jsonBytes, err := jsonv2.Marshal(&t)
	if err != nil {
		return nil, err
	}
	return &jsonBytes, nil
}

func Str2ObjByGoJson[T any, STR string | *string](str STR) (*T, error) {
	t := new(T)
	acStr := ""
	if reflect.TypeOf(str) == reflect.TypeOf("") {
		acStr = reflect.ValueOf(str).String()
	} else {
		acStr = reflect.ValueOf(str).Elem().String()
	}
	err := jsonv2.Unmarshal([]byte(acStr), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func Str2ObjArrByGoJson[T any, STR string | *string](str STR) (*[]*T, error) {
	t := new([]*T)
	acStr := ""
	if reflect.TypeOf(str) == reflect.TypeOf("") {
		acStr = reflect.ValueOf(str).String()
	} else {
		acStr = reflect.ValueOf(str).Elem().String()
	}
	err := jsonv2.Unmarshal([]byte(acStr), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// ---------- 按路径取值 (NwNode) ----------
// 下列 NwNode 及其方法原位于 sonic_ext.go（曾基于 github.com/bytedance/sonic/ast 的
// ast.Node），现迁移至标准库 encoding/json/v2 的 jsontext.Value。
// 由于底层保留原始 JSON 文本，按路径读取 int64 不会经 float64 中转，因此大整数
// （如 >2^53 的值）不会丢失精度。

// errNotFound is returned when a path component does not exist or a value is
// not of the expected JSON kind.
var errNotFound = errors.New("njson: path not found or unexpected kind")

// NwNode wraps a raw JSON value (jsontext.Value), retaining the exact textual
// form of JSON numbers so no precision is lost during path-based access.
type NwNode struct {
	value jsontext.Value
}

// NewNwNodeByJsonStr parses str into a NwNode without converting numbers to
// float64, preserving the original number literals.
func NewNwNodeByJsonStr(str string) (*NwNode, error) {
	root, err := newValueFromBytes([]byte(str))
	if err != nil {
		return nil, err
	}
	return &NwNode{value: root}, nil
}

// NewNwNodeByMap marshals data to JSON text then parses into a NwNode.
func NewNwNodeByMap(data map[string]any) (*NwNode, error) {
	jsonStr, err := Obj2JsonStr(data)
	if nil != err {
		return nil, err
	}
	return NewNwNodeByJsonStr(jsonStr)
}

func newValueFromBytes(in []byte) (jsontext.Value, error) {
	dec := jsontext.NewDecoder(bytes.NewReader(in))
	v, err := dec.ReadValue()
	if err != nil {
		return nil, err
	}
	// ReadValue returns a Value backed by the Decoder's internal buffer, which
	// is invalidated by the next decoder call (and by the decoder being pooled).
	// Clone it so the NwNode owns an independent, stable copy.
	v = v.Clone()
	if _, err := dec.ReadToken(); err == nil {
		return nil, errNotFound
	}
	return v, nil
}

// findPath walks the raw JSON value along the given paths (string keys and
// integer indexes) and returns the selected value.
func findPath(root jsontext.Value, paths []any) (jsontext.Value, error) {
	cur := root
	for _, p := range paths {
		switch key := p.(type) {
		case string:
			var found bool
			var next jsontext.Value
			err := walkObject(cur, key, &found, &next)
			if err != nil {
				return nil, err
			}
			if !found {
				return nil, errNotFound
			}
			cur = next
		case int:
			var next jsontext.Value
			err := walkArrayIndex(cur, key, &next)
			if err != nil {
				return nil, err
			}
			cur = next
		default:
			return nil, errNotFound
		}
	}
	return cur, nil
}

func walkObject(v jsontext.Value, name string, found *bool, out *jsontext.Value) error {
	dec := jsontext.NewDecoder(bytes.NewReader(v))
	if dec.PeekKind() != jsontext.KindBeginObject {
		return errNotFound
	}
	if _, err := dec.ReadToken(); err != nil { // consume '{'
		return err
	}
	for dec.PeekKind() != jsontext.KindEndObject {
		nameTok, err := dec.ReadToken()
		if err != nil {
			return err
		}
		key := nameTok.String()
		if key == name {
			mval, err := dec.ReadValue()
			if err != nil {
				return err
			}
			*found = true
			*out = mval
			return nil
		}
		if err := dec.SkipValue(); err != nil {
			return err
		}
	}
	return errNotFound
}

func walkArrayIndex(v jsontext.Value, idx int, out *jsontext.Value) error {
	dec := jsontext.NewDecoder(bytes.NewReader(v))
	if dec.PeekKind() != jsontext.KindBeginArray {
		return errNotFound
	}
	if _, err := dec.ReadToken(); err != nil { // consume '['
		return err
	}
	i := 0
	for dec.PeekKind() != jsontext.KindEndArray {
		if i == idx {
			elem, err := dec.ReadValue()
			if err != nil {
				return err
			}
			*out = elem
			return nil
		}
		if err := dec.SkipValue(); err != nil {
			return err
		}
		i++
	}
	return errNotFound
}

func (s *NwNode) getByPath(paths ...any) (jsontext.Value, error) {
	if s == nil {
		return nil, errNotFound
	}
	return findPath(s.value, paths)
}

// GetStringByPath returns the string value at the given path.
func (s *NwNode) GetStringByPath(paths ...any) string {
	v, err := s.getByPath(paths...)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	if v.Kind() != jsontext.KindString {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", errNotFound))
	}
	unquoted, err := jsontext.AppendUnquote(nil, v)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return string(unquoted)
}

// Int64FromValue returns the integer value parsed from a raw JSON number
// literal, preserving precision beyond float64's exact integer range.
func Int64FromValue(v jsontext.Value) (int64, error) {
	if len(v) == 0 {
		return 0, errNotFound
	}
	if v.Kind() == jsontext.KindString {
		var err error
		v, err = jsontext.AppendUnquote(nil, v)
		if err != nil {
			return 0, err
		}
	}
	if v.Kind() != jsontext.KindNumber {
		return 0, errNotFound
	}
	// Fast path: try plain int64 literal.
	if i, err := strconv.ParseInt(string(v), 10, 64); err == nil {
		return i, nil
	}
	// Parse arbitrary precision (handles exponents like 1e6 and big literals).
	bf := new(big.Float).SetPrec(256)
	if _, _, err := bf.Parse(string(v), 10); err != nil {
		return 0, err
	}
	if !bf.IsInt() {
		return 0, errNotFound
	}
	bi, acc := bf.Int(nil)
	if acc != big.Exact {
		return 0, errNotFound
	}
	if !bi.IsInt64() {
		return 0, errNotFound
	}
	return bi.Int64(), nil
}

// GetInt64ByPath returns the int64 value at the given path. Numbers are parsed
// from their exact literal representation, avoiding float64 rounding.
func (s *NwNode) GetInt64ByPath(paths ...any) int64 {
	v, err := s.getByPath(paths...)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	val, err := Int64FromValue(v)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return val
}

// Float64FromValue returns the float64 value parsed from a raw JSON number.
func Float64FromValue(v jsontext.Value) (float64, error) {
	if len(v) == 0 {
		return 0, errNotFound
	}
	if v.Kind() == jsontext.KindString {
		var err error
		v, err = jsontext.AppendUnquote(nil, v)
		if err != nil {
			return 0, err
		}
	}
	if v.Kind() != jsontext.KindNumber {
		return 0, errNotFound
	}
	f, err := strconv.ParseFloat(string(v), 64)
	if err != nil && !isUnderflowRangeErr(err) {
		return 0, err
	}
	return f, nil
}

// isUnderflowRangeErr reports whether err is a strconv.ErrRange caused by
// underflow rather than overflow. Underflow yields a representable (subnormal)
// value; only overflow should be treated as a parse failure.
func isUnderflowRangeErr(err error) bool {
	ne, ok := err.(*strconv.NumError)
	if !ok || ne.Err != strconv.ErrRange {
		return false
	}
	// 对同一个 Num 重新解析时，下溢仍会返回 ErrRange，因此不能因为 ferr != nil 就返回 false。
	// 通过返回值的量级区分：溢出得到 ±Inf，下溢得到 0 或最小次正规数（有限值）。
	f, _ := strconv.ParseFloat(ne.Num, 64)
	return !math.IsInf(f, 0)
}

// GetFloat64ByPath returns the float64 value at the given path.
func (s *NwNode) GetFloat64ByPath(paths ...any) float64 {
	v, err := s.getByPath(paths...)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	f, err := Float64FromValue(v)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	return f
}

// GetNumberByPath returns the JSON number as jsontext.Value (the raw JSON text
// literal), avoiding any early conversion for callers who need exact text.
func (s *NwNode) GetNumberByPath(paths ...any) jsontext.Value {
	v, err := s.getByPath(paths...)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	if v.Kind() == jsontext.KindString {
		var err error
		v, err = jsontext.AppendUnquote(nil, v)
		if err != nil {
			panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
		}
	}
	if v.Kind() != jsontext.KindNumber {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", errNotFound))
	}
	return v
}

// GetBoolByPath returns the bool value at the given path.
func (s *NwNode) GetBoolByPath(paths ...any) bool {
	v, err := s.getByPath(paths...)
	if err != nil {
		panic(nerror.NewRunTimeErrorWithError("请检查Json", err))
	}
	switch v.Kind() {
	case jsontext.KindTrue:
		return true
	case jsontext.KindFalse:
		return false
	default:
		panic(nerror.NewRunTimeErrorWithError("请检查Json", errNotFound))
	}
}

// ToString re-encodes the underlying value as a JSON string.
func (s *NwNode) ToString() (string, error) {
	if s == nil || len(s.value) == 0 {
		return "null", nil
	}
	return string(s.value), nil
}

// ---------- 子节点导航与遍历（替代 sonic ast.Node） ----------

// GetNodeByPath 按路径取子节点，返回独立副本。
func (s *NwNode) GetNodeByPath(paths ...any) (*NwNode, error) {
	v, err := s.getByPath(paths...)
	if err != nil {
		return nil, err
	}
	return &NwNode{value: v.Clone()}, nil
}

// GetByPath 与 GetNodeByPath 相同，便于从 sonic 迁移。
func (s *NwNode) GetByPath(paths ...any) (*NwNode, error) {
	return s.GetNodeByPath(paths...)
}

// Get 取对象的单个字段子节点。
func (s *NwNode) Get(key string) (*NwNode, error) {
	return s.GetNodeByPath(key)
}

// HasKey 判断当前对象是否包含指定字段。
func (s *NwNode) HasKey(key string) bool {
	_, err := s.getByPath(key)
	return err == nil
}

// String 读取当前节点的字符串值（非路径）。
func (s *NwNode) String() (string, error) {
	if s == nil {
		return "", errNotFound
	}
	v := s.value
	if v.Kind() == jsontext.KindString {
		unquoted, err := jsontext.AppendUnquote(nil, v)
		if err != nil {
			return "", err
		}
		return string(unquoted), nil
	}
	if v.Kind() == jsontext.KindNumber {
		return string(v), nil
	}
	return "", errNotFound
}

// Int64 读取当前节点的 int64 值。
func (s *NwNode) Int64() (int64, error) {
	if s == nil {
		return 0, errNotFound
	}
	return Int64FromValue(s.value)
}

// Float64 读取当前节点的 float64 值。
func (s *NwNode) Float64() (float64, error) {
	if s == nil {
		return 0, errNotFound
	}
	return Float64FromValue(s.value)
}

// Bool 读取当前节点的 bool 值。
func (s *NwNode) Bool() (bool, error) {
	if s == nil {
		return false, errNotFound
	}
	switch s.value.Kind() {
	case jsontext.KindTrue:
		return true, nil
	case jsontext.KindFalse:
		return false, nil
	default:
		return false, errNotFound
	}
}

// ArrayNodes 将当前数组节点展开为子节点列表。
func (s *NwNode) ArrayNodes() ([]*NwNode, error) {
	if s == nil {
		return nil, errNotFound
	}
	dec := jsontext.NewDecoder(bytes.NewReader(s.value))
	if dec.PeekKind() != jsontext.KindBeginArray {
		return nil, errNotFound
	}
	if _, err := dec.ReadToken(); err != nil {
		return nil, err
	}
	var nodes []*NwNode
	for dec.PeekKind() != jsontext.KindEndArray {
		elem, err := dec.ReadValue()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, &NwNode{value: elem.Clone()})
	}
	return nodes, nil
}

// ArrayUseNode 与 ArrayNodes 相同，便于从 sonic 迁移。
func (s *NwNode) ArrayUseNode() ([]*NwNode, error) {
	return s.ArrayNodes()
}

// ArrayUseNodeByPath 按路径定位数组后展开为子节点列表。
func (s *NwNode) ArrayUseNodeByPath(paths ...any) ([]*NwNode, error) {
	n, err := s.GetNodeByPath(paths...)
	if err != nil {
		return nil, err
	}
	return n.ArrayUseNode()
}

// MapNodes 将当前对象节点展开为 map[key]*NwNode。
func (s *NwNode) MapNodes() (map[string]*NwNode, error) {
	if s == nil {
		return nil, errNotFound
	}
	dec := jsontext.NewDecoder(bytes.NewReader(s.value))
	if dec.PeekKind() != jsontext.KindBeginObject {
		return nil, errNotFound
	}
	if _, err := dec.ReadToken(); err != nil {
		return nil, err
	}
	result := make(map[string]*NwNode)
	for dec.PeekKind() != jsontext.KindEndObject {
		nameTok, err := dec.ReadToken()
		if err != nil {
			return nil, err
		}
		key := nameTok.String()
		mval, err := dec.ReadValue()
		if err != nil {
			return nil, err
		}
		result[key] = &NwNode{value: mval.Clone()}
	}
	return result, nil
}

// MapUseNode 与 MapNodes 相同，便于从 sonic 迁移。
func (s *NwNode) MapUseNode() (map[string]*NwNode, error) {
	return s.MapNodes()
}

// TryGetString 按路径取 string，失败返回 error（不 panic）。
func (s *NwNode) TryGetString(paths ...any) (string, error) {
	v, err := s.getByPath(paths...)
	if err != nil {
		return "", err
	}
	return (&NwNode{value: v}).String()
}

// TryGetInt64 按路径取 int64，失败返回 error（不 panic）。
func (s *NwNode) TryGetInt64(paths ...any) (int64, error) {
	v, err := s.getByPath(paths...)
	if err != nil {
		return 0, err
	}
	return Int64FromValue(v)
}

// TryGetFloat64 按路径取 float64，失败返回 error（不 panic）。
func (s *NwNode) TryGetFloat64(paths ...any) (float64, error) {
	v, err := s.getByPath(paths...)
	if err != nil {
		return 0, err
	}
	return Float64FromValue(v)
}

// TryGetBool 按路径取 bool，失败返回 error（不 panic）。
func (s *NwNode) TryGetBool(paths ...any) (bool, error) {
	v, err := s.getByPath(paths...)
	if err != nil {
		return false, err
	}
	switch v.Kind() {
	case jsontext.KindTrue:
		return true, nil
	case jsontext.KindFalse:
		return false, nil
	default:
		return false, errNotFound
	}
}
