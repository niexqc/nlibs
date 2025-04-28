package ngin

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var zhTrans ut.Translator

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

func InithZhTranslator() {
	// 核实gin的valid引擎
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		slog.Debug("InithZhTranslator")
		// 创建中文翻译器
		localZh := zh.New()
		zhTrans, _ = ut.New(localZh, localZh).GetTranslator("zh")
		// 注册中文翻译
		zh_translations.RegisterDefaultTranslations(v, zhTrans)
		// 注册自定义字段名映射_通过tag[zhdesc]将中文字段标准出来
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			tagJson := fld.Tag.Get("json")
			tagZhdesc := fld.Tag.Get("zhdesc")
			if tagJson != "" && tagZhdesc != "" {
				return fmt.Sprintf("%s[%s]", tagZhdesc, tagJson)
			} else if tagJson != "" {
				return tagJson
			} else if tagZhdesc != "" {
				return tagZhdesc
			}
			return fld.Name
		})
	}
}

func transErr2Zh(err error) {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		validErr := &NiexqValidErr{}
		// 遍历每个错误并翻译
		for _, e := range validationErrors {
			validErr.ErrDescList = append(validErr.ErrDescList, e.Translate(zhTrans)+";")
		}
		panic(validErr)
	}
}
