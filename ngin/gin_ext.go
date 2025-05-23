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
	NValider  *NValider
}

func NewNGin() *NGin {
	return NewNGinWithMaxConcurrent(100)
}

func NewNGinWithMaxConcurrent(max int) *NGin {
	gin.SetMode(gin.ReleaseMode)
	valider := NewNValider("json", "zhdesc")

	ngin := &NGin{GinEngine: gin.New(), NValider: valider}
	ngin.Use(MaxConcurrentHandlerFunc(max))
	return ngin
}

func (nGin *NGin) Use(middleware ...gin.HandlerFunc) gin.IRoutes {
	return nGin.GinEngine.Use(middleware...)
}

func (nGin *NGin) Static(relativePath, root string) gin.IRoutes {
	return nGin.GinEngine.StaticFS(relativePath, gin.Dir(root, false))
}

func (nGin *NGin) GET(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return nGin.GinEngine.GET(relativePath, handlers...)
}

func (nGin *NGin) POST(relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes {
	return nGin.GinEngine.POST(relativePath, handlers...)
}

func (nGin *NGin) Run(addr string) (err error) {
	slog.Info(addr)
	return nGin.GinEngine.Run(addr)
}

func (nGin *NGin) RouterRedirect(redirectPath string, ctx *gin.Context) {
	ctx.Request.URL.Path = redirectPath
	nGin.GinEngine.HandleContext(ctx)
}

func ShouldBindByHeader[T any](headerVo *NiexqGinHeaderVo, ctx *gin.Context, nValider *NValider) *T {
	if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "application/json") {
		return ShouldBindJSON[T](ctx, nValider)
	} else if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "	application/x-www-form-urlencoded") {
		return ShouldBind[T](ctx, nValider)
	} else if strings.HasPrefix(strings.ToLower(headerVo.ContentType), "multipart/form-data") {
		panic(nerror.NewRunTimeError("ContentType:multipart/form-data 还未处理"))
	} else {
		panic(nerror.NewRunTimeError("错误的[ContentType]"))
	}
}

func ShouldBind[T any](ctx *gin.Context, nValider *NValider) *T {
	obj := new(T)
	if err := ctx.ShouldBind(obj); err != nil {
		nValider.TransErr2Zh(err)
		return obj
	}
	return obj
}

func ShouldBindJSON[T any](ctx *gin.Context, nValider *NValider) *T {
	obj := new(T)
	if err := ctx.ShouldBindJSON(obj); err != nil {
		nValider.TransErr2Zh(err)
		return obj
	}
	return obj
}

func ReadHeader(ctx *gin.Context, nValider *NValider) *NiexqGinHeaderVo {
	obj := &NiexqGinHeaderVo{}
	if err := ctx.ShouldBindJSON(obj); err != nil {
		nValider.TransErr2Zh(err)
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
