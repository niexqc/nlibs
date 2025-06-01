package ntools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"

	"os"
	"runtime"
	"strings"
	"sync"

	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/timandy/routine"
)

type NwLogHandler struct {
	Level      slog.Leveler
	PrintMehod int // 0-不打印 ，1-详情,2-仅方法名称
	OutMode    int // 0-不打印,1-控制台,2-文件,3-都打印
	out        io.Writer
}

type nwDayLogWriter struct {
	FileNamePrefix string
	fileWriter     *bufio.Writer
	curFile        *os.File
	curDateStr     string
}

var syncLock sync.Mutex
var threadLocal = routine.NewInheritableThreadLocal[string]()

func SlogSetTraceId(traceId string) {
	threadLocal.Set(traceId)
}

func SlogGetTraceId() string {
	return threadLocal.Get()
}

// printMethod int 方法打印 0-不打印 ，1-详情,2-仅方法名称
// outMode int 日志输出方式 0-不打印,1-控制台,2-文件,3-都打印
func SlogConf(logFilePrefix, confLevel string, outMode int, printMethod int) {
	slogLevel := SlogLevelStr2Level(confLevel)
	logWriter := &nwDayLogWriter{FileNamePrefix: logFilePrefix}
	nwLogHandler := NewNwLogHandlerForSlog(logWriter, slogLevel, outMode, printMethod)
	slogger := slog.New(nwLogHandler)
	slog.SetDefault(slogger)
	slog.Info(fmt.Sprintf("SLog Level:%v", confLevel))
}

func SlogConf4Test() {
	SlogConf("test", "debug", 1, 2)
}

func SlogLevelStr2Level(confLevel string) slog.Level {
	confLevel = strings.ToLower(confLevel)
	var slogLevel slog.Level
	if confLevel == "debug" {
		slogLevel = slog.LevelDebug
	} else if confLevel == "info" {
		slogLevel = slog.LevelInfo
	} else if confLevel == "warn" {
		slogLevel = slog.LevelWarn
	} else {
		slogLevel = slog.LevelError
	}
	return slogLevel
}

func NewNwLogHandlerForSlog(out io.Writer, level slog.Leveler, outMode int, printMethod int) *NwLogHandler {
	h := &NwLogHandler{Level: level, out: out, OutMode: outMode, PrintMehod: printMethod}
	return h
}

func (h *nwDayLogWriter) Write(p []byte) (n int, err error) {
	curDateStr := TimeTo20060102(time.Now())
	if h.curDateStr != curDateStr {
		//重新初始化fileWriter
		return 0, h.syncLockReInitFile(curDateStr)
	}
	n, err = h.fileWriter.Write(p)
	h.fileWriter.Flush()
	return n, err
}

func (h *nwDayLogWriter) syncLockReInitFile(curDateStr string) error {
	syncLock.Lock()
	defer syncLock.Unlock()
	if h.curDateStr != curDateStr {
		//处理同步问题
		h.curDateStr = curDateStr
		if h.curFile != nil {
			h.curFile.Close()
		}
		os.Mkdir("logs", 0755)

		flag := os.O_RDWR | os.O_CREATE | os.O_APPEND
		curFile, err := os.OpenFile(fmt.Sprintf("logs/%s%s.log", h.FileNamePrefix, h.curDateStr), flag, 0755)
		if err != nil {
			return err
		}
		h.curFile = curFile
		h.fileWriter = bufio.NewWriter(h.curFile)
	}
	return nil
}

func (h *NwLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.Level.Level()
}

func (h *NwLogHandler) WithGroup(name string) slog.Handler {
	panic(nerror.NewRunTimeError("未实现 WithGroup"))
}

func (h *NwLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	panic(nerror.NewRunTimeError("未实现 WithAttrs"))
}

func (h *NwLogHandler) Handle(ctx context.Context, r slog.Record) (err error) {
	if h.OutMode == 0 {
		return nil
	}
	sb := strings.Builder{}
	if !r.Time.IsZero() {
		sb.WriteString(fmt.Sprintf("%-23s ", Time2StrMilli(r.Time)))
	}
	sb.WriteString(fmt.Sprintf("%-5s ", r.Level.String()))
	traceId := SlogGetTraceId()
	sb.WriteString(fmt.Sprintf("%s ", If3(len(traceId) > 0, traceId, "traceIdUnSet")))

	callerStr, funcStr := h.caller(r)
	sb.WriteString(fmt.Sprintf("%s ", callerStr))

	if h.PrintMehod > 0 {
		sb.WriteString(fmt.Sprintf("%s ", funcStr))
	}

	sb.WriteString(r.Message + " ")

	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(a.String())
		return true
	})

	sb.WriteString("\n")
	printData := []byte(sb.String())

	if h.OutMode == 1 {
		os.Stdout.Write(printData)
	} else if h.OutMode == 2 {
		_, err = h.out.Write(printData)
	} else {
		os.Stdout.Write(printData)
		_, err = h.out.Write(printData)
	}
	return err
}

func (h *NwLogHandler) caller(r slog.Record) (caller, funcStr string) {
	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		ec, _ := fs.Next()
		idx := strings.LastIndexByte(ec.File, '/')
		pathStr := ec.File
		if idx != -1 {
			idx = strings.LastIndexByte(ec.File[:idx], '/')
			if idx != -1 {
				pathStr = ec.File[idx+1:]
			}
		}
		funcName := ""
		if h.PrintMehod > 0 {
			if h.PrintMehod == 1 {
				funcName = ec.Func.Name()
			} else {
				funcName = ec.Func.Name()
				funcName = funcName[strings.LastIndex(funcName, ".")+1:]
			}
		}
		return fmt.Sprintf("%s:%d", pathStr, int64(ec.Line)), funcName
	} else {
		return "unknown:0", "unknown"
	}
}
