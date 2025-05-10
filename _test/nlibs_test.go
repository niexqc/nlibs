package test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/niexqc/nlibs"
	"github.com/niexqc/nlibs/nerror"
	"github.com/niexqc/nlibs/ntools"
)

func init() {
	ntools.SlogConf("test", "debug", 1, 2)
}
func TestErrorExt(t *testing.T) {
	err := nerror.NewRunTimeError("this is run time error")
	slog.Info(nerror.GenErrDetail(err))
}

func TestReadXlsx(t *testing.T) {
	contents, _ := ntools.XlsxRead("test.xlsx", "user", 1)
	fmt.Println(contents)
}

func TestCronTask(t *testing.T) {
	cron := nlibs.NewCronWithSeconds()
	cron.AddFunc("* * * * * *", func() {
		slog.Info("定时任务执行")
	})
	cron.Start()
	time.Sleep(time.Second * 3)
}
