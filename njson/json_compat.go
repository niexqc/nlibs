package njson

import (
	"log/slog"

	jsonv2 "encoding/json/v2"
)

// GetBool 按路径读取 bool，失败时打日志并返回 false
func GetBool(node *NwNode, path ...any) bool {
	if node == nil {
		slog.Error("json node is nil")
		return false
	}
	result, err := node.TryGetBool(path...)
	if err != nil {
		slog.Error(err.Error())
		return false
	}
	return result
}

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

func ToObj(str *string, t any) error {
	return jsonv2.Unmarshal([]byte(*str), t)
}

func ToStr(t any) (string, error) {
	bytes, err := Obj2JsonBytes(t)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ToStrOk(t any) string {
	str, _ := ToStr(t)
	return str
}

func ToMap(str *string) (map[string]any, error) {
	jsonMap := make(map[string]any)
	err := jsonv2.Unmarshal([]byte(*str), &jsonMap)
	return jsonMap, err
}

func Map2Obj(data map[string]any, t any) error {
	str, err := ToStr(data)
	if err != nil {
		return err
	}
	return jsonv2.Unmarshal([]byte(str), t)
}
