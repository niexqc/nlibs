package ntools

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"os"
	"runtime"
	"strings"
	"sync"

	"time"

	"github.com/niexqc/nlibs/nerror"
	"github.com/timandy/routine"
)

var threadLocal = routine.NewInheritableThreadLocal[string]()

type NwLogHandler struct {
	Level      slog.Leveler
	PrintMehod int // 0-不打印 ，1-详情,2-仅方法名称
	OutMode    int // 0-不打印,1-控制台,2-文件,3-都打印
	out        io.Writer
}

func SlogSetTraceId(traceId string) {
	threadLocal.Set(traceId)
}

func SlogGetTraceId() string {
	return threadLocal.Get()
}

// printMethod int 方法打印 0-不打印 ，1-详情,2-仅方法名称
// outMode int 日志输出方式 0-不打印,1-控制台,2-文件,3-都打印
func SlogConfWithDir(logDir, logFilePrefix, confLevel string, outMode int, printMethod int) {
	slogLevel := SlogLevelStr2Level(confLevel)
	logWriter, err := NewDailyRotatingLogger(logDir, logFilePrefix, 10240, 200*time.Millisecond)
	if nil != err {
		panic(err)
	}
	nwLogHandler := NewNwLogHandlerForSlog(logWriter, slogLevel, outMode, printMethod)
	slogger := slog.New(nwLogHandler)
	slog.SetDefault(slogger)
	slog.Info(fmt.Sprintf("SLog Level:%v", confLevel))
}

// printMethod int 方法打印 0-不打印 ，1-详情,2-仅方法名称
// outMode int 日志输出方式 0-不打印,1-控制台,2-文件,3-都打印
func SlogConf(logFilePrefix, confLevel string, outMode int, printMethod int) {
	SlogConfWithDir("logs", logFilePrefix, confLevel, outMode, printMethod)
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

type DailyRotatingLogger struct {
	filePath      string        // 日志目录路径
	filePrefix    string        // 文件名前缀
	currentDate   string        // 当前日期 (格式: 20060102)
	file          *os.File      // 当前文件句柄
	fileWriter    *bufio.Writer // 缓冲写入器
	buffer        chan []byte   // 内存缓冲队列
	flushInterval time.Duration // 刷盘间隔
	mu            sync.Mutex    // 文件操作互斥锁
	stopChan      chan struct{} // 关闭信号
}

func NewDailyRotatingLogger(dir, prefix string, bufferSize int, flushInterval time.Duration) (*DailyRotatingLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}

	logger := &DailyRotatingLogger{
		filePath:      dir,
		filePrefix:    prefix,
		buffer:        make(chan []byte, bufferSize),
		flushInterval: flushInterval,
		stopChan:      make(chan struct{}),
	}

	if err := logger.rotateIfNeeded(); err != nil {
		return nil, err
	}

	go logger.processBuffer()
	return logger, nil
}

// 核心写入逻辑：日志先进入内存缓冲
func (l *DailyRotatingLogger) Write(p []byte) (n int, err error) {
	// 复制数据避免外部修改
	entry := make([]byte, len(p))
	copy(entry, p)
	l.buffer <- entry
	return len(p), nil
}

// 异步处理缓冲队列
func (l *DailyRotatingLogger) processBuffer() {
	ticker := time.NewTicker(l.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-l.buffer:
			l.safeWrite(entry)
		case <-ticker.C:
			l.mu.Lock()
			if l.fileWriter != nil {
				l.fileWriter.Flush()
			}
			l.mu.Unlock()
		case <-l.stopChan:
			l.mu.Lock()
			if l.fileWriter != nil {
				l.fileWriter.Flush()
				l.file.Close()
			}
			l.mu.Unlock()
			return
		}
	}
}

// 安全写入（含日期检查和轮转）
func (l *DailyRotatingLogger) safeWrite(data []byte) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 检查日期变更
	if err := l.rotateIfNeeded(); err != nil {
		fmt.Printf("轮转失败: %v\n", err)
		return
	}
	// 写入文件
	if _, err := l.fileWriter.Write(data); err != nil {
		fmt.Printf("写入失败: %v\n", err)
	}
}

// 按需轮转文件
func (l *DailyRotatingLogger) rotateIfNeeded() error {
	today := time.Now().Format("20060102")

	// 无需轮转
	if today == l.currentDate && l.file != nil {
		return nil
	}

	// 关闭旧文件
	if l.file != nil {
		l.fileWriter.Flush()
		l.file.Close()
	}

	// 创建新文件
	newFileName := filepath.Join(l.filePath, fmt.Sprintf("%s-%s.log", l.filePrefix, today))
	file, err := os.OpenFile(newFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	l.currentDate = today
	l.file = file
	l.fileWriter = bufio.NewWriterSize(file, 4096) // 4KB缓冲区
	return nil
}

// 优雅关闭
func (l *DailyRotatingLogger) Close() {
	close(l.stopChan)
}
