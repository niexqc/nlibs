package ngin

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/niexqc/nlibs/ndb/sqlext"
	"github.com/niexqc/nlibs/nerror"
)

type NValider struct {
	Validate *validator.Validate
	ZhTrans  ut.Translator
}

func NewNValider(tagJsonName, tagZhdescName string) *NValider {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		slog.Debug("InithZhTranslator")
		// 创建中文翻译器
		localZh := zh.New()
		zhTrans, _ := ut.New(localZh, localZh).GetTranslator("zh")
		// 注册中文翻译
		zh_translations.RegisterDefaultTranslations(v, zhTrans)
		// 注册自定义字段名映射_通过tag[zhdesc]将中文字段标准出来
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			tagJson := fld.Tag.Get(tagJsonName)
			tagZhdesc := fld.Tag.Get(tagZhdescName)
			if tagJson != "" && tagZhdesc != "" {
				return fmt.Sprintf("%s[%s]", tagZhdesc, tagJson)
			} else if tagJson != "" {
				return tagJson
			} else if tagZhdesc != "" {
				return tagZhdesc
			}
			return fld.Name
		})
		// 注册自定义类型
		registerNullFunc(v)

		return &NValider{Validate: v, ZhTrans: zhTrans}
	} else {
		panic(nerror.NewRunTimeError("检查binding.Validator.Engine()是否是*validator.Validate"))
	}

}

func registerNullFunc(validate *validator.Validate) {
	validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(driver.Valuer); ok {
			val, _ := valuer.Value()
			return val
		}
		return nil
	},
		sqlext.NullString{},
		sqlext.NullTime{},
		sqlext.NullInt{},
		sqlext.NullString{},
		sqlext.NullInt64{},
		sqlext.NullFloat64{},
		sqlext.NullBool{})
}

func (nvld *NValider) TransErr2Zh(err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		validErr := &NiexqValidErr{}
		// 遍历每个错误并翻译
		for _, e := range validationErrors {
			validErr.ErrDescList = append(validErr.ErrDescList, e.Translate(nvld.ZhTrans)+";")
		}
		panic(validErr)
	}
}

func (nvld *NValider) TransErr2ZhErr(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		validErr := &NiexqValidErr{}
		// 遍历每个错误并翻译
		for _, e := range validationErrors {
			validErr.ErrDescList = append(validErr.ErrDescList, e.Translate(nvld.ZhTrans)+";")
		}
		return validErr
	}
	return err
}

func (nvld *NValider) ValidStrct(obj any) error {
	return nvld.Validate.Struct(obj)
}

type NiexqValidErr struct {
	ErrDescList []string
}

// Error 实现Error接口
func (e *NiexqValidErr) Error() string {
	errTxt := ""
	if len(e.ErrDescList) > 0 {
		for i, v := range e.ErrDescList {
			errTxt += fmt.Sprintf("%d:%s\n", i+1, v)
		}
	}
	return errTxt
}
