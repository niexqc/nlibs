package njson

import (
	"encoding/json"
)

// ToStrOk 对象转换为Str
// t 任意对象，注意取地址传入
func ToStrOk(t any) string {
	result, _ := ToStr(t)
	return result
}

// ToStr 对象转换为Str
//
//	t 任意对象，注意取地址传入
func ToStr(t any) (string, error) {
	jsonBytes, err := ToJSONBytes(t)
	if err != nil {
		return "", err
	}
	return string(*jsonBytes), nil
}

//	 ToJSONBytes 对象转换为[]byte
//
//		t 任意对象，注意取地址传入
func ToJSONBytes(t any) (*[]byte, error) {
	jsonBytes, err := json.Marshal(&t)
	if err != nil {
		return nil, err
	}
	return &jsonBytes, nil
}

// ToObj Str转换为对象
//
//	str 字符串的引用
//	t  需要转换到的对象的引用
func ToObj(str *string, t any) error {
	err := json.Unmarshal([]byte(*str), &t)
	if err != nil {
		return err
	}
	return nil
}

// ToMap 对象转换为map[string]interface{}
// t 任意对象，注意取地址传入
func ToMap(str *string) (r map[string]any, err error) {
	jsonMap := make(map[string]any, 0)
	err = json.Unmarshal([]byte(*str), &jsonMap)
	return jsonMap, err
}
