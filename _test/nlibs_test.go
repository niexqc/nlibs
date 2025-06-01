package test

import (
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/njson"
	"github.com/niexqc/nlibs/ntools"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}

func TestReadXlsx(t *testing.T) {
	contents, err := ntools.XlsxRead("_file4test/ndb_test_xls_read.xlsx", "user", 1)
	ntools.TestErrPainic(t, "读取Xlsx失败", err)

	contentJson := njson.Obj2StrWithPanicError(contents)
	slog.Info(contentJson)
	ntools.TestEq(t, "读取Xlsx失败", `[["1","niexq","niexq","聂小强","测试","2025-04-28 12:01:59"]]`, contentJson)
}

func TestRunCmd(t *testing.T) {
	err := ntools.CmdRunAndPrintLog(ntools.OsIsWindows(), "cmd", "", "/c", "echo", "cmd命令输出")
	ntools.TestErrPainic(t, "测试Cmd运行", err)

	recv := make(chan string, 1)

	err = ntools.CmdRunWithStdOut(ntools.OsIsWindows(), "cmd", "", recv, "/c", "echo", "中文输出")
	ntools.TestErrPainic(t, "测试Cmd运行", err)
	cmdOutStr := <-recv
	slog.Info(cmdOutStr)
	ntools.TestEq(t, "测试Cmd运行", `中文输出`, cmdOutStr)
}

func TestCronTask(t *testing.T) {
	cron := nlibs.NewCronWithSeconds()
	recv := make(chan int, 1)
	var firstTime time.Time
	runCount := 0

	cron.AddFunc("* * * * * *", func() {
		runCount++
		if firstTime.Year() == 1 {
			firstTime = time.Now()
		}
		if runCount == 2 {
			recv <- time.Now().Second() - firstTime.Second()
		}
	})
	cron.Start()
	intervlTime := <-recv
	ntools.TestEq(t, "测试CronTask,间隔1秒", 1, intervlTime)

}

func TestHttpClientPool(t *testing.T) {

	client := ntools.NewNHttpClientPool(10, 10, 30*time.Second, 30*time.Second)

	req, _ := http.NewRequest("GET", "http://www.baidu.com", nil)

	recv := make(chan string, 1)
	client.RunRequst(req, func(resp *http.Response, err error) {
		data, _ := io.ReadAll(resp.Body)
		recv <- string(data)
	})
	httpRespText := <-recv
	ntools.TestStrContains(t, "测试TestHttpClientPool异步执行", "百度", httpRespText)
}
