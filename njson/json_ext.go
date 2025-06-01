package njson

import (
	"encoding/json"
	"log/slog"
	"reflect"

	"github.com/bytedance/sonic"
	"github.com/niexqc/nlibs/nerror"
)

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
	bytes, err := sonic.Marshal(obj)
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
	err := sonic.UnmarshalString(acStr, t)
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
	err := sonic.UnmarshalString(acStr, tarr)
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

// 以下为GO默认JSON转换
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
	jsonBytes, err := json.Marshal(&t)
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
	err := json.Unmarshal([]byte(acStr), t)
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
	err := json.Unmarshal([]byte(acStr), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
