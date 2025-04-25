package ngin

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"

	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/nredis"
	"github.com/niexqc/nlibs/ntools"

	"github.com/gin-gonic/gin"
)

// GinLogger 接收gin框架默认的日志
func LoggerHandlerFunc() gin.HandlerFunc {
	slog.Debug("Add Middleware LoggerHandlerFunc")
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()
		cost := time.Since(start).Milliseconds()
		agentStr := (&ntools.NString{S: c.Request.UserAgent()}).CutString(32)
		logStr := fmt.Sprintf("%s\t%s\t%d\t%s\t%dms\t%s\t%s", path, c.Request.Method, c.Writer.Status(), c.ClientIP(), cost, query, agentStr)
		slog.Info(logStr)
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
func TraceIdGenHandlerFunc(traceIdPrefix string, redisService *nredis.RedisService) gin.HandlerFunc {
	slog.Debug("Add Middleware TraceIdGenHandlerFunc")
	return func(c *gin.Context) {
		timeStr := time.Now().Format("20060102T150405")
		redisKeyStr := traceIdPrefix + timeStr
		keySeqNo, err := redisService.Int64Incr(redisKeyStr, 1200)
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
	return func(c *gin.Context) {
		ginHeaders := c.Request.Header
		heaerVo := NiexqGinHeaderVo{}
		heaerVo.UserAgent = ginHeaders.Get("user-agent")
		heaerVo.ContentType = ginHeaders.Get("content-type")
		heaerVo.UserToken = ginHeaders.Get("user-token")
		heaerVo.AppType = ginHeaders.Get("app-type")
		heaerVo.AppVer = ginHeaders.Get("app-ver")
		heaerVo.ClientTime = ginHeaders.Get("client-time")
		heaerVo.VistSrc = ginHeaders.Get("vist-src")
		heaerVo.UserIp = c.ClientIP()
		heaerVo.VistTar = c.Request.RequestURI
		c.Set(reflect.TypeOf(heaerVo).Name(), &heaerVo)
		c.Next()
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
