package sqlext

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"strings"

	"github.com/niexqc/nlibs/ntools"
)

type NullString struct{ sql.NullString }
type NullTime struct{ sql.NullTime }
type NullInt struct{ sql.NullInt32 }
type NullInt64 struct{ sql.NullInt64 }
type NullFloat64 struct{ sql.NullFloat64 }
type NullBool struct{ sql.NullBool }

func NewNullString(str string) NullString {
	if str == "" {
		return NullString{sql.NullString{Valid: false}}
	}
	return NullString{sql.NullString{Valid: true, String: str}}
}

func NewNullTime(valid bool, time time.Time) NullTime {
	if !valid {
		return NullTime{sql.NullTime{Valid: false}}
	}
	return NullTime{sql.NullTime{Valid: true, Time: time}}
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	ns.Valid = valid
	ns.String = str
	return nil
}

func nullVlaleStr(data []byte) (string, bool) {
	valStr := string(data)
	valStr = strings.TrimPrefix(valStr, "\"")
	valStr = strings.TrimSuffix(valStr, "\"")
	if valStr == "" || strings.ToLower(string(valStr)) == "null" {
		return "", true
	}
	return valStr, false
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil // 返回null或直接忽略字段
	}
	return json.Marshal(ntools.Time2Str(nt.Time))
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	nt.Valid = valid
	nt.Time = ntools.TimeStr2Time(str)
	return nil
}

func (ns NullInt) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int32)
}

func (ns *NullInt) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	intv, err := strconv.ParseInt(str, 10, 32)
	if nil != err {
		ns.Valid = false
		return err
	}
	ns.Valid = valid
	ns.Int32 = int32(intv)

	return nil
}

func (ns NullInt64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int64)
}

func (ns *NullInt64) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	intv, err := strconv.ParseInt(str, 10, 32)
	if nil != err {
		ns.Valid = false
		return err
	}
	ns.Valid = valid
	ns.Int64 = intv
	return nil
}

func (ns NullFloat64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Float64)
}

func (ns *NullFloat64) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	intv, err := strconv.ParseFloat(str, 64)
	if nil != err {
		ns.Valid = false
		return err
	}
	ns.Valid = valid
	ns.Float64 = intv
	return nil
}

func (ns NullBool) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Bool)
}

func (ns *NullBool) UnmarshalJSON(data []byte) error {
	str, valid := nullVlaleStr(data)
	intv, err := strconv.ParseBool(str)
	if nil != err {
		ns.Valid = false
		return err
	}
	ns.Valid = valid
	ns.Bool = intv
	return nil
}
