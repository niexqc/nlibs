package ngauss

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

type NGaussDyObjFieldInfo struct {
	DbColName       string
	StructFieldName string
	JsonColName     string
	GoColType       string
	DbColType       string
	DbColIsNull     bool
}

type NGaussDyObj struct {
	DbNameFiledsMap map[string]*NGaussDyObjFieldInfo
	Data            any
}

func GetFiledVal[T sqlext.NdbBasicType](dyObj *NGaussDyObj, structFieldName string) (rt *T, err error) {
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

func DyObjList2Json(dyObjList []*NGaussDyObj) (jsonStr string, err error) {
	dataList := []any{}
	for _, dyObj := range dyObjList {
		dataList = append(dataList, dyObj.Data)
	}
	jsonStr, err = njson.SonicObj2Str(dataList)
	return jsonStr, nil
}

func CreateDyStruct(cols []*sql.ColumnType) (dyObjDefine reflect.Type, dbNameFiledsMap map[string]*NGaussDyObjFieldInfo, err error) {
	fields := []reflect.StructField{}
	dbNameFiledsMap = make(map[string]*NGaussDyObjFieldInfo)
	for _, v := range cols {
		DbNameNstr := &ntools.NString{S: v.Name()}
		dbFname := DbNameNstr.S
		structFname := DbNameNstr.Under2Camel(true)
		jsonFname := DbNameNstr.Under2Camel(false)
		//高斯驱动在查询是否不返回字段是否允许为空，所以固定为空
		goType, err := gaussDbUdtNameToGoType(v.DatabaseTypeName(), true)
		if nil != err {
			return nil, nil, err
		}

		tag := reflect.StructTag(fmt.Sprintf(`db:"%s" json:"%s"`, dbFname, jsonFname))

		fields = append(fields, reflect.StructField{Name: structFname, Type: goType, Tag: tag})

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		dbNameFiledsMap[dbFname] = &NGaussDyObjFieldInfo{
			StructFieldName: structFname,
			DbColName:       dbFname,
			JsonColName:     jsonFname,
			GoColType:       goType.String(),
			DbColType:       v.DatabaseTypeName(),
			DbColIsNull:     nullable,
		}
	}
	// 创建动态结构体类型
	return reflect.StructOf(fields), dbNameFiledsMap, nil
}
