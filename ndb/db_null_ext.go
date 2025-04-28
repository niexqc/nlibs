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
