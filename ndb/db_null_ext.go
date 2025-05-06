package ndb

import (
	"database/sql"
	"encoding/json"

	"strings"

	"github.com/niexqc/nlibs/ntools"
)

type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.String)
}

type NullTime struct {
	sql.NullTime
}

func (nt NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return []byte("null"), nil // 返回null或直接忽略字段
	}
	return json.Marshal(ntools.Time2Str(nt.Time))
}

func (nt *NullTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		nt.Valid = false
		return nil
	}
	timeStr := string(data)
	timeStr = strings.ReplaceAll(timeStr, "\"", "")
	nt.Time = ntools.TimeStr2Time(timeStr)
	nt.Valid = true
	return nil
}

type NullInt struct {
	sql.NullInt32
}

func (ns NullInt) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int32)
}

func (ns *NullInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.Int32)
}

type NullInt64 struct {
	sql.NullInt64
}

func (ns NullInt64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Int64)
}

func (ns *NullInt64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.Int64)
}

type NullFloat64 struct {
	sql.NullFloat64
}

func (ns NullFloat64) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Float64)
}

func (ns *NullFloat64) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.Float64)
}

type NullBool struct {
	sql.NullBool
}

func (ns NullBool) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil // 输出为 JSON 的 null
	}
	return json.Marshal(ns.Bool)
}

func (ns *NullBool) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}
	ns.Valid = true
	return json.Unmarshal(data, &ns.Bool)
}
