package ngin

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"

	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/niexqc/nlibs/ncache"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"

	"github.com/gin-gonic/gin"
)

// GinLogger 接收gin框架默认的日志
func LoggerHandlerFunc(showReqBody bool) gin.HandlerFunc {
	slog.Debug("Add Middleware LoggerHandlerFunc")
	return func(ctx *gin.Context) {
		start := time.Now()
		nGinPrintReqLog(ctx, showReqBody)
		ctx.Next()
		slog.Info(fmt.Sprintf("Resp:%s\t%v\t%dms", ctx.Request.RequestURI, ctx.Writer.Status(), time.Since(start).Milliseconds()))
	}
}

func nGinPrintReqLog(ctx *gin.Context, showReqBody bool) {
	headerVo := GetHeaderVoFromCtx(ctx)
	visitTar := headerVo.VisitTar
	rawQuery := ntools.If3(ctx.Request.URL.RawQuery == "", "Nil_NoRawQuery", ctx.Request.URL.RawQuery)
	reqMethod := ctx.Request.Method
	contentType := headerVo.ContentType
	visitSrc := ntools.If3(headerVo.VisitSrc == "", "No_VisitSrc", headerVo.VisitSrc)
	clientIp := ctx.ClientIP()

	agentStr := (&ntools.NString{S: ctx.Request.UserAgent()}).CutString(120)
	agentStr = ntools.If3(agentStr == "", "Nil_UnSetUserAgent", agentStr)

	logStr := fmt.Sprintf("Req:%s\t%s\t%s\t%s\t%s\t%s\t%s", visitTar, rawQuery, reqMethod, contentType, visitSrc, clientIp, agentStr)
	slog.Info(logStr)
	// 打印原始请求参数
	if showReqBody {
		reqBodyStr := ""
		if strings.ContainsAny(contentType, "json") || strings.ContainsAny(contentType, "text") || strings.ContainsAny(contentType, "xml") {
			reqBodyStr = string(*headerVo.ReqBody)
		} else {
			reqBodyStr = "Nil_ParseBody"
		}
		slog.Info(fmt.Sprintf("ReqBody:%s", reqBodyStr))
	}
}

// MaxConcurrentHandlerFunc 掉项目可能出现的panic
func MaxConcurrentHandlerFunc(max int) gin.HandlerFunc {
	slog.Debug("Add Middleware MaxConcurrentHandlerFunc")
	sem := make(chan struct{}, max)
	return func(c *gin.Context) {
		select {
		case sem <- struct{}{}: // 获取信号量
			defer func() { <-sem }() // 处理完成后释放
			c.Next()
		default:
			c.AbortWithStatusJSON(429, gin.H{"error": "服务繁忙，请稍后重试"})
		}
	}
}

// Recovery recover掉项目可能出现的panic
func RecoveryHandlerFunc() gin.HandlerFunc {
	slog.Debug("Add Middleware RecoveryHandlerFunc")
	return func(c *gin.Context) {
		defer recoveryErrorWork(c)
		c.Next()
	}
}

// 日志跟踪ID生成
func TraceIdGenHandlerFunc(traceIdPrefix string, incahce ncache.ICacheService) gin.HandlerFunc {
	slog.Debug("Add Middleware TraceIdGenHandlerFunc")
	return func(c *gin.Context) {
		timeStr := time.Now().Format("20060102T150405")
		redisKeyStr := traceIdPrefix + timeStr
		keySeqNo, err := incahce.Int64Incr(redisKeyStr, 1200)
		if err != nil {
			panic(err)
		}
		traceId := fmt.Sprintf("%s%04d", redisKeyStr, keySeqNo)
		ntools.SlogSetTraceId(traceId)
		c.Next()
	}
}

// Header读取并设置
func HeaderSetHandlerFunc() gin.HandlerFunc {
	slog.Debug("Add Middleware HeaderSetHandlerFunc")

	readAndResetBody := func(c *gin.Context) *[]byte {
		// 1. 读取原始 Body 内容
		body, err := c.GetRawData()
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return &body
		}
		// 2. 重写 GetBody 方法（关键！）
		c.Request.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBuffer(body)), nil
		}
		// 3. 重置 Body 供后续使用
		c.Request.Body, _ = c.Request.GetBody()
		return &body
	}

	return func(ctx *gin.Context) {
		readAndResetBody(ctx)
		ginHeaders := ctx.Request.Header
		heaerVo := NiexqGinHeaderVo{}
		heaerVo.ReqBody = readAndResetBody(ctx)
		heaerVo.UserAgent = ctx.Request.UserAgent()
		heaerVo.ContentType = ginHeaders.Get("content-type")
		heaerVo.UserToken = ginHeaders.Get("user-token")
		heaerVo.AppType = ginHeaders.Get("app-type")
		heaerVo.AppVer = ginHeaders.Get("app-ver")
		heaerVo.ClientTime = ginHeaders.Get("client-time")
		heaerVo.OneceStr = ginHeaders.Get("onece-str")
		heaerVo.VisitSrc = ginHeaders.Get("vist-src")
		heaerVo.UserIp = ctx.ClientIP()
		heaerVo.VisitTar = ctx.Request.RequestURI
		slog.Debug("Headers:", "json", njson.SonicObj2Str(heaerVo))
		ctx.Set(reflect.TypeOf(heaerVo).Name(), &heaerVo)
		ctx.Next()
	}
}

func GetHeaderVoFromCtx(c *gin.Context) *NiexqGinHeaderVo {
	if v, exist := c.Get(reflect.TypeOf(NiexqGinHeaderVo{}).Name()); exist {
		if ne, ok := v.(*NiexqGinHeaderVo); ok {
			return ne
		}
	}
	return &NiexqGinHeaderVo{}
}

func recoveryErrorWork(c *gin.Context) {
	reqPath := c.Request.URL.Path
	if err := recover(); err != nil {
		//检查连接是否已断开
		var brokenPipe bool
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}
		httpRequest, _ := httputil.DumpRequest(c.Request, false)
		if brokenPipe {
			//发生了异常，但是连接已断开
			slog.Error(fmt.Sprintf("%s\t异常:%v\n%s\n%s", reqPath, err, string(httpRequest), "连接已断开"))
			c.Error(err.(error)) // nolint: errcheck
			c.Abort()
			return
		}
		//这里把错误输出出去
		if vze, ok := err.(*NiexqValidErr); ok {
			result := NewErrBaseResp(fmt.Sprintf("%v", vze.Error()))
			result.Code = RespCode_Valid_Err
			c.JSON(http.StatusOK, &result)
		} else if vze, ok := err.(*nerror.RunTimeErr); ok {
			if vze.SrcErr == nil {
				result := NewErrBaseResp(fmt.Sprintf("%v", vze.ErrDesc))
				result.Code = RespCode_RunTime_Err
				c.JSON(http.StatusOK, &result)
			} else {
				errStr := fmt.Sprintf("%v", vze.ErrDesc)
				if vze1, ok := vze.SrcErr.(*nerror.RunTimeErr); ok {
					errStr = fmt.Sprintf("1-%v,2-%v", errStr, vze1.Error())
				} else {
					slog.Error(fmt.Sprintf("%s\t异常:%v\n%s\n%s", reqPath, err, string(httpRequest), debug.Stack()))
				}
				result := NewErrBaseResp(errStr)
				result.Code = RespCode_RunTime_Err2
				c.JSON(http.StatusOK, &result)
			}
		} else {
			slog.Error(fmt.Sprintf("%s\t异常:%v\n%s\n%s", reqPath, err, string(httpRequest), debug.Stack()))
			result := NewErrBaseResp(fmt.Sprintf("%v", err))
			result.Code = RespCode_UnKnown_Err
			c.JSON(http.StatusOK, &result)
		}
		c.Abort()
	}
}
