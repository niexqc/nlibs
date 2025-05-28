package ngin

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type ReqVoInt int
type ReqVoInt64 int64
type ReqVoBool bool
type ReqVoFloat64 float64

func (i *ReqVoInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return fmt.Errorf("错误的int字符串: %s", s)
		}
		*i = ReqVoInt(val)
		return nil
	}
	// 若前端传递的是数字类型，直接解析
	var num int
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*i = ReqVoInt(num)
	return nil
}

func (i *ReqVoInt) Value() int {
	return int(*i)
}

func (i *ReqVoInt64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("错误的int64字符串: %s", s)
		}
		*i = ReqVoInt64(val)
		return nil
	}
	// 若前端传递的是数字类型，直接解析
	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*i = ReqVoInt64(num)
	return nil
}

func (i *ReqVoInt64) Value() int64 {
	return int64(*i)
}

func (i *ReqVoBool) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("错误的bool字符串: %s", s)
		}
		*i = ReqVoBool(val)
		return nil
	}
	// 若前端传递的是数字类型，直接解析
	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*i = ReqVoBool(b)
	return nil
}

func (i *ReqVoBool) Value() bool {
	return bool(*i)
}

func (i *ReqVoFloat64) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("错误的float字符串: %s", s)
		}
		*i = ReqVoFloat64(val)
		return nil
	}
	// 若前端传递的是数字类型，直接解析
	var b float64
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	*i = ReqVoFloat64(b)
	return nil
}

func (i *ReqVoFloat64) Value() float64 {
	return float64(*i)
}
