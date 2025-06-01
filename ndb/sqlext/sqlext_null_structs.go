package sqlext

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"strings"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
	"github.com/shopspring/decimal"
)

type NullString struct{ sql.NullString }
type NullTime struct{ sql.NullTime }
type NullInt struct{ sql.NullInt32 }
type NullInt64 struct{ sql.NullInt64 }
type NullFloat64 struct{ sql.NullFloat64 }
type NullBool struct{ sql.NullBool }

// decimal.NullDecimal

func NewNullDecimal(valid bool, str string) decimal.NullDecimal {
	if !valid {
		return decimal.NullDecimal{Valid: false, Decimal: decimal.Zero}
	}
	d, err := decimal.NewFromString(str)
	if nil != err {
		slog.Warn("字符串不能转为decimal", "str", str, "err", err)
		panic(nerror.NewRunTimeErrorFmt("字符串[%s]不能转为decimal", str))
	}
	return decimal.NewNullDecimal(d)
}

func NewNullString(valid bool, str string) NullString {
	if !valid {
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

func NewNullInt(valid bool, val int) NullInt {
	if !valid {
		return NullInt{sql.NullInt32{Valid: false}}
	}
	return NullInt{sql.NullInt32{Valid: true, Int32: int32(val)}}
}

func NewNullFloat64(valid bool, val float64) NullFloat64 {
	if !valid {
		return NullFloat64{sql.NullFloat64{Valid: false}}
	}
	return NullFloat64{sql.NullFloat64{Valid: true, Float64: val}}
}

func NewNullInt64(valid bool, val int) NullInt64 {
	if !valid {
		return NullInt64{sql.NullInt64{Valid: false}}
	}
	return NullInt64{sql.NullInt64{Valid: true, Int64: int64(val)}}
}

func NewNullBool(valid bool, val bool) NullBool {
	if !valid {
		return NullBool{sql.NullBool{Valid: false}}
	}
	return NullBool{sql.NullBool{Valid: true, Bool: val}}
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(string(valStr)) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	ns.String = valStr
	return nil
}

func valueStrTrim(data []byte) string {
	valStr := string(data)
	valStr = strings.TrimPrefix(valStr, "\"")
	valStr = strings.TrimSuffix(valStr, "\"")
	return valStr
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil // 返回null或直接忽略字段
	}
	return json.Marshal(ntools.Time2Str(nt.Time))
}

func (ns *NullTime) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(valStr) == "null" || valStr == "" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	str2Time, err := ntools.TimeStr2Time(valStr)
	if nil != err {
		return err
	}
	ns.Time = str2Time
	return nil
}

func (ns NullInt) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int32)
}

func (ns *NullInt) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(valStr) == "null" || valStr == "" {
		ns.Valid = false
		return nil
	}
	cv, err := strconv.ParseInt(valStr, 10, 32)
	if nil != err {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	ns.Int32 = int32(cv)
	return nil
}

func (ns NullInt64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int64)
}

func (ns *NullInt64) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(valStr) == "null" || valStr == "" {
		ns.Valid = false
		return nil
	}
	cv, err := strconv.ParseInt(valStr, 10, 32)
	if nil != err {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	ns.Int64 = cv
	return nil
}

func (ns NullFloat64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Float64)
}

func (ns *NullFloat64) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(valStr) == "null" || valStr == "" {
		ns.Valid = false
		return nil
	}
	cv, err := strconv.ParseFloat(valStr, 64)
	if nil != err {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	ns.Float64 = cv
	return nil
}

func (ns NullBool) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Bool)
}

func (ns *NullBool) UnmarshalJSON(data []byte) error {
	if len(data) <= 0 {
		ns.Valid = false
		return nil
	}
	valStr := valueStrTrim(data)
	if strings.ToLower(valStr) == "null" || valStr == "" {
		ns.Valid = false
		return nil
	}
	cv, err := strconv.ParseBool(valStr)
	if nil != err {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	ns.Bool = cv
	return nil
}
