package nmysql

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

type NMysqlDyObjFieldInfo struct {
	DbColName       string
	StructFieldName string
	JsonColName     string
	GoColType       string
	DbColType       string
	DbColIsNull     bool
}

type NMysqlDyObj struct {
	DbNameFiledsMap map[string]*NMysqlDyObjFieldInfo
	Data            any
}

func GetFiledVal[T sqlext.NdbBasicType](dyObj *NMysqlDyObj, structFieldName string) (rt *T, err error) {
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

func DyObjList2Json(dyObjList []*NMysqlDyObj) (jsonStr string, err error) {
	dataList := []any{}
	for _, dyObj := range dyObjList {
		dataList = append(dataList, dyObj.Data)
	}
	jsonStr, err = njson.SonicObj2Str(dataList)
	return jsonStr, nil
}

func createDyStruct(cols []*sql.ColumnType) (dyObjDefine reflect.Type, filedInfos map[string]*NMysqlDyObjFieldInfo, err error) {
	fields := []reflect.StructField{}
	mysqlType2GoType := func(col *sql.ColumnType) (reflect.Type, error) {
		nullable, ok := col.Nullable()
		if !ok {
			nullable = false
		}
		return mysqlTypeToGoType(col.DatabaseTypeName(), nullable)
	}
	filedInfos = make(map[string]*NMysqlDyObjFieldInfo)
	for _, v := range cols {
		DbNameNstr := &ntools.NString{S: v.Name()}
		dbFname := DbNameNstr.S
		structFname := DbNameNstr.Under2Camel(true)
		jsonFname := DbNameNstr.Under2Camel(false)
		goType, err := mysqlType2GoType(v)
		if nil != err {
			return nil, nil, err
		}
		tag := reflect.StructTag(fmt.Sprintf(`db:"%s" json:"%s"`, dbFname, jsonFname))

		fields = append(fields, reflect.StructField{Name: structFname, Type: goType, Tag: tag})

		nullable, ok := v.Nullable()
		if !ok {
			nullable = false
		}
		filedInfos[dbFname] = &NMysqlDyObjFieldInfo{
			StructFieldName: structFname,
			DbColName:       dbFname,
			JsonColName:     jsonFname,
			GoColType:       goType.String(),
			DbColType:       v.DatabaseTypeName(),
			DbColIsNull:     nullable,
		}
	}
	// 创建动态结构体类型
	return reflect.StructOf(fields), filedInfos, nil
}
