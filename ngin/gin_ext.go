package ngin

import (
	"fmt"
	"log/slog"

	"strings"

	"github.com/gin-gonic/gin"
	"github.com/niexqc/nlibs/nerror"
)

type NGin struct {
	GinEngine *gin.Engine
}

func NewNGin() *NGin {
	slog.Debug("CreateGinEngine...")
	gin.SetMode(gin.ReleaseMode)
	return &NGin{GinEngine: gin.New()}
}

func (nGin *NGin) RouterRedirect(redirectPath string, ctx *gin.Context) {
	ctx.Request.URL.Path = redirectPath
	nGin.GinEngine.HandleContext(ctx)
}

func ShouldBindByHeader[T any](headerVo *NiexqGinHeaderVo, ctx *gin.Context) *T {
	if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "application/json") {
		return ShouldBindJSON[T](ctx)
	} else if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "	application/x-www-form-urlencoded") {
		return ShouldBind[T](ctx)
	} else if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "multipart/form-data") {
		panic(nerror.NewRunTimeError("ContentType:multipart/form-data 还未处理"))
	} else {
		panic(nerror.NewRunTimeError("错误的[ContentType]"))
	}
}

func ShouldBind[T any](ctx *gin.Context) *T {
	obj := new(T)
	if err := ctx.ShouldBind(obj); err != nil {
		transErr2Zh(err)
		return obj
	}
	return obj
}

func ShouldBindJSON[T any](ctx *gin.Context) *T {
	obj := new(T)
	if err := ctx.ShouldBindJSON(obj); err != nil {
		transErr2Zh(err)
		return obj
	}
	return obj
}

func ReadHeader(ctx *gin.Context) *NiexqGinHeaderVo {
	obj := &NiexqGinHeaderVo{}
	if err := ctx.ShouldBindJSON(obj); err != nil {
		transErr2Zh(err)
		return obj
	}
	return obj
}

func (nGin *NGin) LogPrintAllRouterInfo() {
	routers := nGin.GinEngine.Routes()
	routersInfo := ""
	for _, v := range routers {
		routersInfo += "\n" + fmt.Sprintf("%s %s %s", v.Method, v.Path, v.Handler)
	}
	slog.Debug(routersInfo)
}
