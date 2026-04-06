package ndnen_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

type TestVoDo struct {
	Id          int64              `schm:"ndb_test" tbn:"tb01" db:"id" json:"id" zhdesc:"主键"`
	T02Int      sqlext.NullInt     `schm:"ndb_test" tbn:"tb01" db:"t02_int" json:"t02Int" zhdesc:"NullInt"`
	T03Varchar  sqlext.NullString  `schm:"ndb_test" tbn:"tb01" db:"t03_varchar" json:"t03Varchar" zhdesc:"NullVarchar"`
	T04Text     sqlext.NullString  `schm:"ndb_test" tbn:"tb01" db:"t04_text" json:"t04Text" zhdesc:"NullText"`
	T05Longtext sqlext.NullString  `schm:"ndb_test" tbn:"tb01" db:"t05_longtext" json:"t05Longtext" zhdesc:"NullLongText"`
	T06Decimal  sqlext.NullDecimal `schm:"ndb_test" tbn:"tb01" db:"t06_decimal" json:"t06Decimal" zhdesc:"NullDecimal"`
	T07Float    sqlext.NullFloat64 `schm:"ndb_test" tbn:"tb01" db:"t07_float" json:"t07Float" zhdesc:"NullFloat"`
	T08Double   sqlext.NullFloat64 `schm:"ndb_test" tbn:"tb01" db:"t08_double" json:"t08Double" zhdesc:"NullDouble"`
	T09Datetime sqlext.NullTime    `schm:"ndb_test" tbn:"tb01" db:"t09_datetime" json:"t09Datetime" zhdesc:"NullDateTime"`
	T10Bool     sqlext.NullBool    `schm:"ndb_test" tbn:"tb01" db:"t10_bool" json:"t10Bool" zhdesc:"NullBool"`
}

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}

// formatInsertValue 按 JSON/Go 实际类型转成 SQL 字面量中的字符串，避免 float64 用 %v 出现科学计数法；时间统一为 yyyy-MM-dd HH:mm:ss。
func formatInsertValue(v interface{}) string {
	switch x := v.(type) {
	case time.Time:
		return formatSQLDateTime(x)
	case *time.Time:
		if x == nil {
			return ""
		}
		return formatSQLDateTime(*x)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case bool:
		return strconv.FormatBool(x)
	case string:
		if out, ok := tryFormatTimeString(x); ok {
			return out
		}
		return x
	case json.Number:
		ns := x.String()
		if out, ok := tryFormatTimeString(ns); ok {
			return out
		}
		return ns
	default:
		return fmt.Sprintf("%v", x)
	}
}

const sqlDateTimeLayout = "2006-01-02 15:04:05" // yyyy-MM-dd HH:mm:ss

func formatSQLDateTime(t time.Time) string {
	return t.Format(sqlDateTimeLayout)
}

// tryFormatTimeString 将 RFC3339、常见本地时间串、纯数字 Unix 秒/毫秒解析为 yyyy-MM-dd HH:mm:ss。
func tryFormatTimeString(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return s, false
	}
	zoneLayouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
	}
	for _, layout := range zoneLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return formatSQLDateTime(t), true
		}
	}
	localLayouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05.000000",
		"2006-01-02T15:04:05.000",
		"2006-01-02T15:04:05.000000",
		"2006-01-02",
	}
	for _, layout := range localLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return formatSQLDateTime(t), true
		}
	}
	if len(s) == 10 || len(s) == 13 {
		if onlyDigits(s) {
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return "", false
			}
			var t time.Time
			if len(s) == 13 {
				t = time.UnixMilli(n)
			} else {
				t = time.Unix(n, 0)
			}
			y := t.Year()
			if y >= 1990 && y <= 2100 {
				return formatSQLDateTime(t), true
			}
		}
	}
	return "", false
}

func onlyDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func TestGoJson(t *testing.T) {
	strTime, _ := ntools.TimeStr2TimeByLayout("2025-02-02", "2006-01-02")
	voNullVal := &TestVoDo{}
	jsonNullValStr := `{"id":0,"t02Int":null,"t03Varchar":null,"t04Text":null,"t05Longtext":null,"t06Decimal":null,"t07Float":null,"t08Double":null,"t09Datetime":null,"t10Bool":null}`

	voFullVal := &TestVoDo{
		Id: 1, T02Int: sqlext.NewNullInt(true, 1),
		T03Varchar: sqlext.NewNullString(true, "1"), T04Text: sqlext.NewNullString(true, "1"),
		T05Longtext: sqlext.NewNullString(true, "1"), T06Decimal: sqlext.NewNullDecimal(true, "1.1"),
		T07Float: sqlext.NewNullFloat64(true, 1.1), T08Double: sqlext.NewNullFloat64(true, 1.1),
		T09Datetime: sqlext.NewNullTime(true, strTime), T10Bool: sqlext.NewNullBool(true, false),
	}
	jsonFullValStr := `{"id":1,"t02Int":1,"t03Varchar":"1","t04Text":"1","t05Longtext":"1","t06Decimal":"1.1","t07Float":1.1,"t08Double":1.1,"t09Datetime":"2025-02-02 00:00:00","t10Bool":false}`

	str, err := njson.ObjToJsonStrByGoJson(voNullVal)
	ntools.TestErrPainic(t, "TestGoJson voNullVal", err)
	ntools.TestEq(t, "TestGoJson voNullVal", jsonNullValStr, str)

	str, err = njson.ObjToJsonStrByGoJson(voFullVal)
	ntools.TestErrPainic(t, "TestGoJson voNullVal", err)
	ntools.TestEq(t, "TestGoJson voNullVal", jsonFullValStr, str)

	vo, err := njson.Str2ObjByGoJson[TestVoDo](jsonFullValStr)
	ntools.TestErrPainic(t, "Str2ObjByGoJson", err)
	ntools.TestEq(t, "Str2ObjByGoJson", "1", vo.T03Varchar.String)

	arrJsonStr := `[{"id":1},{"id":2}]`

	vos, err := njson.Str2ObjArrByGoJson[TestVoDo](&arrJsonStr)
	ntools.TestErrPainic(t, "Str2ArrByGoJson", err)
	ntools.TestEq(t, "Str2ArrByGoJson", 2, len(*vos))

	//基本类型测试
	text := `"1"`
	jsonStr, err := njson.Str2ObjByGoJson[string](&text)
	ntools.TestErrPainic(t, "TestGoJson ", err)
	ntools.TestEq(t, "Str2ObjByGoJson", text, fmt.Sprintf("\"%s\"", *jsonStr))

	text = `["1","2"]`
	resultArr, err := njson.Str2ObjArrByGoJson[string](text)
	ntools.TestErrPainic(t, "Str2ArrByGoJson ", err)
	ntools.TestEq(t, "Str2ArrByGoJson", 2, len(*resultArr))

}

func TestSonicJson(t *testing.T) {
	strTime, _ := ntools.TimeStr2TimeByLayout("2025-02-02", "2006-01-02")
	voNullVal := &TestVoDo{}
	jsonNullValStr := `{"id":0,"t02Int":null,"t03Varchar":null,"t04Text":null,"t05Longtext":null,"t06Decimal":null,"t07Float":null,"t08Double":null,"t09Datetime":null,"t10Bool":null}`

	voFullVal := &TestVoDo{
		Id: 1, T02Int: sqlext.NewNullInt(true, 1),
		T03Varchar: sqlext.NewNullString(true, "1"), T04Text: sqlext.NewNullString(true, "1"),
		T05Longtext: sqlext.NewNullString(true, "1"), T06Decimal: sqlext.NewNullDecimal(true, "1.1"),
		T07Float: sqlext.NewNullFloat64(true, 1.1), T08Double: sqlext.NewNullFloat64(true, 1.1),
		T09Datetime: sqlext.NewNullTime(true, strTime), T10Bool: sqlext.NewNullBool(true, false),
	}
	jsonFullValStr := `{"id":1,"t02Int":1,"t03Varchar":"1","t04Text":"1","t05Longtext":"1","t06Decimal":"1.1","t07Float":1.1,"t08Double":1.1,"t09Datetime":"2025-02-02 00:00:00","t10Bool":false}`

	str, err := njson.Obj2JsonStr(voNullVal)
	ntools.TestErrPainic(t, "Obj2JsonStr voNullVal", err)
	ntools.TestEq(t, "Obj2JsonStr voNullVal", jsonNullValStr, str)

	str, err = njson.Obj2JsonStr(voFullVal)
	ntools.TestErrPainic(t, "Obj2JsonStr voNullVal", err)
	ntools.TestEq(t, "Obj2JsonStr voNullVal", jsonFullValStr, str)

	vo, err := njson.Str2Obj[TestVoDo](&jsonFullValStr)
	ntools.TestErrPainic(t, "Str2Obj ", err)
	ntools.TestEq(t, "Str2Obj ", "1", vo.T03Varchar.String)

	arrJsonStr := `[{"id":1},{"id":2}]`

	vos, err := njson.Str2ObjArr[TestVoDo](&arrJsonStr)
	ntools.TestErrPainic(t, "Str2ObjArr ", err)
	ntools.TestEq(t, "Str2ObjArr ", 2, len(*vos))

	text := `"1"`
	jsonStr, err := njson.Str2Obj[string](&text)
	ntools.TestErrPainic(t, "Str2Obj ", err)
	ntools.TestEq(t, "Str2Obj", text, fmt.Sprintf("\"%s\"", *jsonStr))

	text = `["1","2"]`
	resultArr, err := njson.Str2ObjArr[string](text)
	ntools.TestErrPainic(t, "Str2ObjArr ", err)
	ntools.TestEq(t, "Str2ObjArr", 2, len(*resultArr))

}
