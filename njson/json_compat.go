package njson

import (
	"log/slog"
)

func GetInt64(node *NwNode, path ...any) int64 {
	if node == nil {
		slog.Error("json node is nil")
		return -1
	}
	result, err := node.TryGetInt64(path...)
	if err != nil {
		slog.Error(err.Error())
		return -1
	}
	return result
}

func GetString(node *NwNode, path ...any) string {
	if node == nil {
		slog.Error("json node is nil")
		return ""
	}
	result, err := node.TryGetString(path...)
	if err != nil {
		slog.Error(err.Error())
		return ""
	}
	return result
}

// ToStrOk 将对象序列化为 JSON 字符串，忽略错误。
func ToStrOk(t any) string {
	return Obj2StrWithPanicError(t)
}
