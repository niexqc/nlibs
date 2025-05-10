package sqlext

import (
	"fmt"
	"reflect"
	"time"

	"github.com/niexqc/nlibs/nerror"
)

type NdbBasicType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string | ~bool |
		time.Time | NullBool | NullFloat64 | NullInt | NullInt64 | NullString | NullTime
}

type NdbDyObjFieldInfo struct {
	DbColName       string
	StructFieldName string
	GoColType       string
	DbColType       string
	DbColIsNull     bool
}
type NdbDyObj struct {
	FiledsInfo map[string]*NdbDyObjFieldInfo
	Data       any
}

func GetFiledVal[T NdbBasicType](dyObj *NdbDyObj, structFieldName string) (rt *T, err error) {
	objType := reflect.ValueOf(dyObj.Data)
	if objType.Kind() == reflect.Pointer {
		objType = objType.Elem()
	}
	if objType.Kind() != reflect.Struct {
		return nil, nerror.NewRunTimeError("不能获取非结构的值")
	}
	fieldVal := objType.FieldByName(structFieldName)
	if !fieldVal.IsValid() {
		return nil, nil
	}
	// 检查字段是否可访问
	if !fieldVal.CanInterface() {
		return nil, nerror.NewRunTimeError("字段不可访问")
	}
	// 处理字段指针解引用
	if fieldVal.Kind() == reflect.Pointer {
		if fieldVal.IsNil() {
			return nil, nil
		}
		fieldVal = fieldVal.Elem()
	}
	v := fieldVal.Interface()
	// 如果是指针已经解引用了
	if ns, ok := v.(T); ok {
		return &ns, nil
	} else {
		return nil, nerror.NewRunTimeError(fmt.Sprintf("【%s】的字段类型为【%s】", structFieldName, fieldVal.Type().String()))
	}
}
