package ngin

import (
	"database/sql/driver"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

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
	if validate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		slog.Debug("InithZhTranslator")
		// 创建中文翻译器
		localZh := zh.New()
		//
		zhTrans, _ := ut.New(localZh, localZh).GetTranslator("zh")
		// 注册中文翻译
		zh_translations.RegisterDefaultTranslations(validate, zhTrans)
		// 注册自定义字段名映射_通过tag[zhdesc]将中文字段标准出来
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
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
		registerNullFunc(validate)

		// 覆盖所有数值规则的翻译逻辑：移除千位分隔符
		registerCustomFormat(validate, zhTrans)

		return &NValider{Validate: validate, ZhTrans: zhTrans}
	} else {
		panic(nerror.NewRunTimeError("检查binding.Validator.Engine()是否是*validator.Validate"))
	}

}
func registerCustomFormat(validate *validator.Validate, trans ut.Translator) {
	// 覆盖所有数值规则的翻译逻辑
	numberTagMap := map[string]string{
		"min": "大于等于", "max": "小于等于", "gte": "大于等于", "lte": "小于等于", "eq": "等于", "ne": "不等于", "gt": "大于", "lt": "小于",
	}
	// numericTags := []string{"min", "max", "gte", "lte", "eq", "ne", "gt", "lt"}
	for tag, tip := range numberTagMap {
		_ = validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, fmt.Sprintf("{0}必须满足条件,%s[{1}]", tip), true) // 保持原始模板
		}, func(ut ut.Translator, fe validator.FieldError) string {
			// 关键：移除参数中的千位分隔符
			param := strings.ReplaceAll(fe.Param(), ",", "")
			t, _ := ut.T(tag, fe.Field(), param)
			return t
		})
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
		sqlext.NullFloat64{})

	validate.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(driver.Valuer); ok {
			val := valuer.(sqlext.NullBool)
			return val.Valid
		}
		return nil
	}, sqlext.NullBool{})
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
